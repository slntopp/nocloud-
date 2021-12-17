package main

import (
	"context"
	"crypto/tls"
	"io"
	"time"

	pb "github.com/slntopp/nocloud-tunnel-mesh/pkg/proto"
	"github.com/slntopp/nocloud-tunnel-mesh/pkg/tclient"
	"github.com/slntopp/nocloud/pkg/nocloud"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var (
	log *zap.Logger

	host             string
	secure           bool
)

func init() {
	viper.AutomaticEnv()

	log = nocloud.NewLogger()

	viper.SetDefault("TUNNEL_HOST", "localhost:8080")
	viper.SetDefault("DESTINATION_HOST", "ione")
	viper.SetDefault("SECURE", true)

	host = viper.GetString("TUNNEL_HOST")
	secure = viper.GetBool("SECURE")
}

func main() {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	if secure {
		// Load client cert
		cert, err := tls.LoadX509KeyPair("/cert/client.crt", "/cert/client.key")
		if err != nil {
			log.Fatal("fail to LoadX509KeyPair:", zap.Error(err))
		}

		// Setup HTTPS client
		config := &tls.Config{
			Certificates: []tls.Certificate{cert},
			// InsecureSkipVerify: false,
			InsecureSkipVerify: true,
		}
		cred := credentials.NewTLS(config)

		// cred := credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
		opts[0] = grpc.WithTransportCredentials(cred)
	}

	opts = append(opts, grpc.WithBlock())

	//Reconnection
	for {

		func() {
			defer time.Sleep(5 * time.Second)

			log.Info("Try to connect...", zap.String("host", host), zap.Skip())

			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()

			conn, err := grpc.DialContext(ctx, host, opts...)
			if err != nil {
				log.Error("fail to dial:", zap.Error(err))
				return
			}
			defer conn.Close()

			client := pb.NewSocketConnectionClient(conn)

			stream, err := client.InitConnection(context.Background())
			if err != nil {
				log.Error("Failed InitTunnel", zap.Error(err))
				return
			}

			log.Info("Connected to server:", zap.String("host", host), zap.Skip())

			for {
				in, err := stream.Recv()
				if err == io.EOF {
					log.Info("Connection closed", zap.String("Message", in.Message), zap.Skip())
					return
				}
				if err != nil {
					log.Error("Failed to receive a note:", zap.Error(err))
					return
				}

				log.Debug("Received request from server:", zap.String("Message", in.Message), zap.Skip())

				//TODO send errors from httpClient
				go tclient.HttpClient(log, stream, in.Message, in.Id, in.Json)
			}
		}()

	}
}
