package ecloud

import (
	"context"
	"errors"
	"net"
	"net/url"
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
	ID         int
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
func (c *NetworkClient) GetByID(ctx context.Context, id int) (*Network, *schema.NetworkGetResponse, error) {
	var body schema.NetworkGetResponse
	resp, err := c.client.GetNetworkByID(&body)
	if err != nil {
		if IsError(err, ErrorCodeNotFound) {
			return nil, resp, nil
		}
		return nil, nil, err
	}
	return NetworkFromSchema(body.Network), resp, nil
}

// GetByName retrieves a network by its name. If the network does not exist, nil is returned.
func (c *NetworkClient) GetByName(ctx context.Context, name string) (*Network, *schema.NetworkListResponse, error) {
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

func (l NetworkListOpts) values() url.Values {
	vals := l.ListOpts.Values()
	if l.Name != "" {
		vals.Add("name", l.Name)
	}
	for _, sort := range l.Sort {
		vals.Add("sort", sort)
	}
	return vals
}

// List returns a list of networks for a specific page.
//
// Please note that filters specified in opts are not taken into account
// when their value corresponds to their zero value or when they are empty.
func (c *NetworkClient) List(ctx context.Context, opts NetworkListOpts) ([]*Network, *schema.NetworkListResponse, error) {
	var body schema.NetworkListResponse
	resp, err := c.client.ListNetwork(&body)
	if err != nil {
		return nil, nil, err
	}
	Networks := make([]*Network, 0, len(body.Networks))
	for _, s := range body.Networks {
		if opts.Name != "" && s.Name != opts.Name {
			continue
		}
		Networks = append(Networks, NetworkFromSchema(s))
	}
	return Networks, resp, nil
}

// All returns all networks.
func (c *NetworkClient) All(ctx context.Context) ([]*Network, error) {
	return c.AllWithOpts(ctx, NetworkListOpts{ListOpts: ListOpts{PerPage: 50}})
}

// AllWithOpts returns all networks for the given options.
func (c *NetworkClient) AllWithOpts(ctx context.Context, opts NetworkListOpts) ([]*Network, error) {
	allNetworks := []*Network{}

	err := c.client.all(func(page int) (*Response, error) {
		opts.Page = page
		Networks, _, err := c.List(ctx, opts)
		if err != nil {
			return nil, err
		}
		allNetworks = append(allNetworks, Networks...)
		return &Response{}, nil
	})
	if err != nil {
		return nil, err
	}

	return allNetworks, nil
}

// Delete deletes a network.
func (c *NetworkClient) Delete(ctx context.Context, network *Network) (*Response, error) {
	return c.client.DeleteNetwork(network)
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

// Create creates a new network.
func (c *NetworkClient) Create(ctx context.Context, opts NetworkCreateOpts) (*Network, *Response, error) {
	if err := opts.Validate(); err != nil {
		return nil, nil, err
	}
	reqBody := schema.NetworkCreateRequest{
		Name:    opts.Name,
		IPRange: opts.IPRange.String(),
	}
	for _, subnet := range opts.Subnets {
		s := schema.NetworkSubnet{
			Type:        string(subnet.Type),
			IPRange:     subnet.IPRange,
			NetworkZone: string(subnet.NetworkZone),
		}
		if subnet.VSwitchID != 0 {
			s.VSwitchID = subnet.VSwitchID
		}
		reqBody.Subnets = append(reqBody.Subnets, s)
	}
	if opts.Labels != nil {
		reqBody.Labels = &opts.Labels
	}

	resp, err := c.client.CreateNetwork(reqBody)
	if err != nil {
		return nil, nil, err
	}
	return NetworkFromSchema(resp.Network), &Response{}, nil
}

// NetworkSubnetFromSchema converts a schema.NetworkSubnet to a NetworkSubnet.
func NetworkSubnetFromSchema(s schema.NetworkSubnet) NetworkSubnet {
	return NetworkSubnet{
		Type:        NetworkSubnetType(s.Type),
		IPRange:     s.IPRange,
		NetworkZone: NetworkZone(s.NetworkZone),
		Gateway:     s.Gateway,
		VSwitchID:   s.VSwitchID,
	}
}

// NetworkServerFromSchema converts a schema.Server to a Server.
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

// NetworkFromSchema converts a schema.Network to a Network.
func NetworkFromSchema(s schema.Network) *Network {
	subnets := make([]NetworkSubnet, len(s.Subnets))
	for i, subnet := range s.Subnets {
		subnets[i] = NetworkSubnetFromSchema(subnet)
	}
	servers := make([]*Server, len(s.Servers))
	for i, server := range s.Servers {
		servers[i] = NetworkServerFromSchema(*server)
	}
	return &Network{
		ID:         s.ID,
		Name:       s.Name,
		Created:    s.Created,
		IPRange:    s.IPRange,
		Subnets:    subnets,
		Routes:     s.Routes,
		Servers:    servers,
		Protection: s.Protection,
		Labels:     s.Labels,
	}
}

// // NetworkAddSubnetOpts specifies options for adding a subnet to a network.
// type NetworkAddSubnetOpts struct {
// 	Subnet NetworkSubnet
// }

// // AddSubnet adds a subnet to a network.
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

// // NetworkDeleteSubnetOpts specifies options for deleting a subnet from a network.
// type NetworkDeleteSubnetOpts struct {
// 	Subnet NetworkSubnet
// }

// // DeleteSubnet deletes a subnet from a network.
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
