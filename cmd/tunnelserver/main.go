package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
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
	viper.SetDefault("SECURE", true)
}

type tunnelServerParams struct {
	stream pb.Tunnel_InitTunnelServer
	ch     chan error
}

type tunnelServer struct {
	pb.UnimplementedTunnelServer

	tsps map[string]tunnelServerParams
}

func (s *tunnelServer) SendData(ctx context.Context, req *pb.SendDataRequest) (*pb.SendDataResponse, error) {

	// fmt.Println(s.tsps)
	tsp, ok := s.tsps[req.GetHost()]
	if !ok {
		lg.Error("Connection not found", zap.String("GetHost", req.GetHost()), zap.String("f", "SendData"))
		return nil, status.Error(codes.NotFound, "Connection not found")
	}
	err := tsp.stream.Send(&pb.StreamData{Message: req.GetMessage()})
	if err != nil {
		lg.Error("Failed to send a note:", zap.String("GetHost", req.GetHost()), zap.String("f", "SendData"))
		tsp.ch <- err
		return nil, status.Error(codes.Internal, err.Error())
	}

	in, _ := tsp.stream.Recv()
	if err == io.EOF {
		lg.Error("Connection lost", zap.String("GetHost", req.GetHost()), zap.String("f", "SendData"))
		tsp.ch <- nil
		return nil, status.Error(codes.NotFound, "Connection lost")
	}
	if err != nil {
		lg.Error("stream.Recv()", zap.String("GetHost", req.GetHost()), zap.String("f", "SendData"))
		tsp.ch <- err
		return nil, err
	}

	return &pb.SendDataResponse{Result: "true\n" == in.Host}, nil

}

func (s *tunnelServer) InitTunnel(stream pb.Tunnel_InitTunnelServer) error {

	in, err := stream.Recv()
	if err == io.EOF {
		lg.Error("lost conn InitTunnel:", zap.Error(err), zap.String("f", "InitTunnel"))
		return nil
	}
	if err != nil {
		lg.Error("InitTunnel:", zap.Error(err), zap.String("f", "InitTunnel"))
		return err
	}

	lg.Info("Hello from client to server!", zap.String("note", in.Host), zap.Skip())

	s.tsps[in.Host] = tunnelServerParams{stream, make(chan error)}
	defer delete(s.tsps, in.Host)

	return <-s.tsps[in.Host].ch
}

func newServer() *tunnelServer {
	s := &tunnelServer{tsps: make(map[string]tunnelServerParams)}
	return s
}

// wrappedStream wraps around the embedded grpc.ServerStream, and intercepts the RecvMsg and
// SendMsg method call.
type wrappedStream struct {
	grpc.ServerStream
}

func newWrappedStream(s grpc.ServerStream) grpc.ServerStream {

	// https://stackoverflow.com/questions/27006725/check-fingerprints-of-server-ssl-tls-certificates-in-http-newrequest

	fmt.Println("s:", s)
	return &wrappedStream{s}
}

func streamInterceptor(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	// authentication (token verification)
	md, _ := metadata.FromIncomingContext(ss.Context())

	// 	md: map[:authority:[localhost:8080] content-type:[application/grpc] user-agent:[grpc-go/1.42.0]]
	// info: &{/tunnel.Tunnel/InitTunnel true true}
	fmt.Println("md:", md)
	fmt.Println("info:", info)
	// if !ok {
	// 	return errMissingMetadata
	// }
	// if !valid(md["authorization"]) {
	// 	return errInvalidToken
	// }

	handler(srv, newWrappedStream(ss))
	// err := handler(srv, newWrappedStream(ss))
	// if err != nil {
	// 	logger("RPC failed with error %v", err)
	// }
	return nil
}

func main() {
	port := viper.GetString("PORT")
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		lg.Fatal("failed to listen:", zap.Error(err))
	}

	var opts []grpc.ServerOption

	if viper.GetBool("SECURE") {
		//openssl req -new -newkey rsa:4096 -x509 -sha256 -days 30 -nodes -out server.crt -keyout server.key
		cert, err := tls.LoadX509KeyPair("cert/server.crt", "cert/server.key")
		if err != nil {
			lg.Fatal("server: loadkeys:", zap.Error(err))
		}
		// Note if we don't tls.RequireAnyClientCert client side certs are ignored.
		config := &tls.Config{
			Certificates: []tls.Certificate{cert},
			ClientAuth: tls.RequireAnyClientCert,
			// ClientAuth:   tls.RequireAndVerifyClientCert,
			// InsecureSkipVerify: false,
			InsecureSkipVerify: true,
		}
		cred := credentials.NewTLS(config)

		// cred := credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
		opts = append(opts, grpc.Creds(cred))
		// // Enable TLS for all incoming connections.
		// opts = append(opts, grpc.Creds(credentials.NewServerTLSFromCert(&cert)))
		// failed to complete security handshake on grpc?
		// https://stackoverflow.com/questions/43829022/failed-to-complete-security-handshake-on-grpc
	}

	// opts = append(opts, grpc.StreamInterceptor(streamInterceptor))

	grpcServer := grpc.NewServer(opts...)
	//r.TLS.PeerCertificates[0].Subject.CommonName

	pb.RegisterTunnelServer(grpcServer, newServer())
	// pb.RegisterTunnelServer(grpcServer, &tunnelServer{})
	lg.Info("gRPC-Server Listening on localhost:", zap.String("port", port), zap.Skip())
	if err := grpcServer.Serve(lis); err != nil {
		lg.Fatal("failed to serve:", zap.Error(err))
	}
}
