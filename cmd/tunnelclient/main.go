package main

import (
	"context"
	"crypto/tls"
	"io"

	"github.com/slntopp/nocloud-tunnel-mesh/pkg/logger"
	pb "github.com/slntopp/nocloud-tunnel-mesh/pkg/proto"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	lg     *zap.Logger
	host   string
	secure bool
)

func init() {
	lg = logger.NewLogger()

	viper.AutomaticEnv()
	viper.SetDefault("TUNNEL_HOST", "localhost:8080")
	viper.SetDefault("SECURE", false)

	host = viper.GetString("TUNNEL_HOST")
	secure = viper.GetBool("SECURE")
}

func main() {

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	if secure {
		// cred := credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})

		// Load client cert
		cert, err := tls.LoadX509KeyPair("../cert/certNEW.pem", "../cert/serverNEW.key")
		if err != nil {
			lg.Fatal("fail to LoadX509KeyPair:", zap.Error(err))
		}

		// // Load CA cert
		// //Certification authority, CA
		// //A CA certificate is a digital certificate issued by a certificate authority (CA), so SSL clients (such as web browsers) can use it to verify the SSL certificates sign by this CA.
		// caCert, err := ioutil.ReadFile("../cert/cacerts.cer")
		// if err != nil {
		// 	log.Fatal(err)
		// }
		// caCertPool := x509.NewCertPool()
		// caCertPool.AppendCertsFromPEM(caCert)

		// Setup HTTPS client
		config := &tls.Config{
			Certificates: []tls.Certificate{cert},
			// RootCAs:            caCertPool,
			InsecureSkipVerify: false,
		}
		config.BuildNameToCertificate()
		cred := credentials.NewTLS(config)

		opts[0] = grpc.WithTransportCredentials(cred)
	}

	opts = append(opts, grpc.WithBlock())

	//Reconnection
	for {

		conn, err := grpc.Dial(host, opts...)
		if err != nil {
			lg.Error("fail to dial:", zap.Error(err))
			continue
		}
		defer conn.Close()

		lg.Info("Connected server", zap.String("host", host), zap.Skip())

		client := pb.NewTunnelClient(conn)

		stream, err := client.InitTunnel(context.Background()) //, &pb.InitTunnelRequest{Host: "testo"})
		if err != nil {
			lg.Error("Failed InitTunnel", zap.Error(err))
			continue
		}

		if err := stream.Send(&pb.InitTunnelRequest{Host: "Hello, server!"}); err != nil {
			lg.Error("Failed to send Hello:", zap.Error(err))
			continue
		}

		for {
			in, err := stream.Recv()
			if err == io.EOF {
				lg.Info("Connection closed", zap.String("TUNNEL_HOST", host), zap.Skip())
				break
			}
			if err != nil {
				lg.Error("Failed to receive a note:", zap.Error(err))
				break
			}

			lg.Info("Received StreamData:"+in.Message, zap.Skip())

			if err := stream.Send(&pb.InitTunnelRequest{Host: in.Message}); err != nil {
				lg.Error("Failed to send a note:", zap.Error(err))
				break
			}
		}

	}
}
