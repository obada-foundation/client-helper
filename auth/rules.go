package auth

import (
	_ "embed"
)

// These the current set of rules we have for auth.
const (
	RuleAuthenticate = "auth"
	RuleAny          = "allowAny"
	RuleUserOnly     = "allowOnlyUser"
)

// Package name of our rego code.
const (
	opaPackage string = "obada.rego"
)

// Core OPA policies.
var (
	//go:embed rego/authentication.rego
	opaAuthentication string

	//go:embed rego/authorization.rego
	opaAuthorization string
)
