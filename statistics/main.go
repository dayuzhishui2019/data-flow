package main

import (
	"fmt"
	"github.com/json-iterator/go/extra"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"dyzs/data-flow/logger"
	_ "dyzs/data-flow/operator"
	"dyzs/data-flow/stream"
	"dyzs/data-flow/util"
)

func init() {
	configPath := util.GetAppPath() + "config.yml"
	logger.LOG_INFO("configPath:", configPath)
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		logger.LOG_WARN("Fail to read config file :", err)
	}
}

func main() {

	logger.Init()

	//json模糊匹配
	extra.RegisterFuzzyDecoders()

	myStream, err := stream.Build([]string{"kafkaconsumer", "datatowebsocket", "1400digest"})
	if err != nil {
		panic(fmt.Sprint("流程初始化失败：", err))
		return
	}
	err = myStream.Init()
	if err != nil {
		panic(fmt.Sprint("流程初始化失败：", err))
		return
	}
	myStream.Run()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	select {
	case <-c:
		break
	}
}
