# Exactly-Once Semantics: Debezium + Kafka Connect

> Гайд, не реализация. В демо EOS не настроен.

## Kafka Connect (worker)

```properties
exactly.once.source.support = enabled
offset.storage.partitions   = 25   # меньше bottleneck при rebalance
offset.storage.replication.factor = 3
```

## Debezium connector

```json
{
  "connector.class": "io.debezium.connector.postgresql.PostgresConnector",
  "exactly.once.support": "required",
  "transaction.boundary": "poll"
}
```

С `exactly.once.support = required` коннектор не стартует без EOS на worker. С `transaction.boundary = poll` каждый `poll()` оборачивается в Kafka-транзакцию, offsets коммитятся атомарно с данными.

## Consumers

```properties
isolation.level = read_committed
```

Без этой настройки консюмер увидит сообщения из прерванных транзакций.

## Рестарт

Worker падает посреди `poll()`, транзакция остаётся незакоммиченной. Консюмер с `read_committed` эти сообщения не видит. Worker перезапускается, читает offset из offset-топика и продолжает с того же места. Дубликатов нет.
