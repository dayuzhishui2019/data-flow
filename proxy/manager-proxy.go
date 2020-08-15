package proxy

import (
	"dyzs/data-flow/context"
	"dyzs/data-flow/logger"
	"dyzs/data-flow/model"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"net/http"
)

const (
	URL_PREFIX = "/mapi"
)

var serverIns *gin.Engine

func StartManagerProxy(port string) {
	if port == "" {
		port = "7777"
	}
	server := gin.Default()
	server.Handle(http.MethodGet, "/debug", ChangeLogLevel)
	router := server.Group(URL_PREFIX)
	router.Handle(http.MethodPost, "/init", Init)
	router.Handle(http.MethodPost, "/heart", KeepAlive)
	router.Handle(http.MethodPost, "/assignResource", AssignResource)
	router.Handle(http.MethodPost, "/revokeResource", RevokeResource)
	go server.Run(":" + port)
	//go server.Run(":7777")
	serverIns = server
}


func ChangeLogLevel(ctx *gin.Context) {
	level := ctx.Query("level")
	if level != "" {
		logger.ChangeLevel(level)
	}
}


type ReponseBody struct {
	Code int         `json:"code"`
	Msg  interface{} `json:"msg"`
}

func response(c *gin.Context, code int, data interface{}) {
	if data == nil {
		data = struct{}{}
	}
	c.JSON(code, ReponseBody{
		Code: code,
		Msg:  data,
	})
}

//初始化
func Init(ctx *gin.Context) {
	task := &model.Task{}
	err := ctx.BindJSON(task)
	if err != nil {
		logger.LOG_WARN("任务下发解析失败：", err)
		return
	}
	logger.LOG_WARN("任务下发：", *task)
	//任务信息
	context.Set("$task", task)
	resources := task.GetResources()
	if len(resources) > 0 {
		context.AssignResources(resources)
	}
	config := task.AccessParam
	configMap := make(map[string]interface{})
	logger.LOG_WARN("任务配置参数：", config)
	if config != "" {
		err = jsoniter.Unmarshal([]byte(config), &configMap)
		if err != nil {
			logger.LOG_WARN("任务配置参数解析失败：", err)
		}
		for k, v := range configMap {
			context.Set(k, v)
		}
	}

	response(ctx, http.StatusOK, nil)
}

//心跳接口
func KeepAlive(ctx *gin.Context) {
	response(ctx, http.StatusOK, nil)
}

//下发资源
func AssignResource(ctx *gin.Context) {
	var err error
	resources := make([]*model.Resource, 0)
	err = ctx.BindJSON(&resources)
	if err != nil {
		logger.LOG_WARN("配置参数解析失败", err)
		response(ctx, http.StatusBadRequest, err)
		return
	}
	context.AssignResources(resources)
	response(ctx, http.StatusOK, nil)
}

//移除资源
func RevokeResource(ctx *gin.Context) {
	var err error
	resources := make([]string, 0)
	err = ctx.BindJSON(&resources)
	if err != nil {
		logger.LOG_WARN("配置参数解析失败", err)
		response(ctx, http.StatusBadRequest, err)
		return
	}
	context.RevokeResources(resources)
	response(ctx, http.StatusOK, nil)
}
