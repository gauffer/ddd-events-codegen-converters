# Налог на конвертер: N событий × 3 файла

Каждое новое доменное событие требует правки трёх файлов. Ниже реальный код из продакшн-сервиса с 7 событиями и ~360 строками бойлерплейта.

## 1. Доменная модель дублирует proto

```go
// model/events.go — 160 строк на 7 событий
type Event interface { markEvent() }
type commonEvent struct{}
func (commonEvent) markEvent() {}

type UserDeletedEvent struct {
    commonEvent
    UserID    UserID
    CreatedAt time.Time
}

type KYCStatusChangedEvent struct {
    commonEvent
    UserID        UserID
    OldStatus     KycStatus
    CurrentStatus KycStatus
    CreatedAt     time.Time
}
// ... ещё 5 struct'ов
```

Каждый struct это копия proto-полей в Go-типах. Добавил поле в proto, добавь и сюда.

## 2. Конвертер копирует поля руками

```go
// producer/converter.go — 136 строк на 7 событий
func ConvertUserDeletedEvent(v any) (proto.Message, error) {
    dto, ok := v.(model.UserDeletedEvent)
    if !ok {
        return nil, fmt.Errorf("unexpected type %T", v)
    }
    return &pbEvent.UserDeletedEvent{
        UserId:    dto.UserID.String(),
        EventTime: timestamppb.New(dto.CreatedAt),
    }, nil
}

func ConvertKYCStatusChangedEvent(v any) (proto.Message, error) {
    dto, ok := v.(model.KYCStatusChangedEvent)
    if !ok {
        return nil, fmt.Errorf("unexpected type %T", v)
    }
    oldStatus, ok := _KYCStatusToPbKYCStatus[dto.OldStatus]
    if !ok {
        return nil, fmt.Errorf("unanspect convert %s to %T", dto.OldStatus, ...)
    }
    curStatus, ok := _KYCStatusToPbKYCStatus[dto.CurrentStatus]
    if !ok {
        return nil, fmt.Errorf("unanspect convert %s to %T", dto.OldStatus, ...)
    }
    return &pbKYCEvent.KYCStatusChangedEvent{
        UserId:        dto.UserID.String(),
        OldStatus:     oldStatus,
        CurrentStatus: curStatus,
        EventTime:     timestamppb.New(dto.CreatedAt),
    }, nil
}
// ... ещё 5 функций
```

Все конвертеры одинаковые: assert тип, скопировать поля, вернуть. Enum-поля требуют ещё маппинг-таблицу.

## 3. Dispatch регистрирует каждый тип

```go
// producer/producer.go
var _EventToMessage = map[reflect.Type]func(v any) (proto.Message, error){
    reflect.TypeFor[model.WithdrawalAddressChangedEvent]():      ConvertWithdrawalAddressChanged,
    reflect.TypeFor[model.UserChangedPasswordEvent]():           ConvertUserChangedPasswordEvent,
    reflect.TypeFor[model.UserDeletedEvent]():                   ConvertUserDeletedEvent,
    reflect.TypeFor[model.KYCStatusChangedEvent]():              ConvertKYCStatusChangedEvent,
    reflect.TypeFor[model.UserUpdatedEvent]():                   ConvertUserUpdatedEvent,
    reflect.TypeFor[model.KYBStatusChangedEvent]():              ConvertKYBStatusChangedEvent,
    reflect.TypeFor[model.NotificationPreferenceChangedEvent](): ConvertNotificationPreferenceChangedEvent,
}

func (s *EventProducer) Publish(ctx context.Context, event model.Event) error {
    eventType, eventData := s.getEventDataAndType(ctx, event)
    converter, ok := _EventToMessage[eventType]
    if !ok {
        return fmt.Errorf("not find converter for %T", eventData)
    }
    dtoEvent, err := converter(eventData)
    // ...
}
```

Функционально это `switch` на reflect. Забыл добавить строку и получил рантайм-ошибку.

## Стоимость нового события

| Шаг | Файл | Что добавить |
|-----|------|-------------|
| 1 | `model/events.go` | struct + конструктор (~15 строк) |
| 2 | `converter.go` | `ConvertXxx()` с ручным маппингом (~15 строк) |
| 3 | `producer.go` | строка в `_EventToMessage` |
| 4 | `converter_test.go` | тест на маппинг полей |

Около 30 строк на событие. 7 событий дают 360 строк. 30 событий дадут больше тысячи.

Этот проект убирает все три файла. Proto-struct **является** событием. См. [README](../README.md).
