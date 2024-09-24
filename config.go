package main

import (
	"os"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"gopkg.in/yaml.v3"
)

type ExporterConfig struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type ExporterTargets []struct {
	Targets []string          `yaml:"targets"`
	Labels  map[string]string `yaml:"labels"`
}

type ProbeResult struct {
	status      int
	LabelVaules []string
}

func (et *ExporterTargets) GetLabelNames() []string {
	keys := make([]string, 0)
	for _, target := range *et {
		labels := target.Labels
		for key := range labels {
			keys = append(keys, key)
		}
		break
	}
	keys = append(keys, "instanceip")
	return keys
}

func ReadConfig(configFile *string, targetFile *string, logger log.Logger) (*ExporterConfig, *ExporterTargets) {

	level.Info(logger).Log("msg", "Read Config File")
	ExporterConfig := &ExporterConfig{}
	ReadYaml(*configFile, ExporterConfig, logger)

	exporterTargets := &ExporterTargets{}
	ReadYaml(*targetFile, exporterTargets, logger)
	return ExporterConfig, exporterTargets
}

func ReadYaml(fileName string, entity any, logger log.Logger) {
	content, err := os.ReadFile(fileName)
	if err != nil {
		level.Error(logger).Log("err", err)
	}
	yaml.Unmarshal(content, entity)
}
