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

var (
	defaultHBaseMasterServerLabels            = []string{"host", "role"}
	defaultHBaseMasterServerLabelServerValues = func(masterServer masterServerResponse) []string {
		return []string{
			masterServer.Host,
			strings.ToLower(masterServer.Role),
		}
	}
)

type masterServerMetric struct {
	Type   prometheus.ValueType
	Desc   *prometheus.Desc
	Value  func(masterServer masterServerResponse) float64
	Labels func(masterServer masterServerResponse) []string
}

type MasterServer struct {
	logger log.Logger
	url    *url.URL

	up                              prometheus.Gauge
	totalScrapes, jsonParseFailures prometheus.Counter

	metrics []*masterServerMetric
}

func NewMasterServer(logger log.Logger, url *url.URL) *MasterServer {
	subsystem := "server"

	return &MasterServer{
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

		metrics: []*masterServerMetric{
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "average_load"),
					"The number of average_load.",
					defaultHBaseMasterServerLabels, nil,
				),
				Value: func(masterServer masterServerResponse) float64 {
					return float64(masterServer.AverageLoad)
				},
				Labels: defaultHBaseMasterServerLabelServerValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "num_regionservers"),
					"The number of num_regionservers.",
					defaultHBaseMasterServerLabels, nil,
				),
				Value: func(masterServer masterServerResponse) float64 {
					return float64(masterServer.NumRegionServers)
				},
				Labels: defaultHBaseMasterServerLabelServerValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "num_dead_regionserver"),
					"The number of num_dead_regionserver.",
					defaultHBaseMasterServerLabels, nil,
				),
				Value: func(masterServer masterServerResponse) float64 {
					return float64(masterServer.NumDeadRegionServers)
				},
				Labels: defaultHBaseMasterServerLabelServerValues,
			},
			{
				Type: prometheus.GaugeValue,
				Desc: prometheus.NewDesc(
					prometheus.BuildFQName(namespace, subsystem, "is_active_master"),
					"The number ofis_active_master.",
					defaultHBaseMasterServerLabels, nil,
				),
				Value: func(masterServer masterServerResponse) float64 {
					if masterServer.IsActiveMaster == "true" {
						return float64(1)
					} else {
						return float64(0)
					}
				},
				Labels: defaultHBaseMasterServerLabelServerValues,
			},
		},
	}
}

func (m *MasterServer) Describe(ch chan<- *prometheus.Desc) {
	for _, metric := range m.metrics {
		ch <- metric.Desc
	}

	ch <- m.up.Desc()
	ch <- m.totalScrapes.Desc()
	ch <- m.jsonParseFailures.Desc()
}

func (m *MasterServer) fetchAndDecodeMasterServer() (masterServerResponse, error) {
	var msr masterServerResponse

	u := *m.url
	url := u.String() + "?" + "qry=Hadoop:service=HBase,name=Master,sub=Server"
	res, err := http.Get(url)

	if err != nil {
		return msr, fmt.Errorf("failed to get cluster health from %s://%s:%s%s: %s",
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
		return msr, fmt.Errorf("HTTP Request failed with code %d", res.StatusCode)
	}

	bts, err := ioutil.ReadAll(res.Body)
	if err != nil {
		m.jsonParseFailures.Inc()
		return msr, err
	}

	data := gjson.Get(string(bts), "beans")
	masterServerStr := data.Array()[0].String()

	if err := json.Unmarshal([]byte(masterServerStr), &msr); err != nil {
		m.jsonParseFailures.Inc()
		return msr, err
	}

	return msr, nil
}

func (m *MasterServer) Collect(ch chan<- prometheus.Metric) {
	var err error

	m.totalScrapes.Inc()
	defer func() {
		ch <- m.up
		ch <- m.totalScrapes
		ch <- m.jsonParseFailures
	}()

	masterServerResp, err := m.fetchAndDecodeMasterServer()

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
			metric.Value(masterServerResp),
			metric.Labels(masterServerResp)...,
		)
	}
}
