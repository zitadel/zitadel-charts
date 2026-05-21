package support

import wenv "github.com/mridang/wilhelm/env"

// Env wraps Wilhelm's Env and adds Zitadel-specific CRD getters
// (Gateway API, ServiceMonitor) not covered by Wilhelm's generated dispatch.
type Env struct {
	*wenv.Env
}
