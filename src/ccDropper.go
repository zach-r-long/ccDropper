package main

import (
	v1 "../../minimega/phenix/types/version/v1"
	"../tmpl"
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const NAME = "ccDropper"
const INJECTPATH = "/minimega/"

var universalConfig int = 0

type HostAgentConfig struct {
	Hostname    string `yaml:"hostname"`
	AgentPath   string `yaml:"agent_path"`
	AutoStart   bool   `yaml:"auto_start"`
	Agent       string `yaml:"agent"`
	AgentArgs   string `yaml:"agent_args"`
	ServiceType string `yaml:"service_type"`
}

type DropperConfig struct {
	Hosts []HostAgentConfig `yaml: "cc_hosts"`
}

func Name() string {
	return NAME
}

func agentPath(ext string, agent string, path string) (string, string) {
	dir, err := os.Open(path)
	if err != nil {
		log.Fatal("ccDropper: Error opening agent directory, check scenario \n"+"Path: "+path+"\n", err)
	} else {
		files, er := dir.Readdir(-1)
		dir.Close()
		if er != nil {
			log.Fatal(er)
		}
		for _, file := range files {
			if strings.Contains(file.Name(), agent) {
				if ext != "" && strings.Contains(file.Name(), ".") && strings.HasSuffix(file.Name(), ext) {
					return filepath.Join(path, file.Name()), INJECTPATH + file.Name()
				} else if strings.HasSuffix(file.Name(), ext) && !strings.Contains(file.Name(), ".") {
					return filepath.Join(path, file.Name()), INJECTPATH + file.Name()

				}
			}
		}
	}
	return "", ""
}

func getVms(spec *v1.ExperimentSpec) []*v1.Node {
	vms := spec.Topology.Nodes
	not_vms := spec.Topology.FindNodesWithLabels("hitl")
	var vmList []*v1.Node
	if len(not_vms) > 0 {
		for _, vm := range vms {
			for _, not_vm := range not_vms {
				if vm.General.Hostname != not_vm.General.Hostname {
					vmList = append(vmList, vm)
				}
			}
		}
		return vmList
	}
	for _, vm := range vms {
		vmList = append(vmList, vm)
	}
	return vmList
}

func configure(spec *v1.ExperimentSpec, config DropperConfig, startupDir string) {
	vms := getVms(spec)
	for _, node := range vms {
		log.Printf("Configuring Host %s\n", node.General.Hostname)
		agentCfg := config.Hosts[universalConfig]
		ext := ""
		for _, host := range config.Hosts {
			if host.Hostname == node.General.Hostname {
				log.Printf("Found custom config for host %s\n", host.Hostname)
				agentCfg = host
			}
		}

		switch node.Hardware.OSType {
		case v1.OSType_Windows:
			ext = "exe"

			var (
				startupFile = startupDir + "/" + node.General.Hostname + "-cc_startup.ps1"
				schedFile   = startupDir + "/" + node.General.Hostname + "-cc_scheduler.cmd"
			)

			a := &v1.Injection{
				Src: startupFile,
				Dst: INJECTPATH + "cc_startup.ps1",
			}
			b := &v1.Injection{
				Src: schedFile,
				Dst: "ProgramData/Microsoft/Windows/Start Menu/Programs/StartUp/CommandAndControl.cmd",
			}

			log.Print(" Windows Injections\n")
			log.Printf("%v\n%v\n", a, b)

			node.Injections = append(node.Injections, a, b)

		case v1.OSType_Linux, v1.OSType_RHEL, v1.OSType_CentOS:
			ext = ""

			var (
				startupFile = startupDir + "/" + node.General.Hostname + "-cc_startup.sh"
				svcFile     = startupDir + "/" + node.General.Hostname + "-cc_startup.service"
				svcLink     = startupDir + "/" + node.General.Hostname + "-cc_startup.serviceLink"
			)
			a := &v1.Injection{
				Src:         startupFile,
				Dst:         INJECTPATH + "cc_startup.sh",
				Description: "",
			}
			if strings.ToLower(agentCfg.ServiceType) == "systemd" {
				b := &v1.Injection{
					Src:         svcFile,
					Dst:         "/etc/systemd/system/CommandAndControl.service",
					Description: "",
				}
				c := &v1.Injection{
					Src:         svcLink,
					Dst:         "/etc/systemd/system/multi-user.target.wants/CommandAndControl.service",
					Description: "",
				}

				log.Print(" Linux Injections\n")
				log.Printf("%v\n%v\n", a, b, c)

				node.Injections = append(node.Injections, a, b, c)
			} else {
				b := &v1.Injection{
					Src:         svcFile,
					Dst:         "/etc/init.d/CommandAndControl",
					Description: "",
				}
				c := &v1.Injection{
					Src:         svcLink,
					Dst:         "/etc/rc5.d/S99CommandAndControl",
					Description: "",
				}

				log.Print(" Linux Injections\n")
				log.Printf("%v\n%v\n", a, b, c)

				node.Injections = append(node.Injections, a, b, c)
			}

		}

		agentSrc, agentDst := agentPath(ext, agentCfg.Agent, agentCfg.AgentPath)
		if len(agentSrc) < 2 {
			erro := fmt.Sprintf("Agent not found when looking for %s.%s in %s", agentCfg.Agent, ext, agentCfg.AgentPath)
			log.Fatal(erro)
		}
		a := &v1.Injection{
			Src:         agentSrc,
			Dst:         agentDst,
			Description: "",
		}
		log.Print(" Agent Injection\n")
		log.Printf("%v\n", a)

		node.Injections = append(node.Injections, a)
	}
}

func start(spec *v1.ExperimentSpec, config DropperConfig, startupDir string) {
	vms := getVms(spec)
	for _, vm := range vms {
		agentCfg := config.Hosts[universalConfig]
		for _, host := range config.Hosts {
			if host.Hostname == vm.General.Hostname {
				agentCfg = host
			}
		}
		switch vm.Hardware.OSType {
		case v1.OSType_Linux, v1.OSType_RHEL, v1.OSType_CentOS:

			file := startupDir + "/" + vm.General.Hostname + "-cc_startup.sh"
			if err := tmpl.CreateFileFromTemplate("linux_startup.tmpl", agentCfg, file, 0755); err != nil {
				log.Fatal("generating linux command and control startup script: ", err)
			}
			if strings.ToLower(agentCfg.ServiceType) == "systemd" {

				file = startupDir + "/" + vm.General.Hostname + "-cc_startup.service"
				if err := tmpl.CreateFileFromTemplate("systemd-service.tmpl", agentCfg, file, 0644); err != nil {
					log.Fatal("generating linux command and control service script: ", err)
				}
				file = startupDir + "/" + vm.General.Hostname + "-cc_startup.serviceLink"
				//Symlinks will not overwrite, so remove before attempting to relink
				os.Remove(file)
				if err := os.Symlink("/etc/systemd/system/CommandAndControl.service", file); err != nil {
					log.Fatal("generating linux command and control service link: ", err)
				}
			} else {
				file = startupDir + "/" + vm.General.Hostname + "-cc_startup.service"
				if err := tmpl.CreateFileFromTemplate("sysinitv-service.tmpl", agentCfg, file, 0755); err != nil {
					log.Fatal("generating linux command and control service script: ", err)
				}

				file = startupDir + "/" + vm.General.Hostname + "-cc_startup.serviceLink"
				//Symlinks will not overwrite, so remove before attempting to relink
				os.Remove(file)
				if err := os.Symlink("/etc/init.d/CommandAndControl", file); err != nil {
					log.Fatal("generating linux command and control service link: ", err)
				}

			}

		case v1.OSType_Windows:
			file := startupDir + "/" + vm.General.Hostname + "-cc_startup.ps1"
			if err := tmpl.CreateFileFromTemplate("windows_startup.tmpl", agentCfg, file, 0755); err != nil {
				log.Fatal("generating windows command and control startup script: ", err)
			}
			file = startupDir + "/" + vm.General.Hostname + "-cc_scheduler.cmd"
			if err := tmpl.CreateFileFromTemplate("windows-scheduler.tmpl", agentCfg, file, 0755); err != nil {
				log.Fatal("generating windows command and control service script: ", err)
			}
		}
	}

}

func postStart(spec v1.ExperimentSpec) error {
	return nil
}

func cleanup(spec v1.ExperimentSpec) error {
	return nil
}

func main() {
	logFile, err := os.Create("/tmp/ccDropperLog")
	defer logFile.Close()
	log.SetOutput(logFile)
	if err != nil {
		log.Fatal("ccDropper: Can't create log file ... exiting")
	}
	log.Println("Appliction Start")
	var spec v1.ExperimentSpec
	mode := os.Args[1]
	var config DropperConfig

	//read in exp spec as json blob
	err = json.NewDecoder(os.Stdin).Decode(&spec)
	if err != nil {
		log.Fatal(err.Error())
	}
	//Get the application configuration data
	for _, e := range spec.Scenario.Apps.Experiment {
		if e.Name == NAME {
			log.Print("Found config")
			//log.Print(e.Metadata["cc_hosts"])
			vEncoding := ""
			for i, hostConfig := range e.Metadata["cc_hosts"].([]interface{}) {
				gg := ""
				for k, v := range hostConfig.(map[string]interface{}) {
					switch v.(type) {
					case float32, float64, int8, int16, int32, int64, uint8, uint16, uint32, uint64, int, uint:
						vEncoding = "%v."
					case bool:
						vEncoding = "%v"
					case string:
						vEncoding = "\"%v\""
					}
					gg += fmt.Sprintf("%s: "+vEncoding+"\n", k, v)
					if k == "*" {
						universalConfig = i
					}
					//log.Printf("Key: %s:%v formated as %T",k,v,v)
				}
				//log.Printf("\n%s\n",gg)
				cfg := HostAgentConfig{}
				err = yaml.Unmarshal([]byte(gg), &cfg)
				if err != nil {
					log.Fatal(err)
				}
				//log.Printf("%v", cfg)
				config.Hosts = append(config.Hosts, cfg)

			}
		}
	}
	log.Print("Config: \n")
	log.Print(config)
	//computer the start directory where things will be generated
	startupDir := spec.BaseDir + "/startup"
	switch mode {
	case "configure":
		log.Print(" ----------------------Configure------------------\n")
		configure(&spec, config, startupDir)
	case "start":
		log.Print(" ----------------------Start------------------\n")
		start(&spec, config, startupDir)
	case "postStart":
		log.Print(" ----------------------Post Start------------------\n")
		postStart(spec)
	case "cleanup":
		log.Print(" ----------------------Cleanup------------------\n")
		cleanup(spec)
	}
	data, err := json.Marshal(spec)
	out := bytes.NewBuffer(data)
	if err != nil {
		log.Fatal("ccDropper: marshaling experiment spec to JSON: %w", err)
	}
	fmt.Print(out)
	f, err := os.Create("/tmp/outJson")
	defer f.Close()
	f.Write(data)
	f.Sync()

}
