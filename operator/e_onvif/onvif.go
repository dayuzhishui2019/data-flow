package e_onvif

import (
	"dyzs/data-flow/model"
	"encoding/xml"
	"errors"
	"github.com/yakovlevdmv/goonvif"
	"github.com/yakovlevdmv/goonvif/Media"
	"github.com/yakovlevdmv/goonvif/xsd/onvif"
	"io/ioutil"
	"net/http"
	"strings"
)

//获取通道列表
func LoadResourceChannels(resource *model.Resource) (channels []onvif.Profile, err error) {
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
