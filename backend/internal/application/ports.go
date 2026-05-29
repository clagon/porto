package application

import "errors"

// ErrNoGateway reports that no UPnP gateway could be discovered.
var ErrNoGateway = errors.New("no UPnP gateway discovered")

// DiscoveryClient discovers a router control endpoint for port mapping.
type DiscoveryClient interface {
	Discover() (DiscoveryResult, error)
}

// PortMapper manages port mappings against a discovered router.
type PortMapper interface {
	GetExternalIPAddress() (string, error)
	AddPortMapping(PortMapping) error
	DeletePortMapping(protocol string, externalPort int) error
	GetGenericPortMappingEntry(index int) (PortMapping, error)
}

// PortMapperFactory creates a PortMapper from a discovery result.
type PortMapperFactory func(DiscoveryResult) PortMapper
