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

const (
	_URL_MEDIA_INVITE = "http://mediatransfer:7555/api/invite"
	_URL_MEDIA_BYE    = "http://mediatransfer:7555/api/bye"

	_URL_CENTER_SUBMIT_CHANNEL = "http://{CENTER_IP}:{CENTER_PORT}/management/sensor/submitChannels"
)

func init() {
	stream.RegistEmitter("onvif", func() stream.Emitter {
		return &OnvifEmitter{}
	})
}

type OnvifEmitter struct {
	sync.Mutex

	submitLock  sync.Mutex
	submitedMap map[string]*model.Resource
	sessionMap  map[string]*Session
	previewWs   *PreviewWebsocket

	httpClient *http.Client

	ctx    context2.Context
	cancel context2.CancelFunc
}

type Session struct {
	Resource  *model.Resource
	ChannelNo string

	SessionId interface{}

	Opening bool

	InputChannel  string
	OutputChannel string
}

func (e *OnvifEmitter) Init(emit func(interface{}) error) error {
	e.ctx, e.cancel = context2.WithCancel(context2.Background())
	e.sessionMap = make(map[string]*Session, 0)
	e.submitedMap = make(map[string]*model.Resource, 0)

	e.initHttpClient()
	//启动websocket
	e.initWs()
	//启动通道查询上报
	e.startSubmitChannel()

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

func (e *OnvifEmitter) startSubmitChannel() {
	go func() {
		for {
			time.Sleep(5 * time.Second)
			select {
			case <-e.ctx.Done():
				return
			default:
			}
			e.submitLock.Lock()
			unSubmitResources := make([]*model.Resource, 0)
			context.ReadAllResourceGB(func(gbIdMaps map[string]*model.Resource) {
				for gid, r := range gbIdMaps {
					if _, ok := e.submitedMap[gid]; !ok {
						unSubmitResources = append(unSubmitResources, r)
					}
				}
			})
			e.submitLock.Unlock()
			if len(unSubmitResources) == 0 {
				continue
			}
			//上报通道
			for _, usr := range unSubmitResources {
				channels, err := LoadResourceChannels(usr)
				if err != nil {
					logger.LOG_WARN("onvif获取通道失败,GBID:", usr.GbID, ",err:", err)
					continue
				}
				if len(channels) == 0 {
					logger.LOG_WARN("onvif获取通道个数为0", usr.GbID)
					continue
				}
				//上报
				channelsReq := make([]map[string]interface{}, 0)
				for _, c := range channels {
					channelsReq = append(channelsReq, map[string]interface{}{
						"channelNo": c.Token,
						"id":        "",
						"name":      c.Name,
						"parentId":  "",
					})
				}
				err = util.Retry(func() error {
					return e.request(strings.ReplaceAll(strings.ReplaceAll(_URL_CENTER_SUBMIT_CHANNEL, "{CENTER_IP}", context.GetString("CENTER_IP")), "{CENTER_PORT}", context.GetString("CENTER_PORT")), http.MethodPost, "application/json", map[string]interface{}{
						"channels": channelsReq,
						"gid":      usr.GbID,
					}, nil)
				}, 3, 1*time.Second)
				if err != nil {
					logger.LOG_WARN("上报通道信息异常：", err)
					continue
				}
				e.submitedMap[usr.GbID] = usr
			}
		}
	}()
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
	idx := strings.Index(idChStr, "_")
	if idx < 0 {
		logger.LOG_WARN("错误的设备通道格式：", idChStr)
		return
	}
	id := idChStr[:idx]
	channel := idChStr[idx+1:]
	resource, ok := context.GetResource(id)
	if !ok {
		logger.LOG_WARN("未找到设备资源：", id)
		return
	}

	//查询设备开流地址
	rtsp, err := LoadChannelRTSP(resource, channel)
	if err != nil {
		logger.LOG_WARN("onvif获取rtsp失败：", err)
		return
	}
	s := &Session{
		Resource:  resource,
		ChannelNo: idChStr,

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
	openStreamRes := make(map[string]interface{})
	err := util.Retry(func() error {
		return ce.request(_URL_MEDIA_INVITE, http.MethodPost, "application/json", openStreamReq, &openStreamRes)
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
		return ce.request(_URL_MEDIA_BYE, http.MethodPost, "application/json", closeStreamReq, &closeStreamRes)
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
	logger.LOG_INFO("http-response:", string(resBytes))
	if resPointer != nil {
		return jsoniter.Unmarshal(resBytes, resPointer)
	}
	return nil
}
