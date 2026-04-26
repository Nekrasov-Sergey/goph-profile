package kafka

import (
	"github.com/segmentio/kafka-go"
)

// kafkaHeadersCarrier адаптирует []kafka.Header под интерфейс propagation.TextMapCarrier
// для инжекции и экстракции trace context.
type kafkaHeadersCarrier []kafka.Header

// Get возвращает значение заголовка по ключу.
func (c *kafkaHeadersCarrier) Get(key string) string {
	for _, h := range *c {
		if h.Key == key {
			return string(h.Value)
		}
	}
	return ""
}

// Set добавляет заголовок.
func (c *kafkaHeadersCarrier) Set(key, value string) {
	*c = append(*c, kafka.Header{Key: key, Value: []byte(value)})
}

// Keys возвращает список ключей заголовков.
func (c *kafkaHeadersCarrier) Keys() []string {
	keys := make([]string, len(*c))
	for i, h := range *c {
		keys[i] = h.Key
	}
	return keys
}
