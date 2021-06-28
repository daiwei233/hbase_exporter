package collector

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/tidwall/gjson"
)

const (
	namespace = "hbase"
)

var (
	defaultHBaseJvmLabels      = []string{"host", "role"}
	defaultHBaseLabelJvmValues = func(hbaseJvm hbaseJvmResponse) []string {
		return []string{
			hbaseJvm.Host,
			strings.ToLower(hbaseJvm.Role),
		}
	}
)

type hbaseJvmMetric struct {
	Type   prometheus.ValueType
	Desc   *prometheus.Desc
	Value  func(hbaseJvm hbaseJvmResponse) float64
	Labels func(hbaseJvm hbaseJvmResponse) []string
}

type HBaseJvm struct {
	logger log.Logger
	url    *url.URL

	up                              prometheus.Gauge
	totalScrapes, jsonParseFailures prometheus.Counter

	metrics []*hbaseJvmMetric
}

func NewHBaseJvm(logger log.Logger, url *url.URL) *HBaseJvm {
	subsystem := "jvm"

	return &HBaseJvm{
		logger: logger,
		url:    url,

		up: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: prometheus.BuildFQName(namespace, subsystem, "up"),
			Help: "Was the last scrape of the ElasticSearch cluster health endpoint successful.",
		}),
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, subsystem, "total_scrapes"),
			Help: "Current total ElasticSearch cluster health scrapes.",
		}),
		jsonParseFailures: prometheus.NewCounter(prometheus.CounterOpts{
			Name: prometheus.BuildFQName(namespace, subsystem, "json_parse_failures"),
			Help: "Number of errors while parsing JSON.",
		}),

		metrics: []*hbaseJvmMetric{
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "mem_non_heap_used_m"),
					"The number of mem_non_heap_used_m.",
					defaultHBaseJvmLabels, nil,
				),
				Value: func(hbaseJvm hbaseJvmResponse) float64 {
					return float64(hbaseJvm.MemNonHeapUsedM)
				},
				Labels: defaultHBaseLabelJvmValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "mem_heap_used_m"),
					"The number of mem_heap_used_m.",
					defaultHBaseJvmLabels, nil,
				),
				Value: func(hbaseJvm hbaseJvmResponse) float64 {
					return float64(hbaseJvm.MemHeapUsedM)
				},
				Labels: defaultHBaseLabelJvmValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "mem_heap_mx_m"),
					"The number of mem_heap_mx_m.",
					defaultHBaseJvmLabels, nil,
				),
				Value: func(hbaseJvm hbaseJvmResponse) float64 {
					return float64(hbaseJvm.MemHeapMaxM)
				},
				Labels: defaultHBaseLabelJvmValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "mem_max_m"),
					"The number of mem_max_m.",
					defaultHBaseJvmLabels, nil,
				),
				Value: func(hbaseJvm hbaseJvmResponse) float64 {
					return float64(hbaseJvm.MemMaxM)
				},
				Labels: defaultHBaseLabelJvmValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "gc_time_millis"),
					"The number of gc_time_millis.",
					defaultHBaseJvmLabels, nil,
				),
				Value: func(hbaseJvm hbaseJvmResponse) float64 {
					return float64(hbaseJvm.GcTimeMillis)
				},
				Labels: defaultHBaseLabelJvmValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "gc_count"),
					"The number of gc_count.",
					defaultHBaseJvmLabels, nil,
				),
				Value: func(hbaseJvm hbaseJvmResponse) float64 {
					return float64(hbaseJvm.GcCount)
				},
				Labels: defaultHBaseLabelJvmValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "thread_blocked"),
					"The number of thread_blocked.",
					defaultHBaseJvmLabels, nil,
				),
				Value: func(hbaseJvm hbaseJvmResponse) float64 {
					return float64(hbaseJvm.ThreadsBlocked)
				},
				Labels: defaultHBaseLabelJvmValues,
			},
		},
	}
}

func (m *HBaseJvm) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range m.metrics {
		ch <- metric.Desc
	}

	ch <- m.up.Desc()
	ch <- m.totalScrapes.Desc()
	ch <- m.jsonParseFailures.Desc()
}

func (m *HBaseJvm) fetchAndDecodeHBaseJvm() (hbaseJvmResponse, error) {
	var mjr hbaseJvmResponse

	u := *m.url
	url := u.String() + "?" + "qry=Hadoop:service=HBase,name=JvmMetrics"
	res, err := http.Get(url)

	if err != nil {
		return mjr, fmt.Errorf("failed to get cluster health from %s://%s:%s%s: %s",
			u.Scheme, u.Hostname(), u.Port(), u.Path, err)
	}

	defer func() {
		err = res.Body.Close()
		if err != nil {
			_ = level.Warn(m.logger).Log(
				"msg", "failed to close http.Client",
				"err", err,
			)
		}
	}()

	if res.StatusCode != http.StatusOK {
		return mjr, fmt.Errorf("HTTP Request failed with code %d", res.StatusCode)
	}

	bts, err := ioutil.ReadAll(res.Body)
	if err != nil {
		m.jsonParseFailures.Inc()
		return mjr, err
	}

	data := gjson.Get(string(bts), "beans")
	hbaseJvmStr := data.Array()[0].String()

	if err := json.Unmarshal([]byte(hbaseJvmStr), &mjr); err != nil {
		m.jsonParseFailures.Inc()
		return mjr, err
	}

	return mjr, nil

}

func (m *HBaseJvm) Collect(ch chan<- prometheus.Metric) {
	var err error

	m.totalScrapes.Inc()
	defer func() {
		ch <- m.up
		ch <- m.totalScrapes
		ch <- m.jsonParseFailures
	}()

	hbaseJvmResp, err := m.fetchAndDecodeHBaseJvm()

	if err != nil {
		m.up.Set(0)
		_ = level.Warn(m.logger).Log(
			"msg", "failed to fetch and decode cluster health",
			"err", err,
		)
		return
	}
	m.up.Set(1)

	for _, metric := range m.metrics {

		ch <- prometheus.MustNewConstMetric(
			metric.Desc,
			metric.Type,
			metric.Value(hbaseJvmResp),
			metric.Labels(hbaseJvmResp)...,
		)
	}
}
