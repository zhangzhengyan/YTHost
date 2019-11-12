package host

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/yottachain/P2PHost/pb"
)

// Server implemented server API for P2PHostServer service.
type Server struct {
	Host Host
}

// ID implemented ID function of P2PHostServer
func (server *Server) ID(ctx context.Context, req *pb.Empty) (*pb.StringMsg, error) {
	return &pb.StringMsg{Value: server.Host.ID()}, nil
}

// Addrs implemented Addrs function of P2PHostServer
func (server *Server) Addrs(ctx context.Context, req *pb.Empty) (*pb.StringListMsg, error) {
	return &pb.StringListMsg{Values: server.Host.Addrs()}, nil
}

// Connect implemented Connect function of P2PHostServer
func (server *Server) Connect(ctx context.Context, req *pb.ConnectReq) (*pb.Empty, error) {
	err := server.Host.Connect(req.GetId(), req.GetAddrs())
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &pb.Empty{}, nil
}

// DisConnect implemented DisConnect function of P2PHostServer
func (server *Server) DisConnect(ctx context.Context, req *pb.StringMsg) (*pb.Empty, error) {
	err := server.Host.DisConnect(req.GetValue())
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &pb.Empty{}, nil
}

// SendMsg implemented SendMsg function of P2PHostServer
func (server *Server) SendMsg(ctx context.Context, req *pb.SendMsgReq) (*pb.SendMsgResp, error) {
	bytes, err := server.Host.SendMsg(req.GetId(), req.GetMsgType(), req.GetMsg())
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &pb.SendMsgResp{Value: bytes}, nil
}

// RegisterHandler implemented RegisterHandler function of P2PHostServer
func (server *Server) RegisterHandler(ctx context.Context, req *pb.StringMsg) (*pb.Empty, error) {
	server.Host.RegisterHandler(req.GetValue(), nil)
	return &pb.Empty{}, nil
}

// UnregisterHandler implemented UnregisterHandler function of P2PHostServer
func (server *Server) UnregisterHandler(ctx context.Context, req *pb.StringMsg) (*pb.Empty, error) {
	server.Host.UnregisterHandler(req.GetValue())
	return &pb.Empty{}, nil
}

// Close implemented Close function of P2PHostServer
func (server *Server) Close(ctx context.Context, req *pb.Empty) (*pb.Empty, error) {
	server.Host.Close()
	return &pb.Empty{}, nil
}
