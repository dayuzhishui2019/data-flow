package e_onvif

import (
	"bytes"
	"dyzs/data-flow/context"
	"dyzs/data-flow/logger"
	"dyzs/data-flow/model"
	"dyzs/data-flow/stream"
	"dyzs/data-flow/util"
	"errors"
	jsoniter "github.com/json-iterator/go"
	context2 "golang.org/x/net/context"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
)

func init() {
	stream.RegistEmitter("onvif", func() stream.Emitter {
		return &OnvifEmitter{}
	})
}

type OnvifEmitter struct {
	sync.Mutex

	sessionMap map[string]*Session
	previewWs  *PreviewWebsocket

	httpClient *http.Client

	ctx    context2.Context
	cancel context2.CancelFunc
}

type Session struct {
	Resource  *model.Resource
	ChannelNo string

	SessionId string

	Opening bool

	InputChannel  string
	OutputChannel string
}

func (e *OnvifEmitter) Init(emit func(interface{}) error) error {
	e.ctx, e.cancel = context2.WithCancel(context2.Background())
	e.sessionMap = make( map[string]*Session,0)

	e.initHttpClient()
	//启动websocket
	e.initWs()

	return nil
}

func (e *OnvifEmitter) initHttpClient() {
	e.httpClient = &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives:   false, //false 长链接 true 短连接
			Proxy:               http.ProxyFromEnvironment,
			MaxIdleConns:        10 * 5, //client对与所有host最大空闲连接数总和
			MaxConnsPerHost:     10,
			MaxIdleConnsPerHost: 10,               //连接池对每个host的最大连接数量,当超出这个范围时，客户端会主动关闭到连接
			IdleConnTimeout:     60 * time.Second, //空闲连接在连接池中的超时时间
		},
		Timeout: 5 * time.Second,
	}
}

func (e *OnvifEmitter) initWs() {
	e.previewWs = &PreviewWebsocket{
		ctx: e.ctx,
	}
	e.previewWs.Run(func(subscribe *WsSubject) {
		for _, us := range subscribe.UnSubscribe {
			logger.LOG_INFO("移除预览：", us)
			e.closeSession(us)
		}
		for _, s := range subscribe.Subscribe {
			logger.LOG_INFO("新增预览：", s)
			e.openSession(s, subscribe.RtmpMap[s])
		}
	})
}

func (ce *OnvifEmitter) openSession(idChStr string, outputChannel string) {
	id_ch := strings.Split(idChStr, "_")
	if len(id_ch) < 2 {
		return
	}
	id := id_ch[0]
	channel := id_ch[1]
	resource, ok := context.GetResource(id)
	if !ok {
		logger.LOG_WARN("未找到设备资源：", id)
		return
	}



	cs ,err := LoadResourceChannels(resource)
	channel = string(cs[0].Token)


	//查询设备开流地址
	rtsp, err := LoadChannelRTSP(resource, channel)
	if err != nil {
		logger.LOG_WARN("onvif获取rtsp失败：", err)
		return
	}
	s := &Session{
		Resource:  resource,
		ChannelNo: channel,

		InputChannel:  rtsp,
		OutputChannel: outputChannel,
	}
	ce.Lock()
	ce.sessionMap[idChStr] = s
	ce.Unlock()
	//通知媒体服务开流
	ce.inviteMedia(s)
}

func (ce *OnvifEmitter) inviteMedia(session *Session) {
	//调用媒体服务
	openStreamReq := map[string]interface{}{
		"channel":   session.ChannelNo,
		"recvproto": "rtsp",
		"recvparam": map[string]interface{}{
			// 公共参数，必须给值
			"istcp": false,
			// rtsp  rtsp必填参数，非rtsp可不给该字段
			"rtspurl": session.InputChannel,
		},
		"forwardproto": "rtmp",
		"forwardparam": map[string]interface{}{
			// rtmp rtmp参数，非rtsp可不给该字段
			"publishurl": session.OutputChannel,
		},
	}
	openStreamRes := make(map[string]string)
	err := util.Retry(func() error {
		return ce.request("http://172.16.129.104:7555/api/invite", http.MethodPost, "application/json", openStreamReq, &openStreamRes)
	}, 3, 1*time.Second)
	if err != nil {
		session.Opening = false
		logger.LOG_WARN("调用媒体服务开流异常：", err)
		return
	}
	session.Opening = true
	session.SessionId = openStreamRes["sessionid"]
}

func (ce *OnvifEmitter) closeSession(idChStr string) {
	ce.Lock()
	defer func() {
		ce.Unlock()
	}()
	session, ok := ce.sessionMap[idChStr]
	if !ok {
		logger.LOG_INFO("未找到session:", idChStr)
		return
	}
	//调用媒体服务
	closeStreamReq := map[string]interface{}{
		"channel":   session.ChannelNo,
		"sessionid": session.SessionId,
	}
	closeStreamRes := map[string]interface{}{}

	err := util.Retry(func() error {
		return ce.request("http://172.16.129.104:7555/api/bye", http.MethodPost, "application/json", closeStreamReq, &closeStreamRes)
	}, 3, 1*time.Second)
	if err != nil {
		session.Opening = false
		logger.LOG_WARN("调用媒体服务关流异常：", err)
		return
	}
	delete(ce.sessionMap, idChStr)
}

func (ce *OnvifEmitter) Close() error {
	return nil
}

//http请求
func (ce *OnvifEmitter) request(url, method, contentType string, body interface{}, resPointer interface{}) error {
	var bodyBytes []byte
	var resBytes []byte
	if body != nil {
		bodyBytes, _ = jsoniter.Marshal(body)
	}
	logger.LOG_INFO("http-request:", url)
	if logger.IsDebug() {
		params, err := jsoniter.Marshal(body)
		if err != nil {
			logger.LOG_WARN("http-request-params-error:", err)
		} else {
			logger.LOG_INFO("http-request-params:", string(params))
		}
	}
	err := util.Retry(func() error {
		req, err := http.NewRequest(method, url, bytes.NewReader(bodyBytes))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", contentType)
		res, err := ce.httpClient.Do(req)
		if err != nil {
			return err
		}
		defer func() {
			err := res.Body.Close()
			if err != nil {
				logger.LOG_WARN("关闭res失败", err)
			}
		}()
		resBytes, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		_ = res.Body.Close()
		if res.StatusCode != http.StatusOK {
			return errors.New(string(resBytes))
		}
		return nil
	}, 3, 3*time.Second)
	if err != nil {
		return err
	}
	if resPointer != nil {
		return jsoniter.Unmarshal(resBytes, resPointer)
	}
	return nil
}
