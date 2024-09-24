package main

import (
	"net/http"
	"os"

	"github.com/alecthomas/kingpin/v2"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/promlog"
	"github.com/prometheus/common/promlog/flag"
	"github.com/prometheus/common/version"
)

const (
	sshPingEndpoint = "/sshping"
	metricsEndpoint = "/metrics"
)

var (
	configFile    = kingpin.Flag("config.file", "sshping exporter configuration file.").Default("sshping_exporter.yml").String()
	targetFile    = kingpin.Flag("target.file", "sshping exporter target file.").Default("device.yml").String()
	listenAddress = kingpin.Flag("web.listen-address",
		"Address to listen on for web interface and telemetry",
	).Default(":9966").String()
)

func main() {
	promlogConfig := &promlog.Config{}
	flag.AddFlags(kingpin.CommandLine, promlogConfig)
	kingpin.Version(version.Print("sshping_exporter"))
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	logger := promlog.New(promlogConfig)

	run(logger)
}
func run(logger log.Logger) {
	level.Info(logger).Log("msg", "Starting sshping_exporter", "version", version.Info())
	level.Info(logger).Log("msg", "Build context", "build_context", version.BuildContext())
	level.Info(logger).Log("msg", "Starting Server", "address", *listenAddress)

	exporterConfig, exporterTargets := ReadConfig(configFile, targetFile, logger)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>SSH Ping Exporter</title></head>
             <body>
             <h1>SSH Exporter</h1>
             <p><a href='` + sshPingEndpoint + `'>SSH Ping Metrics</a></p>
             <p><a href='` + metricsEndpoint + `'>Exporter Metrics</a></p>
             </body>
             </html>`))
	})
	http.Handle(sshPingEndpoint, SSHPingHandler(exporterConfig, exporterTargets, logger))
	http.Handle(metricsEndpoint, promhttp.Handler())
	err := http.ListenAndServe(*listenAddress, nil)
	if err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}
}

func SSHPingHandler(exporterConfig *ExporterConfig, exporterTargets *ExporterTargets, logger log.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		level.Info(logger).Log("msg", "ssh ping handler")

		registry := prometheus.NewPedanticRegistry()

		dsw := r.URL.Query().Get("target")
		if dsw == "" {
			http.Error(w, "'target' parameter must be specified", http.StatusBadRequest)
			return
		}

		sshPingCollector := NewSSHPingCollecotr(dsw, exporterConfig, exporterTargets, logger)

		registry.MustRegister(sshPingCollector)

		gatherers := prometheus.Gatherers{registry}

		h := promhttp.HandlerFor(gatherers, promhttp.HandlerOpts{})
		h.ServeHTTP(w, r)
	}
}
