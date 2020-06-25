package gat1400

import (
	"dyzs/data-flow/model/gat1400/base"
	protobuf "dyzs/data-flow/model/proto/proto_model"
	"dyzs/data-flow/util/times"
)

type MotorVehicleModel struct {
	MotorVehicleListObject *MotorVehicleListObject `json:"MotorVehicleListObject"`

	proxyType         string `json:"proxyType"`         //数据格式标识，0:GAT1400 格式，1:私有格式(扩展GAT1400)
	proxyManufacturer string `json:"proxyManufacturer"` //厂商编码
	resJson           string `json:"resJson"`           //扩展字段
}

type MotorVehicleListObject struct {
	MotorVehicleObject []*MotorVehicleObject `json:"MotorVehicleObject"`
}

type MotorVehicleObject struct {
	MotorVehicleID       string `json:"MotorVehicleID"`       //车辆标识 -备注：辆全局唯一标识 R -必选
	TollgateID           string `json:"TollgateID"`           //关联卡口编码 O -可选
	StorageURL1          string `json:"StorageUrl1"`          //近景照片 -卡口相机所拍照片，自动采集必选，图像访问路径采用URI命名规范 R -必选
	StorageURL2          string `json:"StorageUrl2"`          //车辆照片 O -可选
	StorageURL3          string `json:"StorageUrl3"`          //远景照片 全景相机所拍照片 O 可选
	StorageURL4          string `json:"StorageUrl4"`          //合成图 O 可选
	StorageURL5          string `json:"StorageUrl5"`          //缩略图 O 可选
	LaneNo               string `json:"LaneNo"`               //车道号 O 可选
	Direction            string `json:"Direction"`            //行驶方向
	VehicleClass         string `json:"VehicleClass"`         //车辆类型
	VehicleColorDepth    string `json:"VehicleColorDepth"`    //颜色深浅
	HitMarkInfo          string `json:"HitMarkInfo"`          //撞痕信息
	VehicleBodyDesc      string `json:"VehicleBodyDesc"`      //车身描述
	VehicleFrontItem     string `json:"VehicleFrontItem"`     //车前部物品
	VehicleRearItem      string `json:"VehicleRearItem"`      //车前部物品描述
	DescOfRearItem       string `json:"DescOfRearItem"`       //车后部物品
	NumOfPassenger       int    `json:"NumOfPassenger"`       //车后部物品描述
	PassTime             string `json:"PassTime"`             //车内人数
	NameOfPassedRoad     string `json:"NameOfPassedRoad"`     //经过道路名称
	IsSuspicious         string   `json:"IsSuspicious"`         //是否可疑车
	Sunvisor             int    `json:"Sunvisor"`             //遮阳板状态
	SafetyBelt           int    `json:"SafetyBelt"`           //安全带状态
	Calling              int    `json:"Calling"`              //打电话状态
	PlateReliability     string `json:"PlateReliability"`     //号牌识别可信度
	PlateCharReliability string `json:"PlateCharReliability"` //每位号牌号码可信度
	BrandReliability     string `json:"BrandReliability"`     //品牌标识识别可信度
	DriverFaceID         string `json:"DriverFaceID"`         //主驾人脸标识
	CopilotFaceID        string `json:"CopilotFaceID"`        //副驾人脸标识
	VehicleStyles        string `json:"VehicleStyles"`        //车辆年款
	IsModified           string   `json:"IsModified"`           //改装标志

	DescOfFrontItem string `json:"DescOfFrontItem"`
	ResJSON         string `json:"resJson"`

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

func (item *MotorVehicleObject) GetResourceID() string {
	return item.DeviceID
}

func (item *MotorVehicleObject) GetDigest() *protobuf.DigestRecord {
	shotTime := ""
	if item.SubImageList != nil && len(item.SubImageList.SubImageInfoObject) > 0 {
		shotTime = item.SubImageList.SubImageInfoObject[0].ShotTime
	}
	return &protobuf.DigestRecord{
		DataCategory: "GAT1400",
		DataType:     GAT1400_VEHICLE,
		ResourceId:   item.DeviceID,
		EventTime:    times.Str2TimeF(shotTime,GAT1400_TIME_FORMATTER).UnixNano() / 1e6,
	}
}
