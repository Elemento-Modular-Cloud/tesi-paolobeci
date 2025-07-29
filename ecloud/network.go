package ecloud

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/Elemento-Modular-Cloud/tesi-paolobeci/ecloud/schema"
)

// NetworkZone specifies a network zone.
type NetworkZone string

// List of available Network Zones.
const (
	NetworkZoneEUCentral NetworkZone = "eu-south"
	// ...
)

// NetworkSubnetType specifies a type of a subnet.
type NetworkSubnetType string

// List of available network subnet types.
const (
	NetworkSubnetTypeCloud NetworkSubnetType = "cloud"
	// ...
)

// Network represents a network in Elemento Cloud.
type Network struct {
	ID         string
	Name       string
	Created    time.Time
	IPRange    *net.IPNet
	Subnets    []NetworkSubnet
	Routes     string
	Servers    []*Server
	Protection bool
	Labels     map[string]string
}

// NetworkSubnet represents a subnet of a network in the Elemento Cloud.
type NetworkSubnet struct {
	Type        NetworkSubnetType
	IPRange     *net.IPNet
	NetworkZone NetworkZone
	Gateway     net.IP
	VSwitchID   int
}

// NetworkClient is a client for the network API.
type NetworkClient struct {
	client *Client
}

// GetByID retrieves a network by its ID. If the network does not exist, nil is returned.
func (c *NetworkClient) GetByID(ctx context.Context, uuid string) (*Network, *schema.GetNetworkByIDResponse, error) {
	var body schema.GetNetworkByIDRequest
	body.NetworkID = uuid
	resp, err := c.client.GetNetworkByID(body)
	if err != nil {
		if IsError(err, ErrorCodeNotFound) {
			return nil, resp, nil
		}
		return nil, nil, err
	}
	return NetworkFromSchema(resp.Network), resp, nil
}

// GetByName retrieves a network by its name. If the network does not exist, nil is returned.
func (c *NetworkClient) GetByName(ctx context.Context, name string) (*Network, *schema.ListNetworkResponse, error) {
	if name == "" {
		return nil, nil, nil
	}
	Networks, response, err := c.List(ctx, NetworkListOpts{Name: name})
	if len(Networks) == 0 {
		return nil, response, err
	}
	return Networks[0], response, err
}

// NetworkListOpts specifies options for listing networks.
type NetworkListOpts struct {
	ListOpts
	Name string
	Sort []string
}

// List returns a list of networks.
func (c *NetworkClient) List(ctx context.Context, opts NetworkListOpts) ([]*Network, *schema.ListNetworkResponse, error) {
	resp, err := c.client.ListNetwork()
	if err != nil {
		return nil, nil, err
	}
	Networks := make([]*Network, 0, len(*resp))
	for _, s := range *resp {
		if opts.Name != "" && s.Name != opts.Name {
			continue
		}
		Networks = append(Networks, NetworkFromSchema(s))
	}
	return Networks, resp, nil
}

// Delete deletes a network.
func (c *NetworkClient) Delete(ctx context.Context, uuid string) (*Response, *schema.DeleteNetworkResponse, error) {
	var body schema.DeleteNetworkRequest
	body.NetworkID = uuid
	resp, err := c.client.DeleteNetwork(body)
	if err != nil {
		return nil, nil, err
	}
	return &Response{}, resp, nil
}

// NetworkCreateOpts specifies options for creating a new network.
type NetworkCreateOpts struct {
	Name    string
	IPRange *net.IPNet
	Subnets []NetworkSubnet
	Routes  string
	Labels  map[string]string
}

// Validate checks if options are valid.
func (o NetworkCreateOpts) Validate() error {
	if o.Name == "" {
		return errors.New("missing name")
	}
	if o.IPRange == nil || o.IPRange.String() == "" {
		return errors.New("missing IP range")
	}
	return nil
}

// getFirstUsableIP returns the first usable IP address in the given network range (network + 1)
func getFirstUsableIP(ipnet *net.IPNet) net.IP {
	ip := ipnet.IP.To4()
	if ip == nil {
		ip = ipnet.IP.To16()
	}

	// Get the network address
	network := ip.Mask(ipnet.Mask)

	// Create a copy and increment by 1 to get the first usable IP
	firstUsable := make(net.IP, len(network))
	copy(firstUsable, network)

	// Increment the last octet by 1
	for i := len(firstUsable) - 1; i >= 0; i-- {
		firstUsable[i]++
		if firstUsable[i] != 0 {
			break
		}
	}

	return firstUsable
}

// getLastUsableIP returns the last usable IP address in the given network range (broadcast - 1)
func getLastUsableIP(ipnet *net.IPNet) net.IP {
	ip := ipnet.IP.To4()
	if ip == nil {
		ip = ipnet.IP.To16()
	}

	// Get the network address
	network := ip.Mask(ipnet.Mask)

	// Calculate the broadcast address
	broadcast := make(net.IP, len(network))
	copy(broadcast, network)

	// Set all host bits to 1 to get broadcast
	for i := 0; i < len(network); i++ {
		broadcast[i] |= ^ipnet.Mask[i]
	}

	// Subtract 1 from broadcast to get last usable IP
	for i := len(broadcast) - 1; i >= 0; i-- {
		if broadcast[i] > 0 {
			broadcast[i]--
			break
		}
		broadcast[i] = 255
	}

	return broadcast
}

// Create creates a new network.
func (c *NetworkClient) Create(ctx context.Context, opts NetworkCreateOpts) (*Network, *Response, error) {
	if err := opts.Validate(); err != nil {
		return nil, nil, err
	}
	reqBody := schema.CreateNetworkRequest{
		ServerUrl: "https://172.16.24.228:7776",
		Name:      opts.Name,
		Type:      "libvirt",
		Mode:      "",
		Private:   true,
		IP: schema.NetworkIP{
			Address: opts.IPRange.String(),
			DHCP: schema.DHCP{
				Start: getFirstUsableIP(opts.IPRange).String(),
				End:   getLastUsableIP(opts.IPRange).String(),
			},
		},
		Routes: []schema.Route{
			{
				Address: opts.Routes,
			},
		},
	}

	_, err := c.client.CreateNetwork(reqBody)
	if err != nil {
		return nil, nil, err
	}
	// No response from the API as of now, so we return nil

	return nil, &Response{}, nil
}

// NetworkSubnetFromSchema converts a schema.NetworkSubnet to a NetworkSubnet.
// func NetworkSubnetFromSchema(s schema.NetworkSubnet) NetworkSubnet {
// 	return NetworkSubnet{
// 		Type:        NetworkSubnetType(s.Type),
// 		IPRange:     s.IPRange,
// 		NetworkZone: NetworkZone(s.NetworkZone),
// 		Gateway:     s.Gateway,
// 		VSwitchID:   s.VSwitchID,
// 	}
// }

// NetworkServerFromSchema converts a schema.Server to a Server compatible with kOps.
func NetworkServerFromSchema(s schema.Server) *Server {
	return &Server{
		ID:           s.UniqueID,
		Name:         s.Name,
		Status:       ServerStatus(s.Status),
		Created:      s.Created,
		PublicNet:    ServerPublicNet{},
		ServerType:   &ServerType{},
		Datacenter:   Datacenter{},
		BackupWindow: "",
		Labels:       s.Labels,
		Volumes:      []*schema.StorageVolume{},
	}
}

// NetworkFromSchema converts a schema.Network to a Network compatible with kOps.
func NetworkFromSchema(s schema.Network) *Network {
	// subnets := make([]NetworkSubnet, len(s.Subnets))
	// for i, subnet := range s.Subnets {
	// 	subnets[i] = NetworkSubnetFromSchema(subnet)
	// }
	// servers := make([]*Server, len(s.Servers))
	// for i, server := range s.Servers {
	// 	servers[i] = NetworkServerFromSchema(*server)
	// }

	// Parse the IP address string back to *net.IPNet
	var ipRange *net.IPNet
	if s.IP.Address != "" {
		_, parsed, err := net.ParseCIDR(s.IP.Address)
		if err == nil {
			ipRange = parsed
		}
	}

	// Get the first route address if available
	var routesStr string
	if len(s.Routes) > 0 {
		routesStr = s.Routes[0].Address
	}

	return &Network{
		ID:         s.NetworkID,
		Name:       s.Name,
		Created:    time.Now(),
		IPRange:    ipRange,
		Subnets:    nil,
		Routes:     routesStr,
		Servers:    nil,
		Protection: s.Private,
		Labels:     nil,
	}
}

// NetworkAddSubnetOpts specifies options for adding a subnet to a network.
// type NetworkAddSubnetOpts struct {
// 	Subnet NetworkSubnet
// }

// AddSubnet adds a subnet to a network.
// func (c *NetworkClient) AddSubnet(ctx context.Context, network *Network, opts NetworkAddSubnetOpts) (*Response, error) {
// 	reqBody := schema.NetworkActionAddSubnetRequest{
// 		Type:        string(opts.Subnet.Type),
// 		NetworkZone: string(opts.Subnet.NetworkZone),
// 	}
// 	if opts.Subnet.IPRange != nil {
// 		reqBody.IPRange = opts.Subnet.IPRange.String()
// 	}
// 	if opts.Subnet.VSwitchID != 0 {
// 		reqBody.VSwitchID = opts.Subnet.VSwitchID
// 	}
// 	reqBodyData, err := json.Marshal(reqBody)
// 	if err != nil {
// 		return nil, nil, err
// 	}

// 	path := fmt.Sprintf("/networks/%d/actions/add_subnet", network.ID)
// 	req, err := c.client.NewRequest(ctx, "POST", path, bytes.NewReader(reqBodyData))
// 	if err != nil {
// 		return nil, nil, err
// 	}

// 	respBody := schema.NetworkActionAddSubnetResponse{}
// 	resp, err := c.client.Do(req, &respBody)
// 	if err != nil {
// 		return nil, resp, err
// 	}
// 	return resp, nil
// }

// NetworkDeleteSubnetOpts specifies options for deleting a subnet from a network.
// type NetworkDeleteSubnetOpts struct {
// 	Subnet NetworkSubnet
// }

// DeleteSubnet deletes a subnet from a network.
// func (c *NetworkClient) DeleteSubnet(ctx context.Context, network *Network, opts NetworkDeleteSubnetOpts) (*Response, error) {
// 	reqBody := schema.NetworkActionDeleteSubnetRequest{
// 		IPRange: opts.Subnet.IPRange.String(),
// 	}
// 	reqBodyData, err := json.Marshal(reqBody)
// 	if err != nil {
// 		return nil, nil, err
// 	}

// 	path := fmt.Sprintf("/networks/%d/actions/delete_subnet", network.ID)
// 	req, err := c.client.NewRequest(ctx, "POST", path, bytes.NewReader(reqBodyData))
// 	if err != nil {
// 		return nil, nil, err
// 	}

// 	respBody := schema.NetworkActionDeleteSubnetResponse{}
// 	resp, err := c.client.Do(req, &respBody)
// 	if err != nil {
// 		return nil, resp, err
// 	}
// 	return resp, nil
// }
