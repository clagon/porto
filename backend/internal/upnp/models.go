package upnp

import "github.com/clagon/port-mapper/backend/internal/domain"

// PortMapping は、ルーターに対するポート転送の要求情報を表すドメイン型（domain.PortMapping）のエイリアスです。
type PortMapping = domain.PortMapping

// MaxLeaseDurationSeconds は、ポートマッピングの最大リース期間の定数エイリアスです。
const MaxLeaseDurationSeconds = domain.MaxLeaseDurationSeconds

// DiscoveryResult は、探索されたルーター制御エンドポイント情報を表すドメイン型（domain.DiscoveryResult）のエイリアスです。
type DiscoveryResult = domain.DiscoveryResult
