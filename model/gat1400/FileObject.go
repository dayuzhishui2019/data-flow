package gat1400

type FileObject struct {
	FileListObject struct {
		File []struct {
			Data     string `json:"Data"`
			FileInfo struct {
				FileFormat    string `json:"FileFormat"`
				FileHash      string `json:"FileHash"`
				FileID        string `json:"FileID"`
				FileName      string `json:"FileName"`
				FileSize      int64  `json:"FileSize"`
				InfoKind      int64  `json:"InfoKind"`
				SecurityLevel string `json:"SecurityLevel"`
				Source        string `json:"Source"`
				StoragePath   string `json:"StoragePath"`
				SubmiterName  string `json:"SubmiterName"`
				SubmiterOrg   string `json:"SubmiterOrg"`
				Title         string `json:"Title"`
			} `json:"FileInfo"`
		} `json:"File"`
	} `json:"FileListObject"`
}
