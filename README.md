# keenetic-route-add

Утилита для добавления маршрутов в Keenetic через SSH.

## Использование

### Из аргументов командной строки

```bash
keenetic-route-add 23.222.66.0/24 34.91.161.0/24
```

### Опции

- `-host <ip>` — адрес SSH-сервера (по умолчанию `192.168.1.1`)
- `-user <name>` — имя пользователя SSH (по умолчанию `admin`)
- `-port <number>` — порт SSH (по умолчанию `22`)

Пример:
```bash
keenetic-route-add -host 10.0.0.1 -user myuser 23.222.66.0/24
```

### Из txt-файлов

Поместите `.txt` файлы рядом с исполняемым файлом в формате:

```
route add 23.222.66.0 mask 255.255.255.0 0.0.0.0
route add 34.91.161.0 mask 255.255.255.0 0.0.0.0
```

Маршруты будут автоматически преобразованы в CIDR-нотацию (`23.222.66.0/24`).

Оба источника объединяются — можно использовать аргументы и файлы одновременно.

## Настройка

Хост и пользователь SSH задаются в константах в `main.go`:

```go
const (
    sshHost = "192.168.1.1"
    sshUser = "admin"
)
```

Маршруты добавляются на интерфейсы `OpenVPN0` и `OpenVPN1`.

## Сборка

```bash
# Windows
GOOS=windows GOARCH=amd64 go build -o bin/keenetic-route-add.exe .

# Linux
GOOS=linux GOARCH=amd64 go build -o bin/keenetic-route-add-linux .

# macOS
GOOS=darwin GOARCH=amd64 go build -o bin/keenetic-route-add-macos .
```

Готовые бинарники находятся в папке `bin/`.
