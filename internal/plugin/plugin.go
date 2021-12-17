// Package plugin orchestrates the plugin relationship between rules and the hclvet process
//
// It uses hashicorp's go-plugin which implements a local client/server relationship with
// plugins and allows communication to the plugins over grpc.
//
// To read more about hashicorp's plugin system see here:
// https://github.com/hashicorp/go-plugin/blob/master/docs/internals.md
package plugin

import (
	"github.com/clintjedwards/hclvet/internal/plugin/proto"
	"github.com/hashicorp/go-plugin"
)

// TODO(clintjedwards): Clean up and document all of this

// This file contains structures that both the plugin and the plugin host has to implement

// Handshake is a common handshake that is shared by plugin and host.
// If any of the below values do not match for the plugin being run, the handshake will fail.
// This means that the handshake acts as a type of versioning to instantly deprecate plugin
// apis that will no longer work.
//
// More documentation on the HandshakeConfig here:
// https://pkg.go.dev/github.com/hashicorp/go-plugin#HandshakeConfig
var Handshake = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "HCLVET_PLUGIN",
	MagicCookieValue: "26pGPy",
}

// RuleDefinition is the interface in which both the plugin and the host has to implement
type RuleDefinition interface {
	ExecuteRule(request *proto.ExecuteRuleRequest) (*proto.ExecuteRuleResponse, error)
	GetRuleInfo(request *proto.GetRuleInfoRequest) (*proto.GetRuleInfoResponse, error)
}

// HCLvetRulePlugin is just a wrapper so we implement the correct go-plugin interface
// it allows us to serve/consume the plugin
type HCLvetRulePlugin struct {
	plugin.Plugin
	Impl RuleDefinition
}
