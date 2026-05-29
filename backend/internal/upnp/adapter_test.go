package upnp

import (
	"testing"

	"github.com/clagon/port-mapper/backend/internal/application"
)

func TestNewSOAPPortMapperUsesTimeoutClient(t *testing.T) {
	client := NewSOAPPortMapper(application.DiscoveryResult{ServiceType: "urn:schemas-upnp-org:service:WANIPConnection:2", ControlURL: "http://192.168.1.1:1900/control"})
	soap, ok := client.(*SOAPClient)
	if !ok {
		t.Fatalf("NewSOAPPortMapper() type = %T, want *SOAPClient", client)
	}
	if soap.HTTPClient == nil {
		t.Fatal("HTTPClient = nil")
	}
	if soap.HTTPClient.Timeout <= 0 {
		t.Fatalf("HTTPClient timeout = %v, want > 0", soap.HTTPClient.Timeout)
	}
}
