# Plus7 - TCP/UDP прокси

Это VPN-прокси, который работает через [socket.io](https://socket.io/).

## Текущие проблемы

- Реализовать TCP. Работает HTTP, но при HTTPS какие то проблемы в протоколе.
- Проверить UDP

## Сборка

```sh
# Для Mac на Intel
env GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build
```
