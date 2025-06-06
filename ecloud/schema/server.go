package schema

import (
	"time"
)

// -------- HEALTH --------

// HealthCheckComputeResponse represents the response from the compute health check endpoint.
// The response is a plain string, not a JSON object.
type HealthCheckComputeResponse string

// -------- CAN ALLOCATE --------
type Misc struct {
	OsFamily  string `json:"os_family"`
	OsFlavour string `json:"os_flavour"`
}

type CanAllocateComputeRequest struct {
	Slots         int      `json:"slots"`
	Overprovision int      `json:"overprovision"`
	AllowSMT      bool     `json:"allowSMT"`
	Archs         []string `json:"archs"`
	Flags         []string `json:"flags"`
	Ramsize       int      `json:"ramsize"` // MB
	ReqECC        bool     `json:"reqECC"`
	Misc          Misc     `json:"misc"`
	Pci           []string `json:"pci"`
}

type Price struct {
	Hour  float64 `json:"hour"`
	Month float64 `json:"month"`
	Unit  string  `json:"unit"`
}

type ProviderInfo struct {
	Price    Price  `json:"price"`
	Provider string `json:"provider"`
	Region   string `json:"region"`
}

type CanAllocateComputeResponse struct {
	Mesos []ProviderInfo `json:"mesos"`
}

// -------- CREATE COMPUTE --------
type CreateComputeRequest struct {
	// elemento fields
	Name          string              `json:"vm_name"`
	Slots         int                 `json:"slots"`
	Overprovision int                 `json:"overprovision"`
	AllowSMT      bool                `json:"allowSMT"`
	Archs         []string            `json:"archs"`
	Flags         []string            `json:"flags"`
	Ramsize       int                 `json:"ramsize"`
	ReqECC        bool                `json:"reqECC"`
	Misc          Misc                `json:"misc"`
	Pci           []string            `json:"pci"`
	Volumes       []map[string]string `json:"volumes"`
	Netdevs       []string            `json:"netdevs"`

	// kOps required
	UserData   string             `json:"user_data,omitempty"`
	Labels     *map[string]string `json:"labels,omitempty"`
	SSHKeys    []int              `json:"ssh_keys,omitempty"`
	Datacenter string             `json:"datacenter,omitempty"`
	Networks   []int              `json:"networks,omitempty"`
}

type CreateComputeResponse struct {
	Server       Server  `json:"server"`
	RootPassword *string `json:"root_password"`
}

// -------- COMPUTE STATUS --------
type ComputeStatusResponse []Server

type Server struct {
	UniqueID      string            `json:"uniqueID"`
	Name          string            `json:"name"`
	Status        string            `json:"status"`
	Created       time.Time         `json:"created"`
	NetworkConfig NetworkConfig     `json:"network_config"`
	ServerURL     string            `json:"serverurl"`
	IsGateway     bool              `json:"is_gateway"`
	ReqJSON       RequestConfig     `json:"req_json"`
	CreationDate  time.Time         `json:"creation_date"`
	Labels        map[string]string `json:"labels"`
	Volumes       []Volume          `json:"volumes"`
}

type RequestConfig struct {
	Slots         int      `json:"slots"`
	Overprovision int      `json:"overprovision"`
	AllowSMT      bool     `json:"allowSMT"`
	Arch          string   `json:"arch"`
	Flags         []string `json:"flags"`
	RamSize       float64  `json:"ramsize"`
	ReqECC        bool     `json:"reqECC"`
	Volumes       []Volume `json:"volumes"`
	PciDevs       []string `json:"pcidevs"`
	NetDevs       []string `json:"netdevs"`
	OSFamily      string   `json:"os_family"`
	OSFlavour     string   `json:"os_flavour"`
	VMName        string   `json:"vm_name"`
}

type Volume struct {
	Bootable       bool     `json:"bootable"`
	CreatorID      string   `json:"creatorID"`
	Name           string   `json:"name"`
	NumServers     int      `json:"nservers"`
	Own            bool     `json:"own"`
	Private        bool     `json:"private"`
	ReadOnly       bool     `json:"readonly"`
	Server         string   `json:"server"`
	Servers        []string `json:"servers"`
	ServerURL      string   `json:"serverurl"`
	Shareable      bool     `json:"shareable"`
	Size           int64    `json:"size"`
	VolumeID       string   `json:"volumeID"`
	Vid            string   `json:"vid"`
	SelectedServer string   `json:"selected_server"`
	ISCSIName      string   `json:"iscsi_name"`
	Driver         string   `json:"driver"`
}

type NetworkConfig struct {
	Name       string         `json:"name"`
	Interface  string         `json:"interface"`
	Type       string         `json:"type"`
	Source     string         `json:"source"`
	Model      string         `json:"model"`
	MAC        string         `json:"mac"`
	DomDisplay NetworkDisplay `json:"dom_display"`
	IPv4       *string        `json:"ipv4"` // Assuming IPv4 could be null
}

type NetworkDisplay struct {
	Protocol string `json:"protocol"`
	Port     int    `json:"port"`
}

// -------- COMPUTE TEMPLATES --------
type ComputeTemplatesResponse []ComputeTemplate

type ComputeTemplate struct {
	Info struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	} `json:"info"`
	CPU struct {
		Slots         int      `json:"slots"`
		Overprovision int      `json:"overprovision"`
		AllowSMT      bool     `json:"allowSMT"`
		Archs         []string `json:"archs"`
		Flags         []string `json:"flags"`
	} `json:"cpu"`
	RAM struct {
		Ramsize int  `json:"ramsize"`
		ReqECC  bool `json:"reqECC"`
	} `json:"ram"`
}

// -------- COMPUTE DELETE --------
type DeleteComputeRequest struct {
	LocalIndex string `json:"local_index"`
}
type DeleteComputeResponse struct{}

// -------- STORAGE --------
