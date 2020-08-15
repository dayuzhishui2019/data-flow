package proxy

import (
	"bytes"
	"dyzs/data-flow/logger"
	"dyzs/data-flow/model/videocmd"
	"dyzs/data-flow/util"
	"errors"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	_URL_GALAXY_CMD = "http://galaxy:8200/cmd"

	TARGET_GALAXY = "galaxy"
	TARGET_MEDIA  = "media"
)

type VideoCmdOperator interface {
	OpenStream(param *videocmd.CmdOpenStream) (result interface{}, err error)
	CloseStream(param *videocmd.CmdCloseStream) (result interface{}, err error)
	PTZ(param *videocmd.CmdPTZ) (result interface{}, err error)
	HistoryVideo(param *videocmd.CmdHistoryVideo) (result interface{}, err error)
}

func responseSuccess(c *gin.Context, result interface{}) {
	c.JSON(http.StatusOK, map[string]interface{}{
		"Code":    http.StatusOK,
		"Message": "success",
		"Result":  result,
	})
}
func responseFail(c *gin.Context, err error) {
	c.JSON(http.StatusOK, map[string]interface{}{
		"Code":    http.StatusInternalServerError,
		"Message": err.Error(),
	})
}

type ResponseWrap struct {
	Code    int                 `json:"Code"`
	Message string              `json:"Message"`
	Result  jsoniter.RawMessage `json:"Result"`
}

func AttachVideoCmd(operator VideoCmdOperator) {
	serverIns.POST("/cmd", func(c *gin.Context) {
		body, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			responseFail(c, err)
			return
		}
		cmd := &videocmd.Cmd{}
		err = jsoniter.Unmarshal(body, cmd)
		if err != nil {
			responseFail(c, err)
			return
		}
		switch cmd.Cmd {
		case videocmd.CMD_OPEN_STREAM:
			openStream(c, cmd.Param, operator)
		case videocmd.CMD_CLOSE_STREAM:
			closeStream(c, cmd.Param, operator)
		case videocmd.CMD_PTZ:
			ptz(c, cmd.Param, operator)
		case videocmd.CMD_HISTORY_VIDEO:
			historyVideo(c, cmd.Param, operator)
		default:
			logger.LOG_WARN("未找到指令的处理方法：", cmd.Cmd)
		}
	})
}

func SubmitVideoChannelToGalaxy(catalogs []*videocmd.ParamDeviceChannel) error {
	return util.Retry(func() error {
		return request(_URL_GALAXY_CMD, http.MethodPost, "application/json", map[string]string{"Target": TARGET_GALAXY}, map[string]interface{}{
			"Cmd": "ResponseCatalog",
			"Param": map[string]interface{}{
				"CatalogList": catalogs,
			},
		}, nil)
	}, 3, 1*time.Second)
}

//============================================  cmd func  ==================================================

func openStream(c *gin.Context, paramBytes jsoniter.RawMessage, operator VideoCmdOperator) {
	param := &videocmd.CmdOpenStream{}
	err := jsoniter.Unmarshal(paramBytes, param)
	if err != nil {
		responseFail(c, err)
		return
	}
	result, err := operator.OpenStream(param)
	if err != nil {
		responseFail(c, err)
		return
	}
	//调用媒体开流
	openStreamRes := &videocmd.ParamMedieOpenResult{}
	err = util.Retry(func() error {
		return request(_URL_GALAXY_CMD, http.MethodPost, "application/json", map[string]string{"Target": TARGET_MEDIA}, result, openStreamRes)
	}, 3, 1*time.Second)
	if err != nil {
		logger.LOG_WARN("调用媒体服务开流异常：", err)
		responseFail(c, errors.New("调用媒体服务开流异常："+err.Error()))
		return
	}
	responseSuccess(c, openStreamRes)
}

func closeStream(c *gin.Context, paramBytes jsoniter.RawMessage, operator VideoCmdOperator) {
	param := &videocmd.CmdCloseStream{}
	err := jsoniter.Unmarshal(paramBytes, param)
	if err != nil {
		responseFail(c, err)
		return
	}
	result, err := operator.CloseStream(param)
	if err != nil {
		logger.LOG_WARN("调用媒体服务关流异常：", err)
		responseFail(c, errors.New("调用媒体服务关流异常："+err.Error()))
		return
	}
	err = util.Retry(func() error {
		return request(_URL_GALAXY_CMD, http.MethodPost, "application/json", map[string]string{"Target": TARGET_MEDIA}, result, nil)
	}, 3, 1*time.Second)
	if err != nil {
		responseFail(c, err)
		return
	}
	responseSuccess(c, nil)
}

func ptz(c *gin.Context, paramBytes jsoniter.RawMessage, operator VideoCmdOperator) {
	param := &videocmd.CmdPTZ{}
	err := jsoniter.Unmarshal(paramBytes, param)
	if err != nil {
		responseFail(c, err)
		return
	}
	result, err := operator.PTZ(param)
	if err != nil {
		logger.LOG_WARN("调用云台控制异常：", err)
		responseFail(c, errors.New("调用云台控制异常："+err.Error()))
		return
	}
	responseSuccess(c, result)
}

func historyVideo(c *gin.Context, paramBytes jsoniter.RawMessage, operator VideoCmdOperator) {
	param := &videocmd.CmdHistoryVideo{}
	err := jsoniter.Unmarshal(paramBytes, param)
	if err != nil {
		responseFail(c, err)
		return
	}
	result, err := operator.HistoryVideo(param)
	if err != nil {
		logger.LOG_WARN("调用查询录像异常：", err)
		responseFail(c, errors.New("调用查询录像异常："+err.Error()))
		return
	}
	responseSuccess(c, result)
}

//=========================================  http request  ===============================================

var httpClient = &http.Client{
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

//http请求
func request(url, method, contentType string, headers map[string]string, body interface{}, resPointer interface{}) error {
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
		if headers != nil {
			for k, v := range headers {
				req.Header.Set(k, v)
			}
		}
		req.Header.Set("Content-Type", contentType)
		res, err := httpClient.Do(req)
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
		if res.StatusCode != http.StatusOK {
			return errors.New(string(resBytes))
		}
		return nil
	}, 3, 3*time.Second)
	if err != nil {
		return err
	}
	logger.LOG_INFO("http-response:", string(resBytes))
	wrap := &ResponseWrap{}
	err = jsoniter.Unmarshal(resBytes, wrap)
	if err != nil {
		return err
	}
	if resPointer != nil {
		return jsoniter.Unmarshal(wrap.Result, resPointer)
	}
	return nil
}
