package gat1400

import (
	"dyzs/data-flow/model/gat1400/base"
	protobuf "dyzs/data-flow/model/proto/proto_model"
	"dyzs/data-flow/util/times"
)

type FaceModel struct {
	FaceListObject *FaceListObject `json:"FaceListObject"`

	proxyType         string `json:"proxyType"`         //数据格式标识，0:GAT1400 格式，1:私有格式(扩展GAT1400)
	proxyManufacturer string `json:"proxyManufacturer"` //厂商编码
	resJson           string `json:"resJson"`           //扩展字段
}

type FaceListObject struct {
	FaceObject []*FaceObject `json:"FaceObject"`
}

type FaceObject struct {
	FaceID            string `json:"FaceID"`            //人脸标识
	FaceAppearTime    string `json:"FaceAppearTime"`    //人脸出现时间
	FaceDisAppearTime string `json:"FaceDisAppearTime"` //人脸消失时间
	Attitude          string `json:"Attitude"`          //姿态分布
	Similaritydegree  int    `json:"Similaritydegree"`  //相似度
	EyebrowStyle      string `json:"EyebrowStyle"`      //眉形
	NoseStyle         string `json:"NoseStyle"`         //鼻型
	MustacheStyle     string `json:"MustacheStyle"`     //胡型
	LipStyle          string `json:"LipStyle"`          //嘴唇
	WrinklePouch      string `json:"WrinklePouch"`      //皱纹眼袋
	AcneStain         string `json:"AcneStain"`         //痤疮色斑
	FreckleBirthmark  string `json:"FreckleBirthmark"`  //黑痣标记
	ScarDimple        string `json:"ScarDimple"`        //疤痕酒窝
	OtherFeature      string `json:"OtherFeature"`      //其他特征
	PersonID          string `json:"PersonID"`          //人员标识
	MotorVehicleID    string `json:"MotorVehicleID"`    //机动车标识
	NonMotorVehicleID string `json:"NonMotorVehicleID"` //非机动车标识
	ShotTime          string `json:"ShotTime"`          //拍摄时间
	Emotion           int32  `json:"Emotion"`           //人脸表情
	IsGlass           int32  `json:"IsGlass"`           //是否戴眼镜
	IsRespirator      int32  `json:"IsRespirator"`      //是否戴口罩
	IsCap             int32  `json:"IsCap"`             //是否戴帽子

	InfoKind                          string  `json:"InfoKind"`                          //信息分类 -备注：人工采集还是自动采集 R -
	SourceID                          string  `json:"SourceID"`                          //来源标识 -备注：来源图像标识 R -必选
	DeviceID                          string  `json:"DeviceID"`                          //设备编码
	LeftTopX                          int32   `json:"LeftTopX"`                          //左上角X坐标
	LeftTopY                          int32   `json:"LeftTopY"`                          //左上角Y坐标
	RightBtmX                         int32   `json:"RightBtmX"`                         //右下角X坐标
	RightBtmY                         int32   `json:"RightBtmY"`                         //右下角Y坐标
	LocationMarkTime                  string  `json:"LocationMarkTime"`                  //位置标记时间
	IDType                            string  `json:"IDType"`                            //证件种类
	IDNumber                          string  `json:"IDNumber"`                          //证件号码
	Name                              string  `json:"Name"`                              //姓名
	UsedName                          string  `json:"UsedName"`                          // 曾用名
	Alias                             string  `json:"Alias"`                             //绰号
	GenderCode                        string  `json:"GenderCode"`                        //性别代码
	AgeUpLimit                        int32   `json:"AgeUpLimit"`                        //年龄上限
	AgeLowerLimit                     int32   `json:"AgeLowerLimit"`                     //年龄下限
	EthicCode                         string  `json:"EthicCode"`                         //民族代码
	NationalityCode                   string  `json:"NationalityCode"`                   //国籍代码
	NativeCityCode                    string  `json:"NativeCityCode"`                    //籍贯省市县代码
	ResidenceAdminDivision            string  `json:"ResidenceAdminDivision"`            //居住地行政区划
	ChineseAccentCode                 string  `json:"ChineseAccentCode"`                 //汉语口音代码
	PersonOrg                         string  `json:"PersonOrg"`                         //单位名称
	JobCategory                       string  `json:"JobCategory"`                       //职业类别代码
	AccompanyNumber                   int32   `json:"AccompanyNumber"`                   //同行人数
	SkinColor                         string  `json:"SkinColor"`                         //肤色
	HairStyle                         string  `json:"HairStyle"`                         //发型
	HairType                          string  `json:"HairType"`                          //
	HairColor                         string  `json:"HairColor"`                         //发色
	FaceStyle                         string  `json:"FaceStyle"`                         //脸型
	FacialFeature                     string  `json:"FacialFeature"`                     //脸部特征
	PhysicalFeature                   string  `json:"PhysicalFeature"`                   //体貌特征
	RespiratorColor                   string  `json:"RespiratorColor"`                   //口罩颜色
	CapStyle                          string  `json:"CapStyle"`                          //帽子款式
	CapColor                          string  `json:"CapColor"`                          //帽子颜色
	GlassStyle                        string  `json:"GlassStyle"`                        //眼镜款式
	GlassColor                        string  `json:"GlassColor"`                        //眼镜颜色
	IsDriver                          int32   `json:"IsDriver"`                          //是否驾驶员
	IsForeigner                       int32   `json:"IsForeigner"`                       //是否涉外人员
	PassportType                      string  `json:"PassportType"`                      //护照证件种类
	ImmigrantTypeCode                 string  `json:"ImmigrantTypeCode"`                 //出入境人员类别编码
	IsSuspectedTerrorist              int32   `json:"IsSuspectedTerrorist"`              //是否涉恐人员
	SuspectedTerroristNumber          string  `json:"SuspectedTerroristNumber"`          //涉恐人员编号
	IsCriminalInvolved                int32   `json:"IsCriminalInvolved"`                //是否涉案人员
	IsSuspiciousPerson                int32   `json:"IsSuspiciousPerson"`                //是否可疑人
	CriminalInvolvedSpecilisationCode string  `json:"CriminalInvolvedSpecilisationCode"` //涉案人员专长代码
	BodySpeciallMark                  string  `json:"BodySpeciallMark"`                  //体表特殊标记
	CrimeMethod                       string  `json:"CrimeMethod"`                       //作案手段
	CrimeCharacterCode                string  `json:"CrimeCharacterCode"`                //作案特点代码
	EscapedCriminalNumber             string  `json:"EscapedCriminalNumber"`             //在逃人员编号
	IsDetainees                       int32   `json:"IsDetainees"`                       //是否在押人员
	DetentionHouseCode                string  `json:"DetentionHouseCode"`                //看守所编码
	DetaineesIdentity                 string  `json:"DetaineesIdentity"`                 //在押人员身份
	DetaineesSpecialIdentity          string  `json:"DetaineesSpecialIdentity"`          //在押人员特殊身份
	MemberTypeCode                    string  `json:"MemberTypeCode"`                    //成员类型代码
	IsVictim                          int32   `json:"IsVictim"`                          //是否被害人
	VictimType                        string  `json:"VictimType"`                        //被害人种类
	InjuredDegree                     string  `json:"InjuredDegree"`                     //受伤害程度
	CorpseConditionCode               string  `json:"CorpseConditionCode"`               //尸体状况代码
	IsSuspiciousFace                  int32   `json:"IsSuspiciousFace"`                  //是否可疑人
	StorageURL                        string  `json:"StorageURL"`                        //大图（场景图）路径
	TabID                             string  `json:"TabID"`                             //归属分类标签标识
	ResJson                           string  `json:"resJson"`                           //预留扩展字段
	RelatedType                       string  `json:"RelatedType"`                       //关联关系类型【海康提供的标准】 01-人员 02-机动车 03-非机动车 04-物品 05-场景 06-人脸 07-视频图像标签 99-其他
	Longitude                         float64 `json:"Longitude"`                         //设备经度【固定点位设备可选填，移动设备必填】
	Latitude                          float64 `json:"Latitude"`                          //设备纬度【固定点位设备可选填，移动设备必填】

	SubImageList *base.SubImageList `json:"SubImageList"` //图像列表
	FeatureList  *base.FeatureList  `json:"FeatureList"`  //特征值列表
	RelatedList  *base.RelatedList  `json:"RelatedList"`  //关联关系实体

}

func (face *FaceObject) SetFaceID(faceId string) {
	face.FaceID = faceId
}

func (face *FaceObject) getFaceId() string {
	if face != nil {
		return face.FaceID
	}
	return ""
}

func (face *FaceObject) GetResourceID() string {
	return face.DeviceID
}

func (item *FaceObject) GetDigest() *protobuf.DigestRecord {
	shotTime := item.ShotTime
	if item.SubImageList != nil && len(item.SubImageList.SubImageInfoObject) > 0 {
		shotTime = item.SubImageList.SubImageInfoObject[0].ShotTime
	}
	return &protobuf.DigestRecord{
		DataCategory: "GAT1400",
		DataType:     GAT1400_FACE,
		ResourceId:   item.DeviceID,
		EventTime:    times.Str2TimeF(shotTime,GAT1400_TIME_FORMATTER).UnixNano() / 1e6,
	}
}
