package e_onvif

import (
	"context"
	context2 "dyzs/data-flow/context"
	"dyzs/data-flow/logger"
	"dyzs/data-flow/redis"
	"encoding/json"
	"github.com/gorilla/websocket"
	jsoniter "github.com/json-iterator/go"
	"github.com/spf13/viper"
	"net/url"
	"strings"
	"sync"
	"time"
)

const _WS_PATH = "/exchange/ws/{namespace}/{sid}"
const _WS_NAMESPACE = "video_preview"
const _WS_SEND_MULTI = "sensorMultiple"
const _WS_CONTENT_TYPE_JSON = "application/json"

type WsReceiveMessage struct {
	From        string          `json:"from"`
	To          string          `json:"to"`
	SendType    string          `json:"sendType"`
	ContentType string          `json:"contentType"`
	Content     json.RawMessage `json:"content"`
	Timestamp   int64           `json:"timestamp"`
}
type WsSendMessage struct {
	From        string      `json:"from"`
	To          string      `json:"to"`
	SendType    string      `json:"sendType"`
	ContentType string      `json:"contentType"`
	Content     interface{} `json:"content"`
	Timestamp   int64       `json:"timestamp"`
}

type WsSubject struct {
	RtmpMap map[string]string `json:"rtmp"`
	Subscribe   []string `json:"subscribe"`
	UnSubscribe []string `json:"unSubscribe"`
	PTZControl map[string]PTZControl `json:"ptzControl"`
}

type PTZControl struct{
	CMD string `json:"cmd"`
	Speed float64 `json:"speed"`
}

type PreviewWebsocket struct {
	ws     *websocket.Conn
	wsLock sync.RWMutex

	redisClient *redis.Cache
	ctx         context.Context
	handle      func(subscribe *WsSubject)
}

func (pw *PreviewWebsocket) Run(handle func(subscribe *WsSubject)) {
	pw.redisClient = redis.NewRedisCache(0, viper.GetString("redis.addr"), redis.FOREVER)
	pw.handle = handle
	go pw.initWebsocket()
}

func (pw *PreviewWebsocket) initWebsocket() {
LOOP:
	for {
		time.Sleep(5 * time.Second)
		select {
		case <-pw.ctx.Done():
			break LOOP
		default:
		}
		var ws *websocket.Conn
		pw.wsLock.RLock()
		ws = pw.ws
		pw.wsLock.RUnlock()
		if ws != nil {
			continue
		}
		centerAddr := context2.GetString("CENTER_IP")
		wsPort := context2.GetString("CENTER_PORT")
		wsPath := _WS_PATH

		task, err := context2.GetTask()
		if err != nil {
			continue
		}

		logger.LOG_INFO("wsPort:", wsPort)
		logger.LOG_INFO("wsPath:", wsPath)
		wsPath = strings.ReplaceAll(strings.ReplaceAll(wsPath, "{namespace}", _WS_NAMESPACE), "{sid}", task.ID)
		logger.LOG_INFO("wsPath decode:", wsPath)
		u := url.URL{Scheme: "ws", Host: centerAddr + ":" + wsPort, Path: wsPath}
		ws, _, err = websocket.DefaultDialer.Dial(u.String(), nil)
		pw.wsLock.Lock()
		if err != nil {
			logger.LOG_WARN(err)
			pw.ws = nil
			pw.wsLock.Unlock()
			continue
		}
		logger.LOG_INFO("websocket连接成功")
		pw.ws = ws
		pw.wsLock.Unlock()
		//读取数据
	READ_LOOP:
		for {
			select {
			case <-pw.ctx.Done():
				break LOOP
			default:
			}
			wrap := &WsReceiveMessage{}
			err := ws.ReadJSON(wrap)
			if err != nil {
				logger.LOG_WARN(err)
				pw.wsLock.Lock()
				pw.ws = nil
				pw.wsLock.Unlock()
				break READ_LOOP
			}
			logger.LOG_INFO("WS_RES：",string(wrap.Content))
			data := &WsSubject{}
			err = jsoniter.Unmarshal(wrap.Content, data)
			if err != nil {
				logger.LOG_WARN(err)
				continue
			}
			pw.handle(data)
		}
	}
}
