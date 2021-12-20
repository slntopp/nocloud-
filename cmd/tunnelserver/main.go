package main

import (
	"context"
	"crypto/tls"
	"net"

	pb "github.com/slntopp/nocloud-tunnel-mesh/pkg/proto"
	"github.com/slntopp/nocloud-tunnel-mesh/pkg/tserver"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/connectdb"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	port string
	log *zap.Logger

	arangodbHost 	string
	arangodbCred 	string
)

func init() {
	viper.AutomaticEnv()

	log = nocloud.NewLogger()

	viper.SetDefault("DB_HOST", "db:8529")
	viper.SetDefault("DB_CRED", "root:openSesame")
	viper.SetDefault("GRPC_PORT", "8080")

	arangodbHost 	= viper.GetString("DB_HOST")
	arangodbCred 	= viper.GetString("DB_CRED")
	port = viper.GetString("GRPC_PORT")
}

func main() {
	defer func() {
		_ = log.Sync()
	}()
	log.Info("Starting Tunnel Server service")

	log.Info("Setting up DB Connection")
	db := connectdb.MakeDBConnection(log, arangodbHost, arangodbCred)
	log.Info("DB connection established")

	log.Info("Checking collection")
	tserver.EnsureCollectionExists(log, db)
	log.Info("Collection checked")

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
	server := tserver.NewTunnelServer(log, db)
	server.LoadHostFingerprintsFromDB()
	pb.RegisterSocketConnectionServer(grpcServer, server)

	srv := server.StartHttpServer()
	defer srv.Shutdown(context.TODO())

	log.Info("gRPC-Server Listening on 0.0.0.0:", zap.String("port", port), zap.Skip())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("failed to serve grpc:", zap.Error(err))
	}
}
