package h_toalidatahub

import (
	"dyzs/data-flow/context"
	"dyzs/data-flow/logger"
	"dyzs/data-flow/model/gat1400"
	"dyzs/data-flow/stream"
	"errors"
	"fmt"
	"github.com/aliyun/aliyun-datahub-sdk-go/datahub"
	"reflect"
	"time"
)

func init() {
	stream.RegistHandler("toalidatahub", func() stream.Handler {
		return &AliDatahub{}
	})
}

type AliDatahub struct {
	accessId    string
	accessKey   string
	endpoint    string
	projectName string
	topicName   string
	shardId     string

	dh datahub.DataHubApi
}

func (h *AliDatahub) Init(config interface{}) error {
	logger.LOG_WARN("------------------ toalidatahub config ------------------")
	logger.LOG_WARN("toalidatahub_accessId", context.GetString("toalidatahub_accessId"))
	logger.LOG_WARN("toalidatahub_accessKey", context.GetString("toalidatahub_accessKey"))
	logger.LOG_WARN("toalidatahub_endpoint", context.GetString("toalidatahub_endpoint"))
	logger.LOG_WARN("toalidatahub_projectName", context.GetString("toalidatahub_projectName"))
	logger.LOG_WARN("toalidatahub_topicName", context.GetString("toalidatahub_topicName"))
	logger.LOG_WARN("------------------------------------------------------")
	h.accessId = context.GetString("toalidatahub_accessId")
	h.accessKey = context.GetString("toalidatahub_accessKey")
	h.endpoint = context.GetString("toalidatahub_endpoint")
	h.projectName = context.GetString("toalidatahub_projectName")
	h.topicName = context.GetString("toalidatahub_topicName")
	h.dh = datahub.New(h.accessId, h.accessKey, h.endpoint)

	err := h.createProject()
	if err == nil {
		err = h.createTupleTopic()
	}
	if err == nil {
		ls, err := h.dh.ListShard(h.projectName, h.topicName)
		if err != nil {
			logger.LOG_WARN("get shard list failed,", err)
			cerr := h.createTupleTopic()
			if cerr != nil {
				logger.LOG_WARN("创建topic异常：", cerr)
				return cerr
			}
		}
		if len(ls.Shards) == 0 {
			return errors.New("shardId empty")
		}
		h.shardId = ls.Shards[0].ShardId
	}
	if err != nil {
		logger.LOG_WARN("上云初始化失败，", err)
	}
	return err
}

func (h *AliDatahub) Handle(data interface{}, next func(interface{}) error) error {
	wraps, ok := data.([]*gat1400.Gat1400Wrap)
	if !ok {
		return errors.New(fmt.Sprintf("Handle [kafkamsgto1400] 数据格式错误，need []*kafka.KafkaMessage , get %T", reflect.TypeOf(data)))
	}
	if len(wraps) == 0 {
		return nil
	}
	for _, wrap := range wraps {
		bytes, err := wrap.BuildToJson()
		if err != nil {
			logger.LOG_WARN("序列化异常：", err)
			continue
		}
		h.putTupleData(string(bytes))
	}
	return nil
}

func (h *AliDatahub) putTupleData(data string) {
	topic, err := h.dh.GetTopic(h.projectName, h.topicName)
	if err != nil {
		fmt.Println(err)
		return
	}
	records := make([]datahub.IRecord, 3)
	record := datahub.NewTupleRecord(topic.RecordSchema, 0)
	record.ShardId = h.shardId
	record.SetValueByName("data", data)
	records = append(records, record)

	maxReTry := 3
	retryNum := 0
	for retryNum < maxReTry {
		result, err := h.dh.PutRecords(h.projectName, h.topicName, records)
		if err != nil {
			if _, ok := err.(*datahub.LimitExceededError); ok {
				logger.LOG_WARN("maybe qps exceed limit,retry")
				retryNum++
				time.Sleep(5 * time.Second)
				continue
			} else {
				logger.LOG_WARN("put record failed,", err)
				return
			}
		}
		logger.LOG_INFO("put successful num is %d, put records failed num is %d\n", len(records)-result.FailedRecordCount, result.FailedRecordCount)
		for _, v := range result.FailedRecords {
			logger.LOG_ERROR(v)
		}
		break
	}
	if retryNum >= maxReTry {
		logger.LOG_WARN("put records failed ")
	}
}

func (h *AliDatahub) createProject() (err error) {
	if err = h.dh.CreateProject(h.projectName, "project comment"); err != nil {
		if _, ok := err.(*datahub.InvalidParameterError); ok {
			logger.LOG_WARN("invalid parameter,please check your input parameter")
		} else if _, ok := err.(*datahub.ResourceExistError); ok {
			logger.LOG_WARN("project already exists")
		} else if _, ok := err.(*datahub.AuthorizationFailedError); ok {
			logger.LOG_WARN("accessId or accessKey err,please check your accessId and accessKey")
		} else if _, ok := err.(*datahub.LimitExceededError); ok {
			logger.LOG_WARN("limit exceed, so retry")
			for i := 0; i < 5; i++ {
				// wait 5 seconds
				time.Sleep(5 * time.Second)
				if err = h.dh.CreateProject(h.projectName, "project comment"); err != nil {
					logger.LOG_WARN("create project failed:", err)
				} else {
					break
				}
			}
		} else {
			logger.LOG_WARN("unknown error:", err)
		}
	}
	return err
}

func (h *AliDatahub) createTupleTopic() error {
	recordSchema := datahub.NewRecordSchema()
	recordSchema.AddField(datahub.Field{Name: "bigint_field", Type: datahub.BIGINT, AllowNull: true}).
		AddField(datahub.Field{Name: "timestamp_field", Type: datahub.TIMESTAMP, AllowNull: false}).
		AddField(datahub.Field{Name: "string_field", Type: datahub.STRING}).
		AddField(datahub.Field{Name: "double_field", Type: datahub.DOUBLE}).
		AddField(datahub.Field{Name: "boolean_field", Type: datahub.BOOLEAN})
	if err := h.dh.CreateTupleTopic(h.projectName, h.topicName, h.topicName, 5, 7, recordSchema); err != nil {
		return err
	}
	logger.LOG_WARN("create topic successful,", h.projectName, ":", h.topicName)
	return nil
}

func (h *AliDatahub) Close() error {
	return nil
}
