package rsync

// rsync channel
type RsyncModelChan struct {
	Url  string `json:"url"`
	Dest string `jsong:"dest"`
}

// parse channel
type ParseModelChan struct {
	FilePathName string `json:"filePathName"`
}

// rsync and parse end channel, may be end
type RsyncParseEndChan struct {
}
