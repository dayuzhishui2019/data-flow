package e_1400server

import (
	context2 "context"
	"dyzs/data-flow/context"
	"dyzs/data-flow/logger"
	gat1400_model "dyzs/data-flow/model/gat1400"
	"dyzs/data-flow/model/gat1400/base"
	"dyzs/data-flow/stream"
	"dyzs/data-flow/util"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"github.com/satori/go.uuid"
	"net/http"
	"net/textproto"
	"strconv"
	"strings"
	"sync"
	"time"
)

func init() {
	stream.RegistEmitter("1400server", func() stream.Emitter {
		return &Gat1400Server{}
	})
}

var DEFAULT_SESSION_TIMEOUT = 60 * time.Second

type RegisterSession struct {
	ViewLibID     string
	Ip            string
	Port          string
	RegistTime    time.Time
	KeepaliveTime time.Time
}

type Gat1400Server struct {
	sync.Mutex
	viewLibId  string
	serverPort string
	username   string
	password   string
	openAuth   bool
	registers  map[string]*RegisterSession

	httpServer     *http.Server
	emit           func(interface{}) error
	sessionTimeout time.Duration
	ctx            context2.Context
	cancel         context2.CancelFunc
}

func (s *Gat1400Server) Init(emit func(interface{}) error) error {
	viewLibId := context.GetString("1400server_viewLibId")
	serverPort := context.GetString("1400server_serverPort")
	openAuth := context.GetBool("1400server_openAuth")
	username := context.GetString("1400server_username")
	password := context.GetString("1400server_password")
	sessionTimeout := context.GetInt("1400server_sessionTimeout")
	logger.LOG_WARN("------------------ 1400server config ------------------")
	logger.LOG_WARN("1400server_viewLibId : " + viewLibId)
	logger.LOG_WARN("1400server_serverPort : " + serverPort)
	logger.LOG_WARN("1400server_openAuth : " + strconv.FormatBool(openAuth))
	logger.LOG_WARN("1400server_username : " + username)
	logger.LOG_WARN("1400server_password : " + password)
	logger.LOG_WARN("1400server_sessionTimeout : " + strconv.Itoa(sessionTimeout))
	logger.LOG_WARN("------------------------------------------------------")
	if viewLibId == "" {
		return errors.New("[1400server] 视图库id不能为空")
	}
	if serverPort == "" {
		return errors.New("[1400server] 视图库服务端口不能为空")
	}
	if openAuth && (username == "" || password == "") {
		return errors.New("[1400server] 开启认证时，1400server_username 和 1400server_password 不能为空")
	}
	s.emit = emit
	s.registers = make(map[string]*RegisterSession)
	s.viewLibId = viewLibId
	s.serverPort = serverPort
	s.openAuth = openAuth
	s.username = username
	s.password = password
	s.sessionTimeout = DEFAULT_SESSION_TIMEOUT
	configTimeout := context.GetInt("1400server_sessionTimeout")
	if configTimeout > 0 {
		s.sessionTimeout = time.Duration(configTimeout) * time.Second
	}
	ctx, cancel := context2.WithCancel(context2.Background())
	s.ctx = ctx
	s.cancel = cancel
	go s.InitHttpServer()
	go s.checkKeepaliveSession()
	return nil
}

func (s *Gat1400Server) InitHttpServer() {
	engine := gin.Default()

	//注册
	engine.POST(base.URL_REGIST, func(c *gin.Context) {
		s.regist(c)
	})

	//取消注册
	engine.POST(base.URL_UNREGIST, func(c *gin.Context) {
		s.unRegist(c)
	})

	//保活
	engine.POST(base.URL_KEEPALIVE, func(c *gin.Context) {
		s.keepalive(c)
	})

	//校时
	engine.POST(base.URL_TIME, func(c *gin.Context) {
		s.time(c)
	})

	//人脸
	engine.POST(base.URL_FACES, func(c *gin.Context) {
		s.receive(gat1400_model.GAT1400_FACE, c)
	})
	//人体
	engine.POST(base.URL_PERSONS, func(c *gin.Context) {
		s.receive(gat1400_model.GAT1400_BODY, c)
	})
	//车辆
	engine.POST(base.URL_VEHICLE, func(c *gin.Context) {
		s.receive(gat1400_model.GAT1400_VEHICLE, c)
	})
	//非机动车
	engine.POST(base.URL_NOMOTORS, func(c *gin.Context) {
		s.receive(gat1400_model.GAT1400_NONMOTOR, c)
	})
	//图像
	engine.POST(base.URL_IMAGE, func(c *gin.Context) {
		rm := &gat1400_model.ImageObject{}
		err := jsoniter.NewDecoder(c.Request.Body).Decode(rm)
		if err != nil {
			c.JSON(http.StatusOK, base.BuildRespnse(base.BuildResponseObject(base.URL_IMAGE, "", base.JSON_FORMAT_INVALID)))
			return
		}
		objs := make([]*base.ResponseStatusObject, 0)
		if len(rm.ImageListObject.Image) > 0 {
			for _, item := range rm.ImageListObject.Image {
				objs = append(objs, base.BuildResponseObject(base.URL_IMAGE, item.ImageInfo.ImageID, base.OK))
			}
		}
		c.JSON(http.StatusOK, base.BuildRespnse(objs...))
	})
	//图像数据
	engine.POST(base.URL_IMAGE+"/:id/Data", func(c *gin.Context) {
		c.JSON(http.StatusOK, base.BuildSingleResponse(base.BuildResponseObject(c.Request.RequestURI, c.Param("id"), base.OK)))
	})
	//视频
	engine.POST(base.URL_VIDEOSLICE, func(c *gin.Context) {
		rm := &gat1400_model.VideoSliceObject{}
		err := jsoniter.NewDecoder(c.Request.Body).Decode(rm)
		if err != nil {
			c.JSON(http.StatusOK, base.BuildRespnse(base.BuildResponseObject(base.URL_VIDEOSLICE, "", base.JSON_FORMAT_INVALID)))
			return
		}
		objs := make([]*base.ResponseStatusObject, 0)
		if len(rm.VideoSliceListObject.VideoSlice) > 0 {
			for _, item := range rm.VideoSliceListObject.VideoSlice {
				objs = append(objs, base.BuildResponseObject(base.URL_VIDEOSLICE, item.VideoSliceInfo.VideoID, base.OK))
			}
		}
		c.JSON(http.StatusOK, base.BuildRespnse(objs...))
	})
	//视频数据
	engine.POST(base.URL_VIDEOSLICE+"/:id/Data", func(c *gin.Context) {
		c.JSON(http.StatusOK, base.BuildSingleResponse(base.BuildResponseObject(c.Request.RequestURI, c.Param("id"), base.OK)))
	})
	//文件
	engine.POST(base.URL_FILE, func(c *gin.Context) {
		rm := &gat1400_model.FileObject{}
		err := jsoniter.NewDecoder(c.Request.Body).Decode(rm)
		if err != nil {
			c.JSON(http.StatusOK, base.BuildRespnse(base.BuildResponseObject(base.URL_FILE, "", base.JSON_FORMAT_INVALID)))
			return
		}
		objs := make([]*base.ResponseStatusObject, 0)
		if len(rm.FileListObject.File) > 0 {
			for _, item := range rm.FileListObject.File {
				objs = append(objs, base.BuildResponseObject(base.URL_FILE, item.FileInfo.FileID, base.OK))
			}
		}
		c.JSON(http.StatusOK, base.BuildRespnse(objs...))
	})
	//文件数据
	engine.POST(base.URL_FILE+"/:id/Data", func(c *gin.Context) {
		c.JSON(http.StatusOK, base.BuildSingleResponse(base.BuildResponseObject(c.Request.RequestURI, c.Param("id"), base.OK)))
	})

	server := &http.Server{
		Handler: engine,
		Addr:    ":" + s.serverPort,
	}
	s.httpServer = server
	err := server.ListenAndServe()
	if err != nil {
		logger.LOG_ERROR("1400服务启动失败：", err)
		s.httpServer = nil
		return
	}
}

func (s *Gat1400Server) Close() error {
	if s.httpServer != nil {
		_ = s.httpServer.Close()
		s.httpServer = nil
	}
	if s.cancel != nil {
		s.cancel()
	}
	return nil
}

/**
 * 注册
 */
func (s *Gat1400Server) regist(c *gin.Context) {
	var err error
	rm := &gat1400_model.RegisterModel{}
	err = jsoniter.NewDecoder(c.Request.Body).Decode(rm)
	if err != nil {
		c.JSON(http.StatusOK, base.BuildSingleResponse(base.BuildResponseObject(base.URL_REGIST, "", base.JSON_FORMAT_INVALID)))
		return
	}
	viewID := rm.GetViewID() //视图库id & 设备id
	if viewID == "" {
		c.JSON(http.StatusOK, base.BuildSingleResponse(base.BuildResponseObject(base.URL_REGIST, viewID, base.VIEWID_IS_NULL)))
		return
	}
	if viewID != s.viewLibId && !context.ExsitGbId(viewID) {
		c.JSON(http.StatusOK, base.BuildSingleResponse(base.BuildResponseObject(base.URL_REGIST, viewID, base.DEVICEID_IS_NOT_EXIST)))
		return
	}
	//header params
	authorization := c.GetHeader("Authorization")
	if authorization == "" {
		//第一次注册
		nonce := strings.ReplaceAll(uuid.NewV4().String(), "-", "")[0:12]
		authorization = fmt.Sprintf(`Digest realm="myrealm",qop="auth",nonce="%s"`, nonce)
		mh := textproto.MIMEHeader(c.Writer.Header())
		mh["WWW-Authenticate"] = []string{authorization}
		c.JSON(http.StatusUnauthorized, base.BuildSingleResponse(base.BuildResponseObject(base.URL_REGIST, viewID, base.VIEWID_IS_NULL)))
		return
	}

	//第二次注册
	params := make(map[string]string)
	authorization = strings.Trim(strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(authorization, "Digest", ""), "'", ""), `"`, ""), " ")
	args := strings.Split(authorization, ",")
	for _, item := range args {
		idx := strings.Index(item, "=")
		if idx > 0 {
			params[strings.Trim(item[:idx], " ")] = strings.Trim(item[idx+1:], " ")
		}
	}
	keys := []string{
		"username",
		"realm",
		"nonce",
		"cnonce",
		"response",
		"algorithm",
		"qop",
		"nc"}
	for _, k := range keys {
		if params[k] == "" {
			c.JSON(http.StatusUnauthorized, base.BuildSingleResponse(base.BuildResponseObject(base.URL_REGIST, viewID, base.INVALID_OPERATION)))
			return
		}
	}
	if params["username"] != s.username {
		c.JSON(http.StatusUnauthorized, base.BuildSingleResponse(base.BuildResponseObject(base.URL_REGIST, viewID, base.INVALID_OPERATION)))
		return
	}
	//check auth
	h1 := util.MD5(params["username"] + ":" + params["realm"] + ":" + s.password)
	logger.LOG_INFO("注册H1-params:", params["username"]+":"+params["realm"]+":"+s.password)
	logger.LOG_INFO("注册H1:", h1)
	h2 := util.MD5("POST:/VIID/System/Register")
	logger.LOG_INFO("注册H2-params:", "POST:/VIID/System/Register")
	logger.LOG_INFO("注册H2:", h2)
	res := util.MD5(h1 + ":" + params["nonce"] + ":" + params["nc"] + ":" + params["cnonce"] + ":" + params["qop"] + ":" + h2)
	logger.LOG_INFO("注册response-params:", h1+":"+params["nonce"]+":"+params["nc"]+":"+params["cnonce"]+":"+params["qop"]+":"+h2)
	logger.LOG_INFO("注册response:", res)
	logger.LOG_INFO("收到的response:", params["response"])
	if res != params["response"] {
		c.JSON(http.StatusUnauthorized, base.BuildSingleResponse(base.BuildResponseObject(base.URL_REGIST, viewID, base.RESPONSE_NOT_CORRECT)))
		return
	}
	//注册成功
	s.Lock()
	s.registers[viewID] = &RegisterSession{
		ViewLibID:     viewID,
		Ip:            c.ClientIP(),
		RegistTime:    time.Now(),
		KeepaliveTime: time.Now(),
	}
	s.Unlock()
	c.JSON(http.StatusOK, base.BuildSingleResponse(base.BuildResponseObject(base.URL_REGIST, viewID, base.OK)))
}

/**
取消注冊

*/
func (s *Gat1400Server) unRegist(c *gin.Context) {
	viewID := ""
	if c.Request.Header.Get("User-Identify") != "" {
		viewID = c.Request.Header.Get("User-Identify")
	}
	c.JSON(http.StatusOK, base.BuildSingleResponse(base.BuildResponseObject(base.URL_UNREGIST, viewID, base.OK)))
}

/**
 * 保活
 */
func (s *Gat1400Server) keepalive(c *gin.Context) {
	var err error
	km := &gat1400_model.KeepaliveModel{}
	err = jsoniter.NewDecoder(c.Request.Body).Decode(km)
	if err != nil {
		c.JSON(http.StatusOK, base.BuildSingleResponse(base.BuildResponseObject(base.URL_KEEPALIVE, "", base.JSON_FORMAT_INVALID)))
		return
	}
	viewLibId := km.GetViewID()
	if viewLibId == "" {
		c.JSON(http.StatusOK, base.BuildSingleResponse(base.BuildResponseObject(base.URL_KEEPALIVE, viewLibId, base.VIEWID_IS_NULL)))
		return
	}
	s.Lock()
	defer s.Unlock()
	r, ok := s.registers[viewLibId]
	if r == nil || !ok {
		c.JSON(http.StatusOK, base.BuildSingleResponse(base.BuildResponseObject(base.URL_KEEPALIVE, viewLibId, base.INVALID_OPERATION)))
		return
	}
	r.KeepaliveTime = time.Now()
	c.JSON(http.StatusOK, base.BuildSingleResponse(base.BuildResponseObject(base.URL_KEEPALIVE, viewLibId, base.OK)))
	return
}

/**
 * 校时
 */
func (s *Gat1400Server) time(c *gin.Context) {
	c.JSON(http.StatusOK, BuildSystemTime())
	return
}

/**
 * 采集数据
 */
func (s *Gat1400Server) receive(dataType string, c *gin.Context) {
	start := time.Now()
	viewLibId := c.GetHeader("User-Identify")
	if viewLibId == "" {
		c.JSON(http.StatusOK, base.BuildRespnse(base.BuildResponseObject(base.URL_FACES, "", base.VIEWID_IS_NULL)))
		return
	}
	//校验视图库id或设备id是否合法
	if viewLibId != s.viewLibId && !context.ExsitGbId(viewLibId) {
		c.JSON(http.StatusOK, base.BuildRespnse(base.BuildResponseObject(base.URL_REGIST, viewLibId, base.UNAUTHORIZED)))
		return
	}
	//校验认证session
	if s.openAuth {
		if _, ok := s.registers[viewLibId]; !ok {
			c.JSON(http.StatusUnauthorized, base.BuildRespnse(base.BuildResponseObject(base.URL_REGIST, viewLibId, base.UNAUTHORIZED)))
			return
		}
	}
	wrap, err := gat1400_model.BuildFromJson(dataType, c.Request.Body)
	if err != nil {
		c.JSON(http.StatusOK, base.BuildRespnse(base.BuildResponseObject(base.URL_FACES, "", base.JSON_FORMAT_INVALID)))
		return
	}
	logger.LOG_WARN("receive 耗时：" + time.Since(start).String())
	err = s.emit([]*gat1400_model.Gat1400Wrap{wrap})
	if err == nil {
		c.JSON(http.StatusOK, wrap.BuildResponse(base.OK))
	} else if base.ExsitErrCode(err.Error()) {
		c.JSON(http.StatusOK, wrap.BuildResponse(err.Error()))
	} else {
		c.JSON(http.StatusOK, wrap.BuildResponse(base.OTHER_ERROR))
	}
}

//session 失效检测
func (s *Gat1400Server) checkKeepaliveSession() {
	duration := s.sessionTimeout
	interval := duration / 3
LOOP:
	for {
		select {
		case <-s.ctx.Done():
			break LOOP
		default:
		}
		s.Lock()
		if len(s.registers) != 0 {
			for k, r := range s.registers {
				if time.Since(r.KeepaliveTime) > duration {
					delete(s.registers, k)
					logger.LOG_INFO("视图库已超时，移除session：", r.ViewLibID)
				}
			}
		}
		s.Unlock()
		time.Sleep(interval)
	}
}
