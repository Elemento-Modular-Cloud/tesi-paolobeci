package schema

import (
	"net"
	"time"
)

// Network represents a network in the API response
type Network struct {
	ID         int               `json:"id"`
	Name       string            `json:"name"`
	Created    time.Time         `json:"created"`
	IPRange    *net.IPNet        `json:"ip_range"`
	Subnets    []NetworkSubnet   `json:"subnets"`
	Routes     string            `json:"routes"`
	Servers    []*Server         `json:"servers"`
	Protection bool              `json:"protection"`
	Labels     map[string]string `json:"labels"`
}

// NetworkSubnet represents a subnet in the API response
type NetworkSubnet struct {
	Type        string     `json:"type"`
	IPRange     *net.IPNet `json:"ip_range"`
	NetworkZone string     `json:"network_zone"`
	Gateway     net.IP     `json:"gateway"`
	VSwitchID   int        `json:"vswitch_id"`
}

// NetworkRoute represents a route in the API response
type NetworkRoute struct {
	Destination *net.IPNet `json:"destination"`
	Gateway     net.IP     `json:"gateway"`
}

// Get network
type NetworkGetResponse struct {
	Network Network `json:"network"`
}

// List network
type NetworkListResponse struct {
	Networks []Network `json:"networks"`
}

// Create network
type NetworkCreateRequest struct {
	Name    string             `json:"name"`
	IPRange string             `json:"ip_range"`
	Subnets []NetworkSubnet    `json:"subnets,omitempty"`
	Labels  *map[string]string `json:"labels,omitempty"`
}

type NetworkCreateResponse struct {
	Network Network `json:"network"`
}
