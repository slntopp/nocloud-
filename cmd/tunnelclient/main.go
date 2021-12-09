package main

import (
	"context"
	"crypto/tls"
	"io"
	"time"

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
	viper.SetDefault("SECURE", true)

	host = viper.GetString("TUNNEL_HOST")
	secure = viper.GetBool("SECURE")
}

func main() {

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	if secure {
		// Load client cert
		cert, err := tls.LoadX509KeyPair("cert/0client.crt", "cert/0client.key")
		// cert, err := tls.LoadX509KeyPair("cert/1client.crt", "cert/1client.key")
		if err != nil {
			lg.Fatal("fail to LoadX509KeyPair:", zap.Error(err))
		}

		// Setup HTTPS client
		config := &tls.Config{
			Certificates: []tls.Certificate{cert},
			// InsecureSkipVerify: false,
			InsecureSkipVerify: true,
		}
		// config.BuildNameToCertificate()
		cred := credentials.NewTLS(config)

		// cred := credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
		opts[0] = grpc.WithTransportCredentials(cred)
	}

	opts = append(opts, grpc.WithBlock())
	opts = append(opts,  grpc.WithTimeout(15 * time.Second))

	//Reconnection
	for {

		func() {
			defer time.Sleep(5 * time.Second)

			lg.Info("Try to connect...", zap.String("host", host), zap.Skip())
	
			conn, err := grpc.Dial(host, opts...)
			if err != nil {
				lg.Error("fail to dial:", zap.Error(err))
				return
			}
			defer conn.Close()

			lg.Info("Connected:", zap.String("host", host), zap.Skip())

			client := pb.NewTunnelClient(conn)

			stream, err := client.InitTunnel(context.Background()) //, &pb.InitTunnelRequest{Host: "testo"})
			if err != nil {
				lg.Error("Failed InitTunnel", zap.Error(err))
				return
			}

			if err := stream.Send(&pb.InitTunnelRequest{Host: "ClientZero"}); err != nil { //todo Clientname?
				lg.Error("Failed to send Hello:", zap.Error(err))
				return
			}

			for {
				in, err := stream.Recv()
				if err == io.EOF {
					lg.Info("Connection closed", zap.String("TUNNEL_HOST", host), zap.Skip())
					return
				}
				if err != nil {
					lg.Error("Failed to receive a note:", zap.Error(err))
					return
				}

				lg.Info("Received StreamData:"+in.Message, zap.Skip())

				if err := stream.Send(&pb.InitTunnelRequest{Host: in.Message}); err != nil {
					lg.Error("Failed to send a note:", zap.Error(err))
					return
				}
			}
		}()

	}
}
