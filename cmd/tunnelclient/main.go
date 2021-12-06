package main

import (
	"context"
	"fmt"
	"io"

	"github.com/slntopp/nocloud-tunnel-mesh/pkg/logger"
	pb "github.com/slntopp/nocloud-tunnel-mesh/pkg/proto"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	lg *zap.Logger
)

func init() {
	lg = logger.NewLogger()

	viper.AutomaticEnv()
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("HOST", "localhost")
}

func main() {

	host := viper.GetString("HOST")
	port := viper.GetString("PORT")

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())

	conn, err := grpc.Dial(host+":"+port, opts...)
	if err != nil {
		lg.Fatal("fail to dial:", zap.Error(err))
	}
	defer conn.Close()

	lg.Info("Connected server", zap.String("host", host), zap.String("port", port), zap.Skip())

	client := pb.NewTunnelClient(conn)

	stream, err := client.InitTunnel(context.Background()) //, &pb.InitTunnelRequest{Host: "testo"})
	if err != nil {
		lg.Fatal("Failed InitTunnel", zap.Error(err))
	}
	// if err != nil {
	// 	panic(err)
	// }
	// for {
	// 	var msg string
	// 	err := stream.RecvMsg(&msg)
	// 	if err == io.EOF {
	// 		return
	// 	}
	// 	if err != nil {
	// 		panic(err)
	// 	}
	// }

	if err := stream.Send(&pb.InitTunnelRequest{Host: "Hello, server!"}); err != nil {
		lg.Fatal("Failed to send Hello:", zap.Error(err))
	}

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			lg.Info(fmt.Sprintf("Connection closed %s:%s", host, port), zap.Skip())
			return
		}
		if err != nil {
			lg.Fatal("Failed to receive a note:", zap.Error(err))
		}

		lg.Info("Received StreamData:"+in.Message, zap.Skip())

		if err := stream.Send(&pb.InitTunnelRequest{Host: in.Message}); err != nil {
			lg.Fatal("Failed to send a note:", zap.Error(err))
		}
	}

}
