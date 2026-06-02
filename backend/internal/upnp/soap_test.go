package upnp

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/clagon/port-mapper/backend/internal/domain"
)

func TestSOAPEnvelope(t *testing.T) {
	tests := []struct {
		name               string
		action             string
		serviceType        string
		body               map[string]string
		wantBodySubstrings []string // SOAP envelope substrings
	}{
		{
			name:        "get external ip action and namespace",
			action:      "GetExternalIPAddress",
			serviceType: "urn:schemas-upnp-org:service:WANIPConnection:2",
			wantBodySubstrings: []string{
				`<u:GetExternalIPAddress xmlns:u="urn:schemas-upnp-org:service:WANIPConnection:2">`,
				`<s:Envelope`,
			},
		},
		{
			name:        "add port mapping includes all fields",
			action:      "AddPortMapping",
			serviceType: "urn:schemas-upnp-org:service:WANIPConnection:2",
			body: map[string]string{
				"NewRemoteHost":             "",
				"NewExternalPort":           "8080",
				"NewProtocol":               "TCP",
				"NewInternalPort":           "8080",
				"NewInternalClient":         "192.168.1.20",
				"NewEnabled":                "1",
				"NewPortMappingDescription": "test mapping",
				"NewLeaseDuration":          "3600",
			},
			wantBodySubstrings: []string{
				`<u:AddPortMapping xmlns:u="urn:schemas-upnp-org:service:WANIPConnection:2">`,
				`<NewExternalPort>8080</NewExternalPort>`,
				`<NewProtocol>TCP</NewProtocol>`,
				`<NewInternalClient>192.168.1.20</NewInternalClient>`,
				`<NewPortMappingDescription>test mapping</NewPortMappingDescription>`,
				`<NewLeaseDuration>3600</NewLeaseDuration>`,
			},
		},
		{
			name:        "delete port mapping includes protocol and external port",
			action:      "DeletePortMapping",
			serviceType: "urn:schemas-upnp-org:service:WANIPConnection:2",
			body: map[string]string{
				"NewExternalPort": "8080",
				"NewProtocol":     "UDP",
			},
			wantBodySubstrings: []string{
				`<u:DeletePortMapping xmlns:u="urn:schemas-upnp-org:service:WANIPConnection:2">`,
				`<NewExternalPort>8080</NewExternalPort>`,
				`<NewProtocol>UDP</NewProtocol>`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildSOAPEnvelope(tt.action, tt.serviceType, tt.body)
			if err != nil {
				t.Fatalf("buildSOAPEnvelope() error = %v", err)
			}
			body := string(got)
			for _, want := range tt.wantBodySubstrings {
				if !strings.Contains(body, want) {
					t.Fatalf("envelope missing %q\n%s", want, body)
				}
			}
		})
	}
}

func TestSOAPGetExternalIPAddress(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "get external ip",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestSOAPClient(t, func(r *http.Request) *http.Response {
				if r.Method != http.MethodPost {
					t.Fatalf("method = %s, want POST", r.Method)
				}
				if got := r.Header.Get("SOAPAction"); got != `"urn:schemas-upnp-org:service:WANIPConnection:2#GetExternalIPAddress"` {
					t.Fatalf("SOAPAction = %q", got)
				}
				_, _ = io.ReadAll(r.Body)
				return soapTestResponse(r, http.StatusOK, `<?xml version="1.0"?>
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
  <s:Body>
    <u:GetExternalIPAddressResponse xmlns:u="urn:schemas-upnp-org:service:WANIPConnection:2">
      <NewExternalIPAddress>203.0.113.42</NewExternalIPAddress>
    </u:GetExternalIPAddressResponse>
  </s:Body>
</s:Envelope>`)
			})
			got, err := c.GetExternalIPAddress()
			if err != nil {
				t.Fatalf("GetExternalIPAddress() error = %v", err)
			}
			if got != "203.0.113.42" {
				t.Fatalf("GetExternalIPAddress() = %q, want %q", got, "203.0.113.42")
			}
		})
	}
}

func TestSOAPClientDefaultTimeout(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "default timeout",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := (&SOAPClient{}).client(); got == nil {
				t.Fatal("client() = nil")
			} else if got.Timeout <= 0 {
				t.Fatalf("client timeout = %v, want > 0", got.Timeout)
			}
		})
	}
}

func TestSOAPGenericPortMappingEntryIndexInvalid(t *testing.T) {
	c := newTestSOAPClient(t, func(r *http.Request) *http.Response {
		return soapTestResponse(r, http.StatusInternalServerError, `<?xml version="1.0"?>
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
  <s:Body>
    <s:Fault>
      <faultcode>s:Client</faultcode>
      <faultstring>UPnPError</faultstring>
      <detail>
        <UPnPError xmlns="urn:schemas-upnp-org:control-1-0">
          <errorCode>713</errorCode>
          <errorDescription>SpecifiedArrayIndexInvalid</errorDescription>
        </UPnPError>
      </detail>
    </s:Fault>
  </s:Body>
</s:Envelope>`)
	})

	_, err := c.GetGenericPortMappingEntry(0)
	if !errors.Is(err, domain.ErrNoPortMappingEntry) {
		t.Fatalf("GetGenericPortMappingEntry() error = %v, want ErrNoPortMappingEntry", err)
	}
}

func TestSOAPClientRejectsInvalidEndpoint(t *testing.T) {
	c := &SOAPClient{
		Endpoint:    "http://127.0.0.1:1900/control",
		ServiceType: "urn:schemas-upnp-org:service:WANIPConnection:2",
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				return soapTestResponse(r, http.StatusOK, ""), nil
			}),
		},
	}

	_, err := c.GetExternalIPAddress()
	if err == nil {
		t.Fatal("GetExternalIPAddress() error = nil, want invalid endpoint error")
	}
	if !strings.Contains(err.Error(), "invalid soap endpoint") {
		t.Fatalf("error = %v, want invalid soap endpoint", err)
	}
}

func TestSOAPClientRejectsRedirect(t *testing.T) {
	c := newTestSOAPClient(t, func(r *http.Request) *http.Response {
		resp := soapTestResponse(r, http.StatusFound, "")
		resp.Header.Set("Location", "http://192.168.1.2:1900/control")
		return resp
	})

	_, err := c.GetExternalIPAddress()
	if err == nil {
		t.Fatal("GetExternalIPAddress() error = nil, want redirect error")
	}
	if !strings.Contains(err.Error(), "redirect target host") {
		t.Fatalf("error = %v, want redirect host error", err)
	}
}

func TestSOAPClientRejectsOversizedResponse(t *testing.T) {
	c := newTestSOAPClient(t, func(r *http.Request) *http.Response {
		return soapTestResponse(r, http.StatusOK, strings.Repeat("x", maxSOAPResponseBytes+1))
	})

	_, err := c.GetExternalIPAddress()
	if err == nil {
		t.Fatal("GetExternalIPAddress() error = nil, want size limit error")
	}
	if !strings.Contains(err.Error(), "soap response exceeds") {
		t.Fatalf("error = %v, want size limit error", err)
	}
}

func TestSOAPFault(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "soap fault",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := newTestSOAPClient(t, func(r *http.Request) *http.Response {
				return soapTestResponse(r, http.StatusInternalServerError, `<?xml version="1.0"?>
<s:Envelope xmlns:s="http://schemas.xmlsoap.org/soap/envelope/">
  <s:Body>
    <s:Fault>
      <faultcode>s:Client</faultcode>
      <faultstring>UPnPError</faultstring>
      <detail>
        <UPnPError xmlns="urn:schemas-upnp-org:control-1-0">
          <errorCode>401</errorCode>
          <errorDescription>Invalid Action</errorDescription>
        </UPnPError>
      </detail>
    </s:Fault>
  </s:Body>
</s:Envelope>`)
			})
			_, err := c.GetExternalIPAddress()
			if err == nil {
				t.Fatal("GetExternalIPAddress() error = nil, want error")
			}
			if !strings.Contains(err.Error(), "Invalid Action") {
				t.Fatalf("error = %v, want SOAP fault", err)
			}
		})
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func newTestSOAPClient(t *testing.T, handler func(*http.Request) *http.Response) *SOAPClient {
	t.Helper()
	return &SOAPClient{
		Endpoint:    "http://192.168.1.1:1900/control",
		ServiceType: "urn:schemas-upnp-org:service:WANIPConnection:2",
		HTTPClient: &http.Client{
			Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				return handler(r), nil
			}),
		},
	}
}

func soapTestResponse(r *http.Request, statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Status:     http.StatusText(statusCode),
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    r,
	}
}
