package gat1400

type VideoSliceObject struct {
	VideoSliceListObject struct {
		VideoSlice []struct {
			VideoSliceInfo struct {
				AudioCodedFormat        string  `json:"AudioCodedFormat"`
				AudioFlag               int64   `json:"AudioFlag"`
				BeginTime               string  `json:"BeginTime"`
				CodedFormat             string  `json:"CodedFormat"`
				CollectorID             string  `json:"CollectorID"`
				CollectorIDType         string  `json:"CollectorIDType"`
				CollectorName           string  `json:"CollectorName"`
				CollectorOrg            string  `json:"CollectorOrg"`
				ContentDescription      string  `json:"ContentDescription"`
				DeviceID                string  `json:"DeviceID"`
				EndTime                 string  `json:"EndTime"`
				EntryClerk              string  `json:"EntryClerk"`
				EntryClerkID            string  `json:"EntryClerkID"`
				EntryClerkIDType        string  `json:"EntryClerkIDType"`
				EntryClerkOrg           string  `json:"EntryClerkOrg"`
				EntryTime               string  `json:"EntryTime"`
				EventSort               int64   `json:"EventSort"`
				FileFormat              string  `json:"FileFormat"`
				FileHash                string  `json:"FileHash"`
				FileSize                int64   `json:"FileSize"`
				Height                  int64   `json:"Height"`
				HorizontalShotDirection string  `json:"HorizontalShotDirection"`
				InfoKind                int64   `json:"InfoKind"`
				IsAbstractVideo         string  `json:"IsAbstractVideo"`
				Keyword                 string  `json:"Keyword"`
				MainCharacter           string  `json:"MainCharacter"`
				OriginVideoID           string  `json:"OriginVideoID"`
				OriginVideoURL          string  `json:"OriginVideoURL"`
				QualityGrade            string  `json:"QualityGrade"`
				SecurityLevel           string  `json:"SecurityLevel"`
				ShotPlaceCode           string  `json:"ShotPlaceCode"`
				ShotPlaceFullAdress     string  `json:"ShotPlaceFullAdress"`
				ShotPlaceLongitude      float64 `json:"ShotPlaceLongitude"`
				ShotPlacetLatitude      float64 `json:"ShotPlacetLatitude"`
				SpecialName             string  `json:"SpecialName"`
				StoragePath             string  `json:"StoragePath"`
				ThumbnailStoragePath    string  `json:"ThumbnailStoragePath"`
				TimeErr                 int64   `json:"TimeErr"`
				Title                   string  `json:"Title"`
				TitleNote               string  `json:"TitleNote"`
				VerticalShotDirection   string  `json:"VerticalShotDirection"`
				VideoID                 string  `json:"VideoID"`
				VideoLen                int64   `json:"VideoLen"`
				VideoProcFlag           int64   `json:"VideoProcFlag"`
				VideoSource             string  `json:"VideoSource"`
				Width                   int64   `json:"Width"`
			} `json:"VideoSliceInfo"`
		} `json:"VideoSlice"`
	} `json:"VideoSliceListObject"`
}
