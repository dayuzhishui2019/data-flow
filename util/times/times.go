package times

import "time"

const (
	defultTimeLayout string = "2006-01-02 15:04:05"
)

/**字符串->时间对象*/
func Str2Time(formatTimeStr string) time.Time {
	loc, _ := time.LoadLocation("Local")
	theTime, _ := time.ParseInLocation(defultTimeLayout, formatTimeStr, loc) //使用模板在对应时区转化为time.time类型

	return theTime

}

/**字符串->时间戳*/
func Str2Stamp(formatTimeStr string) int64 {
	timeStruct := Str2Time(formatTimeStr)
	millisecond := timeStruct.UnixNano() / 1e6
	return millisecond
}

/**时间对象->字符串*/
func Time2Str() string {
	t := time.Now()
	temp := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.Local)
	str := temp.Format(defultTimeLayout)
	return str
}

/*时间对象->时间戳*/
func Time2Stamp() int64 {
	t := time.Now()
	millisecond := t.UnixNano() / 1e6
	return millisecond
}

/*时间戳->字符串*/
func Stamp2Str(stamp int64, timeLayout string) string {
	str := time.Unix(stamp/1000, 0).Format(timeLayout)
	return str
}

/*时间戳->时间对象*/
func Stamp2Time(stamp int64) time.Time {
	stampStr := Stamp2Str(stamp, defultTimeLayout)
	timer := Str2Time(stampStr)
	return timer
}

/**时间对象->字符串*/
func Time2StrF(t time.Time, formatStr string) string {
	temp := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.Local)
	str := temp.Format(formatStr)
	return str
}


/**字符串->时间对象*/
func Str2TimeF(timeStr string,formatStr string) time.Time {
	loc, _ := time.LoadLocation("Local")
	theTime, _ := time.ParseInLocation(formatStr, timeStr, loc) //使用模板在对应时区转化为time.time类型
	return theTime
}