package plugin

import (
	"context"

	"github.com/clintjedwards/hclvet/internal/plugin/proto"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

// GRPCServer is the implementation that allows the plugin to respond to requests from the main process.
type GRPCServer struct {
	proto.UnimplementedHCLvetRulePluginServer
	Impl RuleDefinition
}

// GRPCServer is the server implementation that allows our plugins to receive RPCs.
func (p *HCLvetRulePlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	proto.RegisterHCLvetRulePluginServer(s, &GRPCServer{Impl: p.Impl})
	return nil
}

// Below are wrappers for how plugins should respond to the RPC in question
// They are all pretty simple since the general flow is to just call the implementation
// of the rpc method for that specific plugin and return the result

// ExecuteRule executes a single rule on a plugin
func (m *GRPCServer) ExecuteRule(ctx context.Context, request *proto.ExecuteRuleRequest) (*proto.ExecuteRuleResponse, error) {
	response, err := m.Impl.ExecuteRule(request)
	return response, err
}

// GetRuleInfo gets information about the plugin
func (m *GRPCServer) GetRuleInfo(ctx context.Context, request *proto.GetRuleInfoRequest) (*proto.GetRuleInfoResponse, error) {
	response, err := m.Impl.GetRuleInfo(request)
	return response, err
}
