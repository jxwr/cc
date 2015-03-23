package meta

type AppConfig struct {
	AppName      string   `json:"appname"`
	AutoFailover bool     `json:"autofailover"`
	MasterRegion string   `json:"master_region"`
	Regions      []string `json:"regions"`
}

type ControllerConfig struct {
	Ip       string `json:"ip"`
	HttpPort int    `json:"http_port"`
	WsPort   int    `json:"ws_port"`
	Region   string `json:"region"`
}
