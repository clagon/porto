package upnp

import "github.com/clagon/port-mapper/backend/internal/application"

// PortMapping represents a single port forwarding request.
type PortMapping = application.PortMapping

const MaxLeaseDurationSeconds = application.MaxLeaseDurationSeconds

// DiscoveryResult is the selected UPnP control endpoint discovered from a root description.
type DiscoveryResult = application.DiscoveryResult
