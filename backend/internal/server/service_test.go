package server

import (
	"errors"
	"testing"

	"github.com/clagon/port-mapper/backend/internal/config"
	"github.com/clagon/port-mapper/backend/internal/upnp"
)

func TestDefaultPortMapperFactoryUsesTimeoutClient(t *testing.T) {
	client := defaultPortMapperFactory(upnp.DiscoveryResult{ServiceType: "urn:schemas-upnp-org:service:WANIPConnection:2", ControlURL: "http://192.168.1.1:1900/control"})
	soap, ok := client.(*upnp.SOAPClient)
	if !ok {
		t.Fatalf("defaultPortMapperFactory() type = %T, want *upnp.SOAPClient", client)
	}
	if soap.HTTPClient == nil {
		t.Fatal("HTTPClient = nil")
	}
	if soap.HTTPClient.Timeout <= 0 {
		t.Fatalf("HTTPClient timeout = %v, want > 0", soap.HTTPClient.Timeout)
	}
}

type recordingSettingsStore struct {
	saved []config.Config
	err   error
}

func (s *recordingSettingsStore) Save(cfg config.Config) error {
	if s.err != nil {
		return s.err
	}
	s.saved = append(s.saved, cfg)
	return nil
}

func TestUpdateSettingsUsesInjectedSettingsStore(t *testing.T) {
	store := &recordingSettingsStore{}
	svc := newService(serviceOptions{
		cfg:           config.DefaultConfig(),
		settingsStore: store,
	})

	next, err := svc.updateSettings(config.Config{ListenAddr: "127.0.0.1:9090", AutoDiscover: config.BoolPtr(false)})
	if err != nil {
		t.Fatalf("updateSettings() error = %v", err)
	}
	if len(store.saved) != 1 {
		t.Fatalf("Save() calls = %d, want 1", len(store.saved))
	}
	if store.saved[0].ListenAddr != next.ListenAddr {
		t.Fatalf("saved ListenAddr = %q, want %q", store.saved[0].ListenAddr, next.ListenAddr)
	}
	if got := svc.settings().ListenAddr; got != next.ListenAddr {
		t.Fatalf("settings().ListenAddr = %q, want %q", got, next.ListenAddr)
	}
}

func TestUpdateSettingsDoesNotMutateConfigWhenSettingsStoreFails(t *testing.T) {
	storeErr := errors.New("save failed")
	svc := newService(serviceOptions{
		cfg:           config.Config{ListenAddr: "127.0.0.1:8080", AutoDiscover: config.BoolPtr(true)},
		settingsStore: &recordingSettingsStore{err: storeErr},
	})

	_, err := svc.updateSettings(config.Config{ListenAddr: "127.0.0.1:9090", AutoDiscover: config.BoolPtr(false)})
	if !errors.Is(err, storeErr) {
		t.Fatalf("updateSettings() error = %v, want %v", err, storeErr)
	}
	if got := svc.settings().ListenAddr; got != "127.0.0.1:8080" {
		t.Fatalf("settings().ListenAddr = %q, want original", got)
	}
}
