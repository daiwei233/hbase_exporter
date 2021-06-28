package collector

type hbaseJvmResponse struct {
	Host            string  `json:"tag.Hostname"`
	Role            string  `json:"tag.ProcessName"`
	SubName         string  `json:"name"`
	MemNonHeapUsedM float64 `json:"MemNonHeapUsedM"`
	MemHeapUsedM    float64 `json:"MemHeapUsedM"`
	MemHeapMaxM     float64 `json:"MemHeapMaxM"`
	MemMaxM         float64 `json:"MemMaxM"`
	GcTimeMillis    int     `json:"GcTimeMillis"`
	GcCount         int     `json:"GcCount"`
	ThreadsBlocked  int     `json:"ThreadsBlocked"`
}
