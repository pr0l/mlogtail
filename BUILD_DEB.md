# Сборка DEB пакета для Ubuntu

## Подготовка среды сборки

### Установка зависимостей:
```bash
sudo apt update
sudo apt install -y build-essential devscripts debhelper golang-go

# Если возникает ошибка с golang buildsystem, также установите:
# sudo apt install -y dh-golang
```

### Проверка версии Go:
```bash
go version  # Должно быть >= 1.18
```

## Сборка пакета

### 1. Клонирование репозитория:
```bash
git clone https://github.com/aadz/mlogtail.git
cd mlogtail
```

### 2. Сборка DEB пакета:
```bash
# Автоматическая сборка
make deb

# Или вручную
dpkg-buildpackage -us -uc -b
```

### 3. Результат:
```bash
# Пакет будет создан в родительской директории
ls -la ../mlogtail_*.deb
```

## Установка пакета

### Установка из DEB файла:
```bash
sudo dpkg -i ../mlogtail_1.2.0-1_amd64.deb
sudo apt-get install -f  # Если есть проблемы с зависимостями
```

### Проверка установки:
```bash
systemctl status mlogtail
curl http://localhost:8080/health
```

### Что устанавливает пакет

### Файлы:
- `/usr/bin/mlogtail` - основной бинарник
- `/lib/systemd/system/mlogtail.service` - systemd сервис
- `/lib/systemd/system/mlogtail-reset.*` - timer для сброса счётчиков
- `/usr/share/doc/mlogtail/` - документация и примеры
- `/usr/share/doc/mlogtail/zabbix/` - конфигурации Zabbix (для ручной установки)

### Автоматически:
- ✅ Регистрирует systemd сервисы
- ✅ Включает mlogtail.service (но не запускает)
- ✅ Создаёт пользователя и группу (если нужно)
- ✅ Настраивает права доступа

## Управление сервисом

```bash
# Запуск
sudo systemctl start mlogtail

# Статус
sudo systemctl status mlogtail

# Логи
sudo journalctl -u mlogtail -f

# Включить автосброс счётчиков
sudo systemctl enable mlogtail-reset.timer
sudo systemctl start mlogtail-reset.timer
```

## Удаление

```bash
# Удаление пакета
sudo apt remove mlogtail

# Полная очистка с конфигурацией
sudo apt purge mlogtail
```

## Разработка

### Локальная установка для тестирования:
```bash
make install
```

### Удаление локальной установки:
```bash
make uninstall
```

### Сборка только бинарника:
```bash
make build
```

### Запуск тестов:
```bash
make test
```