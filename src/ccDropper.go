package ccDropper

import (
	v1 "../../minimega/phenix/types/version/v1"
	"../tmpl"
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const NAME = "ccDropper"

type CCDROPPER struct {
}

type host struct {
	hostname string
	cfg      agentConfig `yaml:",inline"`
}

type agentConfig struct {
	agentPath string `yaml: "agent_path"`
	autoStart bool   `yaml: "auto_start"`
	agent     string `yaml: "agent"`
	agentArgs string `yaml: "agent_args"`
}

type dropperConfig struct {
	generic  agentConfig
	specific []host
}

func (CCDROPPER) Init(...interface{}) error {
	return nil
}

func (CCDROPPER) Name() string {
	return NAME
}

var config dropperConfig

var startupDir = ""

func parseConfig(spec *v1.ExperimentSpec) {
	for _, e := range spec.Scenario.Apps.Experiment {
		if e.Name == NAME {
			metaMarshal, err := yaml.Marshal(e.Metadata)
			err = yaml.Unmarshal(metaMarshal, &config)
			if err != nil {
				log.Fatal("Error parsing configuration \n" + err.Error())
			}
		}
	}
	startupDir = spec.BaseDir + "/startup"
}

func agentPath(ext string, agent string, path string) string {
	dir, err := os.Open(path)
	if err != nil {
		log.Fatal("Error opeing agent directory, check scenario \n" + err.Error())
	} else {
		files, er := dir.Readdir(-1)
		dir.Close()
		if er != nil {
			log.Fatal(er)
		}
		for _, file := range files {
			if strings.Contains(file.Name(), agent) {
				if ext != "" && strings.HasSuffix(file.Name(), ext) {
					return filepath.Join(path, file.Name())
				}
				return filepath.Join(path, file.Name())
			}
		}
	}
	return ""
}

func injectCC(node *v1.Node) error {
	agentDst := "/minimega/"
	agentCfg := config.generic
	ext := ""
	for _, host := range config.specific {
		if host.hostname == node.General.Hostname {
			agentCfg = host.cfg
		}
	}

	switch node.Hardware.OSType {
	case v1.OSType_Windows:
		ext = "exe"

		var (
			startupFile = startupDir + "/" + node.General.Hostname + "-startup.ps1"
			schedFile   = startupDir + "/" + node.General.Hostname + "-scheduler.cmd"
		)

		a := &v1.Injection{
			Src: startupFile,
			Dst: agentDst + "startup.ps1",
		}
		b := &v1.Injection{
			Src: schedFile,
			Dst: "ProgramData/Microsoft/Windows/Start Menu/Programs/StartUp/CommandAndControl.cmd",
		}

		node.Injections = append(node.Injections, a, b)

	case v1.OSType_Linux, v1.OSType_RHEL, v1.OSType_CentOS:
		ext = ""

		var (
			startupFile = startupDir + "/" + node.General.Hostname + "-startup.sh"
			schedFile   = startupDir + "/" + node.General.Hostname + "-startup.service"
		)
		a := &v1.Injection{
			Src:         startupFile,
			Dst:         agentDst + "startup.sh",
			Description: "",
		}
		b := &v1.Injection{
			Src:         schedFile,
			Dst:         "/etc/systemd/system/CommandAndControl.service",
			Description: "",
		}
		node.Injections = append(node.Injections, a, b)

	}
	//inject the actual agent
	a := &v1.Injection{
		Src:         agentPath(ext, agentCfg.agent, agentCfg.agentPath),
		Dst:         agentDst,
		Description: "",
	}
	node.Injections = append(node.Injections, a)
	return nil
}

func (this *CCDROPPER) Configure(spec *v1.ExperimentSpec) error {
	parseConfig(spec)
	not_vms := spec.Topology.FindNodesWithLabels("hitl")
	vms := spec.Topology.Nodes
	for _, not_vm := range not_vms {
		for _, vm := range vms {
			if vm.General.Hostname == not_vm.General.Hostname {
				break
			} else {
				injectCC(vm)
			}
		}
	}

	return nil
}

func (this CCDROPPER) Start(spec *v1.ExperimentSpec) error {
	parseConfig(spec)

	not_vms := spec.Topology.FindNodesWithLabels("hitl")
	vms := spec.Topology.Nodes
	for _, not_vm := range not_vms {
		for _, vm := range vms {
			if vm.General.Hostname == not_vm.General.Hostname {
				break
			} else {
				agentCfg := config.generic
				for _, host := range config.specific {
					if host.hostname == vm.General.Hostname {
						agentCfg = host.cfg
					}
				}
				if vm.Type == "Router" {
					file := startupDir + "/" + vm.General.Hostname + "-startup.sh"
					if err := tmpl.CreateFileFromTemplate("linux_startup.tmpl", agentCfg, file); err != nil {
						return fmt.Errorf("generating linux command and control startup script: %w", err)
					}
					file = startupDir + "/" + vm.General.Hostname + "-startup.service"
					if err := tmpl.CreateFileFromTemplate("linux-service.tmpl", agentCfg, file); err != nil {
						return fmt.Errorf("generating linux command and control service script: %w", err)
					}

				} else if vm.Hardware.OSType == v1.OSType_Linux {
					file := startupDir + "/" + vm.General.Hostname + "-startup.sh"
					if err := tmpl.CreateFileFromTemplate("linux_startup.tmpl", agentCfg, file); err != nil {
						return fmt.Errorf("generating linux command and control startup script: %w", err)
					}
					file = startupDir + "/" + vm.General.Hostname + "-startup.service"
					if err := tmpl.CreateFileFromTemplate("linux-service.tmpl", agentCfg, file); err != nil {
						return fmt.Errorf("generating linux command and control service script: %w", err)
					}

				} else if vm.Hardware.OSType == v1.OSType_Windows {
					file := startupDir + "/" + vm.General.Hostname + "-startup.ps1"
					if err := tmpl.CreateFileFromTemplate("windows_startup.tmpl", agentCfg, file); err != nil {
						return fmt.Errorf("generating windows command and control startup script: %w", err)
					}
					file = startupDir + "/" + vm.General.Hostname + "-scheduler.cmd"
					if err := tmpl.CreateFileFromTemplate("windows-scheduler.tmpl", agentCfg, file); err != nil {
						return fmt.Errorf("generating windows command and control service script: %w", err)
					}
				}
			}
		}
	}

	return nil
}

func (CCDROPPER) PostStart(spec *v1.ExperimentSpec) error {
	return nil
}

func (CCDROPPER) Cleanup(spec *v1.ExperimentSpec) error {
	return nil
}
