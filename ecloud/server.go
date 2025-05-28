package ecloud

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/Elemento-Modular-Cloud/tesi-paolobeci/ecloud/schema"
)

// VM representation inside Elemento
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
	Volumes      []*Volume
}

// ServerType represents a server type in the Elemento.
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
	// ArchitectureX86 is the architecture for Intel/AMD x86 CPUs.
	ArchitectureX86 Architecture = "x86"

	// ArchitectureARM is the architecture for ARM CPUs.
	ArchitectureARM Architecture = "arm"
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
	// Call ComputeStatus and get the full status response
	statusResp, err := c.client.ComputeStatus()
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

// ServerListOpts specifies options for listing servers.
type ServerListOpts struct {
	ListOpts
	Name   string
	Status []ServerStatus
	Sort   []string
}

func (l ServerListOpts) values() url.Values {
	vals := l.ListOpts.Values()
	if l.Name != "" {
		vals.Add("name", l.Name)
	}
	for _, status := range l.Status {
		vals.Add("status", string(status))
	}
	for _, sort := range l.Sort {
		vals.Add("sort", sort)
	}
	return vals
}

// AllWithOpts returns all servers for the given options.
func (c *ServerClient) AllWithOpts(ctx context.Context, opts ServerListOpts) ([]*Server, error) {
	allServers := []*Server{}

	err := c.client.all(func(page int) (*Response, error) {
		opts.Page = page
		servers, resp, err := c.List(ctx, opts)
		if err != nil {
			return resp, err
		}
		allServers = append(allServers, servers...)
		return resp, nil
	})
	if err != nil {
		return nil, err
	}

	return allServers, nil
}

// List returns a list of servers for a specific page.
//
// Please note that filters specified in opts are not taken into account
// when their value corresponds to their zero value or when they are empty.
func (c *ServerClient) List(ctx context.Context, opts ServerListOpts) ([]*Server, *Response, error) {
	body, err := c.client.ComputeStatus()
	if err != nil {
		return nil, nil, err
	}

	servers := make([]*Server, 0, len(*body))
	for _, s := range *body {
		servers = append(servers, ServerFromSchema(s))
	}
	return servers, &Response{}, nil
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

// Update updates a server.
func (c *ServerClient) Update(ctx context.Context, server *Server, opts ServerUpdateOpts) (*Server, *Response, error) {
	// reqBody := schema.ServerUpdateRequest{
	// 	Name: opts.Name,
	// }
	// if opts.Labels != nil {
	// 	reqBody.Labels = &opts.Labels
	// }
	// reqBodyData, err := json.Marshal(reqBody)
	// if err != nil {
	// 	return nil, nil, err
	// }

	// path := fmt.Sprintf("/servers/%d", server.ID)
	// req, err := c.client.NewRequest(ctx, "PUT", path, bytes.NewReader(reqBodyData))
	// if err != nil {
	// 	return nil, nil, err
	// }

	// respBody := schema.ServerUpdateResponse{}
	// resp, err := c.client.Do(req, &respBody)
	// if err != nil {
	// 	return nil, resp, err
	// }
	// return ServerFromSchema(respBody.Server), resp, nil
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
	ServerType       *ServerType
	SSHKeys          []*SSHKey
	Datacenter       *Datacenter
	UserData         string
	StartAfterCreate *bool
	Labels           map[string]string
	Automount        *bool
	Volumes          []*Volume
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

	// Prepare the request body according to the schema
	reqBody := schema.CreateComputeRequest{
		Name:       opts.Name,
		UserData:   opts.UserData,
		Labels:     &opts.Labels,
		Datacenter: opts.Datacenter.Name,
	}

	// Add server type configuration
	if opts.ServerType != nil {
		reqBody.Slots = opts.ServerType.Cores
		reqBody.Ramsize = int(opts.ServerType.Memory * 1024) // Convert GB to MB
		reqBody.Archs = []string{string(opts.ServerType.Architecture)}
	}

	// Add SSH keys if provided
	if len(opts.SSHKeys) > 0 {
		reqBody.SSHKeys = make([]int, len(opts.SSHKeys))
		for i, sshKey := range opts.SSHKeys {
			reqBody.SSHKeys[i] = sshKey.ID
		}
	}

	// Add volumes if provided
	if len(opts.Volumes) > 0 {
		reqBody.Volumes = make([]map[string]string, len(opts.Volumes))
		for i, volume := range opts.Volumes {
			reqBody.Volumes[i] = map[string]string{
				"vid": strconv.Itoa(volume.ID),
			}
		}
	}

	// Add networks if provided
	if len(opts.Networks) > 0 {
		reqBody.Networks = make([]int, len(opts.Networks))
		for i, network := range opts.Networks {
			networkID, err := strconv.Atoi(network.ID)
			if err != nil {
				return ServerCreateResult{}, nil, fmt.Errorf("invalid network ID format: %w", err)
			}
			reqBody.Networks[i] = networkID
		}
	}

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
	if o.ServerType == nil || (o.ServerType.ID == 0 && o.ServerType.Name == "") {
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
