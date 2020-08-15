package e_onvif

import (
	"dyzs/data-flow/logger"
	"dyzs/data-flow/model"
	"encoding/xml"
	"errors"
	"github.com/yakovlevdmv/goonvif"
	"github.com/yakovlevdmv/goonvif/Media"
	"github.com/yakovlevdmv/goonvif/PTZ"
	"github.com/yakovlevdmv/goonvif/xsd"
	"github.com/yakovlevdmv/goonvif/xsd/onvif"
	"io/ioutil"
	"net/http"
	"strings"
)

//获取通道列表
func LoadResourceChannels(resource *model.Resource) (channels []onvif.Profile, err error) {
	logger.LOG_INFO("获取设备通道：", resource.MvcIP, ",", resource.MvcPort, ",", resource.MvcUsername, ",", resource.MvcPassword)
	if resource.MvcIP == "" || resource.MvcPort == "" || resource.MvcUsername == "" || resource.MvcPassword == "" {
		return nil, errors.New("设备信息缺失")
	}
	device, err := goonvif.NewDevice(resource.MvcIP + ":" + resource.MvcPort)
	if err != nil {
		return nil, err
	}
	device.Authenticate(resource.MvcUsername, resource.MvcPassword)
	//获取通道列表
	res, err := device.CallMethod(Media.GetProfiles{})
	if err != nil {
		return nil, err
	}
	gp := &Media.GetProfilesResponse{}
	err = decodeSoap(res, gp)
	if err != nil {
		return nil, err
	}
	return gp.Profiles, nil
}

//获取rtsp流地址
func LoadChannelRTSP(resource *model.Resource, channelToken string) (rtsp string, err error) {
	device, err := goonvif.NewDevice(resource.MvcIP + ":" + resource.MvcPort)
	if err != nil {
		return "", err
	}
	device.Authenticate(resource.MvcUsername, resource.MvcPassword)
	res, err := device.CallMethod(Media.GetStreamUri{
		ProfileToken: onvif.ReferenceToken(channelToken),
	})
	if err != nil {
		return
	}
	gp := &Media.GetStreamUriResponse{}
	err = decodeSoap(res, gp)
	if err != nil {
		return "", err
	}
	if gp.MediaUri.Uri == "" {
		return "", errors.New("获取RTSP地址为空")
	}
	return prependUsername(string(gp.MediaUri.Uri), resource.MvcUsername, resource.MvcPassword), nil
}

//控制云台
func ControlPTZ(resource *model.Resource, channelToken string, cmd string, speed float64) (err error) {
	device, err := goonvif.NewDevice(resource.MvcIP + ":" + resource.MvcPort)
	if err != nil {
		return err
	}
	device.Authenticate(resource.MvcUsername, resource.MvcPassword)

	panTilt := onvif.Vector2D{
		Space: xsd.AnyURI("http://www.onvif.org/ver10/tptz/PanTiltSpaces/VelocityGenericSpace"),
	}
	zoom := onvif.Vector1D{
		Space: xsd.AnyURI("http://www.onvif.org/ver10/tptz/ZoomSpaces/VelocityGenericSpace"),
	}
	stop := false
	switch cmd {
	case "LEFT":
		panTilt.X = -speed
		panTilt.Y = 0
	case "RIGHT":
		panTilt.X = speed
		panTilt.Y = 0
	case "UP":
		panTilt.X = 0
		panTilt.Y = speed
	case "DOWN":
		panTilt.X = 0
		panTilt.Y = -speed
	case "LEFTUP":
		panTilt.X = -speed
		panTilt.Y = speed
	case "RIGHTUP":
		panTilt.X = speed
		panTilt.Y = speed
	case "LEFTDOWN":
		panTilt.X = -speed
		panTilt.Y = -speed
	case "RIGHTDOWN":
		panTilt.X = speed
		panTilt.Y = -speed
	case "ZOOMIN":
		zoom.X = speed
	case "ZOOMOUT":
		zoom.X = -speed
	case "STOP":
		stop = true
	default:
	}
	if stop {
		_, err = device.CallMethod(PTZ.Stop{
			ProfileToken: onvif.ReferenceToken(channelToken),
			PanTilt:      true,
			Zoom:         true,
		})
	} else {
		httpres, err := device.CallMethod(PTZ.ContinuousMove{
			ProfileToken: onvif.ReferenceToken(channelToken),
			Velocity: onvif.PTZSpeed{
				PanTilt: panTilt,
				Zoom:    zoom,
			},
			Timeout: "PT00H00M10S",
		})
		data, _ := ioutil.ReadAll(httpres.Body)

		if logger.IsDebug() {
			logger.LOG_DEBUG("onvif云台请求响应：", string(data), err)
		}
	}
	return err
}

func prependUsername(uri, username, password string) string {
	index := strings.Index(uri, "//")
	if index+2 > len(uri) {
		return ""
	}
	return uri[:index+2] + username + ":" + password + "@" + uri[index+2:]
}

type Envelope struct {
	XMLName      xml.Name
	EnvelopeBody EnvelopeBody `xml:"Body"`
}

type EnvelopeBody struct {
	XMLName  xml.Name
	Response []byte `xml:",innerxml"`
}

func decodeSoap(res *http.Response, ptr interface{}) error {
	evp := &Envelope{
		EnvelopeBody: EnvelopeBody{},
	}
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if logger.IsDebug() {
		logger.LOG_DEBUG("onvif请求响应：", string(data))
	}
	err = xml.Unmarshal(data, evp)
	if err != nil {
		return err
	}
	err = xml.Unmarshal(evp.EnvelopeBody.Response, ptr)
	if err != nil {
		return err
	}
	return nil
}
