package h_datatowebsocket

import (
	context2 "context"
	"dyzs/data-flow/logger"
	"dyzs/data-flow/model/gat1400"
	"dyzs/data-flow/model/kafka"
	"dyzs/data-flow/stream"
	"errors"
	"fmt"
	"reflect"
	"sync"
)

func init() {
	stream.RegistHandler("datatowebsocket", func() stream.Handler {
		return &HubmsgWebSocket{}
	})
}

type HubmsgWebSocket struct {
	resourceLock sync.RWMutex
	resourceMap  map[string]bool

	dataHandler *DataHandler

	previewC  chan []*gat1400.Gat1400Wrap
	previewWs *PreviewWebsocket

	ctx    context2.Context
	cancel context2.CancelFunc
}

func (h *HubmsgWebSocket) Init(config interface{}) error {
	h.resourceMap = make(map[string]bool)
	h.ctx, h.cancel = context2.WithCancel(context2.Background())

	//数据处理器
	h.dataHandler = NewDataHandler()
	//启动websocket
	h.previewC = make(chan []*gat1400.Gat1400Wrap, 10)
	h.previewWs = &PreviewWebsocket{
		ctx:      h.ctx,
		previewC: h.previewC,
		dh:       h.dataHandler,
	}
	h.previewWs.Run(func(subscribe []string, unSubscribe []string) {
		h.resourceLock.Lock()
		for _, s := range subscribe {
			logger.LOG_INFO("新增预览：", s)
			h.resourceMap[s] = true
		}
		for _, us := range unSubscribe {
			logger.LOG_INFO("移除预览：", us)
			delete(h.resourceMap, us)
		}
		h.resourceLock.Unlock()
	})
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
	//处理数据
	msgs := h.dataHandler.castKafkaMsgToGat1400(kafkaMsgs)
	//过滤出需要预览的数据
	h.resourceLock.RLock()
	previewMsgs := h.dataHandler.filterPreviewMsgs(msgs, func(id string) bool {
		return h.resourceMap[id]
	})
	h.resourceLock.RUnlock()
	if len(previewMsgs) > 0 {
		select {
		case h.previewC <- previewMsgs:
		default:
			logger.LOG_INFO("预览队列满")
			break
		}

	}
	return next(data)
}

func (h *HubmsgWebSocket) Close() error {
	h.cancel()
	return nil
}
