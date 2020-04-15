package gat1400

import (
	"dyzs/data-flow/model/gat1400/base"
	protobuf "dyzs/data-flow/model/proto/proto_model"
	"errors"
	"github.com/golang/protobuf/proto"
	jsoniter "github.com/json-iterator/go"
	"io"
	"time"
)

/**
1 1 案（事）件目录
2 2 单个案（事）件内容
3 3 采集设备目录
4 4 采集设备状态
5 5 采集系统目录
6 6 采集系统状态
7 7 视频卡口目录
8 8 单个卡口记录
9 9 车道目录
10 10 单个车道记录
11 11 自动采集的人员信息
12 12 自动采集的人脸信息
13 13 自动采集的车辆信息
14 14 自动采集的非机动车辆信息
15 15 自动采集的物品信息
16 16 自动采集的文件信息
*/
const (
	GAT1400_FACE     = "12"
	GAT1400_BODY     = "11"
	GAT1400_VEHICLE  = "13"
	GAT1400_NONMOTOR = "14"
)

const GAT1400_TIME_FORMATTER = "20060102150405"

type DigestType int

type ResourceData interface {
	GetResourceID() string
}

const (
	DIGEST_ACCESS    DigestType = 1
	DIGEST_TRANSIMIT DigestType = 2
)

type Gat1400Wrap struct {
	DataType             string
	FaceModel            *FaceModel
	PersonModel          *PersonModel
	MotorVehicleModel    *MotorVehicleModel
	NonMotorVehicleModel *NonMotorVehicleModel
}

func BuildFromJson(dataType string, reader io.Reader) (*Gat1400Wrap, error) {
	var err error
	var wrap = &Gat1400Wrap{
		DataType: dataType,
	}
	switch wrap.DataType {
	case GAT1400_FACE:
		m := &FaceModel{}
		err = jsoniter.NewDecoder(reader).Decode(m)
		if err == nil {
			wrap.FaceModel = m
		}
	case GAT1400_BODY:
		m := &PersonModel{}
		err = jsoniter.NewDecoder(reader).Decode(m)
		if err == nil {
			wrap.PersonModel = m
		}
	case GAT1400_VEHICLE:
		m := &MotorVehicleModel{}
		err = jsoniter.NewDecoder(reader).Decode(m)
		if err == nil {
			wrap.MotorVehicleModel = m
		}
	case GAT1400_NONMOTOR:
		m := &NonMotorVehicleModel{}
		err = jsoniter.NewDecoder(reader).Decode(m)
		if err == nil {
			wrap.NonMotorVehicleModel = m
		}
	}
	return wrap, err
}

func (wrap *Gat1400Wrap) BuildToJson() ([]byte, error) {
	switch wrap.DataType {
	case GAT1400_FACE:
		return jsoniter.Marshal(wrap.FaceModel)
	case GAT1400_BODY:
		return jsoniter.Marshal(wrap.PersonModel)
	case GAT1400_VEHICLE:
		return jsoniter.Marshal(wrap.MotorVehicleModel)
	case GAT1400_NONMOTOR:
		return jsoniter.Marshal(wrap.NonMotorVehicleModel)
	}
	return nil, errors.New("未知的数据类型：" + wrap.DataType)
}

func (wrap *Gat1400Wrap) FlatResourceData() []ResourceData {
	var rds []ResourceData
	switch wrap.DataType {
	case GAT1400_FACE:
		if wrap.FaceModel != nil && wrap.FaceModel.FaceListObject != nil && len(wrap.FaceModel.FaceListObject.FaceObject) > 0 {
			for _, item := range wrap.FaceModel.FaceListObject.FaceObject {
				rds = append(rds, item)
			}
		}
	case GAT1400_BODY:
		if wrap.PersonModel != nil && wrap.PersonModel.PersonListObject != nil && len(wrap.PersonModel.PersonListObject.PersonObject) > 0 {
			for _, item := range wrap.PersonModel.PersonListObject.PersonObject {
				rds = append(rds, item)
			}
		}
	case GAT1400_VEHICLE:
		if wrap.MotorVehicleModel != nil && wrap.MotorVehicleModel.MotorVehicleListObject != nil && len(wrap.MotorVehicleModel.MotorVehicleListObject.MotorVehicleObject) > 0 {
			for _, item := range wrap.MotorVehicleModel.MotorVehicleListObject.MotorVehicleObject {
				rds = append(rds, item)
			}
		}
	case GAT1400_NONMOTOR:
		if wrap.NonMotorVehicleModel != nil && wrap.NonMotorVehicleModel.NonMotorVehicleListObject != nil && len(wrap.NonMotorVehicleModel.NonMotorVehicleListObject.NonMotorVehicleObject) > 0 {
			for _, item := range wrap.NonMotorVehicleModel.NonMotorVehicleListObject.NonMotorVehicleObject {
				rds = append(rds, item)
			}
		}
	}
	return rds
}

func (wrap *Gat1400Wrap) BuildResponse(code string) *base.Response {
	status := make([]*base.ResponseStatusObject, 0)
	switch wrap.DataType {
	case GAT1400_FACE:
		for _, item := range wrap.FaceModel.FaceListObject.FaceObject {
			status = append(status, base.BuildResponseObject(base.URL_FACES, item.FaceID, code))
		}
	case GAT1400_BODY:
		for _, item := range wrap.PersonModel.PersonListObject.PersonObject {
			status = append(status, base.BuildResponseObject(base.URL_FACES, item.PersonID, code))
		}
	case GAT1400_VEHICLE:
		for _, item := range wrap.MotorVehicleModel.MotorVehicleListObject.MotorVehicleObject {
			status = append(status, base.BuildResponseObject(base.URL_FACES, item.MotorVehicleID, code))
		}
	case GAT1400_NONMOTOR:
		for _, item := range wrap.NonMotorVehicleModel.NonMotorVehicleListObject.NonMotorVehicleObject {
			status = append(status, base.BuildResponseObject(base.URL_FACES, item.NonMotorVehicleID, code))
		}
	}
	return base.BuildRespnse(status...)
}

//获取图片对象集合
func (wrap *Gat1400Wrap) GetSubImageInfos() []*base.SubImageInfo {
	imgs := make([]*base.SubImageInfo, 0)
	switch wrap.DataType {
	case GAT1400_FACE:
		if wrap.FaceModel != nil && wrap.FaceModel.FaceListObject != nil && wrap.FaceModel.FaceListObject.FaceObject != nil {
			for _, item := range wrap.FaceModel.FaceListObject.FaceObject {
				if item.SubImageList != nil && item.SubImageList.SubImageInfoObject != nil {
					imgs = append(imgs, item.SubImageList.SubImageInfoObject...)
				}
			}
		}
	case GAT1400_BODY:
		if wrap.PersonModel != nil && wrap.PersonModel.PersonListObject != nil && wrap.PersonModel.PersonListObject.PersonObject != nil {
			for _, item := range wrap.PersonModel.PersonListObject.PersonObject {
				if item.SubImageList != nil && item.SubImageList.SubImageInfoObject != nil {
					imgs = append(imgs, item.SubImageList.SubImageInfoObject...)
				}
			}
		}
	case GAT1400_VEHICLE:
		if wrap.MotorVehicleModel != nil && wrap.MotorVehicleModel.MotorVehicleListObject != nil && wrap.MotorVehicleModel.MotorVehicleListObject.MotorVehicleObject != nil {
			for _, item := range wrap.MotorVehicleModel.MotorVehicleListObject.MotorVehicleObject {
				if item.SubImageList != nil && item.SubImageList.SubImageInfoObject != nil {
					imgs = append(imgs, item.SubImageList.SubImageInfoObject...)
				}
			}
		}
	case GAT1400_NONMOTOR:
		if wrap.NonMotorVehicleModel != nil && wrap.NonMotorVehicleModel.NonMotorVehicleListObject != nil && wrap.NonMotorVehicleModel.NonMotorVehicleListObject.NonMotorVehicleObject != nil {
			for _, item := range wrap.NonMotorVehicleModel.NonMotorVehicleListObject.NonMotorVehicleObject {
				if item.SubImageList != nil && item.SubImageList.SubImageInfoObject != nil {
					imgs = append(imgs, item.SubImageList.SubImageInfoObject...)
				}
			}
		}
	}
	return imgs
}

func (wrap *Gat1400Wrap) FilterByDeviceID(filter func(id string) bool) *Gat1400Wrap {
	newWrap := &Gat1400Wrap{
		DataType: wrap.DataType,
	}
	var empty bool
	switch wrap.DataType {
	case GAT1400_FACE:
		if wrap.FaceModel != nil && wrap.FaceModel.FaceListObject != nil && wrap.FaceModel.FaceListObject.FaceObject != nil {
			newFaces := make([]*FaceObject, 0)
			for _, item := range wrap.FaceModel.FaceListObject.FaceObject {
				if filter(item.DeviceID) {
					newFaces = append(newFaces, item)
				}
			}
			if empty = len(newFaces) == 0; !empty {
				newWrap.FaceModel = &FaceModel{
					FaceListObject: &FaceListObject{FaceObject: newFaces},
				}
			}
		}
	case GAT1400_BODY:
		if wrap.PersonModel != nil && wrap.PersonModel.PersonListObject != nil && wrap.PersonModel.PersonListObject.PersonObject != nil {
			newObjs := make([]*PersonObject, 0)
			for _, item := range wrap.PersonModel.PersonListObject.PersonObject {
				if filter(item.DeviceID) {
					newObjs = append(newObjs, item)
				}
			}
			if empty = len(newObjs) == 0; !empty {
				newWrap.PersonModel = &PersonModel{
					PersonListObject: &PersonListObject{PersonObject: newObjs},
				}
			}
		}
	case GAT1400_VEHICLE:
		if wrap.MotorVehicleModel != nil && wrap.MotorVehicleModel.MotorVehicleListObject != nil && wrap.MotorVehicleModel.MotorVehicleListObject.MotorVehicleObject != nil {
			newObjs := make([]*MotorVehicleObject, 0)
			for _, item := range wrap.MotorVehicleModel.MotorVehicleListObject.MotorVehicleObject {
				if filter(item.DeviceID) {
					newObjs = append(newObjs, item)
				}
			}
			if empty = len(newObjs) == 0; !empty {
				newWrap.MotorVehicleModel = &MotorVehicleModel{
					MotorVehicleListObject: &MotorVehicleListObject{MotorVehicleObject: newObjs},
				}
			}
		}
	case GAT1400_NONMOTOR:
		if wrap.NonMotorVehicleModel != nil && wrap.NonMotorVehicleModel.NonMotorVehicleListObject != nil && wrap.NonMotorVehicleModel.NonMotorVehicleListObject.NonMotorVehicleObject != nil {
			newObjs := make([]*NonMotorVehicleObject, 0)
			for _, item := range wrap.NonMotorVehicleModel.NonMotorVehicleListObject.NonMotorVehicleObject {
				if filter(item.DeviceID) {
					newObjs = append(newObjs, item)
				}
			}
			if empty = len(newObjs) == 0; !empty {
				newWrap.NonMotorVehicleModel = &NonMotorVehicleModel{
					NonMotorVehicleListObject: &NonMotorVehicleListObject{NonMotorVehicleObject: newObjs},
				}
			}
		}
	}
	if empty {
		return nil
	}
	return newWrap
}

func (wrap *Gat1400Wrap) BuildDigest(dt DigestType, accessPlatformId, transmitPlatformId string) ([]byte, error) {
	digests := make([]*protobuf.DigestRecord, 0)
	now := time.Now().UnixNano() / 1e6
	switch wrap.DataType {
	case GAT1400_FACE:
		for _, item := range wrap.FaceModel.FaceListObject.FaceObject {
			digests = append(digests, item.GetDigest())
		}
	case GAT1400_BODY:
		for _, item := range wrap.PersonModel.PersonListObject.PersonObject {
			digests = append(digests, item.GetDigest())
		}
	case GAT1400_VEHICLE:
		for _, item := range wrap.MotorVehicleModel.MotorVehicleListObject.MotorVehicleObject {
			digests = append(digests, item.GetDigest())
		}
	case GAT1400_NONMOTOR:
		for _, item := range wrap.NonMotorVehicleModel.NonMotorVehicleListObject.NonMotorVehicleObject {
			digests = append(digests, item.GetDigest())
		}
	}
	for _, d := range digests {
		if dt == DIGEST_ACCESS {
			d.AccessTime = now
			d.SourcePlatformId = accessPlatformId
		} else {
			d.TransmitTime = now
			d.TargetPlatformId = transmitPlatformId
		}
	}
	return proto.Marshal(&protobuf.DigestRecordList{
		RecordList: digests,
	})
}
