package h_datatowebsocket

import (
	context2 "context"
	"errors"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/viper"
	"golang.org/x/net/websocket"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"sunset/data-stream/concurrent"
	"sunset/data-stream/constants"
	"sunset/data-stream/logger"
	"sunset/data-stream/model/gat1400"
	"sunset/data-stream/model/gat1400/base"
	"sunset/data-stream/model/kafka"
	"sunset/data-stream/redis"
	"sunset/data-stream/stream"
	"sunset/data-stream/util"
	"sunset/data-stream/util/base64"
	"sync"
	"time"
)

func init() {
	stream.RegistHandler("datatowebsocket", &HubmsgWebSocket{})
}

type HubmsgWebSocket struct {
	wsLock sync.RWMutex

	ws          *websocket.Conn
	redisClient *redis.Cache
	resourceMap map[string]bool
	previewC    chan []*gat1400.Gat1400Wrap

	executor *concurrent.Executor
	client   *http.Client
	ctx      context2.Context
	cancel   context2.CancelFunc
}

func (h *HubmsgWebSocket) Init(config interface{}) error {
	h.ctx, h.cancel = context2.WithCancel(context2.Background())
	h.executor = concurrent.NewExecutor(10)
	h.client = &http.Client{
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
	h.redisClient = redis.NewRedisCache(0, viper.GetString("redis.addr"), redis.FOREVER)
	h.previewC = make(chan []*gat1400.Gat1400Wrap, 10)
	go h.initWebsocket()
	go h.previewData(h.previewC)
	return nil
}
func (h *HubmsgWebSocket) Handle(data interface{}, next func(interface{}) error) error {
	kafkaMsgs, ok := data.([]*kafka.KafkaMessage)
	if !ok {
		return errors.New(fmt.Sprintf("Handle [datatowebsocket] 数据格式错误，need []*kafka.KafkaMessage , get %T", reflect.TypeOf(data)))
	}
	if len(kafkaMsgs) == 0 {
		return nil
	}
	//kafkamsg -> 1400
	msgs := h.castKafkaMsgToGat1400(kafkaMsgs)
	//filter
	previewMsgs := h.filterPreviewMsgs(msgs)
	if len(previewMsgs) > 0 {
		h.previewC <- previewMsgs
	}
	return next(data)
}

func (h *HubmsgWebSocket) previewData(previewC <-chan []*gat1400.Gat1400Wrap) {
	for {
		select {
		case <-h.ctx.Done():
			return
		case msgs := <-previewC:
			var ws *websocket.Conn
			h.wsLock.Lock()
			ws = h.ws
			h.wsLock.Unlock()
			if ws != nil {
				h.downloadWrapImage(msgs)
				msgBytes, err := jsoniter.Marshal(msgs)
				if err != nil {
					logger.LOG_WARN(err)
					continue
				}
				_, err = h.ws.Write(msgBytes)
				if err != nil {
					logger.LOG_WARN("发送预览数据失败：", err)
					h.wsLock.Lock()
					h.ws = nil
					h.wsLock.Unlock()
				}
			}
		}
	}
}

func (h *HubmsgWebSocket) Close() error {
	h.cancel()
	return nil
}

func (h *HubmsgWebSocket) initWebsocket() {
LOOP:
	for {
		time.Sleep(5 * time.Second)
		select {
		case <-h.ctx.Done():
			break LOOP
		default:
		}
		var ws *websocket.Conn
		h.wsLock.RLock()
		ws = h.ws
		h.wsLock.RUnlock()
		if ws != nil {
			continue
		}
		var centerAddr string
		err := h.redisClient.StringGet(constants.REDIS_KEY_CENTERADDR, &centerAddr)
		logger.LOG_INFO("centerAddr:", centerAddr)
		if err != nil || centerAddr == "" {
			h.ws = nil
			continue
		}
		previewWebsocketPath := viper.GetString("previewWebsocketPath")
		logger.LOG_INFO("previewWebsocketPath:", previewWebsocketPath)
		u := url.URL{Scheme: "ws", Host: centerAddr, Path: previewWebsocketPath}

		ws, err = websocket.Dial(u.String(), "", "http://"+centerAddr+"/")
		h.wsLock.Lock()
		if err != nil {
			logger.LOG_WARN(err)
			h.ws = nil
			h.wsLock.Unlock()
			continue
		}
		h.ws = ws
		h.wsLock.Unlock()
		//读取数据
		for {
			var data interface{} // data的类型为接收的JSON类型struct
			err := websocket.Message.Receive(ws, data)
			if err != nil {
				logger.LOG_WARN("websocket读取数据失败", err)
				h.wsLock.Lock()
				h.ws = nil
				h.wsLock.Unlock()
			}
		}
	}
}

func (h *HubmsgWebSocket) castKafkaMsgToGat1400(kafkaMsgs []*kafka.KafkaMessage) []*gat1400.Gat1400Wrap {
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

func (h *HubmsgWebSocket) filterPreviewMsgs(msgs []*gat1400.Gat1400Wrap) []*gat1400.Gat1400Wrap {
	previewMsgs := make([]*gat1400.Gat1400Wrap, 0)

	for _, msg := range msgs {
		previewMsg := msg.FilterByDeviceID(func(id string) bool {
			return h.resourceMap[id]
		})
		if previewMsg != nil {
			previewMsgs = append(previewMsgs, previewMsg)
		}
	}
	return previewMsgs
}

func (h *HubmsgWebSocket) downloadWrapImage(wraps []*gat1400.Gat1400Wrap) {
	tasks := make([]func(), 0)
	for _, wrap := range wraps {
		for _, item := range wrap.GetSubImageInfos() {
			func(img *base.SubImageInfo) {
				tasks = append(tasks, func() {
					h.downloadImage(img)
				})
			}(item)
		}
	}
	err := h.executor.SubmitSyncBatch(tasks)
	if err != nil {
		logger.LOG_WARN("预览数据下载图片异常", err)
	}
}

func (h *HubmsgWebSocket) downloadImage(image *base.SubImageInfo) {
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

		res, err := h.client.Get(url)
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
