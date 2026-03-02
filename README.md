# GoExercises - Что здесь

В этом проекте я держу разные тесты/упражнения и тп по Go
ЗЫ. Может кому будет интересно :-)

## Сравнение Hash-функций: KeyHash, HashMap, HashFnv для выбора шарда
[Детали](./keyhash/README.md)

## Сравнение работы channel и mutex на базе шардированого key-value хранилища
[Детали](./keyvaluestore/README.md)

## Testcontainers
[Разные эксперименты с testcontainers](./testcontainers/README.md)
- redis_test - запуск редиса в докере и проверка базовых функций
- timeline_test - тестирование логики: создание таргет объекта из сорца если его нет разными вариантами: Lua скрипт, SetNX и Redlock
- kafka_test - чисто вспомнить работу с кафкой (testcontainers/testcontainers-go/modules/kafka и клиент segmentio/kafka-go)
