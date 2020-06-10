package h_1400tokafkamsg

import (
	"dyzs/data-flow/logger"
	"dyzs/data-flow/model/gat1400"
	"dyzs/data-flow/model/kafka"
	"dyzs/data-flow/stream"
	"errors"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"reflect"
)

var _data_topic = "gat1400"

func init() {
	stream.RegistHandler("1400tokafkamsg", func() stream.Handler {
		return &stream.HandlerWrapper{
			InitFunc:   Init,
			HandleFunc: Handle,
		}
	})
}

func Init(config interface{}) error {
	logger.LOG_WARN("------------------ 1400tokafkamsg config ------------------")
	logger.LOG_WARN("------------------------------------------------------")
	return nil
}

func Handle(data interface{}, next func(interface{}) error) error {
	wraps, ok := data.([]*gat1400.Gat1400Wrap)
	if !ok {
		return errors.New(fmt.Sprintf("Handle [1400tokafkamsg] 数据格式错误，need []*daghub.Gat1400Wrap , get %T", reflect.TypeOf(data)))
	}
	if len(wraps) == 0 {
		return nil
	}
	kafkaMsgs := make([]*kafka.KafkaMessage, 0)
	for _, wrap := range wraps {
		bytes, err := jsoniter.Marshal(wrap)
		if err != nil {
			logger.LOG_WARN("hubmsgtobyte 转换失败", err)
			continue
		}
		kafkaMsg := &kafka.KafkaMessage{
			Topic: _data_topic,
			Value: bytes,
		}
		digestBytes, err := wrap.BuildDigest(gat1400.DIGEST_ACCESS, "", "")
		if err == nil {
			kafkaMsg.SetHeader("digest", digestBytes)
		}
		kafkaMsgs = append(kafkaMsgs, kafkaMsg)
	}
	if len(kafkaMsgs) > 0 {
		return next(kafkaMsgs)
	}
	return nil
}
