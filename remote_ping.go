package main

import (
	"fmt"
	"strings"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"golang.org/x/crypto/ssh"
)

func RemotePing(dsw string, exporterConfig *ExporterConfig, exporterTargets *ExporterTargets, labelNames []string, logger log.Logger) []ProbeResult {
	level.Info(logger).Log("msg", "remote ping start")
	cmds := FilterBatchPingByDSW(dsw, exporterTargets, logger)
	config := &ssh.ClientConfig{
		User: exporterConfig.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(exporterConfig.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	level.Info(logger).Log("msg", "connecting to SSH server", "DSW", dsw)
	client, err := ssh.Dial("tcp", dsw, config)
	if err != nil {
		level.Error(logger).Log("err", err)

	}
	defer client.Close()

	level.Info(logger).Log("msg", "creating new SSH session")
	session, err := client.NewSession()
	if err != nil {
		level.Error(logger).Log("err", err)
	}
	defer session.Close()

	level.Info(logger).Log("msg", "ping start")

	output, err := session.CombinedOutput(*cmds)
	if err != nil {
		level.Error(logger).Log("err", err)
	}
	pingResult := string(output)
	probeMap := ParsePingResult(&pingResult, logger)
	probeResults := LabelResult(probeMap, exporterTargets, labelNames, logger)
	return probeResults
}

func FilterBatchPingByDSW(dsw string, exporterTargets *ExporterTargets, logger log.Logger) *string {

	level.Info(logger).Log("msg", "filter batch ping by DSW")
	cmds := ""
	pingformat := "ping -c 1 -vpn-instance vpn-default %s ; "
	dswip := strings.Split(dsw, ":")[0]
	for _, exporterTarget := range *exporterTargets {
		if exporterTarget.Labels["DSW_IP"] == dswip {
			for _, target := range exporterTarget.Targets {
				cmd := fmt.Sprintf(pingformat, target)
				cmds += cmd
			}
		}
	}

	return &cmds
}

func ParsePingResult(results *string, logger log.Logger) map[string]int {
	level.Info(logger).Log("msg", "parse ping result")
	probeMap := make(map[string]int)
	lines := strings.Split(*results, "\n")
	var ip string
	for _, line := range lines {
		if strings.Contains(line, "Ping statistics for") {
			ip = strings.Split(line, " ")[4]

		} else if strings.Contains(line, ", 0.0% packet loss") {

			probeMap[ip] = 1
		} else if strings.Contains(line, ", 100.0% packet loss") {

			probeMap[ip] = 0
		}
	}
	return probeMap
}

func LabelResult(probeMap map[string]int, exporterTargets *ExporterTargets, labelNames []string, logger log.Logger) []ProbeResult {
	level.Info(logger).Log("msg", "label result")
	probeResults := []ProbeResult{}

	for ip, status := range probeMap {
	XL:
		for _, exporterTarget := range *exporterTargets {
			for _, target := range exporterTarget.Targets {
				if target == ip {
					labelValues := make([]string, 0)
					for _, labelName := range labelNames {
						if labelName == "instanceip" {
							labelValues = append(labelValues, ip)
						} else {
							labelValue := exporterTarget.Labels[labelName]
							labelValues = append(labelValues, labelValue)
						}
					}
					probeResults = append(probeResults, ProbeResult{status, labelValues})
					break XL
				}
			}
		}
	}
	return probeResults
}
