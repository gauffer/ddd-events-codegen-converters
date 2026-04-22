package kafka

import (
	"github.com/gauffer/ddd-events-codegen-converters/pb/common"

	"errors"
	"fmt"

	"google.golang.org/protobuf/proto"
)

func GetTopic(m proto.Message) (string, error) {
	eKafkaDesc := common.E_Topic.TypeDescriptor()
	msg := m.ProtoReflect().Descriptor().Options().ProtoReflect()
	if !msg.Has(eKafkaDesc) {
		return "", fmt.Errorf("no option field for type %T", m)
	}
	kafka, ok := common.E_Topic.InterfaceOf(msg.Get(eKafkaDesc)).(*common.Topic)
	if !ok {
		return "", fmt.Errorf("cannot cast %T to %T", msg.Get(eKafkaDesc), common.E_Topic)
	}

	topic := kafka.GetValue()
	if topic == "" {
		return "", errors.New("empty topic")
	}
	return topic, nil
}
