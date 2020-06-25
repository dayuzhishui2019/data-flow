package gat1400

type ImageObject struct {
	ImageListObject struct {
		Image []struct {
			Data      string `json:"Data"`
			ImageInfo struct {
				CollectorName       string `json:"CollectorName"`
				CollectorOrg        string `json:"CollectorOrg"`
				ContentDescription  string `json:"ContentDescription"`
				EventSort           int64  `json:"EventSort"`
				FileFormat          string `json:"FileFormat"`
				FileHash            string `json:"FileHash"`
				FileSize            int64  `json:"FileSize"`
				Height              int64  `json:"Height"`
				ImageID             string `json:"ImageID"`
				ImageSource         string `json:"ImageSource"`
				InfoKind            int64  `json:"InfoKind"`
				SecurityLevel       string `json:"SecurityLevel"`
				ShotPlaceCode       string `json:"ShotPlaceCode"`
				ShotPlaceFullAdress string `json:"ShotPlaceFullAdress"`
				ShotTime            string `json:"ShotTime"`
				StoragePath         string `json:"StoragePath"`
				Title               string `json:"Title"`
				Width               int64  `json:"Width"`
			} `json:"ImageInfo"`
		} `json:"Image"`
	} `json:"ImageListObject"`
}