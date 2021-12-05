package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/slntopp/nocloud-tunnel-mesh/pkg/proto"
)

var (
	port = flag.Int("port", 10000, "The server port")
)

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

func (s *tunnelServer) InitTunnel(req *pb.InitTunnelRequest, stream pb.Tunnel_InitTunnelServer) error {
	fmt.Println("Got streaming connection request")
	s.conns[req.GetHost()] = stream
	for {
		err := stream.RecvMsg(nil)
		fmt.Printf("Possible Error receiving from Stream: %w\n", err)
		if err == io.EOF {
			delete(s.conns, req.GetHost())
			return nil
		}
		if err != nil {
			delete(s.conns, req.GetHost())
			return err
		}
	}
}

func newServer() *tunnelServer {
	s := &tunnelServer{conns: make(map[string]pb.Tunnel_InitTunnelServer)}
	return s
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	pb.RegisterTunnelServer(grpcServer, newServer())
	fmt.Printf("gRPC-Server Listening on localhost:%d\n", *port)
	grpcServer.Serve(lis)	
}
