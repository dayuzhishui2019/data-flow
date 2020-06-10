package h_1400digest

import (
	"dyzs/data-flow/logger"
	"dyzs/data-flow/model/gat1400"
	"dyzs/data-flow/model/kafka"
	protobuf "dyzs/data-flow/model/proto/proto_model"
	"dyzs/data-flow/redis"
	"dyzs/data-flow/stream"
	"dyzs/data-flow/util"
	"errors"
	"fmt"
	"github.com/golang/protobuf/proto"
	"os"
	"path"
	"reflect"
	"time"
)

var c redis.Cache
var cstSh *time.Location //上海
var _REDIS_KEY_STATISTICS = "data-hub-statistics"
var _ACCESS_KEY = "access"
var _TRANSMIT_KEY_PREFIX = "transmit_"

var _GAT1400_DATATYPE = map[string]string{
	gat1400.GAT1400_FACE:     "face",
	gat1400.GAT1400_BODY:     "body",
	gat1400.GAT1400_VEHICLE:  "vehicle",
	gat1400.GAT1400_NONMOTOR: "nomotor",
}

func init() {
	_ = os.Setenv("ZONEINFO", path.Join(".", "zonieinfo.zip"))
	cstSh, _ = time.LoadLocation("Asia/Shanghai") //上海
	stream.RegistHandler("1400digest", func() stream.Handler {
		return &stream.HandlerWrapper{
			InitFunc:   Init,
			HandleFunc: Handle,
			CloseFunc:  Close,
		}
	})
}

func Init(config interface{}) error {
	logger.LOG_WARN("------------------ 1400digest config ------------------")
	logger.LOG_WARN("------------------------------------------------------")
	//c = redis.NewRedisCache(0, context.GetString("1400digesttoredis_redisAddr"), redis.FOREVER)
	return nil
}

func Handle(data interface{}, next func(interface{}) error) error {
	kafkaMsgs, ok := data.([]*kafka.KafkaMessage)
	if !ok {
		return errors.New(fmt.Sprintf("Handle [1400digesttoredis] 数据格式错误，need []*kafka.KafkaMessage , get %T", reflect.TypeOf(data)))
	}
	if len(kafkaMsgs) == 0 {
		return nil
	}
	var digests []*protobuf.DigestRecord

	for _, kafkaMsg := range kafkaMsgs {
		digestBytes := kafkaMsg.Header("digest")
		if len(digestBytes) == 0 {
			continue
		}
		digestList := &protobuf.DigestRecordList{}
		err := proto.Unmarshal(digestBytes, digestList)
		if err != nil {
			logger.LOG_ERROR("proto 转化 DigestRecordList 失败", err)
			continue
		}
		digests = append(digests, digestList.RecordList...)
	}
	if len(digests) <= 0 {
		return nil
	}
	//statisticsToRedis(digests)
	return next(digests)
}

func Close() error {
	return c.Close()
}

type DataTypeStatistics struct {
	Name string             `json:"name"`
	M    map[string][]int64 `json:"m"`
	Time int64              `json:"time"`
}

func NewDataTypeStatistics(name string, time time.Time) *DataTypeStatistics {
	return &DataTypeStatistics{
		Name: name,
		M:    make(map[string][]int64),
		Time: time.UnixNano() / 1e6,
	}
}
func (dts *DataTypeStatistics) Add(dataType string, mtInDay int, count int64) {
	list, ok := dts.M[dataType]
	if !ok {
		list = make([]int64, 60*24)
		dts.M[dataType] = list
	}
	list[mtInDay] += count
}
func (dts *DataTypeStatistics) Merge(newDts *DataTypeStatistics) {
	oldDay := time.Unix(dts.Time/1e3, 0).In(cstSh)
	newDay := time.Unix(newDts.Time/1e3, 0).In(cstSh)
	dts.Time = newDts.Time
	if oldDay.Year() != newDay.Year() || oldDay.Month() != newDay.Month() || oldDay.Day() != newDay.Day() {
		dts.M = newDts.M
		return
	}
	//merge
	for k, newS := range newDts.M {
		oldS, ok := dts.M[k]
		if ok {
			for i, rv := range newS {
				oldS[i] += rv
			}
		} else {
			dts.M[k] = newS
		}
	}
}

func statisticsToRedis(digests []*protobuf.DigestRecord) {
	if len(digests) == 0 {
		return
	}

	today := time.Now().In(cstSh)
	statisticsMap := make(map[string]*DataTypeStatistics)

	var modified bool
	for _, d := range digests {
		t, ok := _GAT1400_DATATYPE[d.DataType]
		if !ok {
			t = d.DataType
		}
		var tm int64
		if d.AccessTime > 0 {
			tm = d.AccessTime
		} else if d.TransmitTime > 0 {
			tm = d.TransmitTime
		}
		mtInDay := calcMinuteInToday(today, tm)
		if mtInDay < 0 {
			continue
		}
		modified = true
		//接入
		if d.AccessTime > 0 {
			access, ok := statisticsMap[_ACCESS_KEY]
			if !ok {
				access = NewDataTypeStatistics(_ACCESS_KEY, today)
				statisticsMap[_ACCESS_KEY] = access
			}
			access.Add(t, mtInDay, 1)
		} else {
			//转出
			if d.TargetPlatformId == "" {
				continue
			}
			dts, ok := statisticsMap[_TRANSMIT_KEY_PREFIX+d.TargetPlatformId]
			if !ok {
				dts = NewDataTypeStatistics(_TRANSMIT_KEY_PREFIX+d.TargetPlatformId, today)
				statisticsMap[_TRANSMIT_KEY_PREFIX+d.TargetPlatformId] = dts
			}
			dts.Add(t, mtInDay, 1)
		}
	}
	if !modified {
		return
	}

	//get from redis
	raws := make(map[string]*DataTypeStatistics)
	util.Retry(func() error {
		err := c.StringGet(_REDIS_KEY_STATISTICS, &raws)
		return err
	}, -1, time.Second*3)
	//merge
	for k, newS := range statisticsMap {
		oldS, ok := raws[k]
		if ok {
			oldS.Merge(newS)
		} else {
			raws[k] = newS
		}
	}
	//save to redis
	util.Retry(func() error {
		return c.StringSet(_REDIS_KEY_STATISTICS, raws)
	}, -1, time.Second*3)
}

func calcMinuteInToday(today time.Time, tms int64) int {
	t := time.Unix(tms/1e3, 0).In(cstSh)
	if t.Year() != today.Year() || t.Month() != today.Month() || t.Day() != today.Day() {
		return -1
	}
	return t.Hour()*60 + t.Minute()
}
