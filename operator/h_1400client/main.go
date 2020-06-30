package dag_plugin_1400client

import (
	"bytes"
	context2 "context"
	"dyzs/data-flow/concurrent"
	"dyzs/data-flow/context"
	"dyzs/data-flow/logger"
	"dyzs/data-flow/model/gat1400"
	"dyzs/data-flow/model/gat1400/base"
	"dyzs/data-flow/model/kafka"
	"dyzs/data-flow/stream"
	"dyzs/data-flow/util"
	"errors"
	"fmt"
	json "github.com/json-iterator/go"
	"io/ioutil"
	"math/rand"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"
)

func init() {
	stream.RegistHandler("1400client", func() stream.Handler {
		return &Gat1400Client{}
	})
}

var SEND_DATA_URLS = map[string]string{
	gat1400.GAT1400_FACE:     base.URL_FACES,
	gat1400.GAT1400_BODY:     base.URL_PERSONS,
	gat1400.GAT1400_VEHICLE:  base.URL_VEHICLE,
	gat1400.GAT1400_NONMOTOR: base.URL_NOMOTORS,
}

var ERR_CDX = errors.New("1400重定向")

type Gat1400Client struct {
	sync.Mutex

	targetPlatformId    string //目标平台id
	userIdentify        string //视图库id
	viewLibAddr         string //上级ip
	locationViewLibAddr string //重定向上级ip
	openAuth            bool   //是否开启注册
	username            string //注册用户名
	password            string //注册密码
	keepaliveInterval   int    //保活时间间隔
	RegistMatsrAddr     string //注册重定向使用
	executor            *concurrent.Executor
	httpClient          *http.Client

	linking bool //是否已注册连接

	ctx    context2.Context
	cancel context2.CancelFunc
}

func (c *Gat1400Client) Init(config interface{}) error {
	targetPlatformId := context.GetString("1400client_platformId")
	userIdentify := context.GetString("1400client_userIdentify")
	viewLibAddr := context.GetString("1400client_viewLibAddr")
	openAuth := context.GetBool("1400client_openAuth")
	username := context.GetString("1400client_username")
	password := context.GetString("1400client_password")
	keepaliveInterval := context.GetInt("1400client_keepaliveInterval")
	httpPoolsize := context.GetInt("1400client_httpPoolsize")
	if httpPoolsize <= 0 {
		httpPoolsize = 20
	}
	logger.LOG_WARN("------------------ 1400client config ------------------")
	logger.LOG_WARN("1400client_platformId : " + targetPlatformId)
	logger.LOG_WARN("1400client_userIdentify : " + userIdentify)
	logger.LOG_WARN("1400client_viewLibAddr : " + viewLibAddr)
	logger.LOG_WARN("1400client_openAuth : " + strconv.FormatBool(openAuth))
	logger.LOG_WARN("1400client_username : " + username)
	logger.LOG_WARN("1400client_password : " + password)
	logger.LOG_WARN("1400client_keepaliveInterval : " + strconv.Itoa(keepaliveInterval))
	logger.LOG_WARN("1400client_httpPoolsize : " + strconv.Itoa(httpPoolsize))
	logger.LOG_WARN("------------------------------------------------------")
	if targetPlatformId == "" {
		return errors.New("[1400client] targetPlatformId不能为空")
	}
	if userIdentify == "" {
		return errors.New("[1400client] userIdentify不能为空")
	}
	if viewLibAddr == "" {
		return errors.New("[1400client] viewLibAddr不能为空")
	}
	if openAuth && (username == "" || password == "") {
		return errors.New("[1400client] 开启认证时，1400client_username 和 1400client_password 不能为空")
	}
	c.linking = false
	c.targetPlatformId = targetPlatformId
	c.viewLibAddr = viewLibAddr
	c.userIdentify = userIdentify
	c.openAuth = openAuth
	c.username = username
	c.password = password
	c.keepaliveInterval = keepaliveInterval
	c.executor = concurrent.NewExecutor(httpPoolsize)
	c.httpClient = &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives:   false, //false 长链接 true 短连接
			Proxy:               http.ProxyFromEnvironment,
			MaxIdleConns:        httpPoolsize * 5, //client对与所有host最大空闲连接数总和
			MaxConnsPerHost:     httpPoolsize,
			MaxIdleConnsPerHost: httpPoolsize,     //连接池对每个host的最大连接数量,当超出这个范围时，客户端会主动关闭到连接
			IdleConnTimeout:     60 * time.Second, //空闲连接在连接池中的超时时间
		},
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	c.ctx, c.cancel = context2.WithCancel(context2.Background())

	//开启注册保活定时任务
	if c.openAuth {
		c.mustRegist()
		go c.loopKeepalive()
	} else {
		c.linking = true
	}

	return nil
}

func (c *Gat1400Client) getViewLibAddr() string {
	if c.locationViewLibAddr != "" {
		return c.locationViewLibAddr
	}
	return c.viewLibAddr
}

func (c *Gat1400Client) mustRegist() {
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}
		if c.linking {
			continue
		}
		err := c.regist()
		if err != nil {
			if err == ERR_CDX {
				continue
			}
			logger.LOG_ERROR("注册失败：", err)
			time.Sleep(3 * time.Second)
			//注册失败时，将重定向地址重置
			if c.locationViewLibAddr != "" {
				c.locationViewLibAddr = ""
			}
			continue
		}
		return
	}
}

func (c *Gat1400Client) regist() error {
	currentViewLibAddr := c.getViewLibAddr()
	//注册
	registParams, _ := json.Marshal(gat1400.BuildGat1400RegisterObj(c.userIdentify))
	registParamsBytes := bytes.NewBuffer(registParams)
	req, err := http.NewRequest("POST", "http://"+currentViewLibAddr+base.URL_REGIST, registParamsBytes)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", base.CONTENT_TYPE)
	req.Header.Set("User-Identify", c.userIdentify)
	req.Header.Set("Connection", "keep-alive")
	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	resBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if res.StatusCode == http.StatusOK {
		//注册成功（上级未实现二次注册）
		return nil
	}
	if res.StatusCode == http.StatusMovedPermanently || res.StatusCode == http.StatusFound {
		//重定向，再次注册
		location := res.Header.Get("Location")
		if location == "" {
			return errors.New("重定向地址为空")
		}
		logger.LOG_INFO("重定向地址为： ", location)
		uri := strings.Split(strings.Trim(strings.ReplaceAll(location, "http://", ""), " "), ":")
		if len(uri) != 2 {
			return errors.New("重定向地址错误:" + location)
		}
		c.locationViewLibAddr = uri[0] + ":" + uri[1]
		return ERR_CDX
	}
	if res.StatusCode != http.StatusUnauthorized {
		//注册失败
		return errors.New(strconv.Itoa(res.StatusCode) + " : " + string(resBody))
	}
	//一次注册成功，开始二次注册
	//第二次注册
	firstAuthorization := res.Header.Get(base.REGIST_RETURN_AUTHORIZATION)
	//Digest realm="myrealm",qop="auth" Digest realm="myrealm",qop="auth",nonce="6a0526b08a2a"
	logger.LOG_DEBUG("第一次注册返回的AUTHORIZATION : ", firstAuthorization)
	params := make(map[string]string)
	args := strings.Split(strings.ReplaceAll(strings.ReplaceAll(strings.Trim(firstAuthorization, "Digest"), `"`, ""), `'`, ""), ",")
	for _, item := range args {
		idx := strings.Index(item, "=")
		if idx > 0 {
			params[strings.Trim(item[:idx], " ")] = strings.Trim(item[idx+1:], " ")
		}
	}
	//使用md5加密
	params["uri"] = base.URL_REGIST
	params["username"] = c.username
	params[base.ALGORITHM] = "MD5"
	params[base.NC] = "00000001"
	params[base.CNONCE] = GetRandomString(12)
	// 计算response
	//HA1 = MD5(A1) = MD5(name:realm:password)
	HA1 := util.MD5(c.username + ":" + params[base.REALM] + ":" + c.password)
	logger.LOG_INFO("注册 - H1：", HA1)
	//HA2 = MD5(A2) = MD5(method:uri)
	HA2 := util.MD5("POST" + ":" + base.URL_REGIST)
	logger.LOG_INFO("注册 - H2：", HA1)
	//response = MD5(HA1:nonce:nc:cnonce:qop:HA2)
	responseStr := HA1 + ":" + params[base.NONCE] + ":" + params[base.NC] + ":" + params[base.CNONCE] + ":" + params[base.QOP] + ":" + HA2
	logger.LOG_INFO("注册 - 加密前response：", responseStr)
	responseMD5 := util.MD5(responseStr)
	logger.LOG_INFO("注册 - 加密后response：", responseMD5)
	params[base.RESPONSE] = responseMD5
	// 构建返回的认证头Authorization
	//eg:Digest username='chongqing', realm='myrealm', nonce='0efb398cf026',
	//uri='/VIID/System/Register', response='66254bd928831fd5e5a72a1a509d8235', algorithm='MD5',
	//qop=auth, nc=00000001, cnonce='7cb15baf061f',opaque='df387b3a3e534e8c801c3a6ebc9d65d4'
	keys := []string{
		"username",
		"realm",
		"nonce",
		"uri",
		"cnonce",
		"response",
		"algorithm",
		"qop",
		"nc"}
	var authStr strings.Builder
	authStr.WriteString("Digest ")
	for _, k := range keys {
		if k == "nc" || k == "qop" {
			authStr.WriteString(fmt.Sprintf(`%s=%s,`, k, params[k]))
		} else {
			authStr.WriteString(fmt.Sprintf(`%s="%s",`, k, params[k]))
		}
	}
	secondAuthorization := strings.Trim(authStr.String(), ",")
	logger.LOG_INFO("第二次请求Authorization :", secondAuthorization)
	req, err = http.NewRequest("POST", "http://"+currentViewLibAddr+base.URL_REGIST, bytes.NewBuffer(registParams))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", base.CONTENT_TYPE)
	req.Header.Set("User-Identify", c.userIdentify)
	req.Header.Set("Authorization", secondAuthorization)
	req.Header.Set("Connection", "keep-alive")
	res, err = c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	resBody, err = ioutil.ReadAll(res.Body)
	if res.StatusCode != http.StatusOK {
		//二次注册失败
		return errors.New(strconv.Itoa(res.StatusCode) + " : " + string(resBody))
	}
	err = isGat1400ResponseSuccess(resBody)
	if err != nil {
		return err
	}
	//二次注册成功
	logger.LOG_INFO("注册成功")
	c.Lock()
	c.linking = true
	c.Unlock()
	return nil
}

/**
* 循环保活
 */
func (c *Gat1400Client) loopKeepalive() {
	interval := c.keepaliveInterval
	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}
		time.Sleep(time.Duration(interval) * time.Second)
		//保活
		err := func() error {
			keepaliveParams, _ := json.Marshal(gat1400.BuildGat1400KeepaliveObject(c.userIdentify))
			req, err := http.NewRequest("POST", "http://"+c.viewLibAddr+base.URL_KEEPALIVE, bytes.NewBuffer(keepaliveParams))
			if err != nil {
				return err
			}
			req.Header.Set("Content-Type", base.CONTENT_TYPE)
			req.Header.Set("User-Identify", c.userIdentify)
			req.Header.Set("Connection", "keep-alive")
			res, err := c.httpClient.Do(req)
			if err != nil {
				return err
			}
			defer res.Body.Close()
			resBody, err := ioutil.ReadAll(res.Body)
			if res.StatusCode != http.StatusOK {
				//保活失败
				return errors.New(strconv.Itoa(res.StatusCode) + " : " + string(resBody))
			}
			err = isGat1400ResponseSuccess(resBody)
			if err != nil {
				return err
			}
			return nil
		}()
		if err != nil {
			logger.LOG_ERROR("保活失败：", err)
			c.Lock()
			c.linking = false
			c.Unlock()
		}
	}
}

func isGat1400ResponseSuccess(resBody []byte) error {
	gat1400res, err := base.DecodeBytesToReponse(resBody)
	if err != nil {
		return err
	}
	var statusCode string
	//list
	if gat1400res.ResponseStatusListObject != nil && len(gat1400res.ResponseStatusListObject.ResponseStatusObject) > 0 {
		statusCode = gat1400res.ResponseStatusListObject.ResponseStatusObject[0].StatusCode
	}
	//single
	if statusCode == "" {

	}
	//statusobj
	if gat1400res.ResponseStatusListObject.ResponseStatusObject[0].StatusCode != base.OK {
		return errors.New("注册响应错误：" + string(resBody))
	}
	return nil
}

// GetRandomString 生成 数字和小写字母
func GetRandomString(l int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyz"
	randomBytes := []byte(str)
	var result []byte
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < l; i++ {
		result = append(result, randomBytes[r.Intn(len(randomBytes))])
	}
	return string(result)
}

func (c *Gat1400Client) Handle(data interface{}, next func(interface{}) error) error {
	wraps, ok := data.([]*gat1400.Gat1400Wrap)
	if !ok {
		logger.LOG_ERROR("1400client 转换数据异常", nil)
		return errors.New(fmt.Sprintf("Handle [1400client] 数据格式错误，need []*daghub.Gat1400Wrap , get %T", reflect.TypeOf(data)))
	}
	if len(wraps) == 0 {
		return nil
	}
	if !c.linking {
		c.mustRegist()
	}
	//发送数据
	successMsgs := make([]*kafka.KafkaMessage, 0)
	var lock sync.Mutex
	tasks := make([]func(), 0)
	for _, w := range wraps {
		func(wrap *gat1400.Gat1400Wrap) {
			tasks = append(tasks, func() {
				err := util.Retry(func() error {
					json, err := wrap.BuildToJson()
					if err != nil {
						return err
					}
					req, err := http.NewRequest("POST", "http://"+c.viewLibAddr+SEND_DATA_URLS[w.DataType], bytes.NewBuffer(json))
					if err != nil {
						return err
					}
					req.Header.Set("Content-Type", base.CONTENT_TYPE)
					req.Header.Set("User-Identify", c.userIdentify)
					req.Header.Set("Connection", "keep-alive")
					res, err := c.httpClient.Do(req)
					if err != nil {
						return err
					}
					defer res.Body.Close()
					resBody, err := ioutil.ReadAll(res.Body)
					if err != nil {
						return err
					}
					if res.StatusCode != http.StatusOK {
						return errors.New(strconv.Itoa(res.StatusCode) + " : " + string(resBody))
					}
					gat1400Res, err := base.DecodeBytesToReponse(resBody)
					if err != nil {
						return err
					}
					if len(gat1400Res.ResponseStatusListObject.ResponseStatusObject) > 0 && gat1400Res.ResponseStatusListObject.ResponseStatusObject[0].StatusCode == base.OK {
						//all success
						kafkaMsg := &kafka.KafkaMessage{
							Topic: "transmit",
							Value: nil,
						}
						digestBytes, err := wrap.BuildDigest(gat1400.DIGEST_TRANSIMIT, "", c.targetPlatformId)
						if err == nil {
							kafkaMsg.SetHeader("digest", digestBytes)
							lock.Lock()
							successMsgs = append(successMsgs, kafkaMsg)
							lock.Unlock()
						}
						return nil
					}
					return errors.New("发送失败：" + string(resBody))
				}, 3, time.Second*3)
				if err != nil {
					logger.LOG_WARN(err)
				}
			})
		}(w)
	}

	_ = c.executor.SubmitSyncBatch(tasks)
	if len(successMsgs) > 0 {
		return next(successMsgs)
	}
	return nil
}

func (c *Gat1400Client) Close() error {
	if c.cancel != nil {
		c.cancel()
	}
	if c.executor != nil {
		c.executor.Close()
		c.executor = nil
	}
	c.linking = false
	return nil
}
