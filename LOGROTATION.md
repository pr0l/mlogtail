# Автоматический сброс счётчиков при ротации логов

## Проблема
При ротации логов Postfix (обычно в 00:00) создаётся новый файл `/var/log/mail.log`, но счётчики mlogtail продолжают накапливать данные. Это может привести к некорректной статистике в системах мониторинга.

## Решение
Автоматически сбрасывать счётчики mlogtail каждый день в полночь, синхронно с ротацией логов.

### Установка:
```bash
# Копировать файлы timer и service
sudo cp mlogtail-reset.timer /etc/systemd/system/
sudo cp mlogtail-reset.service /etc/systemd/system/

# Перезагрузить systemd
sudo systemctl daemon-reload

# Включить и запустить timer
sudo systemctl enable mlogtail-reset.timer
sudo systemctl start mlogtail-reset.timer

# Проверить статус
sudo systemctl status mlogtail-reset.timer
sudo systemctl list-timers | grep mlogtail
```

### Преимущества:
- ✅ Основной сервис не перезапускается (нет простоя)
- ✅ HTTP API продолжает работать
- ✅ Быстрый сброс через HTTP API
- ✅ Минимальное влияние на мониторинг

### Как работает:
- Каждый день в 00:00:05 (через 5 сек после ротации)
- Отправляет `POST /reset` на HTTP API
- Счётчики обнуляются, сервис продолжает работу

## Проверка работы

```bash
# Посмотреть когда сработает в следующий раз
sudo systemctl list-timers | grep mlogtail

# Проверить логи сброса
sudo journalctl -u mlogtail-reset.service -f

# Ручной запуск для тестирования
sudo systemctl start mlogtail-reset.service
```

## Кастомизация времени

Если ротация логов происходит не в 00:00, измените время в `mlogtail-reset.timer`:

```ini
# Для ротации в 01:30
OnCalendar=*-*-* 01:30:05

# Для ротации каждые 6 часов  
OnCalendar=*-*-* 00,06,12,18:00:05

# Только по воскресеньям в 02:00
OnCalendar=Sun *-*-* 02:00:05
```