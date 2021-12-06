package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/slntopp/nocloud-tunnel-mesh/pkg/logger"
	pb "github.com/slntopp/nocloud-tunnel-mesh/pkg/proto"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	lg *zap.Logger
)

func init() {
	lg = logger.NewLogger()

	viper.AutomaticEnv()
	viper.SetDefault("PORT", "8080")
}

type tunnelServer struct {
	pb.UnimplementedTunnelServer

	conns map[string]pb.Tunnel_InitTunnelServer
}

func (s *tunnelServer) SendData(ctx context.Context, req *pb.SendDataRequest) (*pb.SendDataResponse, error) {
	stream, ok := s.conns[req.GetHost()]
	if !ok {
		return nil, status.Error(codes.NotFound, "Connection not found")
	}
	err := stream.SendMsg(req.GetMessage())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.SendDataResponse{Result: true}, nil
}

// func (s *tunnelServer) InitTunnel(req *pb.InitTunnelRequest, stream pb.Tunnel_InitTunnelServer) error {
func (*tunnelServer) InitTunnel(stream pb.Tunnel_InitTunnelServer) error {
	fmt.Println("Got streaming connection request")
	// s.conns[req.GetHost()] = stream
	// for {
	// 	err := stream.RecvMsg(nil)
	// 	fmt.Printf("Possible Error receiving from Stream: %w\n", err)
	// 	if err == io.EOF {
	// 		delete(s.conns, req.GetHost())
	// 		return nil
	// 	}
	// 	if err != nil {
	// 		delete(s.conns, req.GetHost())
	// 		return err
	// 	}
	// }

	go func() {
		stdreader := bufio.NewReader(os.Stdin)

		for {
			note, _ := stdreader.ReadString('\n')

			if err := stream.Send(&pb.StreamData{Message: note}); err != nil {
				lg.Fatal("Failed to send a note:", zap.Error(err))
			}

		}
	}()

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		lg.Info("Hello from client to server!", zap.String("note", in.Host), zap.Skip())
		fmt.Print("m2c > ")
	}

}

// func newServer() *tunnelServer {
// 	s := &tunnelServer{conns: make(map[string]pb.Tunnel_InitTunnelServer)}
// 	return s
// }

func main() {
	port := viper.GetString("PORT")
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		lg.Fatal("failed to listen:", zap.Error(err))
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	// pb.RegisterTunnelServer(grpcServer, newServer())
	pb.RegisterTunnelServer(grpcServer, &tunnelServer{})
	lg.Info("gRPC-Server Listening on localhost:", zap.String("port", port), zap.Skip())
	grpcServer.Serve(lis)
}
