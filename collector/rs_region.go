package collector

import (
	"fmt"
	"io/ioutil"
	"sync"

	"net/http"
	"net/url"
	"strings"

	"../utils"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tidwall/gjson"
)

var (
	defaultHBaseRsRegionLabels = []string{"host", "role", "namespace", "htable", "hregion"}
)

type hbaseRegionJmxMetric struct {
	Namespace string
	Table     string
	Region    string
	Metric    string
	Value     float64
}

type RsRegion struct {
	logger log.Logger
	url    *url.URL

	metrics map[string]*prometheus.Desc
	mutex   sync.Mutex

	jmxs []*hbaseRegionJmxMetric
}

func newMetric(metric string, doc string) *prometheus.Desc {
	subsystem := "region"

	return prometheus.NewDesc(
		prometheus.BuildFQName(namespace, subsystem, metric),
		doc,
		defaultHBaseRsRegionLabels,
		nil,
	)
}

func NewRsRegion(logger log.Logger, url *url.URL) *RsRegion {
	return &RsRegion{
		logger: logger,
		url:    url,

		metrics: map[string]*prometheus.Desc{
			"read_request_count":          newMetric("read_request_count", "The description of read_request_count"),
			"write_request_count":         newMetric("write_request_count", "The description of write_request_count"),
			"store_count":                 newMetric("store_count", "The description of store_count"),
			"store_file_count":            newMetric("store_file_count", "The description of store_file_count"),
			"mem_store_size":              newMetric("mem_store_size", "The description of mem_store_size"),
			"store_file_size":             newMetric("store_file_size", "The description of store_file_size"),
			"compactions_completed_count": newMetric("compactions_completed_count", "The description of compactions_completed_count"),
			"num_bytes_compacted_count":   newMetric("num_bytes_compacted_count", "The description of numBytes_compacted_count"),
			"num_files_compacted_count":   newMetric("num_files_compacted_count", "The description of num_files_compacted_count"),
		},

		jmxs: []*hbaseRegionJmxMetric{},
	}
}

func (m *RsRegion) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range m.metrics {
		ch <- metric
	}
}

func (r *RsRegion) fetchAndDecodeRsRegion() (string, string, error) {
	u := *r.url
	r.jmxs = r.jmxs[0:0]
	url := u.String() + "?" + "qry=Hadoop:service=HBase,name=RegionServer,sub=Regions"
	res, err := http.Get(url)

	if err != nil {
		return "", "", fmt.Errorf("failed to get cluster health from %s://%s:%s%s: %s",
			u.Scheme, u.Hostname(), u.Port(), u.Path, err)
	}

	defer func() {
		err = res.Body.Close()
		if err != nil {
			_ = level.Warn(r.logger).Log(
				"msg", "failed to close http.Client",
				"err", err,
			)
		}
	}()

	if res.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("HTTP Request failed with code %d", res.StatusCode)
	}

	bts, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", "", err
	}

	data := gjson.Get(string(bts), "beans").Array()[0].Map()

	host := data["tag.Hostname"].String()
	role := data["tag.Context"].String()

	for k, v := range data {
		if strings.HasPrefix(k, "Namespace") {
			keys := utils.SplitHBaseRegionStr(k)
			values := v.Float()

			r.jmxs = append(r.jmxs, &hbaseRegionJmxMetric{
				keys[0],
				keys[1],
				keys[2],
				keys[3],
				values,
			})

		}
	}

	return host, role, nil

}

func (r *RsRegion) Collect(ch chan<- prometheus.Metric) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	var err error

	host, role, err := r.fetchAndDecodeRsRegion()

	if err != nil {
		_ = level.Warn(r.logger).Log(
			"msg", "failed to fetch and decode cluster health",
			"err", err,
		)
		return
	}

	for _, jmx := range r.jmxs {
		Type := prometheus.GaugeValue
		Value := jmx.Value
		Labels := []string{host, role, jmx.Namespace, jmx.Table, jmx.Region}

		if jmx.Metric == "storeCount" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "store_count"),
					"The number of store_count.",
					defaultHBaseRsRegionLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)
		} else if jmx.Metric == "storeFileCount" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "store_file_count"),
					"The number of store_file_count.",
					defaultHBaseRsRegionLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "memStoreSize" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "mem_store_size"),
					"The number of mem_store_size.",
					defaultHBaseRsRegionLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)
		} else if jmx.Metric == "storeFileSize" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "store_file_size"),
					"The number of store_file_size.",
					defaultHBaseRsRegionLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "compactionsCompletedCount" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "compactions_completed_count"),
					"The number of compactions_completed_count.",
					defaultHBaseRsRegionLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "readRequestCount" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "read_request_count"),
					"The number of read_request_count.",
					defaultHBaseRsRegionLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "writeRequestCount" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "write_request_count"),
					"The number of write_request_count.",
					defaultHBaseRsRegionLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "numFilesCompactedCount" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "num_files_compacted_count"),
					"The number of num_files_compacted_count.",
					defaultHBaseRsRegionLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		} else if jmx.Metric == "numBytesCompactedCount" {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					prometheus.BuildFQName(namespace, "", "num_bytes_compacted_count"),
					"The number of num_bytes_compacted_count.",
					defaultHBaseRsRegionLabels, nil,
				),
				Type,
				Value,
				Labels...,
			)

		}
	}
}
