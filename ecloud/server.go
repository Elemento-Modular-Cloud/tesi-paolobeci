package ecloud

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"os"

	"github.com/Elemento-Modular-Cloud/tesi-paolobeci/ecloud/schema"
)

type Server struct {
	ID           string
	Name         string
	Status       ServerStatus
	Created      time.Time
	PublicNet    ServerPublicNet
	ServerType   *ServerType
	Datacenter   Datacenter
	BackupWindow string
	Labels       map[string]string
	Volumes      []*schema.StorageVolume
}

type ServerType struct {
	ID           int
	Name         string
	Description  string
	Cores        int
	Memory       float32
	Disk         int
	Architecture Architecture
}

// Architecture specifies the architecture of the CPU.
type Architecture string

const (
	// ArchitectureX86_32 is the architecture for Intel/AMD x86 32 bit CPUs.
	ArchitectureX86_32 Architecture = "X86_32"

	// ArchitectureX86_64 is the architecture for Intel/AMD x86 64 bit CPUs.
	ArchitectureX86_64 Architecture = "X86_64"

	// ArchitectureARM_7 is the architecture for ARM 7 CPUs.
	ArchitectureARM_7 Architecture = "ARM_7"

	// ArchitectureARM_8 is the architecture for ARM 8 CPUs.
	ArchitectureARM_8 Architecture = "ARM_8"

	// ArchitecturePPC_64 is the architecture for PowerPC 64 bit CPUs.
	ArchitecturePPC_64 Architecture = "PPC_64"

	// ArchitectureS390X is the architecture for IBM S390 32 bit CPUs.
	ArchitectureS390X Architecture = "S390X"
)

// ServerStatus specifies a server's status.
type ServerStatus string

const (
	// ServerStatusInitializing is the status when a server is initializing.
	ServerStatusInitializing ServerStatus = "initializing"

	// ServerStatusOff is the status when a server is off.
	ServerStatusOff ServerStatus = "off"

	// ServerStatusRunning is the status when a server is running.
	ServerStatusRunning ServerStatus = "running"

	// ServerStatusStarting is the status when a server is being started.
	ServerStatusStarting ServerStatus = "starting"

	// ServerStatusStopping is the status when a server is being stopped.
	ServerStatusStopping ServerStatus = "stopping"

	// ServerStatusMigrating is the status when a server is being migrated.
	ServerStatusMigrating ServerStatus = "migrating"

	// ServerStatusRebuilding is the status when a server is being rebuilt.
	ServerStatusRebuilding ServerStatus = "rebuilding"

	// ServerStatusDeleting is the status when a server is being deleted.
	ServerStatusDeleting ServerStatus = "deleting"

	// ServerStatusUnknown is the status when a server's state is unknown.
	ServerStatusUnknown ServerStatus = "unknown"
)

// ServerPublicNet represents a server's public network.
type ServerPublicNet struct {
	IPv4        string
	IPv6        string
	FloatingIPs []*string
	Firewalls   []*string
}

// ServerClient is a client for the servers API.
type ServerClient struct {
	client *Client
}

// GetByID retrieves a server by its ID. If the server does not exist, nil is returned.
func (c *ServerClient) GetByID(ctx context.Context, id string) (*schema.Server, error) {
	// Call GetCompute and get the full status response
	statusResp, err := c.client.GetCompute()
	if err != nil {
		return nil, err
	}

	// Search for the server with the matching ID
	for _, server := range *statusResp {
		if server.UniqueID == id {
			return &server, nil
		}
	}

	// Server not found
	return nil, nil
}

// GetByName retrieves a server by its name. If the server does not exist, nil is returned.
func (c *ServerClient) GetByName(ctx context.Context, name string) (*Server, *Response, error) {
	if name == "" {
		return nil, nil, nil
	}
	servers, response, err := c.List(ctx, ServerListOpts{Name: name})
	if len(servers) == 0 {
		return nil, response, err
	}
	return servers[0], response, err
}

// ServerListOpts specifies options for listing servers.
type ServerListOpts struct {
	ListOpts
	Name   string
	Status []ServerStatus
}

// List returns a list of servers.
func (c *ServerClient) List(ctx context.Context, opts ServerListOpts) ([]*Server, *Response, error) {
	body, err := c.client.GetCompute()
	if err != nil {
		return nil, nil, err
	}

	servers := make([]*Server, 0, len(*body))
	for _, s := range *body {
		server := ServerFromSchema(s)

		// Filter by name if specified
		if opts.Name != "" && server.Name != opts.Name {
			continue
		}

		// Filter by status if specified
		if len(opts.Status) > 0 {
			statusMatch := false
			for _, status := range opts.Status {
				if server.Status == status {
					statusMatch = true
					break
				}
			}
			if !statusMatch {
				continue
			}
		}

		servers = append(servers, server)
	}
	return servers, &Response{}, nil
}

// Update updates a server.
func (c *ServerClient) Update(ctx context.Context, server *Server, opts ServerUpdateOpts) (*Server, *Response, error) {
	return nil, nil, fmt.Errorf("server update operation is not yet supported")
}

// ServerUpdateOpts specifies options for updating a server.
type ServerUpdateOpts struct {
	Name   string
	Labels map[string]string
}

// ServerCreateOpts specifies options for creating a new server.
type ServerCreateOpts struct {
	Name             string
	ServerType       *ServerType // Size name (e.g., "neon", "argon", "kripton")
	Image            string      // Image name (e.g., "ubuntu-24-04", "debian-12")
	SSHKeys          []*SSHKey
	Datacenter       *Datacenter
	UserData         string
	StartAfterCreate *bool
	Labels           map[string]string
	Automount        *bool
	Volumes          []*schema.StorageVolume
	Networks         []*Network
}

// Create creates a new server.
func (c *ServerClient) Create(ctx context.Context, opts ServerCreateOpts) (ServerCreateResult, *Response, error) {
	if err := opts.Validate(); err != nil {
		return ServerCreateResult{}, nil, err
	}

	// Prepare the can allocate request according to the schema
	reqBodyCanAllocate := schema.CanAllocateComputeRequest{
		// TODO
	}

	// First check if we can allocate the compute instance
	_, err := c.client.CanAllocateCompute(reqBodyCanAllocate)
	if err != nil {
		return ServerCreateResult{}, nil, fmt.Errorf("the config provided cannot be allocated: %w", err)
	}

	// Parse image to OS family and flavor
	osFamily, osFlavour := parseImageToOS(opts.Image)

	// Prepare the request body according to the schema
	reqBody := schema.CreateComputeRequest{
		Info:    schema.Info{Name: opts.Name},
		Flags:   []string{"sse2"},
		Misc:    schema.Misc{OsFamily: osFamily, OsFlavour: osFlavour},
		Pci:     []string{},
		Volumes: []map[string]string{},
		HasNetwork: true,
		Networks: []map[string]string{},
	}

	// Add server type configuration
	if opts.ServerType != nil {
		// Check if ServerType has zero cores/memory but a valid size name
		if opts.ServerType.Cores == 0 && opts.ServerType.Memory == 0 && opts.ServerType.Name != "" {
			fmt.Printf("DEBUG: ServerType has zero cores/memory but name '%s', attempting size conversion\n", opts.ServerType.Name)
			// Try to convert the ServerType name to a size configuration
			sizeConfig, err := ConvertServerSize(opts.ServerType.Name)
			if err == nil {
				fmt.Printf("DEBUG: Successfully converted ServerType name '%s' to size config - Slots: %d, Ramsize: %d MB\n", opts.ServerType.Name, sizeConfig.Slots, sizeConfig.Ramsize)
				reqBody.Slots = sizeConfig.Slots
				reqBody.Overprovision = sizeConfig.Slots
				reqBody.Ramsize = sizeConfig.Ramsize
				// Use the ServerType's architecture if specified, otherwise default to x86_64
				if opts.ServerType.Architecture != "" {
					reqBody.Archs = []string{string(opts.ServerType.Architecture)}
				} else {
					reqBody.Archs = []string{string(ArchitectureX86_64)}
				}
			} else {
				fmt.Printf("DEBUG: Failed to convert ServerType name '%s' to size config: %v\n", opts.ServerType.Name, err)
				// Fall back to using the ServerType as-is (will result in 0 slots/ramsize)
				reqBody.Slots = opts.ServerType.Cores
				reqBody.Overprovision = opts.ServerType.Cores
				reqBody.Ramsize = int(opts.ServerType.Memory * 1024) // Convert GB to MB
				reqBody.Archs = []string{string(opts.ServerType.Architecture)}
			}
		} else {
			// Use ServerType as-is
			reqBody.Slots = opts.ServerType.Cores
			reqBody.Overprovision = opts.ServerType.Cores
			reqBody.Ramsize = int(opts.ServerType.Memory * 1024) // Convert GB to MB
			reqBody.Archs = []string{string(opts.ServerType.Architecture)}
		}
	}

	// Add default boot volume to the vm
	// TODO: configure cloudinit con default SSH key
	sshKeyStrings := make([]string, len(opts.SSHKeys))
	for i, k := range opts.SSHKeys {
		sshKeyStrings[i] = k.PublicKey
	}
	bootvolumeIDs, err := createBootVolume(ctx, c.client, opts.Name, osFlavour, sshKeyStrings, opts.UserData)
	if err != nil {
		return ServerCreateResult{}, nil, fmt.Errorf("failed to create boot volume: %w", err)
	}
	for _, vid := range bootvolumeIDs {
		reqBody.Volumes = append(reqBody.Volumes, map[string]string{"vid": vid})
	}

	// Add volumes if asked
	// TODO: should i add a default boot volume even if not specified by kops?
	if opts.ServerType.Disk > 0 {
		volumeID, err := createVolume(ctx, c.client, opts.Name, opts.ServerType.Disk)
		if err != nil {
			return ServerCreateResult{}, nil, fmt.Errorf("failed to create volume: %w", err)
		}
		reqBody.Volumes = append(reqBody.Volumes, map[string]string{"vid": volumeID})
	}

	// TODO: Add networks if provided
	if len(opts.Networks) > 0 {
		reqBody.Networks = make([]map[string]string, len(opts.Networks))
		for i, network := range opts.Networks {
			reqBody.Networks[i] = map[string]string{"network_uid": network.ID}
		}
	}

	// Wait 15 seconds to allow the volumes to be fully initialized
	time.Sleep(15 * time.Second)

	// Create the compute instance
	resp, err := c.client.CreateCompute(reqBody)
	if err != nil {
		return ServerCreateResult{}, nil, fmt.Errorf("failed to create compute instance: %w", err)
	}

	result := ServerCreateResult{
		Server: ServerFromSchema(resp.Server),
	}
	if resp.RootPassword != nil {
		result.RootPassword = *resp.RootPassword
	}
	return result, &Response{}, nil
}

// Validate checks if options are valid.
func (o ServerCreateOpts) Validate() error {
	if o.Name == "" {
		return errors.New("missing name")
	}
	if o.ServerType == nil {
		return errors.New("missing server type")
	}
	if o.Datacenter == nil {
		return errors.New("missing datacenter")
	}
	return nil
}

// ServerCreateResult is the result of a create server call.
type ServerCreateResult struct {
	Server       *Server
	RootPassword string
}

// Deletes a server
func (c *ServerClient) Delete(ctx context.Context, server *Server) (*schema.DeleteComputeResponse, error) {
	reqBody := schema.DeleteComputeRequest{
		VolumeID: server.ID,
	}
	resp, err := c.client.DeleteCompute(reqBody)
	return resp, err
}

// ServerSizeConfig represents the configuration for a server size
type ServerSizeConfig struct {
	Slots   int // vCPUs
	Ramsize int // RAM in MB
}

// ConvertServerSize converts a server size falvour name to its corresponding
// slots and ramsize configuration
//
// Supported sizes:
//   - "helium": 1 vCPU, 0.5 GB RAM (512 MB)
//   - "neon": 2 vCPUs, 2 GB RAM (2048 MB)
//   - "argon2": 4 vCPUs, 4 GB RAM (4096 MB)
//   - "argon": 6 vCPUs, 4 GB RAM (4096 MB)
//   - "kripton": 8 vCPUs, 8 GB RAM (8192 MB)
//
// The size parameter is case-insensitive and will be trimmed of whitespace.
// Returns an error if the size is not supported.
func ConvertServerSize(size string) (*ServerSizeConfig, error) {
	// Normalize the size string to lowercase for case-insensitive matching
	normalizedSize := strings.ToLower(strings.TrimSpace(size))

	switch normalizedSize {
	case "helium":
		return &ServerSizeConfig{
			Slots:   1,   // 1 vCPU
			Ramsize: 512, // 0.5 GB = 512 MB
		}, nil
	case "neon":
		return &ServerSizeConfig{
			Slots:   2,    // 2 vCPUs
			Ramsize: 2048, // 2 GB = 2048 MB
		}, nil
	case "argon2":
		return &ServerSizeConfig{
			Slots:   4,    // 4 vCPUs
			Ramsize: 4096, // 4 GB = 4096 MB
		}, nil
	case "argon":
		return &ServerSizeConfig{
			Slots:   6,    // 6 vCPUs
			Ramsize: 4096, // 4 GB = 4096 MB
		}, nil
	case "kripton":
		return &ServerSizeConfig{
			Slots:   8,    // 8 vCPUs
			Ramsize: 8192, // 8 GB = 8192 MB
		}, nil
	default:
		return nil, fmt.Errorf("unsupported server size: %s. Supported sizes are: helium, neon, argon2, argon, kripton", size)
	}
}

// parses an image name and returns Elemento OS family and flavor
func parseImageToOS(image string) (string, string) {
	normalizedImage := strings.ToLower(strings.TrimSpace(image))

	// Ubuntu variants
	if strings.Contains(normalizedImage, "ubuntu") {
		return "linux", "ubuntu"
	}

	// Debian variants
	if strings.Contains(normalizedImage, "debian") {
		return "linux", "debian"
	}

	// CentOS/RHEL variants
	if strings.Contains(normalizedImage, "centos") || strings.Contains(normalizedImage, "rhel") || strings.Contains(normalizedImage, "redhat") {
		return "linux", "centos"
	}

	// Fedora
	if strings.Contains(normalizedImage, "fedora") {
		return "linux", "fedora"
	}

	// Alpine
	if strings.Contains(normalizedImage, "alpine") {
		return "linux", "alpine"
	}

	// Windows variants
	if strings.Contains(normalizedImage, "windows") {
		return "windows", "windows"
	}

	// Default fallback
	return "linux", "ubuntu"
}

// Creates a volume to provide into the vm in the creation phase
func createVolume(ctx context.Context, client *Client, serverName string, diskSizeGB int) (string, error) {
	volumeClient := &VolumeClient{client: client}

	// Create volume options
	volumeOpts := VolumeCreateOpts{
		Name:      serverName,
		Size:      diskSizeGB,
		Bootable:  true,
		Readonly:  false,
		Shareable: false,
		Private:   true,
	}

	// Create the volume
	volumeID, _, err := volumeClient.Create(ctx, volumeOpts)
	if err != nil {
		return "", fmt.Errorf("failed to create volume: %w", err)
	}

	return volumeID, nil
}

// Creates the default boot volume with the image requested, returns the volumeID of:
// - boot volume with the image of the requested OS
// - volume containing the cloudinit
func createBootVolume(ctx context.Context, client *Client, serverName string, osFlavour string, sshKey []string, userData string) ([]string, error) {
	volumeClient := &VolumeClient{client: client}
	volumeIDs := []string{}

	// Create boot volume with specified image
	//! TODO: only ubuntu image supported for now, add more os and version support in the future
	volumeOpts := VolumeCreateOpts{
		Name: fmt.Sprintf("%s-boot", serverName),
		Size: 50,
		Url:  "https://cloud-images.ubuntu.com/jammy/current/jammy-server-cloudimg-amd64.img",
	}

	volumeIDboot, _, err := volumeClient.Create(ctx, volumeOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create volume: %w", err)
	}
	volumeIDs = append(volumeIDs, volumeIDboot)

	// Create volume with cloud-init
	// TODO: set custom ssh-key to user-data config file, now it is always Paolo's SSH key
	// TODO: insert kOps scipt on variable userData into user-data cloudinit file
	// filePath := "cloud-config/user-data"
	// input, err := os.ReadFile(filePath)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to read user-data: %w", err)
	// }

	// lines := strings.Split(string(input), "\n")
	// for i, line := range lines {
	// 	if strings.TrimSpace(line) == "data" {
	// 		lines[i] = userData
	// 	}
	// }

	// output := strings.Join(lines, "\n")
	// err = os.WriteFile(filePath, []byte(output), 0644)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to write user-data: %w", err)
	// }

	// CloudInit volume creation
	cloudinitOpts := CloudInitCreateOpts{
		Name: fmt.Sprintf("%s-cloudinit", serverName),
	}
	volumeIDcloudinit, _, err := volumeClient.CreateCloudInit(ctx, cloudinitOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create cloud-init volume: %w", err)
	}
	volumeIDs = append(volumeIDs, volumeIDcloudinit)

	// Wait 5 seconds to allow the volume to be fully initialized
	time.Sleep(5 * time.Second)

	// Feed other file inside cloud-init volume
	_, _, err = volumeClient.FeedFileIntoCloudInitStorage(ctx, volumeIDcloudinit)
	if err != nil {
		return nil, err
	}

	return volumeIDs, nil
}

func SaveCloudInitToFile(userData string, fileName string) error {
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(userData)
	if err != nil {
		return err
	}
	return nil
}
