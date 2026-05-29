package upnp

import (
	"net/http"
	"time"

	"github.com/clagon/port-mapper/backend/internal/application"
)

// DiscoveryClient adapts UPnP discovery to the application service port.
type DiscoveryClient struct{}

// NewDiscoveryClient creates a UPnP discovery adapter.
func NewDiscoveryClient() DiscoveryClient {
	return DiscoveryClient{}
}

// Discover finds a supported UPnP gateway.
func (DiscoveryClient) Discover() (application.DiscoveryResult, error) {
	return Discover()
}

// NewSOAPPortMapper adapts a UPnP discovery result to a SOAP-backed port mapper.
func NewSOAPPortMapper(result application.DiscoveryResult) application.PortMapper {
	return &SOAPClient{
		Endpoint:    result.ControlURL,
		ServiceType: result.ServiceType,
		HTTPClient:  &http.Client{Timeout: 5 * time.Second},
	}
}
