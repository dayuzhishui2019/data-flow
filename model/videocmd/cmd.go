package videocmd

import jsoniter "github.com/json-iterator/go"

const (
	CMD_OPEN_STREAM   = "OpenStream"
	CMD_CLOSE_STREAM  = "CloseStream"
	CMD_PTZ           = "DeviceControlPTZCmd"
	CMD_HISTORY_VIDEO = "QueryRecord"
)

type Cmd struct {
	Cmd   string              `json:"Cmd"`
	Param jsoniter.RawMessage `json:"Param"`
}

//开流
type CmdOpenStream struct {
	FromID       string `json:"FromId"` //转发组件的ID，异步的开流响应消息的Target字段使用。
	DeviceId     string `json:"DeviceId"`
	Channel      string `json:"Channel"`
	ForwardProto string `json:"ForwardProto"`
	ForwardIp    string `json:"Ip"`
	ForwardPort  int    `json:"Port"`
	IsTcp        bool   `json:"IsTcp"`
	IsActive     bool   `json:"IsActive"`
	PublishUrl   string `json:"PublishUrl"`
}

//关流
type CmdCloseStream struct {
	DeviceId  string `json:"DeviceId"`
	Channel   string `json:"Channel"`
	SessionId int    `json:"SessionId"`
}

//云台控制
type CmdPTZ struct {
	DeviceId string `json:"DeviceId"`
	Channel  string `json:"Channel"`
	//Ptz      string `json:"Ptz"`
	Zoom      int `json:"Zoom"`      //0-stop 1-small 2-big
	RightLeft int `json:"RightLeft"` //0-stop 1-right 2-left
	DownUp    int `json:"DownUp"`    //0-stop 1-down 2-up
	Speed     int `json:"Speed"`     //0-255
}

func (ptz *CmdPTZ) Direction() string {
	if ptz.Zoom == 0 && ptz.RightLeft == 0 && ptz.DownUp == 0 {
		return "STOP"
	} else if ptz.RightLeft == 2 && ptz.DownUp == 0 {
		return "LEFT"
	} else if ptz.RightLeft == 1 && ptz.DownUp == 0 {
		return "RIGHT"
	} else if ptz.RightLeft == 0 && ptz.DownUp == 2 {
		return "UP"
	} else if ptz.RightLeft == 0 && ptz.DownUp == 1 {
		return "DOWN"
	} else if ptz.RightLeft == 2 && ptz.DownUp == 2 {
		return "LEFTUP"
	} else if ptz.RightLeft == 1 && ptz.DownUp == 2 {
		return "RIGHTUP"
	} else if ptz.RightLeft == 2 && ptz.DownUp == 1 {
		return "LEFTDOWN"
	} else if ptz.RightLeft == 1 && ptz.DownUp == 1 {
		return "RIGHTDOWN"
	} else if ptz.Zoom == 1 {
		return "ZOOMIN"
	} else if ptz.Zoom == 2 {
		return "ZOOMOUT"
	}
	return "STOP"
}
func (ptz *CmdPTZ) GetSpeed() float64 {
	return float64(ptz.Speed) / 255.0
}

//录像查询
type CmdHistoryVideo struct {
	DeviceId  string `json:"DeviceId"`
	Channel   string `json:"Channel"`
	StartTime string `json:"startTime"` //"2020-07-22 22:22:22"
	EndTime   string `json:"endTime"`   //"2020-07-22 23:22:22"
}

type ParamReceiveStream struct {
	RecvProto string `json:"RecvProto"`
	Istcp     bool   `json:"Istcp"`
	// gb28181 国标必填参数，非国标可不给该字段
	IsActive bool `json:"IsActive"`
	// rtsp  rtsp必填参数，非rtsp可不给该字段
	RtspUrl string `json:"RtspUrl"`
}

type ParamMedieOpenResult struct {
	SessionId int `json:"SessionId"`
	RecvPort  int `json:"RecvPort"` //本机收流端口
}

type ParamDeviceChannel struct {
	FromID       string `json:"FromId"`   // 该字段标识来源设备ID，即下级国标编号
	DeviceID     string `json:"DeviceId"` // 该字段为通道ID
	Name         string `json:"Name"`
	Manufacturer string `json:"Manufacturer"`
	Model        string `json:"Model"`
	Owner        string `json:"Owner"`
	CivilCode    string `json:"CivilCode"`
	Address      string `json:"Address"`
	Parental     int    `json:"Parental"`
	ParentID     string `json:"ParentId"`
	SafetyWay    int    `json:"SafetyWay"`
	RegisterWay  string `json:"RegisterWay"`
	Secrecy      int    `json:"Secrecy"`

	Status    string  `json:"Status"`
	Longitude float64 `json:"Longitude"`
	Latitude  float64 `json:"Latitude"`
	PTZType   int     `json:"PtzType"`
}

//媒体开流
func NewMediaOpenStream(deviceId, channel string, receiveParam *ParamReceiveStream, forwardParam *CmdOpenStream) interface{} {
	return map[string]interface{}{
		"Cmd": "OpenStream",
		"Param": map[string]interface{}{
			"DeviceId":     deviceId,
			"Channel":      channel,
			"RecvParam":    receiveParam,
			"ForwardParam": forwardParam,
		},
	}
}
