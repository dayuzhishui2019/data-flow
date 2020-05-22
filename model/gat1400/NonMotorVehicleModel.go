package gat1400

import (
	"dyzs/data-flow/model/gat1400/base"
	protobuf "dyzs/data-flow/model/proto/proto_model"
	"dyzs/data-flow/util/times"
)

type NonMotorVehicleModel struct {
	NonMotorVehicleListObject *NonMotorVehicleListObject `json:"NonMotorVehicleListObject"`

	proxyType         string `json:"proxyType"`         //数据格式标识，0:GAT1400 格式，1:私有格式(扩展GAT1400)
	proxyManufacturer string `json:"proxyManufacturer"` //厂商编码
	resJson           string `json:"resJson"`           //扩展字段
}

type NonMotorVehicleListObject struct {
	NonMotorVehicleObject []*NonMotorVehicleObject `json:"NonMotorVehicleObject"`
}

type NonMotorVehicleObject struct {
	NonMotorVehicleID string                      `json:"NonMotorVehicleID"` //车辆标识 -备注：辆全局唯一标识 R -必选
	TollgateID        string                      `json:"TollgateID"`        //关联卡口编码 O -可选
	VehicleType       string                      `json:"VehicleType"`       //车辆款型
	FaceID            string                      `json:"FaceID"`            //人脸标识
	PersonID          string                      `json:"PersonID"`          //人员标识
	VehicleClass      string                      `json:"VehicleClass"`      //车辆类型
	LaneNo            int                         `json:"LaneNo"`            //车道号
	Direction         string                      `json:"Direction"`         //行驶方向
	BreakRuleMode     string                      `json:"BreakRuleMode"`     //违章类型
	PersonInfoList    PersonInfoOfVehicleTypeList `json:"PersonInfoList"`    //人体属性信息
	CyclingType       int                         `json:"CyclingType"`       //骑车类型
	IsModified        string                        `json:"IsModified"`        //改装标志

	InfoKind            string             `json:"InfoKind"`            //信息分类 -备注：人工采集还是自动采集 R -
	SourceID            string             `json:"SourceID"`            //来源标识 -备注：来源图像标识 R -必选
	DeviceID            string             `json:"DeviceID"`            //设备编码 R/O -自动采集时必选
	LeftTopX            int                `json:"LeftTopX"`            //左上角X坐标 车的轮廓外界矩形在画面中的位置，记录坐标 R/O 自动采集时必选
	LeftTopY            int                `json:"LeftTopY"`            //左上角Y坐标 车的轮廓外界矩形在画面中的位置，记录坐标 R/O 自动采集时必选
	RightBtmX           int                `json:"RightBtmX"`           //右上角X坐标 车的轮廓外界矩形在画面中的位置，记录坐标 R/O 自动采集时必选
	RightBtmY           int                `json:"RightBtmY"`           //右上角Y坐标 车的轮廓外界矩形在画面中的位置，记录坐标 R/O 自动采集时必选;
	MarkTime            string             `json:"MarkTime"`            //位置标记时间 O 人工采集时有效;
	AppearTime          string             `json:"AppearTime"`          //车辆出现时间 O 人工采集时有效
	DisAppearTime       string             `json:"DisAppearTime"`       //车辆消失时间 O 人工采集时有效;
	HasPlate            string             `json:"HasPlate"`            //有无车牌号;
	PlateClass          string             `json:"PlateClass"`          //号牌种类;
	PlateColor          string             `json:"PlateColor"`          //车牌颜色;
	PlateNo             string             `json:"PlateNo"`             //车牌号;
	PlateNoAttach       string             `json:"PlateNoAttach"`       //挂车牌号;
	PlateDescribe       string             `json:"PlateDescribe"`       //车牌描述;
	IsDecked            string               `json:"IsDecked"`            //是否套牌;
	IsAltered           string               `json:"IsAltered"`           //是否涂改;
	IsCovered           string               `json:"IsCovered"`           //是否遮挡;
	Speed               float64            `json:"Speed"`               //行驶速度;
	DrivingStatusCode   string             `json:"DrivingStatusCode"`   //行驶状态代码;
	UsingPropertiesCode int                `json:"UsingPropertiesCode"` //车辆使用性质代码;
	VehicleBrand        string             `json:"VehicleBrand"`        //车辆品牌;
	VehicleModel        string             `json:"VehicleModel"`        //车辆型号;
	VehicleLength       int                `json:"VehicleLength"`       //车辆长度;
	VehicleWidth        int                `json:"VehicleWidth"`        //车辆宽度;
	VehicleHeight       int                `json:"VehicleHeight"`       //车辆高度;
	VehicleColor        string             `json:"VehicleColor"`        //车身颜色 R 必选;
	VehicleHood         string             `json:"VehicleHood"`         //车前盖;
	VehicleTrunk        string             `json:"VehicleTrunk"`        //车后盖;
	VehicleWheel        string             `json:"VehicleWheel"`        //车轮;
	WheelPrintedPattern string             `json:"WheelPrintedPattern"` //车轮印花纹;
	VehicleWindow       string             `json:"VehicleWindow"`       //车窗;
	VehicleRoof         string             `json:"VehicleRoof"`         //车顶;
	VehicleDoor         string             `json:"VehicleDoor"`         //车门;
	SideOfVehicle       string             `json:"SideOfVehicle"`       //车侧;
	CarOfVehicle        string             `json:"CarOfVehicle"`        //车厢;
	RearviewMirror      string             `json:"RearviewMirror"`      //后视镜;
	VehicleChassis      string             `json:"VehicleChassis"`      //底盘;
	VehicleShielding    string             `json:"VehicleShielding"`    //遮挡;
	FilmColor           string             `json:"FilmColor"`           //贴膜颜色;
	StorageURL          string             `json:"StorageURL"`          //大图（场景图）路径;
	NationalityCode     string             `json:"NationalityCode"`     //NationalityCode;
	TabID               string             `json:"TabID"`               //TabID;
	RelatedType         string             `json:"RelatedType"`         //关联关系类型【海康提供的标准】;
	Longitude           float64            `json:"Longitude"`           //设备经度【固定点位设备可选填，移动设备必填】;
	Latitude            float64            `json:"Latitude"`            //设备纬度【固定点位设备可选填，移动设备必填】
	resJson             string             `json:"resJson"`             //预留扩展字段
	SubImageList        *base.SubImageList `json:"SubImageList"`        //图像列表;
	FeatureList         *base.FeatureList  `json:"FeatureList"`         //FeatureList;
	RelatedList         *base.RelatedList  `json:"RelatedList"`         //关联关系实体;
}

type PersonInfoOfVehicleTypeList struct {
	PersonInfoList []*PersonInfoOfVehicleType `json:"PersonInfoList"`
}

type PersonInfoOfVehicleType struct {
	IDNumber            string `json:"IDNumber"`
	Name                string `json:"Name"`
	UsedName            string `json:"UsedName"`
	CyclingPersonNumber int    `json:"CyclingPersonNumber"`
	PersonDirection     int    `json:"PersonDirection"`
	Things              int    `json:"Things"`
	HeightUpLimit       int    `json:"HeightUpLimit"`
	HeightLowerLimit    int    `json:"HeightLowerLimit"`
	TargetSize          int    `json:"TargetSize"`
	Direction           int    `json:"Direction"`
	Speed               int    `json:"Speed"`
	Backup              int    `json:"Backup"`
	IsBackup            int    `json:"IsBackup"`
	IsGlass             int    `json:"IsGlass"`
	IsCap               int    `json:"IsCap"`
}

func (item *NonMotorVehicleObject) GetResourceID() string {
	return item.DeviceID
}

func (item *NonMotorVehicleObject) GetDigest() *protobuf.DigestRecord {
	shotTime := ""
	if item.SubImageList != nil && len(item.SubImageList.SubImageInfoObject) > 0 {
		shotTime = item.SubImageList.SubImageInfoObject[0].ShotTime
	}
	return &protobuf.DigestRecord{
		DataCategory: "GAT1400",
		DataType:     GAT1400_NONMOTOR,
		ResourceId:   item.DeviceID,
		EventTime:    times.Str2TimeF(shotTime,GAT1400_TIME_FORMATTER).UnixNano() / 1e6,
	}
}
