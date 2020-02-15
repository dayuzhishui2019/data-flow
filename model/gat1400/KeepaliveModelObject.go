package gat1400

import "strings"

type KeepaliveModel struct {
	KeepaliveObject *KeepaliveObject `json:"KeepaliveObject"`
}

type KeepaliveObject struct {
	DeviceID string `json:"DeviceID"`
}

func BuildGat1400KeepaliveObject(deviceID string) *KeepaliveModel {
	return &KeepaliveModel{
		KeepaliveObject: &KeepaliveObject{
			DeviceID: deviceID,
		},
	}
}

func (km *KeepaliveModel) GetViewID() string {
	if km == nil || km.KeepaliveObject == nil {
		return ""
	}
	return strings.Trim(km.KeepaliveObject.DeviceID, " ")
}
