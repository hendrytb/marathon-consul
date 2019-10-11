package consul

type Service struct {
	ID      string   `json:"ID"`
	Name    string   `json:"Name"`
	Tags    []string `json:"Tags"`
	Address string   `json:"Address"`
	Port    int      `json:"Port"`
	Check   Check    `json:"Check"`
}

type Check struct {
	HTTP     string `json:"HTTP"`
	Interval string `json:"Interval"`
	//TTL      string `json:"TTL"`
	Timeout string `json:"Timeout"`
}
