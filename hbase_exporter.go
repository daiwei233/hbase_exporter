package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"

	"./collector"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/version"
	"gopkg.in/alecthomas/kingpin.v2"
)

func getLogger(loglevel, logoutput, logfmt string) log.Logger {
	var out *os.File
	switch strings.ToLower(logoutput) {
	case "stderr":
		out = os.Stderr
	case "stdout":
		out = os.Stdout
	default:
		out = os.Stdout
	}
	var logCreator func(io.Writer) log.Logger
	switch strings.ToLower(logfmt) {
	case "json":
		logCreator = log.NewJSONLogger
	case "logfmt":
		logCreator = log.NewLogfmtLogger
	default:
		logCreator = log.NewLogfmtLogger
	}

	// create a logger
	logger := logCreator(log.NewSyncWriter(out))

	// set loglevel
	var loglevelFilterOpt level.Option
	switch strings.ToLower(loglevel) {
	case "debug":
		loglevelFilterOpt = level.AllowDebug()
	case "info":
		loglevelFilterOpt = level.AllowInfo()
	case "warn":
		loglevelFilterOpt = level.AllowWarn()
	case "error":
		loglevelFilterOpt = level.AllowError()
	default:
		loglevelFilterOpt = level.AllowInfo()
	}
	logger = level.NewFilter(logger, loglevelFilterOpt)
	logger = log.With(logger,
		"ts", log.DefaultTimestampUTC,
		"caller", log.DefaultCaller,
	)
	return logger
}

func main() {
	var (
		Name          = "hbase_exporter"
		listenAddress = kingpin.Flag("web.listen-address",
			"Address to listen on for web interface and telemetry.").
			Default(":9115").Envar("WEB_LISTEN_ADDRESS").String()
		metricsPath = kingpin.Flag("web.telemetry-path",
			"Path under which to expose metrics.").
			Default("/metrics").Envar("WEB_TELEMETRY_PATH").String()
		hbaseMasterURI = kingpin.Flag("hbase.master.uri",
			"HTTP jmx address of an HBase master node.").
			Default("http://localhost:60010/jmx").Envar("HBASE_MASTER_URL").String()
		hbaseRegionserverURI = kingpin.Flag("hbase.regionserver.uri",
			"HTTP jmx address of an HBase regionserver node.").
			Default("http://localhost:60030/jmx").Envar("HBASE_REGIONSERVER_URL").String()
		hbaseIsMaster = kingpin.Flag("hbase.master",
			"Is hbase master.").
			Default("false").Envar("HBASE_IS_MASTER").Bool()
		logLevel = kingpin.Flag("log.level",
			"Sets the loglevel. Valid levels are debug, info, warn, error").
			Default("info").Envar("LOG_LEVEL").String()
		logFormat = kingpin.Flag("log.format",
			"Sets the log format. Valid formats are json and logfmt").
			Default("logfmt").Envar("LOG_FMT").String()
		logOutput = kingpin.Flag("log.output",
			"Sets the log output. Valid outputs are stdout and stderr").
			Default("stdout").Envar("LOG_OUTPUT").String()
	)

	kingpin.Version(version.Print(Name))
	kingpin.CommandLine.HelpFlag.Short('h')
	kingpin.Parse()

	logger := getLogger(*logLevel, *logOutput, *logFormat)

	hbaseMasterURL, err := url.Parse(*hbaseMasterURI)
	if err != nil {
		_ = level.Error(logger).Log(
			"msg", "failed to parse hbase.master.uri",
			"err", err,
		)
		os.Exit(1)
	}

	hbaseRegionserverURL, err := url.Parse(*hbaseRegionserverURI)
	if err != nil {
		_ = level.Error(logger).Log(
			"msg", "failed to parse hbase.regionserver.uri",
			"err", err,
		)
		os.Exit(1)
	}

	versionMetric := version.NewCollector(Name)
	prometheus.MustRegister(versionMetric)

	if *hbaseIsMaster {
		prometheus.MustRegister(collector.NewHBaseJvm(logger, hbaseMasterURL))
		prometheus.MustRegister(collector.NewMasterServer(logger, hbaseMasterURL))
	} else {
		prometheus.MustRegister(collector.NewHBaseJvm(logger, hbaseRegionserverURL))
		prometheus.MustRegister(collector.NewRsServer(logger, hbaseRegionserverURL))

		prometheus.MustRegister(collector.NewRsRegion(logger, hbaseRegionserverURL))
	}
	level.Info(logger).Log("msg", "Build context", "build_context", version.BuildContext())
	level.Info(logger).Log("msg", "Starting hbase_exporter", "version", version.Info())

	fmt.Println(len(*hbaseRegionserverURI))
	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`<html>
				   <head><title>HBase Exporter</title></head>
				   <body>
				   <h1>HBase Exporter</h1>
				   <p><a href='` + *metricsPath + `'>Metrics</a></p>
				   </body>
				   </html>`))
	})
	level.Info(logger).Log("msg", "Listening on", "address", *listenAddress)
	_ = http.ListenAndServe(*listenAddress, nil)

}
