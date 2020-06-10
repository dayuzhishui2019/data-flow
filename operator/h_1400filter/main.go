package h_1400filter

import (
	"dyzs/data-flow/context"
	"dyzs/data-flow/logger"
	"dyzs/data-flow/model/gat1400"
	"dyzs/data-flow/model/gat1400/base"
	"dyzs/data-flow/stream"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

func init() {
	stream.RegistHandler("1400filter", func() stream.Handler {
		return &stream.HandlerWrapper{
			InitFunc:   Init,
			HandleFunc: Handle,
		}
	})
}

func Init(config interface{}) error {
	dataType := context.GetString("dataType")
	logger.LOG_WARN("------------------ 1400filter config ------------------")
	logger.LOG_WARN("1400filter_dataType : " + dataType)
	logger.LOG_WARN("------------------------------------------------------")
	return nil
}

func Handle(data interface{}, next func(interface{}) error) error {
	wraps, ok := data.([]*gat1400.Gat1400Wrap)
	if !ok {
		return errors.New(fmt.Sprintf("Handle [1400tokafkamsg] 数据格式错误，need []*daghub.Gat1400Wrap , get %T", reflect.TypeOf(data)))
	}
	if len(wraps) == 0 {
		return nil
	}
	//data type
	var isFilterByDataType bool
	filterMap := make(map[string]bool)
	dataType := context.GetString("dataType")
	if dataType != "" {
		isFilterByDataType = true
		for _, d := range strings.Split(dataType, ",") {
			filterMap[d] = true
		}
	}
	//resource id
	var isFilterByResource bool
	task, err := context.GetTask()
	if err == nil && !task.AllResource {
		isFilterByResource = true
	}
	if !isFilterByDataType && !isFilterByResource {
		return next(wraps)
	}
	filtedWraps := make([]*gat1400.Gat1400Wrap, 0)
	var isFilterData bool
	for _, wrap := range wraps {
		//filter by dataType
		if isFilterByDataType && !filterMap[wrap.DataType] {
			logger.LOG_INFO("数据类型过滤：", dataType, " -filter- ", wrap.DataType)
			isFilterData = true
			continue
		}
		//filter by resourceId
		if isFilterByResource {
			wrap = wrap.FilterByDeviceID(func(deviceID string) bool {
				flag := context.ExsitResource(deviceID)
				if !flag {
					logger.LOG_INFO("数据所属资源过滤：", " -filter- ", deviceID)
				}
				return flag
			})
			if wrap == nil {
				isFilterData = true
				continue
			}
		}
		filtedWraps = append(filtedWraps, wrap)
	}
	if isFilterData {
		return errors.New(base.DEVICEID_IS_NOT_EXIST)
	}
	if len(filtedWraps) > 0 {
		return next(filtedWraps)
	}
	return nil
}
