# О программе

Основное назначение программы - мониториг почтового сервиса (MTA) путем чтения новых данных, появляющихся в лог-файле, и подсчета значений некоторых параметров, характеризующих работу почтового сервера. В настоящее время поддерживаются только логи Postfix.

У программы два основных режима использования. В первом случае (команда `tail`) программа в фоновом режиме читает новые данные из лог-файла и ведет несколько счетчиков. `mlogtail` самостоятельно ослеживает состояние лог-файла, с которым он работает, поэтому при ротации логов не нужно ничего предпринимать.

Во втором - `mlogtail` используется для обращения в процессу, читающему лог, и получения (и/или обнуления) текущих значений счетчиков.

## Новое в версии 1.2.0

✨ **HTTP API с JSON-ответами** - теперь статистика доступна через REST API для удобной интеграции с современными системами мониторинга (Prometheus, Grafana, и др.)

## Как пользоваться

```none
# mlogtail -h
Usage:
  mlogtail [OPTIONS] tail
  mlogtail [OPTIONS] "stats | stats_reset | reset"
  mlogtail [OPTIONS] <COUNTER_NAME>
  mlogtail -f <LOG_FILE_NAME>

Options:
  -f string
        Mail log file path, if path is "-" then read from STDIN (default "/var/log/mail.log")
  -h    Show this help
  -http string
        HTTP server address (e.g., :37412 or 0.0.0.0:37412) to serve stats as JSON
  -init-from-file
        Read entire log file on startup to initialize counters, then continue tailing
  -l string
        Log reader process is listening for commands on a socket file, or IPv4:PORT,
        or [IPv6]:PORT (default "unix:/var/run/mlogtail.sock")
  -o string
        Set a socket OWNER[:GROUP] while listening on a socket file
  -p int
        Set a socket access permissions while listening on a socket file (default 666)
  -t string
        Mail log type. It is "postfix" only allowed for now (default "postfix")
  -v    Show version information and exit
```

### Запуск в режиме чтения лога

К сожалению, в Go у процесса нет хороших способов стать демоном, поэтому запускаем "читателя" просто в фоновом режиме:

```none
# mlogtail tail &
```

или при помощи `systemctl`. Если процесс, читающий лог, должен слушать сокет, то указание типа `unix` обязательно, например `unix:/tmp/some.sock`.

### Запуск с HTTP API

Для доступа к статистике через HTTP API (JSON):

```bash
# Только HTTP API (рекомендуется: использовать TCP socket вместо Unix)
# HTTP API + чтение конкретного лог-файла
mlogtail -f /var/log/mail.log -http :37412 -l 127.0.0.1:3333 tail

# HTTP API + инициализация счётчиков из всего файла
mlogtail -f /var/log/mail.log -http :37412 -l 127.0.0.1:3333 -init-from-file tail

# HTTP API на всех интерфейсах
mlogtail -f /var/log/mail.log -http 0.0.0.0:37412 -l 127.0.0.1:3333 tail

# HTTP API + Unix socket (для Zabbix, требует прав на /var/run)
mlogtail -f /var/log/mail.log -http :37412 -l unix:/var/run/mlogtail.sock tail

# HTTP API + альтернативный Unix socket в /tmp
mlogtail -f /var/log/mail.log -http :37412 -l unix:/tmp/mlogtail.sock tail
```

> **⚠️ Важно:** По умолчанию программа пытается создать Unix socket `/var/run/mlogtail.sock`. 
> Если получаете ошибку "address already in use" или "permission denied", используйте TCP socket: `-l 127.0.0.1:3333`

#### HTTP эндпоинты:

```bash
# Проверка работоспособности
curl http://localhost:37412/health
# {"status":"ok","version":"1.2.0"}

# Получение всей статистики
curl http://localhost:37412/stats
# {"bytes_received":1059498852,"bytes_delivered":1039967394,"received":2733,...,"queue_size":42}

# Получение конкретного счетчика
curl http://localhost:37412/counter/received
# {"counter":"received","value":2733}

# Сброс счетчиков (POST)
curl -X POST http://localhost:37412/reset
# {"status":"ok","message":"Counters reset"}

# Получение статистики и сброс (POST)
curl -X POST http://localhost:37412/stats_reset
# {"bytes_received":1234,"bytes_delivered":5678,...,"queue_size":15}

# Примечание: queue_size показывает текущий размер очереди Postfix (mailq)

### ⚡ Флаг -init-from-file

**Проблема:** При перезапуске mlogtail все счётчики сбрасываются в 0, что создаёт "скачки" в мониторинге.

**Решение:** Используйте флаг `-init-from-file`:

```bash
# Сначала прочитать весь файл, потом следить за новыми записями
mlogtail -f /var/log/mail.log -http :37412 -init-from-file tail
```

**Как работает:**
1. При запуске читает **весь** лог-файл и подсчитывает статистику
2. Затем переключается в режим `tail` и отслеживает только новые записи
3. Счётчики уже содержат историю, нет скачков при перезапуске

**Идеально для:**
- ✅ Системы мониторинга (Zabbix, Prometheus)  
- ✅ Серверы с большими лог-файлами  
- ✅ Ситуации с частыми перезапусками

### 🔄 Автоматический сброс при ротации логов

Для синхронизации с ротацией логов Postfix можно настроить автоматический сброс счётчиков каждые сутки в 00:00:

```bash
# Установить systemd timer
sudo cp mlogtail-reset.timer mlogtail-reset.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable mlogtail-reset.timer
sudo systemctl start mlogtail-reset.timer
```

Подробнее: [LOGROTATION.md](LOGROTATION.md)
```

### Получение значений счетчиков

```none
# mlogtail stats
bytes-received  1059498852
bytes-delivered 1039967394
received        2733
delivered       2944
forwarded       4
deferred        121
bounced         105
rejected        4
held            0
discarded       0
```

Необходимо обратить внимание на то, что если "читатель" запущен с опцией `-l`, с указанием сокета или IP-адреса и порта, на котором процесс ждет запросов, то и с командой получения значений счетчиков должен использоваться тот же парамер командной строки.

Вероятно, более частый случай обращения к счетчикам - это получение текущего значения одного из них, например:

```none
# mlogtail bytes-received
1059498852
```
```none
# mlogtail rejected
4
```

### Статистика по лог-файлу

Кроме работы в "реальном времени" `mlogtail` может использоваться и со статичным лог-файлом:
```none
# mlogtail -f /var/log/mail.log
```
или STDIN:
```none
# grep '^Apr  1' /var/log/mail.log | mlogtail -f -
```
например, для получения данных на определенное число.

## Установка

```none
go get -u github.com/hpcloud/tail go get golang.org/x/sys/unix &&
  go build && strip mlogtail &&
  cp mlogtail /usr/local/sbin &&
  chown root:bin /usr/local/sbin/mlogtail &&
  chmod 0711 /usr/local/sbin/mlogtail
```
