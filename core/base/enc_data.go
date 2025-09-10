package base

type EncData struct {
	Time     int64  `json:"time"`
	CheckKey string `json:"checkKey"`
	EncData  string `json:"encData"`
}

type DeviceInfo struct {
	Gaid           string `json:"gaid"`
	AndroidId      string `json:"androidId"`
	AppVersion     string `json:"appVersion"`
	AndroidVersion string `json:"androidVersion"`
	SdkVersion     string `json:"sdkVersion"`
	Brand          string `json:"brand"`
	Model          string `json:"model"`
	Extension      string `json:"extension"`
	Manufacturer   string `json:"manufacturer"`
	Data           string `json:"data"`
}
