package server

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"strings"
	"sync"

	"github.com/clagon/port-mapper/backend/internal/application"
	"github.com/clagon/port-mapper/backend/internal/config"
)

type DiscoveryClient = application.DiscoveryClient

type PortMapper = application.PortMapper

type PortMapperFactory = application.PortMapperFactory

type serviceOptions struct {
	configPath        string
	cfg               config.Config
	discovery         DiscoveryClient
	portMapperFactory PortMapperFactory
	logger            *slog.Logger
}

type service struct {
	mu                sync.RWMutex
	cfg               config.Config
	configPath        string
	discovery         DiscoveryClient
	portMapperFactory PortMapperFactory
	gateway           *application.DiscoveryResult
	externalIP        string
	localIP           string
	ports             []application.PortMapping
	logger            *slog.Logger
}

// service 内で gateway 未選択を表すエラー。UPnP discovery 自体の失敗は application.ErrNoGateway を使う。
var errNoGateway = errors.New("no UPnP gateway discovered")

func newService(opts serviceOptions) *service {
	logger := opts.logger
	if logger == nil {
		logger = slog.Default()
	}
	cfg := opts.cfg.WithDefaults()
	if opts.configPath == "" {
		opts.configPath = config.DefaultPath()
	}
	return &service{
		cfg:               cfg,
		configPath:        opts.configPath,
		discovery:         opts.discovery,
		portMapperFactory: opts.portMapperFactory,
		logger:            logger,
	}
}

func (s *service) settings() config.Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg.WithDefaults()
}

func (s *service) updateSettings(next config.Config) (config.Config, error) {
	next = next.WithDefaults()
	if err := config.ValidateLocalListenAddr(next.ListenAddr); err != nil {
		return config.Config{}, err
	}
	if err := config.Save(s.configPath, next); err != nil {
		return config.Config{}, err
	}

	s.mu.Lock()
	s.cfg = next
	s.mu.Unlock()

	if s.logger != nil {
		s.logger.Info("settings saved",
			"config_path", s.configPath,
			"listen_addr", next.ListenAddr,
			"auto_discover", boolValue(next.AutoDiscover),
		)
	}
	return next, nil
}

func (s *service) status() StatusResponse {
	s.mu.RLock()
	defer s.mu.RUnlock()

	resp := StatusResponse{
		Discovered: s.gateway != nil,
		Ports:      append([]application.PortMapping{}, s.ports...),
	}
	if s.gateway != nil {
		resp.ServiceType = s.gateway.ServiceType
		resp.ControlURL = s.gateway.ControlURL
		resp.ExternalIP = s.externalIP
		resp.LocalIP = s.localIP
	}
	return resp
}

func (s *service) discover() (StatusResponse, error) {
	if s.discovery == nil {
		return s.status(), nil
	}
	result, err := s.discovery.Discover()
	if err != nil {
		// discovery 未検出は UI 上の通常状態として扱い、現在の status を返す。
		if errors.Is(err, application.ErrNoGateway) {
			if s.logger != nil {
				s.logger.Info("router not discovered", "reason", err.Error())
			}
			return s.status(), nil
		}
		return StatusResponse{}, err
	}

	var mapper PortMapper
	if s.portMapperFactory != nil {
		mapper = s.portMapperFactory(result)
	}
	var externalIP string
	if mapper != nil {
		if ip, err := mapper.GetExternalIPAddress(); err == nil {
			externalIP = ip
		} else if s.logger != nil {
			s.logger.Warn("external IP lookup failed", "error", err)
		}
	}

	// 自身のローカルIPアドレスの特定
	var localIP string
	if u, err := url.Parse(result.ControlURL); err == nil {
		if conn, err := net.Dial("udp", u.Host); err == nil {
			if localAddr, ok := conn.LocalAddr().(*net.UDPAddr); ok && !localAddr.IP.IsUnspecified() {
				localIP = localAddr.IP.String()
			}
			conn.Close()
		}
	}
	if localIP == "" {
		if ip, err := s.fallbackLocalIP(); err == nil {
			localIP = ip
		}
	}

	s.mu.Lock()
	resultCopy := result
	s.gateway = &resultCopy
	s.externalIP = externalIP
	s.localIP = localIP
	s.mu.Unlock()

	// ルーターからのポートマッピング自動同期
	if mapper != nil && localIP != "" {
		s.syncActivePorts(mapper, localIP)
	}

	if s.logger != nil {
		s.logger.Info("router discovered",
			"service_type", result.ServiceType,
			"control_url", result.ControlURL,
			"external_ip", externalIP,
			"local_ip", localIP,
		)
	}
	return s.status(), nil
}

func (s *service) openPort(mapping application.PortMapping) (StatusResponse, error) {
	resolvedIP, err := s.resolveInternalIP(mapping.InternalIP)
	if err != nil {
		return StatusResponse{}, fmt.Errorf("failed to resolve internal IP: %w", err)
	}
	mapping.InternalIP = resolvedIP

	if err := application.ValidatePortMapping(mapping); err != nil {
		return StatusResponse{}, err
	}
	mapper, err := s.currentPortMapper()
	if err != nil {
		return StatusResponse{}, err
	}
	if err := mapper.AddPortMapping(mapping); err != nil {
		return StatusResponse{}, err
	}

	s.mu.Lock()
	s.ports = upsertMapping(s.ports, mapping)
	s.mu.Unlock()

	if s.logger != nil {
		s.logger.Info("port mapping opened",
			"protocol", normalizeProtocol(mapping.Protocol),
			"external_port", mapping.ExternalPort,
			"internal_ip", mapping.InternalIP,
			"internal_port", mapping.InternalPort,
		)
	}
	return s.status(), nil
}

func (s *service) closePort(mapping application.PortMapping) (StatusResponse, error) {
	if err := validateDeleteRequest(mapping); err != nil {
		return StatusResponse{}, err
	}
	mapper, err := s.currentPortMapper()
	if err != nil {
		return StatusResponse{}, err
	}
	if err := mapper.DeletePortMapping(mapping.Protocol, mapping.ExternalPort); err != nil {
		return StatusResponse{}, err
	}

	s.mu.Lock()
	s.ports = removeMapping(s.ports, mapping.Protocol, mapping.ExternalPort)
	s.mu.Unlock()

	if s.logger != nil {
		s.logger.Info("port mapping closed",
			"protocol", normalizeProtocol(mapping.Protocol),
			"external_port", mapping.ExternalPort,
		)
	}
	return s.status(), nil
}

func (s *service) currentPortMapper() (PortMapper, error) {
	s.mu.RLock()
	gateway := s.gateway
	s.mu.RUnlock()
	if gateway == nil {
		return nil, errNoGateway
	}
	if s.portMapperFactory == nil {
		return nil, fmt.Errorf("port mapper factory is not configured")
	}
	mapper := s.portMapperFactory(*gateway)
	if mapper == nil {
		return nil, fmt.Errorf("port mapper factory returned nil")
	}
	return mapper, nil
}

func upsertMapping(existing []application.PortMapping, next application.PortMapping) []application.PortMapping {
	for i, current := range existing {
		if sameMappingKey(current, next) {
			existing[i] = next
			return existing
		}
	}
	return append(existing, next)
}

func removeMapping(existing []application.PortMapping, protocol string, externalPort int) []application.PortMapping {
	filtered := existing[:0]
	for _, current := range existing {
		if sameMappingIdentity(current.Protocol, current.ExternalPort, protocol, externalPort) {
			continue
		}
		filtered = append(filtered, current)
	}
	return filtered
}

func sameMappingKey(a, b application.PortMapping) bool {
	return sameMappingIdentity(a.Protocol, a.ExternalPort, b.Protocol, b.ExternalPort)
}

func sameMappingIdentity(aProtocol string, aPort int, bProtocol string, bPort int) bool {
	return normalizeProtocol(aProtocol) == normalizeProtocol(bProtocol) && aPort == bPort
}

func normalizeProtocol(protocol string) string {
	return strings.ToUpper(strings.TrimSpace(protocol))
}

func validateDeleteRequest(mapping application.PortMapping) error {
	if err := validateDeleteProtocol(mapping.Protocol); err != nil {
		return err
	}
	if mapping.ExternalPort < 1 || mapping.ExternalPort > 65535 {
		return fmt.Errorf("external port %d out of range: must be 1-65535", mapping.ExternalPort)
	}
	return nil
}

func validateDeleteProtocol(protocol string) error {
	switch normalizeProtocol(protocol) {
	case "TCP", "UDP":
		return nil
	default:
		return fmt.Errorf("invalid protocol %q: must be TCP or UDP", protocol)
	}
}

func boolValue(v *bool) bool {
	return v != nil && *v
}

func (s *service) resolveInternalIP(providedIP string) (string, error) {
	if ip := strings.TrimSpace(providedIP); ip != "" {
		return ip, nil
	}

	s.mu.RLock()
	gateway := s.gateway
	s.mu.RUnlock()

	if gateway == nil {
		return "", errNoGateway
	}

	u, err := url.Parse(gateway.ControlURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse gateway control URL: %w", err)
	}

	// ゲートウェイのホストに向けて一時的にUDPソケットを開くことでローカルIPを特定
	conn, err := net.Dial("udp", u.Host)
	if err != nil {
		// 失敗時はフォールバックとして最初の非ループバックIPv4を探す
		return s.fallbackLocalIP()
	}
	defer conn.Close()

	localAddr, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok || localAddr.IP.IsUnspecified() {
		return s.fallbackLocalIP()
	}

	return localAddr.IP.String(), nil
}

func (s *service) fallbackLocalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip != nil && !ip.IsLoopback() && ip.To4() != nil {
				return ip.String(), nil
			}
		}
	}
	return "", fmt.Errorf("no suitable local IP address found")
}

func (s *service) syncActivePorts(mapper PortMapper, localIP string) {
	if mapper == nil || localIP == "" {
		return
	}

	var syncedPorts []application.PortMapping
	for i := 0; ; i++ {
		entry, err := mapper.GetGenericPortMappingEntry(i)
		if err != nil {
			// インデックス範囲外など、ルーターがこれ以上エントリーを持っていない場合は走査を終了
			if s.logger != nil {
				s.logger.Debug("finished fetching port mapping entries from router", "index", i, "error", err)
			}
			break
		}

		// このローカルPCのIPアドレス宛の転送ルールのみを抽出
		if entry.InternalIP == localIP {
			syncedPorts = append(syncedPorts, entry)
		}

		// 安全のための上限 (無限ループ防止)
		if i >= 256 {
			break
		}
	}

	s.mu.Lock()
	for _, entry := range syncedPorts {
		s.ports = upsertMapping(s.ports, entry)
	}
	s.mu.Unlock()
}
