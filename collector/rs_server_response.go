package collector

type rsServerResponse struct {
	Host                  string  `json:"tag.Hostname"`
	Role                  string  `json:"tag.Context"`
	MemStoreSize          int     `json:"memStoreSize"`
	RegionCount           int     `json:"regionCount"`
	StoreCount            int     `json:"storeCount"`
	StoreFileCount        int     `json:"storeFileCount"`
	StoreFileSize         int     `json:"storeFileSize"`
	TotalRequestCount     int     `json:"totalRequestCount"`
	SplitQueueLength      int     `json:"splitQueueLength"`
	CompactionQueueLength int     `json:"compactionQueueLength"`
	FlushQueueLength      int     `json:"flushQueueLength"`
	BlockCountHitPercent  float64 `json:"blockCountHitPercent"`
	SlowAppendCount       int     `json:"slowAppendCount"`
	SlowDeleteCount       int     `json:"slowDeleteCount"`
	SlowGetCount          int     `json:"slowGetCount"`
	SlowPutCount          int     `json:"slowPutCount"`
	SlowIncrementCount    int     `json:"slowIncrementCount"`
}
