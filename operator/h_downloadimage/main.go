package h_downloadimage

import (
	"dyzs/data-flow/concurrent"
	"dyzs/data-flow/context"
	"dyzs/data-flow/logger"
	"dyzs/data-flow/model/gat1400"
	"dyzs/data-flow/model/gat1400/base"
	"dyzs/data-flow/stream"
	"dyzs/data-flow/util"
	"dyzs/data-flow/util/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func init() {
	stream.RegistHandler("downloadimage", func() stream.Handler {
		return &stream.HandlerWrapper{
			InitFunc:   Init,
			HandleFunc: Handle,
			CloseFunc:  Close,
		}
	})
}

var executor *concurrent.Executor
var client *http.Client

func Init(config interface{}) error {
	capacity := 20
	configCapacity := context.GetInt("downloadimage_capacity")
	if configCapacity > 0 {
		capacity = configCapacity
	}
	logger.LOG_WARN("------------------ downloadimage config ------------------")
	logger.LOG_WARN("downloadimage_capacity : " + strconv.Itoa(capacity))
	logger.LOG_WARN("------------------------------------------------------")
	executor = concurrent.NewExecutor(capacity)
	client = &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives:   false, //false 长链接 true 短连接
			Proxy:               http.ProxyFromEnvironment,
			MaxIdleConns:        capacity * 5, //client对与所有host最大空闲连接数总和
			MaxConnsPerHost:     capacity,
			MaxIdleConnsPerHost: capacity,         //连接池对每个host的最大连接数量,当超出这个范围时，客户端会主动关闭到连接
			IdleConnTimeout:     60 * time.Second, //空闲连接在连接池中的超时时间
		},
		Timeout: 5 * time.Second, //粗粒度 时间计算包括从连接(Dial)到读完response body
	}
	return nil
}

func Handle(data interface{}, next func(interface{}) error) error {
	wraps, ok := data.([]*gat1400.Gat1400Wrap)
	if !ok {
		return errors.New(fmt.Sprintf("Handle [imagedeal] 数据格式错误，need []*daghub.StandardModelWrap , get %T", reflect.TypeOf(data)))
	}
	if len(wraps) == 0 {
		return nil
	}
	tasks := make([]func(), 0)
	for _, wrap := range wraps {
		for _, item := range wrap.GetSubImageInfos() {
			func(img *base.SubImageInfo) {
				tasks = append(tasks, func() {
					downloadImage(img)
				})
			}(item)
		}
	}
	err := executor.SubmitSyncBatch(tasks)
	if err != nil {
		logger.LOG_ERROR("批量下载图片失败：", err)
	}
	return next(wraps)
}

func downloadImage(image *base.SubImageInfo) {
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

		res, err := client.Get(url)
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

func Close() error {
	if executor != nil {
		executor.Close()
	}
	return nil
}
