package stream

import (
	"dyzs/data-flow/logger"
	"errors"
	"strconv"
	"time"
)

var emitter_map = make(map[string]func() Emitter)
var handler_map = make(map[string]func() Handler)

func RegistEmitter(name string, ef func() Emitter) {
	emitter_map[name] = ef
}
func RegistHandler(name string, hf func() Handler) {
	handler_map[name] = hf
}

func GetEmitter(name string) (emitter Emitter, err error) {
	a, b := emitter_map[name]
	if !b {
		return nil, errors.New("未注册的Emitter:" + name)
	}
	return a(), nil
}

func GetHandler(name string) (handler Handler, err error) {
	a, b := handler_map[name]
	if !b {
		return nil, errors.New("未注册的Handler:" + name)
	}
	return a(), nil
}

type Emitter interface {
	Init(func(interface{}) error) error
	Close() error
}

type Handler interface {
	Init(interface{}) error
	Handle(interface{}, func(interface{}) error) error
	Close() error
}

//type Emitter struct {
//	Init  func(func(interface{})) error
//	Close func() error
//}
//
//type Handler struct {
//	Init   func(interface{}) error
//	Handle func(interface{}, func(interface{}))
//	Close  func() error
//}

type Stream struct {
	inited   bool
	running  bool
	emitter  Emitter
	handlers []Handler
}

func (s *Stream) linkHandle(index int, data interface{}) error {
	if len(s.handlers) > index {
		h := s.handlers[index]
		start := time.Now()
		return h.Handle(data, func(ndata interface{}) error {
			logger.LOG_WARN(strconv.Itoa(index) + "_handle - 耗时：" + time.Since(start).String())
			index++
			return s.linkHandle(index, ndata)
		})
	}
	return nil
}

func Build(flow []string) (s *Stream, err error) {
	if len(flow) == 0 {
		return nil, errors.New("未定义流程处理环节")
	}
	//check
	var emitter Emitter
	var handlers = make([]Handler, 0)

	for i, name := range flow {
		if i == 0 {
			e, err := GetEmitter(name)
			if err != nil {
				return nil, err
			}
			emitter = e
		} else {
			h, err := GetHandler(name)
			if err != nil {
				return nil, err
			}
			handlers = append(handlers, h)
		}
	}

	myStream := New(emitter)
	for i := 0; i < len(handlers); i++ {
		myStream.Pipe(handlers[i])
	}
	return myStream, nil
}

func New(emitter Emitter) *Stream {
	s := &Stream{}
	s.handlers = make([]Handler, 0)
	s.emitter = emitter
	return s
}

func (s *Stream) Pipe(h Handler) *Stream {
	s.handlers = append(s.handlers, h)
	return s
}

func (s *Stream) Init() error {
	var err error
	s.inited = false
	for _, h := range s.handlers {
		err = h.Init(nil)
		if err != nil {
			break
		}
	}
	if err != nil {
		logger.LOG_ERROR("处理流程初始化异常,启动失败！：", err)
	} else {
		s.inited = true
	}
	return err
}

func (s *Stream) Run() {
	if !s.inited {
		return
	}
	err := s.emitter.Init(func(data interface{}) error {
		start := time.Now()
		e := s.linkHandle(0, data)
		logger.LOG_INFO("单轮耗时：%v", time.Since(start))
		if e != nil {
			logger.LOG_WARN("流程异常：", e)
		}
		return e
	})

	if err != nil {
		logger.LOG_ERROR("采集器初始化异常,启动失败！：", err)
	} else {
		s.running = true
	}
}

func (s *Stream) Close() {
	var err error
	if s.emitter != nil {
		err = s.emitter.Close()
	}
	if err != nil {
		logger.LOG_WARN("关闭stream异常：", err)
	}
	for _, h := range s.handlers {
		err = h.Close()
		if err != nil {
			logger.LOG_WARN("关闭stream异常：", err)
		}
	}
}

//func EmiterFromPlugin(path string) Emitter {
//	fmt.Println(path)
//	p, _ := plugin.Open(path)
//	init, _ := p.Lookup("Init")
//	closem, _ := p.Lookup("Close")
//	return struct{
//		Init func(func(interface{})) error
//		Close func() error
//	}{
//		Init: init.(func(func(interface{})) error),
//		Close: closem.(func() error),
//	}
//}
//
//func HandlerFromPlugin(path string) Handler {
//	fmt.Println(path)
//	p, _ := plugin.Open(path)
//	init, _ := p.Lookup("Init")
//	handle, _ := p.Lookup("Handle")
//	return Handler{
//		Init:   init.(func(interface{}) error),
//		Handle: handle.(func(interface{}, func(interface{}))),
//	}
//}

type EmitterWrapper struct {
	InitFunc  func(func(interface{})) error
	CloseFunc func() error
}

func (ew *EmitterWrapper) Init(emit func(interface{})) error {
	return ew.InitFunc(emit)
}
func (ew *EmitterWrapper) Close() error {
	if ew.InitFunc == nil {
		return nil
	}
	return ew.CloseFunc()
}

type HandlerWrapper struct {
	InitFunc   func(interface{}) error
	HandleFunc func(interface{}, func(interface{}) error) error
	CloseFunc  func() error
}

func (ew *HandlerWrapper) Init(config interface{}) error {
	if ew.InitFunc == nil {
		return nil
	}
	return ew.InitFunc(config)
}
func (ew *HandlerWrapper) Handle(data interface{}, next func(interface{}) error) error {
	return ew.HandleFunc(data, next)
}
func (ew *HandlerWrapper) Close() error {
	if ew.CloseFunc == nil {
		return nil
	}
	return ew.CloseFunc()
}
