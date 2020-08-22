package h_datatowebsocket

import (
	"context"
	"dyzs/data-flow/constants"
	"dyzs/data-flow/logger"
	"dyzs/data-flow/model/gat1400"
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

const _WS_NAMESPACE = "data_preview"
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
	Subscribe   string `json:"subscribe"`
	UnSubscribe string `json:"unSubscribe"`
}

type PreviewWebsocket struct {
	ws       *websocket.Conn
	wsLock   sync.RWMutex
	previewC chan []*gat1400.Gat1400Wrap
	dh       *DataHandler

	redisClient *redis.Cache
	ctx         context.Context
	handle      func([]string, []string)
}

func (pw *PreviewWebsocket) Run(handle func([]string, []string)) {
	pw.redisClient = redis.NewRedisCache(0, viper.GetString("redis.addr"), redis.FOREVER)
	pw.handle = handle
	//handle([]string{"610125287514449048597855",},nil)
	go pw.initWebsocket()
	go pw.previewData()
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
		var centerAddr string
		err := pw.redisClient.StringGet(constants.REDIS_KEY_CENTERHOST, &centerAddr)
		logger.LOG_INFO("centerAddr:", centerAddr)
		if err != nil || centerAddr == "" {
			pw.ws = nil
			continue
		}
		var boxId string
		err = pw.redisClient.StringGet(constants.REDIS_KEY_BOXID, &boxId)
		logger.LOG_INFO("boxId:", boxId)
		if err != nil || boxId == "" {
			logger.LOG_INFO("未找到盒子的id:", err)
			continue
		}
		wsPort := viper.GetString("center.wsPort")
		wsPath := viper.GetString("center.wsPath")
		logger.LOG_INFO("wsPort:", wsPort)
		logger.LOG_INFO("wsPath:", wsPath)
		wsPath = strings.ReplaceAll(strings.ReplaceAll(wsPath, "{namespace}", _WS_NAMESPACE), "{sid}", boxId)
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
			data := &WsSubject{}
			err = jsoniter.Unmarshal(wrap.Content, data)
			if err != nil {
				logger.LOG_WARN(err)
				continue
			}
			if len(data.Subscribe) == 0 && len(data.UnSubscribe) == 0 {
				continue
			}
			pw.handle([]string{data.Subscribe}, []string{data.UnSubscribe})
		}
	}
}

func (pw *PreviewWebsocket) previewData() {
	for {
		select {
		case <-pw.ctx.Done():
			return
		case msgs := <-pw.previewC:
			var ws *websocket.Conn
			pw.wsLock.Lock()
			ws = pw.ws
			pw.wsLock.Unlock()
			if ws == nil {
				logger.LOG_INFO("ws链接不存在")
				continue
			}
			pw.dh.downloadWrapImage(msgs)
			var wsMsgs []*WsSendMessage
			for _, m := range msgs {
				rds := m.FlatResourceData()
				if len(rds) == 0 {
					continue
				}
				for _, rd := range rds {
					wsMsgs = append(wsMsgs, &WsSendMessage{
						From:        rd.GetResourceID(),
						Timestamp:   time.Now().UnixNano() / 1e6,
						ContentType: _WS_CONTENT_TYPE_JSON,
						SendType:    _WS_SEND_MULTI,
						Content:     rd,
					})
				}
			}
			for _, wm := range wsMsgs {
				msgStr, _ := jsoniter.Marshal(wm)
				//logger.LOG_INFO("发送数据" + string(msgStr))
				err := ws.WriteMessage(websocket.TextMessage, msgStr)
				if err != nil {
					logger.LOG_WARN("发送预览数据失败：", err)
					pw.wsLock.Lock()
					pw.ws = nil
					pw.wsLock.Unlock()
					break
				} else {
					logger.LOG_INFO("发送预览数据成功")
				}
			}
		}
	}
}
