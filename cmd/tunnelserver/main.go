package main

import (
	"context"
	"crypto/tls"
	"net"
	"time"

	pb "github.com/slntopp/nocloud-tunnel-mesh/pkg/proto"
	"github.com/slntopp/nocloud-tunnel-mesh/pkg/tserver"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/slntopp/nocloud/pkg/nocloud/connectdb"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"
)

var (
	port string
	log  *zap.Logger

	arangodbHost      string
	arangodbCred      string
	keepalive_ping    int
	keepalive_timeout int
)

func init() {
	viper.AutomaticEnv()

	log = nocloud.NewLogger()

	viper.SetDefault("DB_HOST", "localhost:8529")
	viper.SetDefault("DB_CRED", "root:openSesame")
	viper.SetDefault("GRPC_PORT", "8080")
	viper.SetDefault("DB_GRPC_PORT", "8000")
	viper.SetDefault("KEEPALIVE_PINGS_EVERY", "5")
	viper.SetDefault("KEEPALIVE_TIMEOUT", "2")

	arangodbHost = viper.GetString("DB_HOST")
	arangodbCred = viper.GetString("DB_CRED")
	port = viper.GetString("GRPC_PORT")
	keepalive_ping = viper.GetInt("KEEPALIVE_PINGS_EVERY")
	keepalive_timeout = viper.GetInt("KEEPALIVE_TIMEOUT")
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
	cert, err := tls.LoadX509KeyPair("/cert/tls.crt", "/cert/tls.key")
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

	var kaep = keepalive.EnforcementPolicy{
		MinTime:             time.Duration(keepalive_ping) * time.Second, // If a client pings more than once every 5 seconds, terminate the connection
		PermitWithoutStream: true,                                         // Allow pings even when there are no active streams           // send pings even without active streams
	}

	opts = append(opts, grpc.KeepaliveEnforcementPolicy(kaep))

	var kasp = keepalive.ServerParameters{
		MaxConnectionIdle:     0,                                              // 15 * time.Second, // If a client is idle for 15 seconds, send a GOAWAY
		MaxConnectionAge:      0,                                              //30 * time.Second, // If any connection is alive for more than 30 seconds, send a GOAWAY
		MaxConnectionAgeGrace: 0,                                              //15 * time.Second,  // Allow 5 seconds for pending RPCs to complete before forcibly closing connections
		Time:                  time.Duration(keepalive_ping) * time.Second,    // send pings every keepalive_ping seconds if there is no activity
		Timeout:               time.Duration(keepalive_timeout) * time.Second, // wait timeout second for ping back
	}

	opts = append(opts, grpc.KeepaliveParams(kasp))

	opts = append(opts, grpc.Creds(cred))

	grpcServer := grpc.NewServer(opts...)
	server := tserver.NewTunnelServer(log, db)
	server.LoadHostFingerprintsFromDB()
	pb.RegisterSocketConnectionServer(grpcServer, server)

	srv := server.StartHttpServer()
	defer srv.Shutdown(context.TODO())

	dbsrv := server.StartDBgRPCServer()
	defer dbsrv.Stop()

	log.Info("gRPC-Server Listening on 0.0.0.0:", zap.String("port", port), zap.Skip())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatal("failed to serve grpc:", zap.Error(err))
	}
}
