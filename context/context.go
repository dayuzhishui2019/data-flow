package context

import (
	"dyzs/data-flow/logger"
	"dyzs/data-flow/model"
	"errors"
	"fmt"
	"github.com/spf13/viper"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

const (
	PARAM_KEY_NODE_ID             = "NODE_ID"
	PARAM_KEY_COMPONENT_ID        = "COMPONENT_ID"
	PARAM_KEY_NODE_ADDR           = "NODE_ADDR"
	PARAM_KEY_FORWARD_PLATFORM_ID = "FORWARD_PLATFORM_ID"
)
const CONFIG_RESOURCE_REFRESH_DELAY = 3 //1秒延迟触发变更
const CONFIG_KEY_PLATFORM_ID = "platformId"

var delayRefreshConfig int32
var delayRefreshResource int32
var rwLock sync.RWMutex

func InitConfig(configPath string) {
	logger.LOG_INFO("configPath:", configPath)
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		logger.LOG_ERROR("Fail to read config file :", err)
	}
}
func Init() {
	go func() {
		for {
			time.Sleep(CONFIG_RESOURCE_REFRESH_DELAY * time.Second)
			if atomic.LoadInt32(&delayRefreshConfig) == 1 {
				if len(configWatchers) > 0 {
					for _, cb := range configWatchers {
						cb()
					}
				}
				atomic.StoreInt32(&delayRefreshConfig, 0)
			}
			if atomic.LoadInt32(&delayRefreshResource) == 1 {
				if len(resourceWatchers) > 0 {
					for _, cb := range resourceWatchers {
						cb()
					}
				}
				atomic.StoreInt32(&delayRefreshResource, 0)
			}
		}
	}()
	initEnv()
}

//环境变量
func initEnv() {
	envs := []string{"CENTER_IP", "CENTER_PORT"}
	for _, k := range envs {
		Set(k, os.Getenv(k))
		fmt.Println("[ENV]", k, ":", os.Getenv(k))
	}
}

func GetTask() (*model.Task, error) {
	task, ok := viper.Get("$task").(*model.Task)
	if task == nil || !ok {
		return nil, errors.New("无任务")
	}
	return task, nil
}

//配置
func Set(key string, value interface{}) {
	rwLock.Lock()
	viper.Set(key, value)
	rwLock.Unlock()
	atomic.StoreInt32(&delayRefreshConfig, 1)
}
func Exsit(keys ...string) (unExsitKeys []string) {
	unExsitKeys = make([]string, 0)
	for _, k := range keys {
		v := viper.IsSet(k)
		if !v {
			unExsitKeys = append(unExsitKeys, k)
		}
	}
	return unExsitKeys
}
func IsExsit(keys ...string) bool {
	return len(Exsit(keys...)) == 0
}
func GetString(key string) string {
	rwLock.RLock()
	defer rwLock.RUnlock()
	return viper.GetString(key)
}
func GetInt(key string) int {
	rwLock.RLock()
	defer rwLock.RUnlock()
	return viper.GetInt(key)
}
func GetInt32(key string) int32 {
	rwLock.RLock()
	defer rwLock.RUnlock()
	return viper.GetInt32(key)
}
func GetInt64(key string) int64 {
	rwLock.RLock()
	defer rwLock.RUnlock()
	return viper.GetInt64(key)
}
func GetBool(key string) bool {
	rwLock.RLock()
	defer rwLock.RUnlock()
	return viper.GetBool(key)
}
func AssignConfig(cfgs []map[string]interface{}) {
	if len(cfgs) > 0 {
		for _, m := range cfgs {
			if len(m) > 0 {
				for k, v := range m {
					Set(k, v)
					if k == CONFIG_KEY_PLATFORM_ID {
						Set(PARAM_KEY_FORWARD_PLATFORM_ID, v)
					}
				}
			}
		}
	}
}

//资源
//设备
var RESOURCE_ID_EQ = make(map[string]*model.Resource)
var RESOURCE_GBID_EQ = make(map[string]*model.Resource)

func AssignResources(eqs []*model.Resource) {
	if len(eqs) > 0 {
		rwLock.Lock()
		defer rwLock.Unlock()
		for _, eq := range eqs {
			RESOURCE_ID_EQ[eq.ID] = eq
			RESOURCE_GBID_EQ[eq.GbID] = eq
		}
		logger.LOG_WARN("共接收资源 ===>", len(eqs))
		logger.LOG_WARN("当前资源 ===>", len(RESOURCE_ID_EQ))
		refreshResource()
	}
}

func RevokeResources(ids []string) {
	if len(ids) > 0 {
		rwLock.Lock()
		defer rwLock.Unlock()
		for _, id := range ids {
			eq, ok := RESOURCE_ID_EQ[id]
			if ok && eq != nil {
				delete(RESOURCE_GBID_EQ, eq.GbID)
			}
			delete(RESOURCE_ID_EQ, id)
		}
		logger.LOG_WARN("共移除资源 ===>", len(ids))
		logger.LOG_WARN("当前资源 ===>", len(RESOURCE_ID_EQ))
		refreshResource()
	}
}

func refreshResource() {
	atomic.StoreInt32(&delayRefreshResource, 1)
}

func GetResource(id string) (res *model.Resource, ok bool) {
	rwLock.RLock()
	defer rwLock.RUnlock()
	res, ok = RESOURCE_ID_EQ[id]
	if ok {
		return res, ok
	}
	res, ok = RESOURCE_GBID_EQ[id]
	if ok {
		return res, ok
	}
	return nil, false
}


func ReadAllResourceGB(cb func(res map[string]*model.Resource)){
	rwLock.RLock()
	defer rwLock.RUnlock()
	cb(RESOURCE_GBID_EQ)
}

func ExsitResource(id string) bool {
	return ExsitGbId(id) || ExsitResourceId(id)
}
func ExsitGbId(gbId string) bool {
	rwLock.RLock()
	defer rwLock.RUnlock()
	_, isExsit := RESOURCE_GBID_EQ[gbId]
	return isExsit
}

func ExsitResourceId(resourceId string) bool {
	rwLock.RLock()
	defer rwLock.RUnlock()
	_, isExsit := RESOURCE_ID_EQ[resourceId]
	return isExsit
}

//监听
var configWatchers = make([]func(), 0)
var resourceWatchers = make([]func(), 0)

func WatchConfig(cb func()) {
	configWatchers = append(configWatchers, cb)
	atomic.StoreInt32(&delayRefreshConfig, 1)
}
func WatchResource(cb func()) {
	resourceWatchers = append(resourceWatchers, cb)
	refreshResource()
}
func GetManagePort() string {
	var port string
	if len(os.Args) > 1 && os.Args[1] == "-p" {
		port = os.Args[2]
	} else {
		port = "7777"
	}
	return port
}
