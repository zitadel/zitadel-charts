// Package assert provides partial assertion helpers for Kubernetes resource
// types. Assertion structs use Opt[T] to mark which fields should be compared,
// skipping any field left at zero value.
//
//go:generate go run ../../internal/gen/assertgen -out zz_generated.go -package assert
package assert
