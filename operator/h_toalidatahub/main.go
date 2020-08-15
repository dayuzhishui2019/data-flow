package h_toalidatahub

import (
	"crypto/tls"
	"dyzs/data-flow/concurrent"
	"dyzs/data-flow/context"
	"dyzs/data-flow/logger"
	"dyzs/data-flow/stream"
	"dyzs/data-flow/util"
	"errors"
	"fmt"
	"github.com/aliyun/aliyun-datahub-sdk-go/datahub"
	"net/http"
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

	aesKey []byte

	dh           datahub.DataHubApi
	recordSchema *datahub.RecordSchema

	executor *concurrent.Executor
}

func (h *AliDatahub) Init(config interface{}) error {
	logger.LOG_WARN("------------------ toalidatahub config ------------------")
	logger.LOG_WARN("toalidatahub_accessId", context.GetString("toalidatahub_accessId"))
	logger.LOG_WARN("toalidatahub_accessKey", context.GetString("toalidatahub_accessKey"))
	logger.LOG_WARN("toalidatahub_endpoint", context.GetString("toalidatahub_endpoint"))
	logger.LOG_WARN("toalidatahub_projectName", context.GetString("toalidatahub_projectName"))
	logger.LOG_WARN("toalidatahub_topicName", context.GetString("toalidatahub_topicName"))
	logger.LOG_WARN("toalidatahub_shardId", context.GetString("toalidatahub_shardId"))
	logger.LOG_WARN("toalidatahub_aeskey", context.GetString("toalidatahub_aeskey"))
	logger.LOG_WARN("------------------------------------------------------")
	h.accessId = context.GetString("toalidatahub_accessId")
	h.accessKey = context.GetString("toalidatahub_accessKey")
	h.endpoint = context.GetString("toalidatahub_endpoint")
	h.projectName = context.GetString("toalidatahub_projectName")
	h.topicName = context.GetString("toalidatahub_topicName")
	h.shardId = context.GetString("toalidatahub_shardId")
	h.aesKey = []byte(context.GetString("toalidatahub_aeskey"))
	h.dh = datahub.New(h.accessId, h.accessKey, h.endpoint)
	h.executor = concurrent.NewExecutor(20)

	h.initDatahubApi()
	err := h.initTupleTopic()
	if err != nil {
		logger.LOG_WARN("上云初始化失败，", err)
	}
	return err
}

func (h *AliDatahub) Handle(data interface{}, next func(interface{}) error) error {
	paths, ok := data.([][]string)
	if !ok {
		return errors.New(fmt.Sprintf("Handle [toalidatahub] 数据格式错误，need [][]string , get %T", reflect.TypeOf(data)))
	}
	if len(paths) == 0 {
		return nil
	}

	batchs := make([][]datahub.IRecord, 0)
	records := make([]datahub.IRecord, 0)
	now := time.Now().UnixNano() / 1e6
	for _, path := range paths {
		if len(path) != 2 {
			continue
		}
		record := datahub.NewTupleRecord(h.recordSchema, now)
		record.ShardId = h.shardId
		record.SetValueByName("type", datahub.String(path[0]))
		record.SetValueByName("path", datahub.String(path[1]))
		records = append(records, record)
		if len(records) >= 20 {
			batchs = append(batchs, records)
			records = make([]datahub.IRecord, 0)
		}
	}
	if len(records) > 0 {
		batchs = append(batchs, records)
	}
	if len(batchs) == 0 {
		return nil
	}

	tasks := make([]func(), 0)
	for _, batch := range batchs {
		b := batch
		tasks = append(tasks, func() {
			err := util.Retry(func() error {
				result, err := h.dh.PutRecords(h.projectName, h.topicName, b)
				if err != nil {
					return err
				}
				logger.LOG_INFO(fmt.Sprintf("put success %d, put fail %d\n", len(records)-result.FailedRecordCount, result.FailedRecordCount))
				return nil
			}, 3, 3*time.Second)
			if err != nil {
				logger.LOG_WARN("数据推送datahub失败：", err)
			}
		})
	}

	err := h.executor.SubmitSyncBatch(tasks)
	if err != nil {
		logger.LOG_ERROR("上传ali datahub error：", err)
		return errors.New("上传ali datahub error：" + err.Error())
	}
	return nil
}

func (h *AliDatahub) initDatahubApi() {
	account := datahub.NewAliyunAccount(h.accessId, h.accessKey)
	config := &datahub.Config{
		EnableBinary: false,
		HttpClient: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	}
	h.dh = datahub.NewClientWithConfig(h.endpoint, config, account)
}

func (h *AliDatahub) initTupleTopic() error {
	topic, err := h.dh.GetTopic(h.projectName, h.topicName)
	if err != nil {
		//创建topic
		recordSchema := datahub.NewRecordSchema()
		recordSchema.AddField(datahub.Field{Name: "type", Type: datahub.STRING, AllowNull: false}).AddField(datahub.Field{Name: "path", Type: datahub.STRING, AllowNull: false})
		if err := h.dh.CreateTupleTopic(h.projectName, h.topicName, h.topicName, 5, 7, recordSchema); err != nil {
			return err
		}
		logger.LOG_WARN("create topic successful,", h.projectName, ":", h.topicName)
		topic, err = h.dh.GetTopic(h.projectName, h.topicName)
		if err != nil {
			return errors.New("获取topic失败:" + err.Error())
		}
	}
	h.recordSchema = topic.RecordSchema
	return nil
}

func (h *AliDatahub) Close() error {
	return nil
}
