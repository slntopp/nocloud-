package main

import (
	"context"
	"crypto/tls"
	"io"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
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
	viper.SetDefault("SECURE", false)
}

type tunnelServer struct {
	pb.UnimplementedTunnelServer

	conns map[string]pb.Tunnel_InitTunnelServer

	c chan error
}

func (s *tunnelServer) SendData(ctx context.Context, req *pb.SendDataRequest) (*pb.SendDataResponse, error) {

	stream, ok := s.conns[req.GetHost()]
	if !ok {
		return nil, status.Error(codes.NotFound, "Connection not found")
	}
	err := stream.Send(&pb.StreamData{Message: req.GetMessage()})
	if err != nil {
		s.c <- err
		return nil, status.Error(codes.Internal, err.Error())
	}

	in, _ := s.conns["0"].Recv()
	if err == io.EOF {
		s.c <- nil
		return nil, status.Error(codes.NotFound, "Connection lost")
	}
	if err != nil {
		s.c <- err
		return nil, err
	}

	return &pb.SendDataResponse{Result: "true\n" == in.Host}, nil

	// return &pb.SendDataResponse{Result: true}, nil

	// ms := req.GetMessage()
	// log.Printf("Greetings s: %v", ms)
	// if "true\n" == ms {

	// 	return &pb.SendDataResponse{Result: true}, nil
	// } else {
	// 	return &pb.SendDataResponse{Result: false}, nil

	// }

	// s.conns["0"].Send(&pb.StreamData{Message: req.GetMessage()})
	// in, _ := s.conns["0"].Recv()
	// // if err == io.EOF {
	// // 	return nil
	// // }
	// // if err != nil {
	// // 	return err
	// // }

	// return &pb.SendDataResponse{Result: "true\n" == in.Host}, nil

}

// func (s *tunnelServer) InitTunnel(req *pb.InitTunnelRequest, stream pb.Tunnel_InitTunnelServer) error {
func (s *tunnelServer) InitTunnel(stream pb.Tunnel_InitTunnelServer) error {

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

	// fmt.Println("Got streaming connection request")
	// go func() {
	// 	stdreader := bufio.NewReader(os.Stdin)

	// 	for {
	// 		note, _ := stdreader.ReadString('\n')

	// 		if err := stream.Send(&pb.StreamData{Message: note}); err != nil {
	// 			lg.Fatal("Failed to send a note:", zap.Error(err))
	// 		}
	// 		log.Printf("Greetings s: %v", s.conns)

	// 	}
	// }()

	in, err := stream.Recv()
	if err == io.EOF {
		return nil
	}
	if err != nil {
		return err
	}

	s.conns["0"] = stream
	s.c = make(chan error)

	lg.Info("Hello from client to server!", zap.String("note", in.Host), zap.Skip())

	return <-s.c
}

func newServer() *tunnelServer {
	s := &tunnelServer{conns: make(map[string]pb.Tunnel_InitTunnelServer)}
	return s
}

func main() {
	port := viper.GetString("PORT")
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		lg.Fatal("failed to listen:", zap.Error(err))
	}

	var opts []grpc.ServerOption

	if viper.GetBool("SECURE") {
		cert, err := tls.LoadX509KeyPair("certs/server.pem", "certs/server.key")
		if err != nil {
			log.Fatalf("server: loadkeys: %s", err)
		}
		// Note if we don't tls.RequireAnyClientCert client side certs are ignored.
		config := &tls.Config{
			Certificates: []tls.Certificate{cert},
			ClientAuth:   tls.RequireAnyClientCert,
			// ClientAuth:   tls.RequireAndVerifyClientCert,
			InsecureSkipVerify: false,
		}
		cred := credentials.NewTLS(config)

		// cred := credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
		opts = append(opts, grpc.WithTransportCredentials(cred))
	}

	grpcServer := grpc.NewServer(opts...)

	pb.RegisterTunnelServer(grpcServer, newServer())
	// pb.RegisterTunnelServer(grpcServer, &tunnelServer{})
	lg.Info("gRPC-Server Listening on localhost:", zap.String("port", port), zap.Skip())
	if err := grpcServer.Serve(lis); err != nil {
		lg.Fatal("failed to serve:", zap.Error(err))
	}
}
