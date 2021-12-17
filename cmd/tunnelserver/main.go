package main

import (
	"context"
	"crypto/tls"
	"net"

	pb "github.com/slntopp/nocloud-tunnel-mesh/pkg/proto"
	"github.com/slntopp/nocloud-tunnel-mesh/pkg/tserver"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	log *zap.Logger
)

func init() {
	viper.AutomaticEnv()

	log = nocloud.NewLogger()

	viper.SetDefault("GRPC_PORT", "8080")
}

func main() {
	log.Info("Starting Tunnel Server service")

	port := viper.GetString("GRPC_PORT")
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal("failed to listen:", zap.Error(err))
	}
	
	var opts []grpc.ServerOption

	//openssl req -new -newkey rsa:4096 -x509 -sha256 -days 30 -nodes -out server.crt -keyout server.key
	cert, err := tls.LoadX509KeyPair("/cert/server.crt", "/cert/server.key")
	if err != nil {
		log.Fatal("server: loadkeys:", zap.Error(err))
	}

	// Note if we don't tls.RequireAnyClientCert client side certs are ignored.
	config := &tls.Config{
		Certificates: []tls.Certificate{cert},
		// ClientCAs:    caCertPool,//It work! Peer client sertificates autenification
		ClientAuth: tls.RequireAnyClientCert,
		// VerifyPeerCertificate: is called only after normal certificate verification https://pkg.go.dev/crypto/tls#Config
		// InsecureSkipVerify: false,
		InsecureSkipVerify: true,
	}
	cred := credentials.NewTLS(config)

	opts = append(opts, grpc.Creds(cred))
	
	grpcServer := grpc.NewServer(opts...)
	server := tserver.NewTunnelServer(log)
	server.LoadHostFingerprintsFromDB()
	pb.RegisterSocketConnectionServer(grpcServer, server)

	srv := server.StartHttpServer()
	defer srv.Shutdown(context.TODO())

	log.Info("gRPC-Server Listening on 0.0.0.0:", zap.String("port", port), zap.Skip())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("failed to serve grpc:", zap.Error(err))
	}
}
