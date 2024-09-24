package main

import (
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/prometheus/client_golang/prometheus"
)

type SSHPingCollecotr struct {
	DSW              string
	ExporterConfig   *ExporterConfig
	ExporterTargets  *ExporterTargets
	LabelNames       []string
	ProbeSuccessDesc *prometheus.Desc
	logger           log.Logger
}

func NewSSHPingCollecotr(dsw string, exporterConfig *ExporterConfig, exporterTargets *ExporterTargets, logger log.Logger) *SSHPingCollecotr {
	labelNames := exporterTargets.GetLabelNames()

	probeSuccessDesc := prometheus.NewDesc(
		"probe_success",
		"Returns 1 if probe succeeded, 0 failed",
		labelNames,
		nil,
	)

	return &SSHPingCollecotr{
		DSW:              dsw,
		ExporterConfig:   exporterConfig,
		ExporterTargets:  exporterTargets,
		LabelNames:       labelNames,
		ProbeSuccessDesc: probeSuccessDesc,
		logger:           logger,
	}
}

func (spc *SSHPingCollecotr) Describe(ch chan<- *prometheus.Desc) {
	ch <- spc.ProbeSuccessDesc
}
func (spc *SSHPingCollecotr) Collect(ch chan<- prometheus.Metric) {
	level.Info(spc.logger).Log("msg", "SSHPingCollecotr collect")
	results := RemotePing(spc.DSW, spc.ExporterConfig, spc.ExporterTargets, spc.LabelNames, spc.logger)

	for _, result := range results {

		ch <- prometheus.MustNewConstMetric(
			spc.ProbeSuccessDesc,
			prometheus.GaugeValue,
			float64(result.status),
			result.LabelVaules...,
		)
	}
}
