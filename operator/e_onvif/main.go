package e_onvif

import (
	"dyzs/data-flow/context"
	"dyzs/data-flow/logger"
	"dyzs/data-flow/model"
	"dyzs/data-flow/model/videocmd"
	"dyzs/data-flow/proxy"
	"dyzs/data-flow/stream"
	"errors"
	context2 "golang.org/x/net/context"
	"net/http"
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

	submitLock  sync.Mutex
	submitedMap map[string]*model.Resource

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
	OutputChannel *videocmd.CmdOpenStream
}

func (e *OnvifEmitter) Init(emit func(interface{}) error) error {
	e.ctx, e.cancel = context2.WithCancel(context2.Background())
	e.submitedMap = make(map[string]*model.Resource, 0)

	//绑定 指令处理到 server
	proxy.AttachVideoCmd(e)
	//启动通道查询上报
	e.startSubmitChannel()

	return nil
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
				catalogs := make([]*videocmd.ParamDeviceChannel, 0)
				for _, c := range channels {
					catalogs = append(catalogs, &videocmd.ParamDeviceChannel{
						FromID:   usr.ID,
						DeviceID: string(c.Token),
						Name:     string(c.Name),
					})
				}
				err = proxy.SubmitVideoChannelToGalaxy(catalogs)
				if err != nil {
					logger.LOG_WARN("上报通道信息异常：", err)
					continue
				}
				e.submitedMap[usr.GbID] = usr
			}
		}
	}()
}

func (ce *OnvifEmitter) OpenStream(forwardParam *videocmd.CmdOpenStream) (result interface{}, err error) {
	id := forwardParam.DeviceId
	channel := forwardParam.Channel
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
	//通知媒体服务开流
	return videocmd.NewMediaOpenStream(id, channel, &videocmd.ParamReceiveStream{
		RecvProto: "rtsp",
		RtspUrl:   rtsp,
	}, forwardParam), nil
}

func (ce *OnvifEmitter) CloseStream(param *videocmd.CmdCloseStream) (result interface{}, err error) {
	return map[string]interface{}{
		"Cmd":   "CloseStream",
		"Param": param,
	}, nil
}

func (ce *OnvifEmitter) PTZ(ptzParam *videocmd.CmdPTZ) (result interface{}, err error) {
	id := ptzParam.DeviceId
	channel := ptzParam.Channel
	resource, ok := context.GetResource(id)

	if !ok {
		logger.LOG_WARN("未找到设备资源：", id)
		return
	}
	//err := ControlPTZ(resource, channel, ptzControl.CMD, ptzControl.Speed)
	err = ControlPTZ(resource, channel, ptzParam.Direction(), ptzParam.GetSpeed())
	if err != nil {
		logger.LOG_WARN("云台控制异常：", err)
		return nil, err
	}
	return nil, nil
}

func (ce *OnvifEmitter) HistoryVideo(param *videocmd.CmdHistoryVideo) (result interface{}, err error) {
	return nil, errors.New("onvif 暂未实现录像接口")
}

func (ce *OnvifEmitter) Close() error {
	return nil
}
