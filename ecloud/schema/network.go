package schema

import (
	"net"
)

// Network represents a network in the API response
type Network struct {
	CreatorID      string    `json:"creator_uid"`
	DeviceName     string    `json:"device_name"`
	IP             NetworkIP `json:"ip"`
	LibvirtNetwork string    `json:"libvirt_network"`
	Name           string    `json:"network_name"`
	NetworkID      string    `json:"network_uid"`
	Private        bool      `json:"private"`
	Routes         []Route   `json:"routes,omitempty"`
	ServerUrl      string    `json:"serverurl"`
	Type           string    `json:"type"`
}

type Route struct {
	Address string `json:"address"`
	Prefix  int    `json:"prefix"`
	Gateway string `json:"gateway"`
}

// NetworkIP represents an IP in the API response
type NetworkIP struct {
	Address string `json:"address"`
	DHCP    DHCP   `json:"dhcp"`
}

type DHCP struct {
	End   string `json:"end"`
	Start string `json:"start"`
	Hosts []Host `json:"hosts,omitempty"`
}

type Host struct {
	Mac     string `json:"mac"`
	Name    string `json:"name,omitempty"`
	Address string `json:"address,omitempty"`
}

// NetworkRoute represents a route in the API response
type NetworkRoute struct {
	Destination *net.IPNet `json:"destination"`
	Gateway     net.IP     `json:"gateway"`
}

// -------- NETWORK GET BY ID --------
type GetNetworkByIDRequest struct {
	NetworkID string `json:"network_uid"`
}

type GetNetworkByIDResponse struct {
	Network Network `json:"network"`
}

// -------- NETWORK LIST --------
type ListNetworkResponse []Network

// -------- NETWORK CREATE --------
type CreateNetworkRequest struct {
	ServerUrl string    `json:"serverurl"`
	Name      string    `json:"network_name"`
	Type      string    `json:"type"`
	Mode      string    `json:"mode"`
	Private   bool      `json:"private"`
	IP        NetworkIP `json:"ip"`
	Routes    []Route   `json:"routes"`
}

type CreateNetworkResponse struct{}

// -------- NETWORK DELETE --------
type DeleteNetworkRequest struct {
	NetworkID string `json:"network_uid"`
}

type DeleteNetworkResponse struct{}
