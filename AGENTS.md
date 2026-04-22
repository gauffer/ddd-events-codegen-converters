# AGENTS.md

Proto-struct IS the domain event (cosmos-sdk approach). No converters, no dispatch maps, no per-event boilerplate. MongoDB outbox, Debezium, Kafka (Redpanda), typed `Consumer[T proto.Message]`. README and Go comments are in Russian.

## Commands

```bash
cd cmd/api && docker compose up -d              # infra + api container
sleep 30 && ./scripts/debezium.sh create-all    # register Debezium CDC connector
go run ./cmd/demo                               # UI :3000 (proxies /api/* → :8080)
easyp generate                                  # regenerate pb/ from proto/
go build ./... && go vet ./...                  # no tests, no CI
deadcode ./...                                  # unreachable funcs (golang.org/x/tools)
```

Redeploy code to docker: `docker compose up -d --build api` from `cmd/api/`.

## Running API locally (`go run ./cmd/api`)

Requires `JWT_PRIVATE_KEY_B64`, otherwise fatals with "failed to create JWT signer". Do not `source .env.example` because `&` in `MONGO_URI` breaks zsh; use `env VAR=val` or compose instead. Conflicts with docker `api` on `:8080`, stop it first with `docker compose stop api`.

## Core pattern

`domain.Event` = `proto.Message` + metadata. Five files carry the idea:

| File | Role |
|------|------|
| `proto/common/kafka.proto` | Extends `MessageOptions` with `(common.topic).value` |
| `proto/identity/events.proto` | Event contract + topic annotation |
| `pb/identity/events_ext.go` | **Hand-written** `EventID()`, `AggregateID()`, `OccurredAtTime()` on proto struct |
| `pkg/kafka/topic.go` | `GetTopic(proto.Message)` — topic from descriptor at runtime |
| `pkg/outbox/publisher.go` | `proto.Marshal` + `GetTopic` — zero per-event code |

After editing `.proto`: `easyp generate`, commit `pb/` (including `_ext.go` for new events).

## Testing philosophy

Field-mapping assertions (`assert.Equal(t, phone, event.GetPhone())`) are unnecessary because there is no converter to mistype. Test behavior instead: `events := user.PopEvents(); require.NotEmpty(t, events)`. See README section "Без конвертера тестируем поведение, а не маппинг" and [docs/converter-tax.md](docs/converter-tax.md) for real production code showing the cost.

## Layout

```
proto/                    source of truth
pb/                       generated; pb/*/events_ext.go is HAND-WRITTEN
pkg/{domain,outbox,kafka} Event interface, outbox publisher, generic Consumer[T]
internal/identity/        hex arch: domain/ application/ adapter/{inbound,outbound}
internal/profile/         Kafka consumer — creates profile on UserCreatedEvent
cmd/api/                  docker-compose.yaml, scripts/, debezium/
cmd/demo/                 static UI proxying /api/* → localhost:8080
```

Each module has `module.go` (wiring) and `adapters.go` (DI via functional options). OAuth clients repo is `inmemory` with seeded data. Debezium connector must use `ByteArrayConverter`, otherwise proto bytes get re-encoded to JSON.

## Conventions

Comments in Go code are **Russian**, dense, without section-header noise like `// User errors`. No emojis unless asked. Test creds: phone `+79991112233`, OTP `112233`, Client ID `1`.
