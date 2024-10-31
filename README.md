GophKeeper is a client-server secret manager.


## Текущий стату

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
cat profile.temp | egrep -v "mock_store.go|test/main|staticlint|keygen|proto|.pb.go" > profile.cov

```

```bash
go tool cover -func profile.cov
```

Use html view to observe code lines coverage with testing:

```bash
go tool cover -html=profile.cov -o coverage.html
```