package collector

type masterServerResponse struct {
	Host                 string  `json:"tag.Hostname"`
	Role                 string  `json:"tag.Context"`
	NumRegionServers     int     `json:"numRegionServers"`
	NumDeadRegionServers int     `json:"numDeadRegionServers"`
	IsActiveMaster       string  `json:"tag.isActiveMaster"`
	AverageLoad          float64 `json:"averageLoad"`
}
