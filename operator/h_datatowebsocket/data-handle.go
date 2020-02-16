package h_datatowebsocket

import (
	"dyzs/data-flow/concurrent"
	"dyzs/data-flow/logger"
	"dyzs/data-flow/model/gat1400"
	"dyzs/data-flow/model/gat1400/base"
	"dyzs/data-flow/model/kafka"
	"dyzs/data-flow/util"
	"dyzs/data-flow/util/base64"
	jsoniter "github.com/json-iterator/go"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type DataHandler struct {
	executor *concurrent.Executor
	client   *http.Client
}

func NewDataHandler() *DataHandler {
	dh := &DataHandler{}
	dh.executor = concurrent.NewExecutor(10)
	dh.client = &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives:   false, //false 长链接 true 短连接
			Proxy:               http.ProxyFromEnvironment,
			MaxIdleConns:        10 * 5, //client对与所有host最大空闲连接数总和
			MaxConnsPerHost:     10,
			MaxIdleConnsPerHost: 10,               //连接池对每个host的最大连接数量,当超出这个范围时，客户端会主动关闭到连接
			IdleConnTimeout:     60 * time.Second, //空闲连接在连接池中的超时时间
		},
		Timeout: 5 * time.Second, //粗粒度 时间计算包括从连接(Dial)到读完response body
	}
	return dh
}

func (dh *DataHandler) castKafkaMsgToGat1400(kafkaMsgs []*kafka.KafkaMessage) []*gat1400.Gat1400Wrap {
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
	return wraps
}

func (dh *DataHandler) filterPreviewMsgs(msgs []*gat1400.Gat1400Wrap, filterFunc func(id string) bool) []*gat1400.Gat1400Wrap {
	previewMsgs := make([]*gat1400.Gat1400Wrap, 0)

	for _, msg := range msgs {
		previewMsg := msg.FilterByDeviceID(func(id string) bool {
			return filterFunc(id)
		})
		if previewMsg != nil {
			previewMsgs = append(previewMsgs, previewMsg)
		}
	}
	return previewMsgs
}

func (dh *DataHandler) downloadWrapImage(wraps []*gat1400.Gat1400Wrap) {
	tasks := make([]func(), 0)
	for _, wrap := range wraps {
		for _, item := range wrap.GetSubImageInfos() {
			func(img *base.SubImageInfo) {
				tasks = append(tasks, func() {
					dh.downloadImage(img)
				})
			}(item)
		}
	}
	err := dh.executor.SubmitSyncBatch(tasks)
	if err != nil {
		logger.LOG_WARN("预览数据下载图片异常", err)
	}
}

func (dh *DataHandler) downloadImage(image *base.SubImageInfo) {
	url := image.Data
	if url == "" {
		logger.LOG_WARN("图片路径缺失：", nil)
		return
	}
	if strings.Index(url, "http://") != 0 {
		return
	}
	err := util.Retry(func() error {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		req.Header.Set("Connection", "keep-alive")

		res, err := dh.client.Get(url)
		if err != nil {
			return err
		}
		defer func() {
			err := res.Body.Close()
			if err != nil {
				logger.LOG_WARN("下载图片,关闭res失败：url - "+url, err)
			}
		}()
		bytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		image.Data = base64.Encode(bytes)
		return nil
	}, 3, 100*time.Millisecond)

	if err != nil {
		logger.LOG_WARN("下载图片失败：url - "+url, err)
		image.Data = ""
	}
}
