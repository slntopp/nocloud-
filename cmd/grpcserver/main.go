package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"

	"google.golang.org/grpc"

	pb "github.com/slntopp/nocloud-tunnel-mesh/grpcstream"
)

var (
	port = flag.Int("port", 10000, "The server port")
)

type routeGuideServer struct {
	pb.UnimplementedRouteGuideServer

	mu         sync.Mutex // protects routeNotes
	routeNotes map[string][]*pb.RouteNote
}

// RouteChat receives a stream of message/location pairs, and responds with a stream of all
// previous messages at each of those locations.
func (*routeGuideServer) RouteChat(stream pb.RouteGuide_RouteChatServer) error {

	go func() {
		stdreader := bufio.NewReader(os.Stdin)

		for {
			myFPass, _ := stdreader.ReadString('\n')
			notes := []*pb.RouteNote{
				{Message: myFPass},
			}
			for _, note := range notes {
				if err := stream.Send(note); err != nil {
					log.Fatalf("Failed to send a note: %v", err)
				}
			}

		}
	}()

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		log.Printf("Hello from client to server! %s", in.Message)
		fmt.Print("m2c > ")
	}
}

func newServer() *routeGuideServer {
	s := &routeGuideServer{routeNotes: make(map[string][]*pb.RouteNote)}
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
	pb.RegisterRouteGuideServer(grpcServer, newServer())
	grpcServer.Serve(lis)
}