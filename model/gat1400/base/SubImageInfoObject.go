package base

type SubImageInfo struct {
	ImageID string `json:"ImageID"`

	InfoKind string `json:"InfoKind"`

	ImageSource string `json:"ImageSource"`

	SourceVideoID string `json:"SourceVideoID"`

	OriginImageID string `json:"OriginImageID"`

	EventSort int32 `json:"EventSort"`

	DeviceID string `json:"DeviceID"`

	StoragePath string `json:"StoragePath"`

	FileHash string `json:"FileHash"`

	FileFormat string `json:"FileFormat"`

	ShotTime string `json:"ShotTime"`

	Title string `json:"Title"`

	TitleNote string `json:"TitleNote"`

	SpecialIName string `json:"SpecialIName"`

	Keyword string `json:"Keyword"`

	ContentDescription string `json:"ContentDescription"`

	SubjectCharacter string `json:"SubjectCharacter"`

	ShotPlaceCode string `json:"ShotPlaceCode"`

	ShotPlaceFullAdress string `json:"ShotPlaceFullAdress"`

	ShotPlaceLongitude string `json:"ShotPlaceLongitude"`

	ShotPlaceLatitude string `json:"ShotPlaceLatitude"`

	HorizontalShotDirection string `json:"HorizontalShotDirection"`

	VerticalShotDirection string `json:"VerticalShotDirection"`

	SecurityLevel string `json:"SecurityLevel"`

	Width int32 `json:"Width"`

	Height int32 `json:"Height"`

	CameraManufacturer string `json:"CameraManufacturer"`

	CameraVersion string `json:"CameraVersion"`

	ApertureValue int `json:"ApertureValue"`

	ISOSensitivity int `json:"ISOSensitivity"`

	FocalLength int `json:"FocalLength"`

	QualityGrade string `json:"QualityGrade"`

	CollectorName string `json:"CollectorName"`

	CollectorOrg string `json:"CollectorOrg"`

	CollectorIDType string `json:"CollectorIDType"`

	CollectorID string `json:"CollectorID"`

	EntryClerk string `json:"EntryClerk"`

	EntryClerkOrg string `json:"EntryClerkOrg"`

	EntryClerkIDType string `json:"EntryClerkIDType"`

	EntryClerkID string `json:"EntryClerkID"`

	EntryTime string `json:"EntryTime"`

	ImageProcFlag int `json:"ImageProcFlag"`

	FileSize int `json:"FileSize"`

	Data string `json:"Data"`

	Type string `json:"Type"`
}

type SubImageList struct {
	SubImageInfoObject []*SubImageInfo `json:"SubImageInfoObject"`
}

func (subImageList *SubImageList) BuildSubImageList(subImageInfo SubImageInfo) {
	subImageList.SubImageInfoObject = append(subImageList.SubImageInfoObject, &subImageInfo)
}
