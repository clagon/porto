package upnp

import (
	"errors"
	"fmt"

	"github.com/clagon/port-mapper/backend/internal/domain"
)

var (
	// ErrNoGateway は、ネットワーク上で適合するUPnPゲートウェイ（ルーター）が発見できなかったことを表すエラーです。
	ErrNoGateway      = domain.ErrNoGateway
	errOnlyWFADevices = errors.New("UPnP discovery found only WPS/WFA devices; no InternetGatewayDevice/WAN service responded")
)

// Discover は、ネットワーク上のすべてのアクティブなネットワークインターフェース（IPv4/IPv6）に対して SSDP M-SEARCH マルチキャスト要求を送信し、
// 最初に応答があり、適合するインターネットゲートウェイ（IGD）を検出・解決して返します。
// SSDP応答がない、またはパースに失敗した場合は、ルーターの既知の制御パスに対するフォールバックプローブ（直接走査）を実行します。
func Discover() (DiscoveryResult, error) {
	ifaces, err := discoverInterfaces()
	if err != nil {
		return DiscoveryResult{}, err
	}

	sawOnlyWFA := false
	var ssdpErr error
	for _, iface := range ifaces {
		result, err := discoverFromInterface(iface)
		if err == nil {
			return result, nil
		}
		if errors.Is(err, errOnlyWFADevices) {
			sawOnlyWFA = true
			continue
		}
		if errors.Is(err, ErrNoGateway) {
			ssdpErr = err
		}
	}

	ipv6Ifaces, err := discoverIPv6Interfaces()
	if err == nil {
		for _, iface := range ipv6Ifaces {
			result, err := discoverFromIPv6Interface(iface)
			if err == nil {
				return result, nil
			}
			if errors.Is(err, errOnlyWFADevices) {
				sawOnlyWFA = true
				continue
			}
			if errors.Is(err, ErrNoGateway) {
				ssdpErr = err
			}
		}
	}

	result, err := probeGatewayControlURLs(ifaces)
	if err == nil {
		return result, nil
	}

	result, err = probeGatewayDescriptions(ifaces)
	if err == nil {
		return result, nil
	}
	return DiscoveryResult{}, discoveryFailureError(ssdpErr, sawOnlyWFA, err)
}

func discoveryFailureError(ssdpErr error, sawOnlyWFA bool, fallbackErr error) error {
	if ssdpErr != nil {
		return fmt.Errorf("%w: SSDP discovery failed: %v; fallback probes failed: %v", ErrNoGateway, ssdpErr, fallbackErr)
	}
	if sawOnlyWFA {
		return fmt.Errorf("%w: %v; fallback probes failed: %v", ErrNoGateway, errOnlyWFADevices, fallbackErr)
	}
	return fmt.Errorf("%w: %v", ErrNoGateway, fallbackErr)
}

func discoverFromInterface(iface discoverInterface) (DiscoveryResult, error) {
	responses, err := collectSSDPResponses(iface)
	if err != nil {
		return DiscoveryResult{}, err
	}
	return discoverFromSSDPResponses(responses, "no matching SSDP responses")
}

func discoverFromIPv6Interface(iface discoverIPv6Interface) (DiscoveryResult, error) {
	responses, err := collectSSDPResponsesIPv6(iface)
	if err != nil {
		return DiscoveryResult{}, err
	}
	return discoverFromSSDPResponses(responses, "no matching IPv6 SSDP responses")
}

func discoverFromSSDPResponses(responses []ssdpResponse, noMatchMessage string) (DiscoveryResult, error) {
	var lastErr error
	wfaCount := 0
	for _, response := range responses {
		// WPS/WFA は UPnP ではあるが、ポートマッピング用 IGD ではないので候補から外す。
		if isWFAResponse(response) {
			wfaCount++
			continue
		}
		result, err := discoverFromLocation(response.Location, defaultLocationFetcher)
		if err == nil {
			return result, nil
		}
		lastErr = fmt.Errorf("SSDP location %q for search target %q and service %q: %w", response.Location, response.SearchTarget, response.ST, err)
	}

	if lastErr != nil {
		return DiscoveryResult{}, fmt.Errorf("%w: %w", ErrNoGateway, lastErr)
	}
	if len(responses) > 0 && wfaCount == len(responses) {
		return DiscoveryResult{}, fmt.Errorf("%w: %w", ErrNoGateway, errOnlyWFADevices)
	}
	return DiscoveryResult{}, errors.New(noMatchMessage)
}
