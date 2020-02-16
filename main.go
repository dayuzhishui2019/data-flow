package main

import (
	"github.com/json-iterator/go/extra"
	"os"
	"os/signal"
	"dyzs/data-flow/context"
	"dyzs/data-flow/logger"
	"dyzs/data-flow/model"
	_ "dyzs/data-flow/operator"
	"dyzs/data-flow/proxy"
	"dyzs/data-flow/stream"
)

var TASK_FLOW = map[string][]string{
	"1400server": []string{"1400server", "1400filter", "uploadimage", "1400tokafkamsg", "kafkaproducer"},
	"1400client": []string{"kafkaconsumer", "kafkamsgto1400", "1400filter", "downloadimage", "1400client", "kafkaproducer"},
	"statistics": []string{"kafkaconsumer", "1400digesttoredis"},
	"1400servertest": []string{"1400server", "1400filter", "uploadimage", "1400tokafkamsg",},
}

func main() {

	//initStatistics()

	context.Set("$manage_port", os.Getenv("MANAGE_PORT"))
	context.Set("$host", os.Getenv("HOST"))
	context.Set("$logLevel", os.Getenv("LOG_LEVEL"))

	logger.Init()

	//json模糊匹配
	extra.RegisterFuzzyDecoders()

	//启动组件管理服务代理
	proxy.StartManagerProxy(context.GetString("$manage_port"))

	var currentStream *stream.Stream
	context.WatchConfig(func() {
		if currentStream != nil {
			currentStream.Close()
		}

		task, err := context.GetTask()
		if err != nil {
			logger.LOG_WARN("未定义任务")
			return
		}
		flow, ok := TASK_FLOW[task.AccessType]
		if !ok {
			logger.LOG_ERROR("未定义的taskType:", task.AccessType)
			return
		}

		myStream, err := stream.Build(flow)
		if err != nil {
			logger.LOG_WARN("流程初始化失败：", err)
			return
		}
		err = myStream.Init()
		if err != nil {
			logger.LOG_WARN("流程初始化失败：", err)
			myStream.Close()
			return
		}
		currentStream = myStream
		myStream.Run()
	})

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	select {
	case <-c:
		break
	}
}

func initStatistics() {
	context.Set("$task", &model.Task{
		AccessType: "statistics",
	})
	context.Set("kafkaconsumer_bootstrap", os.Getenv("KAFKA_HOST"))
	context.Set("kafkaconsumer_bootstrap", os.Getenv("KAFKA_HOST"))
	context.Set("kafkaconsumer_topics", os.Getenv("KAFKA_TOPICS"))
	context.Set("kafkaconsumer_groupId", os.Getenv("KAFKA_GROUP_ID"))
	context.Set("kafkaconsumer_fromEarliestOffset", true)
	context.Set("kafkaconsumer_batchSize", os.Getenv("BATCH_SIZE"))
	context.Set("kafkaconsumer_batchDelay", os.Getenv("BATCH_DELAY"))
	context.Set("1400digesttoredis_redisAddr", os.Getenv("REDIS_ADDR"))
}
