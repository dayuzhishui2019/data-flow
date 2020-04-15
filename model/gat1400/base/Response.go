package base

import (
	"encoding/json"
	"errors"
	jsoniter "github.com/json-iterator/go"
	"dyzs/data-flow/util/times"
	"time"
)

const (
	OK                      = "0"
	OTHER_ERROR             = "1"
	DEVICE_BUSY             = "2"
	DEVICE_ERROR            = "3"
	INVALID_OPERATION       = "4"
	XML_FORMAT_INVALID      = "5"
	XML_CONTENT_INVALID     = "6"
	JSON_FORMAT_INVALID     = "7"
	JSON_CONTENT_INVALID    = "8"
	SYSTEM_REBOOT           = "9"
	VIEWID_IS_NULL          = "10"
	PASSWORD_IS_NULL        = "11"
	RESPONSE_NOT_CORRECT    = "12"
	AUTHORIZATION_IS_NULL   = "13"
	KEEPALIVEOBJECT_IS_NULL = "14"
	DEVICEID_IS_NOT_EXIST   = "15"
	UAERNAEM_NOT_CORRECT    = "16"
	UNAUTHORIZED            = "17"
)

var statusText = map[string]string{
	OK:                      "正常",
	OTHER_ERROR:             "其他未知错误",
	DEVICE_BUSY:             "设备忙",
	DEVICE_ERROR:            "设备错误",
	INVALID_OPERATION:       "无效操作",
	XML_FORMAT_INVALID:      "XML格式无效",
	XML_CONTENT_INVALID:     "XML内容无效",
	JSON_FORMAT_INVALID:     "JSON格式无效",
	JSON_CONTENT_INVALID:    "JSON内容无效",
	SYSTEM_REBOOT:           "系统重启中",
	VIEWID_IS_NULL:          "视图库id为空",
	PASSWORD_IS_NULL:        "密码为空",
	RESPONSE_NOT_CORRECT:    "response 不正确",
	AUTHORIZATION_IS_NULL:   "Authorization 不能为空",
	KEEPALIVEOBJECT_IS_NULL: "keepalive 对象不能为空",
	DEVICEID_IS_NOT_EXIST:   "DeviceID is not exist",
	UAERNAEM_NOT_CORRECT:    "注册时用户名或密码不正确",
	UNAUTHORIZED:            "Authentication failure,Please to register",
}

func ExsitErrCode(code string) bool {
	_, ok := statusText[code]
	return ok
}

type Response struct {
	ResponseStatusListObject *ResponseStatusListObject `json:"ResponseStatusListObject"`
}

type ResponseStatusListObject struct {
	ResponseStatusObject []*ResponseStatusObject `json:"ResponseStatusObject"`
}
type ResponseStatusSingleObj struct {
	ResponseStatusObject *ResponseStatusObject `json:"ResponseStatusObject"`
}
type ResponseStatusObject struct {
	ID           string `json:"Id"`
	LocalTime    string `json:"LocalTime"`
	RequestURL   string `json:"RequestURL"`
	StatusCode   string `json:"StatusCode"`
	StatusString string `json:"StatusString"`
}


func BuildResponseObject(url string, recordId string, code string) *ResponseStatusObject {
	return &ResponseStatusObject{
		ID:           recordId,
		LocalTime:    times.Time2StrF(time.Now(), "20060102150405"),
		RequestURL:   url,
		StatusCode:   code,
		StatusString: statusText[code],
	}
}

func BuildRespnse(objs ...*ResponseStatusObject) *Response {
	return &Response{
		ResponseStatusListObject: &ResponseStatusListObject{
			ResponseStatusObject: objs,
		},
	}
}

func BuildSingleResponse(obj *ResponseStatusObject) *ResponseStatusSingleObj {
	return &ResponseStatusSingleObj{
		ResponseStatusObject: obj,
	}
}


func DecodeBytesToReponse(bytes []byte) (*Response, error) {
	res := &Response{}
	err := jsoniter.Unmarshal(bytes, res)
	if err != nil {
		return nil, err
	}
	// list status
	if res.ResponseStatusListObject != nil && len(res.ResponseStatusListObject.ResponseStatusObject) > 0 {
		return res, nil
	}
	// single status
	var responseStatusSingleObject *ResponseStatusSingleObj
	err = json.Unmarshal(bytes, &responseStatusSingleObject)
	if err != nil {
		return nil, err
	}
	if responseStatusSingleObject != nil && responseStatusSingleObject.ResponseStatusObject != nil {
		return &Response{
			ResponseStatusListObject: &ResponseStatusListObject{
				ResponseStatusObject: []*ResponseStatusObject{responseStatusSingleObject.ResponseStatusObject},
			},
		}, nil
	}
	//status
	var responseStatusObject *ResponseStatusObject
	err = json.Unmarshal(bytes, &responseStatusObject)
	if err != nil {
		return nil, err
	}
	if responseStatusObject.StatusCode != "" {
		return &Response{
			ResponseStatusListObject: &ResponseStatusListObject{
				ResponseStatusObject: []*ResponseStatusObject{responseStatusObject},
			},
		}, nil
	}
	return nil, errors.New("response解析失败：" + string(bytes))
}
