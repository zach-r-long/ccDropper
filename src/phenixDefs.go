package main


type Host struct {
	Hostname string                 `json:"hostname" yaml:"hostname"`
	Metadata map[string]interface{} `json:"metadata" yaml:"metadata"`
}

type App struct {
	Name     string                 `json:"name" yaml:"name"`
	AssetDir string                 `json:"assetDir" yaml:"assetDir" structs:"assetDir" mapstructure:"assetDir"`
	Metadata map[string]interface{} `json:"metadata" yaml:"metadata"`
	Hosts    []Host `json:"hosts" yaml:"hosts"`
}


type ScenarioSpec struct {
	Apps []App `json:"apps" yaml:"apps"`
}

const (
	VMType_NotSet    VMType = ""
	VMType_KVM       VMType = "kvm"
	VMType_Container VMType = "container"
)

type CPU string

const (
	CPU_NotSet    CPU = ""
	CPU_Broadwell CPU = "Broadwell"
	CPU_Haswell   CPU = "Haswell"
	CPU_Core2Duo  CPU = "core2duo"
	CPU_Pentium3  CPU = "pentium3"
)

type OSType string

const (
	OSType_NotSet  OSType = ""
	OSType_Windows OSType = "windows"
	OSType_Linux   OSType = "linux"
	OSType_RHEL    OSType = "rhel"
	OSType_CentOS  OSType = "centos"
)

type Labels map[string]string

type VMType string

type General struct {
	Hostname    string `json:"hostname" yaml:"hostname"`
	Description string `json:"description" yaml:"description"`
	VMType      VMType `json:"vm_type" yaml:"vm_type" mapstructure:"vm_type"`
	Snapshot    *bool  `json:"snapshot" yaml:"snapshot"`
	DoNotBoot   *bool  `json:"do_not_boot" yaml:"do_not_boot" structs:"do_not_boot" mapstructure:"do_not_boot"`
}

type Hardware struct {
	CPU    CPU     `json:"cpu" yaml:"cpu" structs:"cpu" mapstructure:"cpu"`
	VCPU   int     `json:"vcpus" yaml:"vcpus" structs:"vcpus" mapstructure:"vcpus"`
	Memory int     `json:"memory" yaml:"memory"`
	OSType OSType  `json:"os_type" yaml:"os_type" mapstructure:"os_type"`
	Drives []Drive `json:"drives" yaml:"drives"`
}

type Drive struct {
	Image           string `json:"image" yaml:"image"`
	Interface       string `json:"interface" yaml:"interface"`
	CacheMode       string `json:"cache_mode" yaml:"cache_mode" structs:"cache_mode" mapstructure:"cache_mode"`
	InjectPartition *int   `json:"inject_partition" yaml:"inject_partition" mapstructure:"inject_partition"`
}

type Injection struct {
	Src         string `json:"src" yaml:"src"`
	Dst         string `json:"dst" yaml:"dst"`
	Description string `json:"description" yaml:"description"`
	Permissions string `json:"permissions" yaml:"permissions"`
}
type Interface struct {
	Name       string `json:"name" yaml:"name"`
	Type       string `json:"type" yaml:"type"`
	Proto      string `json:"proto" yaml:"proto"`
	UDPPort    int    `json:"udp_port" yaml:"udp_port" mapstructure:"udp_port"`
	BaudRate   int    `json:"baud_rate" yaml:"baud_rate" mapstructure:"baud_rate"`
	Device     string `json:"device" yaml:"device"`
	VLAN       string `json:"vlan" yaml:"vlan"`
	Bridge     string `json:"bridge" yaml:"bridge"`
	Autostart  bool   `json:"autostart" yaml:"autostart"`
	MAC        string `json:"mac" yaml:"mac"`
	MTU        int    `json:"mtu" yaml:"mtu"`
	Address    string `json:"address" yaml:"address"`
	Mask       int    `json:"mask" yaml:"mask"`
	Gateway    string `json:"gateway" yaml:"gateway"`
	RulesetIn  string `json:"ruleset_in" yaml:"ruleset_in" mapstructure:"ruleset_in"`
	RulesetOut string `json:"ruleset_out" yaml:"ruleset_out" mapstructure:"ruleset_out"`
}

type Route struct {
	Destination string `json:"destination" yaml:"destination"`
	Next        string `json:"next" yaml:"next"`
	Cost        *int   `json:"cost" yaml:"cost"`
}

type OSPF struct {
	RouterID               string `json:"router_id" yaml:"router_id" mapstructure:"router_id"`
	Areas                  []Area `json:"areas" yaml:"areas" mapstructure:"areas"`
	DeadInterval           *int   `json:"dead_interval" yaml:"dead_interval" mapstructure:"dead_interval"`
	HelloInterval          *int   `json:"hello_interval" yaml:"hello_interval" mapstructure:"hello_interval"`
	RetransmissionInterval *int   `json:"retransmission_interval" yaml:"retransmission_interval" mapstructure:"retransmission_interval"`
}

type Area struct {
	AreaID       *int          `json:"area_id" yaml:"area_id" mapstructure:"area_id"`
	AreaNetworks []AreaNetwork `json:"area_networks" yaml:"area_networks" mapstructure:"area_networks"`
}

type AreaNetwork struct {
	Network string `json:"network" yaml:"network" mapstructure:"network"`
}

type Ruleset struct {
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description" yaml:"description"`
	Default     string `json:"default" yaml:"default"`
	Rules       []Rule `json:"rules" yaml:"rules"`
}

type Rule struct {
	ID          int       `json:"id" yaml:"id"`
	Description string    `json:"description" yaml:"description"`
	Action      string    `json:"action" yaml:"action"`
	Protocol    string    `json:"protocol" yaml:"protocol"`
	Source      *AddrPort `json:"source" yaml:"source"`
	Destination *AddrPort `json:"destination" yaml:"destination"`
}

type AddrPort struct {
	Address string `json:"address" yaml:"address"`
	Port    int    `json:"port" yaml:"port"`
}

type Network struct {
	Interfaces []Interface `json:"interfaces" yaml:"interfaces"`
	Routes     []Route     `json:"routes" yaml:"routes"`
	OSPF       *OSPF       `json:"ospf" yaml:"ospf" mapstructure:"ospf"`
	Rulesets   []Ruleset   `json:"rulesets" yaml:"rulesets"`
}


type Node struct {
	Labels     Labels       `json:"labels" yaml:"labels"`
	Type       string       `json:"type" yaml:"type"`
	General    General      `json:"general" yaml:"general"`
	Hardware   Hardware     `json:"hardware" yaml:"hardware"`
	Network    Network      `json:"network" yaml:"network"`
	Injections []*Injection `json:"injections" yaml:"injections"`
}

type TopologySpec struct {
	Nodes []*Node `json:"nodes" yaml:"nodes"`
}

type Schedule map[string]string
type VLANAliases map[string]int

type VLANSpec struct {
	Aliases VLANAliases `json:"aliases" yaml:"aliases" structs:"aliases" mapstructure:"aliases"`
	Min     int         `json:"min" yaml:"min" structs:"min" mapstructure:"min"`
	Max     int         `json:"max" yaml:"max" structs:"max" mapstructure:"max"`
}

type ExperimentSpec struct {
	ExperimentName string        `json:"experimentName" yaml:"experimentName" structs:"experimentName"`
	BaseDir        string        `json:"baseDir" yaml:"baseDir" structs:"baseDir"`
	Topology       *TopologySpec `json:"topology" yaml:"topology"`
	Scenario       *ScenarioSpec `json:"scenario" yaml:"scenario"`
	VLANs          *VLANSpec     `json:"vlans" yaml:"vlans" structs:"vlans" mapstructure:"vlans"`
	Schedules      Schedule      `json:"schedules" yaml:"schedules"`
	RunLocal       bool          `json:"runLocal" yaml:"runLocal" structs:"runLocal"`
}

type ExperimentStatus struct {
	StartTime string                 `json:"startTime" yaml:"startTime" structs:"startTime" mapstructure:"startTime"`
	Schedules Schedule               `json:"schedules" yaml:"schedules"`
	Apps      map[string]interface{} `json:"apps" yaml:"apps"`
	VLANs     VLANAliases            `json:"vlans" yaml:"vlans" structs:"vlans" mapstructure:"vlans"`
}

type Annotations map[string]string

type ConfigMetadata struct {
	Name        string      `json:"name" yaml:"name"`
	Created     string      `json:"created" yaml:"created"`
	Updated     string      `json:"updated" yaml:"updated"`
	Annotations Annotations `json:"annotations,omitempty" yaml:"annotations,omitempty"`
}

type Experiment struct {
	Metadata ConfigMetadata       `json:"metadata" yaml:"metadata"` // experiment configuration metadata
	Spec     *ExperimentSpec   `json:"spec" yaml:"spec"`         // reference to latest versioned experiment spec
	Status   *ExperimentStatus `json:"status" yaml:"status"`     // reference to latest versioned experiment status
}


// FindNodesWithLabels finds all nodes in the topology containing at least one
// of the labels provided. Take note that the node does not have to have all the
// labels provided, just one.
func (this TopologySpec) FindNodesWithLabels(labels ...string) []*Node {
	var nodes []*Node

	for _, n := range this.Nodes {
		for _, l := range labels {
			if _, ok := n.Labels[l]; ok {
				nodes = append(nodes, n)
				break
			}
		}
	}

	return nodes
}
