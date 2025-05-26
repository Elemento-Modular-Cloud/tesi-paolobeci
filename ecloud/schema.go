package ecloud

import (
	"strconv"

	"github.com/Elemento-Modular-Cloud/tesi-paolobeci/ecloud/schema"
)

// ServerPublicNetFromSchema converts a schema.NetworkConfig to a ServerPublicNet structure compatible with kOps.
func ServerPublicNetFromSchema(nc schema.NetworkConfig) ServerPublicNet {
	var ipv4 string
	if nc.IPv4 != nil {
		ipv4 = *nc.IPv4
	}
	return ServerPublicNet{
		IPv4: ipv4,
	}
}

// ServerFromSchema converts a schema.Server to a Server structure compatible with kOps.
func ServerFromSchema(s schema.Server) *Server {
	server := &Server{
		ID:        s.UniqueID,
		Name:      s.Name,
		Status:    ServerStatus(s.Status),
		Created:   s.Created,
		PublicNet: ServerPublicNetFromSchema(s.NetworkConfig),
		Labels:    s.Labels,
	}

	// Convert volumes
	for _, vol := range s.Volumes {
		id, _ := strconv.Atoi(vol.VolumeID)
		server.Volumes = append(server.Volumes, &Volume{
			ID:   id,
			Name: vol.Name,
		})
	}

	// Add optional fields here

	return server
}
