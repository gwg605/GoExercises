# Разные эксперименты с testcontainers

## redis_test
Запуск редиса в докере и проверка базовых функций

## timeline_test
Тестирование логики: создание таргет объекта из сорца если его нет разными вариантами: Lua скрипт, SetNX и Redlock

[RAW Бенчмаркинг тестов](../docs/reports/timeline.benchmark.raw.log)

## kafka_test
Чисто вспомнить работу с кафкой (testcontainers/testcontainers-go/modules/kafka и клиент segmentio/kafka-go)
