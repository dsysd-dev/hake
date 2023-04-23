# Hake

Highly Available Kafka Equivalent or HAKE is a go based library for managing a highly available kafka equivalent on aws using SNS and SQS queue combinations.

## Kafka Architecture

https://dsysd-dev.medium.com/the-only-blog-post-you-will-ever-need-to-understand-kafka-473bb21cd784

A topic has multiple partitions and each partition can have a single consumer reading from it (from `a consumer group`).



## Requirements

- can create `Topics` (sns topics)
- can create `Consumer Groups` (consumer group = sqs queue)
- can register those `consumer groups` to `topics`

## Todos

- Remove Topic and associated consumer group
- Remove Consumer Groups
- Tests