# GoExercises - Что здесь

В этом проекте я держу разные тесты/упражнения и тп по Go
ЗЫ. Может кому будет интересно :-)

## Содержимое
- docs - документация, репорты и тп
- keyhash - тестирование хешей
- keyvaluestore - тестирование разных подходов к key-value хранилищам
- lib - общий код
- services - разные сервисы
- testcontainers - различные тесты testcontainers пакета и разные сервисы
- .github - CI/CD вещи связанные с github-ом
- .gitlab-cl.yml - CI/CD для gitlab-а

## Сравнение Hash-функций: KeyHash, HashMap, HashFnv для выбора шарда
[Детали](./keyhash/README.md)

## Сравнение работы channel и mutex на базе шардированого key-value хранилища
[Детали](./keyvaluestore/README.md)

## Testcontainers
[Разные эксперименты с testcontainers](./testcontainers/README.md)
- redis_test - запуск редиса в докере и проверка базовых функций
- timeline_test - тестирование логики: создание таргет объекта из сорца если его нет разными вариантами: Lua скрипт, SetNX и Redlock
- kafka_test - чисто вспомнить работу с кафкой (testcontainers/testcontainers-go/modules/kafka и клиент segmentio/kafka-go)
- postgresql_test - реализовать конкурентные update-ы объектов с версионингом и получить тайминги (github.com/testcontainers/testcontainers-go/modules/postgres и клиент github.com/jackc/pgx/v5/stdlib)

## Полезные команды
- Запуск тестов с проверкой на data races и созданием файла (coverage.txt) с test coverage:
```
go test -race -coverprofile=coverage.txt -v ./...
```
- Конвертация test coverage данных в HTML файл, для просмотра в браузере (можно опустить "-o ./coverage.html" тогда страница должна открыться в браузере):
```
go tool cover -html=./coverage.txt -o ./coverage.html
```
- Запустить тесты с CPU профайлингом:
```
go test -run <test name> -bench=<test name> -cpuprofile <cpu profile output file>
```
- Преобразовать \<cpu profile output file\> файл в .svg файл:
```
go tool pprof -svg <cpu profile output file> > <cpu profile svg file>.svg
```