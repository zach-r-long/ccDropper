package main

import( 
	"../tmpl"
	"bytes"
	"encoding/json"
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path/filepath"
	"strings"
	"regexp"
	"io/ioutil"
)

const NAME = "ccDropper"
const INJECTPATH = "/minimega/"


type HostAgentConfig struct {
	Hostname    string `yaml:"hostname"`
	AgentPath   string `yaml:"agent_path"`
	AutoStart   bool   `yaml:"auto_start"`
	Agent       string `yaml:"agent"`
	AgentArgs   string `yaml:"agent_args"`
	ServiceType string `yaml:"service_type"`
	CustomService struct {
		ScriptPath string `yaml:"script_path"`
		InjectPath string `yaml:"inject_path"`
	} `yaml:"service_custom"` 
}

var universalConfig HostAgentConfig

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

func getVms(spec *Experiment) []*Node {
	vms := spec.Spec.Topology.Nodes
	not_vms := spec.Spec.Topology.FindNodesWithLabels("hitl")
	var vmList []*Node
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

func getAgentConfig(hostname string, config DropperConfig) HostAgentConfig { 
	hostAgentConfig := universalConfig
	log.Println("Using univeral config")
	for _,hostConfig := range config.Hosts { 
		hostRegx ,_ := regexp.Compile(hostConfig.Hostname)
		if hostRegx.MatchString(hostname) {
			log.Printf("Found custom config %s using instead of universal\n",hostConfig.Hostname)
			hostAgentConfig = hostConfig
		}
	}
	return hostAgentConfig
}

func configure(spec *Experiment, config DropperConfig, startupDir string) {
	vms := getVms(spec)
	for _, node := range vms {
		log.Printf("\t\tConfiguring Host %s\n", node.General.Hostname)
		agentCfg := getAgentConfig(node.General.Hostname,config)
		ext := ""
		switch node.Hardware.OSType {
		case OSType_Windows:
			ext = "exe"

			var (
				startupFile = startupDir + "/" + node.General.Hostname + "-cc_startup.ps1"
				schedFile   = startupDir + "/" + node.General.Hostname + "-cc_scheduler.cmd"
			)

			a := &Injection{
				Src: startupFile,
				Dst: INJECTPATH + "cc_startup.ps1",
			}
			b := &Injection{
				Src: schedFile,
				Dst: "ProgramData/Microsoft/Windows/Start Menu/Programs/StartUp/CommandAndControl.cmd",
			}

			log.Print(" Windows Injections\n")
			log.Printf("%v\n%v\n", a, b)

			node.Injections = append(node.Injections, a, b)

		case OSType_Linux, OSType_RHEL, OSType_CentOS:
			ext = ""

			var (
				startupFile = startupDir + "/" + node.General.Hostname + "-cc_startup.sh"
				svcFile     = startupDir + "/" + node.General.Hostname + "-cc_startup.service"
				svcLink     = startupDir + "/" + node.General.Hostname + "-cc_startup.serviceLink"
			)
			a := &Injection{
				Src:         startupFile,
				Dst:         INJECTPATH + "cc_startup.sh",
				Description: "",
			}
			if strings.ToLower(agentCfg.ServiceType) == "systemd" {
				b := &Injection{
					Src:         svcFile,
					Dst:         "/etc/systemd/system/CommandAndControl.service",
					Description: "",
				}
				c := &Injection{
					Src:         svcLink,
					//Dst:         "/etc/systemd/system/multi-user.target.wants/CommandAndControl.service",
					Dst:	    "/lib/systemd/system/multi-user.target.wants/CommandAndControl.service",
					Description: "",
				}

				log.Print(" Linux Injections\n")
				log.Printf("%v\n%v\n%v\n", a, b, c)

				node.Injections = append(node.Injections, a, b, c)
			} else if strings.ToLower(agentCfg.ServiceType) == "custom" {
				b := &Injection{
					Src:         startupDir + "/" + node.General.Hostname + "-cc_startup.sh",
					Dst:         agentCfg.CustomService.InjectPath,
					Description: "",
				}

				log.Print(" Linux Injections\n")
				log.Printf("%v\n%v\n", a, b)

				node.Injections = append(node.Injections, a, b)

			} else {
				b := &Injection{
					Src:         svcFile,
					Dst:         "/etc/init.d/CommandAndControl",
					Description: "",
				}
				c := &Injection{
					Src:         svcLink,
					Dst:         "/etc/rc5.d/S99CommandAndControl",
					Description: "",
				}

				log.Print(" Linux Injections\n")
				log.Printf("%v\n%v\n%vi\n", a, b, c)

				node.Injections = append(node.Injections, a, b, c)
			}

		}

		agentSrc, agentDst := agentPath(ext, agentCfg.Agent, agentCfg.AgentPath)
		if len(agentSrc) < 2 {
			erro := fmt.Sprintf("Agent not found when looking for %s.%s in %s", agentCfg.Agent, ext, agentCfg.AgentPath)
			log.Fatal(erro)
		}
		a := &Injection{
			Src:         agentSrc,
			Dst:         agentDst,
			Description: "",
		}
		log.Print(" Agent Injection \n")
		log.Printf("%v\n", a)

		node.Injections = append(node.Injections, a)
	log.Println("")
	}
}

func start(spec *Experiment, config DropperConfig, startupDir string) {
	vms := getVms(spec)
	for _, vm := range vms {
		log.Print("Host: "+ vm.General.Hostname)
		agentCfg := getAgentConfig(vm.General.Hostname,config) 

		switch vm.Hardware.OSType {
		case OSType_Linux, OSType_RHEL, OSType_CentOS:

			file := startupDir + "/" + vm.General.Hostname + "-cc_startup.sh"
			//Set the hostname so machine shows up in CC with proper hostname
			agentCfg.Hostname = vm.General.Hostname
			if err := tmpl.CreateFileFromTemplate("linux_startup.tmpl", agentCfg, file, 0755); err != nil {
				log.Fatal("generating linux command and control startup script: ", err)
			}
			log.Print("Generated file: " + file)
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
			} else if strings.ToLower(agentCfg.ServiceType) == "custom" {
				customStartText := ""
				fileCPath := startupDir + "/" + vm.General.Hostname + "-cc_startup.sh"
				_, existsCheck := os.Stat(agentCfg.CustomService.ScriptPath)
				if (agentCfg.CustomService.ScriptPath != "" && existsCheck == nil) {
					filerc, err := os.Open(agentCfg.CustomService.ScriptPath)
					if err != nil{
						log.Fatal(err)
					}
					defer filerc.Close()
				        buf := new(bytes.Buffer)
					buf.ReadFrom(filerc)
					customStartText += buf.String()
				} else {
					customStartText += "#!/bin/bash\n"
				}
				customStartText += "##### Gernerated mt ccDropper ######\n"
				agentCfg.Hostname = vm.General.Hostname
				//Generate the template into a string
				runCmds := new(bytes.Buffer)
				if err := tmpl.GenerateFromTemplate("linux_startup.tmpl", agentCfg, runCmds); err != nil {
					log.Fatal("generating linux command and control startup script: ", err)
				}
				//Grab all the startup stuff for CC and add it the the users script 
				for i,line := range strings.Split(runCmds.String(),"\n") {
					if i>0 {
						customStartText += line +"\n"
					}
				}
				file, err := os.Create(fileCPath)
				if err != nil {
					log.Printf("Error creating %s %v",fileCPath,err)
				}
				defer file.Close()

				err = ioutil.WriteFile(fileCPath,[]byte(customStartText),0755)
				if err != nil {
					log.Printf("Error writing %s %v",fileCPath,err)
				}
				log.Print("Generated file: " + fileCPath)

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

		case OSType_Windows:
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

func postStart(spec *Experiment) error {
	return nil
}

func cleanup(spec *Experiment) error {
	return nil
}

func main() {
	logFile, err := os.Create("/tmp/ccDropperLog")
	defer logFile.Close()
	log.SetOutput(logFile)
	if err != nil {
		log.Fatal("ccDropper: Can't create log file ... exiting")
	}
	log.Println("Application Start")
	var spec Experiment
	log.Println("Loaded Spec")
	mode := os.Args[1]
	var config DropperConfig
	log.Println("App Performing " +mode+" step")
	err = json.NewDecoder(os.Stdin).Decode(&spec)
	if err != nil {
		log.Fatal(err.Error())
	}
	//log.Printf("%+v\n",spec)
	//Get the application configuration data
	for _, e := range spec.Spec.Scenario.Apps {
		if e.Name == NAME {
			log.Print("Found config")
			//log.Print(e.Metadata["cc_hosts"])
			vEncoding := ""
			for _, hostConfig := range e.Metadata["cc_hosts"].([]interface{}) {
				log.Printf("--------Parsing---------\n%v",hostConfig)
				gg := ""
				for k, v := range hostConfig.(map[string]interface{}) {
					switch v.(type) {
					case float32, float64, int8, int16, int32, int64, uint8, uint16, uint32, uint64, int, uint:
						vEncoding = "%v."
					case bool:
						vEncoding = "%v"
					case string:
						vEncoding = "\"%v\""
					case map[string]interface{}:
						vM := ""
						if k == "service_custom" {
							vM += "\n"
							for kk,vv := range v.(map[string]interface{}) {
								vM += fmt.Sprintf("  %s: \"%v\"\n", kk, vv)
								log.Printf("Encoding %s\n",v)
							}
							vM = strings.TrimSuffix(vM,"\n")
							//vEncoding += "\n"
							v=vM //pass the decoded data into v
							vEncoding = "%s" // pass throught formating above:wq!

						}
					} //end switch
					gg += fmt.Sprintf("%s: "+vEncoding+"\n", k,v)
					//log.Printf("Key: %s:%v formated as %T",k,v,v)
				}
				log.Printf("----------Marshaled Config ---------\n%v",gg)
				cfg := HostAgentConfig{}
				err = yaml.Unmarshal([]byte(gg), &cfg)
				if err != nil {
					log.Fatal(err)
					return
				}
				//log.Printf("----------Unmarshaled Config-----------\n%v", cfg)
				if cfg.Hostname == "*" {
					universalConfig = cfg
					log.Printf("Univeral config found\n")
				} else {
					config.Hosts = append(config.Hosts, cfg)
				}

			}
		}
	}
	//computer the start directory where things will be generated
	startupDir := spec.Spec.BaseDir + "/startup"
	switch mode {
	case "configure":
		log.Print(" ----------------------Configure------------------\n")
		configure(&spec, config, startupDir)
	case "pre-start":
		log.Print(" ----------------------Start------------------\n")
		start(&spec, config, startupDir)
	case "post-start":
		log.Print(" ----------------------Post Start------------------\n")
		postStart(&spec)
	case "cleanup":
		log.Print(" ----------------------Cleanup------------------\n")
		cleanup(&spec)
	}
	data, err := json.Marshal(spec)
	out := bytes.NewBuffer(data)
	if err != nil {
		log.Fatal("Error marshaling experiment spec to JSON: %w", err)
	}
	fmt.Print(out)
	f, err := os.Create("/tmp/ccDropper.out")
	defer f.Close()
	f.Write(data)
	f.Sync()

}
