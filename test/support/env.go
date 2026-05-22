package support

import wenv "github.com/mridang/wilhelm/env"

// Env is a thin wrapper around Wilhelm's Env, providing a Zitadel-specific
// handle for namespace-scoped test helpers.
type Env struct {
	*wenv.Env
}
