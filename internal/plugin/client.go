package plugin

import (
	"context"

	"github.com/clintjedwards/hclvet/internal/plugin/proto"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

// GRPCClient represents the implementation for a client that can talk to plugins. The client
// in this case is the main HCLvet process.
type GRPCClient struct{ client proto.HCLvetRulePluginClient }

// GRPCClient is the client implementation that allows our main process to send RPCs to plugins.
func (p *HCLvetRulePlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &GRPCClient{client: proto.NewHCLvetRulePluginClient(c)}, nil
}

// Below are wrappers for how plugins should respond to the RPC in question
// They are all pretty simple since the general flow is to just call the implementation
// of the rpc method for that specific plugin and return the result

// ExecuteRule calls the corresponding ExecuteRule on the plugin through the GRPC client
func (m *GRPCClient) ExecuteRule(request *proto.ExecuteRuleRequest) (*proto.ExecuteRuleResponse, error) {
	response, err := m.client.ExecuteRule(context.Background(), request)
	if err != nil {
		return &proto.ExecuteRuleResponse{}, err
	}
	return response, nil
}

// GetRuleInfo calls the corresponding GetRuleInfo method on the plugin through the GRPC client
func (m *GRPCClient) GetRuleInfo(request *proto.GetRuleInfoRequest) (*proto.GetRuleInfoResponse, error) {
	response, err := m.client.GetRuleInfo(context.Background(), request)
	if err != nil {
		return &proto.GetRuleInfoResponse{}, err
	}
	return response, nil
}
