package kafka

import "time"

type KafkaMessage struct {
	Key, Value  []byte
	Topic       string
	Partition   int32
	Offset      int64
	Headers     []*MessageHeader
	headerCache map[string][]byte

	Timestamp      time.Time
	BlockTimestamp time.Time
}

type MessageHeader struct {
	Key   []byte
	Value []byte
}

func (msg *KafkaMessage) SetHeader(key string, value []byte) {
	msg.Headers = append(msg.Headers, &MessageHeader{
		Key:   []byte(key),
		Value: value,
	})
	if len(msg.headerCache) > 0 {
		delete(msg.headerCache, key)
	}
}

func (msg *KafkaMessage) Header(key string) []byte {
	if msg.headerCache == nil {
		msg.headerCache = make(map[string][]byte)
	}
	v, exsit := msg.headerCache[key]
	if exsit {
		return v
	}
	for _, h := range msg.Headers {
		keyStr := string(h.Key)
		if keyStr == key {
			msg.headerCache[keyStr] = h.Value
		}
		return h.Value
	}
	return nil
}
