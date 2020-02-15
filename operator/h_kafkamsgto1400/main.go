package h_kafkamsgto1400

import (
	"errors"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"reflect"
	"sunset/data-stream/logger"
	"sunset/data-stream/model/gat1400"
	"sunset/data-stream/model/kafka"
	"sunset/data-stream/stream"
)

var _data_topic = "gat1400"

func init() {
	stream.RegistHandler("kafkamsgto1400", &stream.HandlerWrapper{
		InitFunc:   Init,
		HandleFunc: Handle,
	})
}

func Init(config interface{}) error {
	logger.LOG_INFO("------------------ kafkamsgto1400 config ------------------")
	logger.LOG_INFO("------------------------------------------------------")
	return nil
}

func Handle(data interface{}, next func(interface{}) error) error {
	kafkaMsgs, ok := data.([]*kafka.KafkaMessage)
	if !ok {
		return errors.New(fmt.Sprintf("Handle [kafkamsgto1400] 数据格式错误，need []*kafka.KafkaMessage , get %T", reflect.TypeOf(data)))
	}
	if len(kafkaMsgs) == 0 {
		return nil
	}
	wraps := make([]*gat1400.Gat1400Wrap, 0)

	for _, kafkaMsg := range kafkaMsgs {
		w := &gat1400.Gat1400Wrap{}
		err := jsoniter.Unmarshal(kafkaMsg.Value, w)
		if err != nil {
			logger.LOG_ERROR("kafkamsgto1400 消息转化失败", err)
			continue
		}
		wraps = append(wraps, w)
	}
	if len(wraps) <= 0 {
		return nil
	}
	return next(wraps)
}
