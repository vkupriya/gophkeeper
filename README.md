GophKeeper is a client-server secret manager.


## Текущий статуc

* Тесты покрытие 35.9% (надо добавить тесты для клиента)
* Надо добавить doc коментарии
* Хотелось бы попробовать сделать шифрование больших файлов с min.io

## Проверка доли покрытия кода тестами

В проекте используются интеграционные тесты, которые конфликтуют с авто-тестами темплейта (metrictests). Для раздельного тестирования, интеграционные тесты используют build tag 'integration'. При запуске локального тестирования и для проверки доли покрытия кода тестами, необходимо указать опцию '-tags=integration', чтобы включить интеграционные тесты.

```bash
go test -v -coverpkg=./... -coverprofile=profile.temp ./... -tags=integration
```

Удаляем сгенерированные файлы mock*.go из профиля, чтобы не влияли на результат вычисления покрытия

```bash
cat profile.temp | egrep -v "mock_|staticlint|keygen|proto|.pb.go" > profile.cov

```

```bash
go tool cover -func profile.cov
```

Use html view to observe code lines coverage with testing:

```bash
go tool cover -html=profile.cov -o coverage.html
```

## GRPC Client Mocks generation

```bash
mockgen -destination=internal/proto/mocks/mock_service_client.go -package=mocks -source=internal/proto/service_grpc.pb.go GophKeeperClient
```

## Флаги сборки для вывода версии клиента

```bash
export PGK_PATH="github.com/vkupriya/gophkeeper/internal/client/cmd"
go build -ldflags "-X $PKG_PATH.BuildVersion=v1.0.1 -X '$PKG_PATH.BuildDate=$(date +'%Y/%m/%d')' -X $PKG_PATH.BuildCommit=cb92c23" -o gkcli
```

При успешной сборке:
```bash
$ ./gkcli version
Using config file: ~/.gk.yaml
Build version: v1.0.1
Build date: 2024/11/03
Build commit: cb92c23
```
