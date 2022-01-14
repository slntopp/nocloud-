package tserver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net"

	driver "github.com/arangodb/go-driver"
	tpb "github.com/slntopp/nocloud-tunnel-mesh/pkg/proto"
	pb "github.com/slntopp/nocloud/pkg/services_providers/proto"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const HOSTS_COLLECTION = "TunnelHosts"

type DBServerAPI struct {
	pb.UnimplementedServicesProvidersExtentionsServiceServer
	*TunnelServer
}

func EnsureCollectionExists(logger *zap.Logger, db driver.Database) {
	log := logger.Named("EnsureCollectionExists")

	options := &driver.CreateCollectionOptions{
		KeyOptions: &driver.CollectionKeyOptions{AllowUserKeys: true, Type: "uuid"},
	}
	log.Debug("Checking Collection existence", zap.String("collection", HOSTS_COLLECTION))
	exists, err := db.CollectionExists(context.TODO(), HOSTS_COLLECTION)
	if err != nil {
		log.Fatal("Failed to check collection exists", zap.Error(err))
	}
	if exists {
		return
	}

	log.Debug("Collection doesn't exist, creating")
	_, err = db.CreateCollection(context.TODO(), HOSTS_COLLECTION, options)
	if err != nil {
		log.Fatal("Error creating Collection", zap.String("collection", HOSTS_COLLECTION), zap.Any("options", options), zap.Error(err))
	}
	log.Debug("Collection existence ensured")

}

//Make hash od raw sertificate as sha256
func MakeFingerprint(c []byte) string {
	sum := sha256.Sum256(c)
	return hex.EncodeToString(sum[:])
}

type HostsFingerprintsPair struct {
	Fingerprint string `json:"_key"`
	Host        string `json:"host"`
}

//Gets fingerprint-host chain from DB
func (s *TunnelServer) LoadHostFingerprintsFromDB() {
	log := s.log.Named("LoadHostFingerprintsFromDB")
	// fp := "419d3335b2b533526d4e7f6f1041b3c492d086cad0f5876739800ffd51659545"
	// s.fingerprints_hosts[fp] = "demo.ione.local"
	// s.hosts_fingerprints[fp] ="reqbin.com"
	// s.hosts_fingerprints[fp] ="zero.client.net"
	// s.hosts_fingerprints[fp] ="ione-cloud.net"

	s.mutex.Lock()
	//clear
	s.fingerprints_hosts = make(map[string]string)

	ctx := context.Background()
	query := "FOR d IN TunnelHosts RETURN d"
	cursor, err := s.col.Database().Query(ctx, query, nil)
	if err != nil {
		// handle error
		fmt.Println(err)
	}
	defer cursor.Close()

	for {
		var doc HostsFingerprintsPair
		meta, err := cursor.ReadDocument(ctx, &doc)
		if driver.IsNoMoreDocuments(err) {
			break
		} else if err != nil {
			log.Error("ReadDocument", zap.Error(err))
		}

		s.fingerprints_hosts[doc.Fingerprint] = doc.Host

		log.Info("Load HostFingerprints From DB", zap.String("key", meta.Key))
	}

	s.mutex.Unlock()
}

func (s *DBServerAPI) GetType(ctx context.Context, req *pb.GetTypeRequest) (*pb.GetTypeResponse, error) {
	return &pb.GetTypeResponse{Type: "nocloud-tunnelmesh"}, nil
}

func ExtractDataFromExtData(req *pb.ServicesProvidersExtentionData) (hostname string, fingerprint string, err error) {
	data := req.Data.AsMap()
	if _, ok := data["hostname"]; !ok {
		return "", "", errors.New("Hostname not given(key: hostname)")
	}
	if _, ok := data["fingerprint"]; !ok {
		return "", "", errors.New("Fingerprint not given(key: fingerprint)")
	}
	hostname    = data["hostname"].(string)
	fingerprint = data["fingerprint"].(string)
	return hostname, fingerprint, nil
}

func (s *DBServerAPI) Test(ctx context.Context, req *pb.ServicesProvidersExtentionData) (*pb.GenericResponse, error) {
	hostname, fingerprint, err := ExtractDataFromExtData(req)
	if err != nil {
		return &pb.GenericResponse{Result: false, Error: err.Error()}, nil
	}
	for k, v := range s.fingerprints_hosts {
		if k == fingerprint {
			return &pb.GenericResponse{Result: false, Error: "Fingerprint exists"}, nil
		} else if v == hostname {
			return &pb.GenericResponse{Result: false, Error: "Hostname already registered"}, nil
		}
	}
	return &pb.GenericResponse{Result: true}, nil
}

//Add host/Fingerprint to DB
func (s *DBServerAPI) Register(ctx context.Context, in *pb.ServicesProvidersExtentionData) (*pb.GenericResponse, error) {
	log := s.log.Named("Register")

	hostname, fingerprint, err := ExtractDataFromExtData(in)
	if err != nil {
		return &pb.GenericResponse{Result: false, Error: err.Error()}, nil
	}
	for k, v := range s.fingerprints_hosts {
		if k == fingerprint {
			return &pb.GenericResponse{Result: false, Error: "Fingerprint exists"}, nil
		} else if v == hostname {
			return &pb.GenericResponse{Result: false, Error: "Hostname already registered"}, nil
		}
	}

	doc := HostsFingerprintsPair{
		fingerprint, hostname,
	}

	meta, err := s.col.CreateDocument(context.Background(), doc)
	if err != nil {
		log.Error("Error Creating Document", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error Creating Document")
	}
	log.Debug("Got doc with key from query", zap.String("key", meta.Key))

	s.LoadHostFingerprintsFromDB()
	// s.resetConn(in)
	err = s.WaitForConnection(hostname)
	if err != nil {
		return &pb.GenericResponse{Result: false, Error: err.Error()}, nil
	}
	return &pb.GenericResponse{Result: true}, nil
}

//Edit host/Fingerprint to DB
func (s *DBServerAPI) Update(ctx context.Context, in *pb.ServicesProvidersExtentionData) (*pb.GenericResponse, error) {
	log := s.log.Named("Update")

	hostname, fingerprint, err := ExtractDataFromExtData(in)
	if err != nil {
		return &pb.GenericResponse{Result: false, Error: err.Error()}, nil
	}

	old_host, ok := s.fingerprints_hosts[fingerprint]
	if !ok {
		return &pb.GenericResponse{Result: false, Error: "Hostname isn't registered"}, nil
	}

	for k, v := range s.fingerprints_hosts {
		if v == hostname && k != fingerprint {
			return &pb.GenericResponse{Result: false, Error: "Hostname already registered"}, nil
		}
	}

	patch := map[string]interface{}{
		"host": hostname,
	}

	meta, err := s.col.UpdateDocument(context.Background(), fingerprint, patch)
	if err != nil {
		log.Error("Error Updating Document", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error Updating Document")
	}

	log.Info("Edit doc with key from query", zap.String("key", meta.Key))

	s.LoadHostFingerprintsFromDB()

	req := &tpb.HostFingerprint{
		Fingerprint: fingerprint,
		Host:        old_host,
	}
	s.resetConn(req)

	return &pb.GenericResponse{Result: true}, nil
}

//Delete host/Fingerprint to DB
func (s *DBServerAPI) Unregister(ctx context.Context, in *pb.ServicesProvidersExtentionData) (*pb.GenericResponse, error) {
	log := s.log.Named("Unregister")

	_, fingerprint, err := ExtractDataFromExtData(in)
	if err != nil {
		return &pb.GenericResponse{Result: false, Error: err.Error()}, nil
	}

	old_host, ok := s.fingerprints_hosts[fingerprint]
	if !ok {
		return &pb.GenericResponse{Result: false, Error: "Hostname isn't registered"}, nil
	}

	meta, err := s.col.RemoveDocument(context.Background(), fingerprint)
	if err != nil {
		log.Error("Error Removing Document", zap.Error(err))
		return nil, status.Error(codes.Internal, "Error Removing Document")
	}

	log.Info("Remove doc with key from query", zap.String("key", meta.Key))

	s.LoadHostFingerprintsFromDB()
	
	req := &tpb.HostFingerprint{
		Fingerprint: fingerprint,
		Host:        old_host,
	}
	s.resetConn(req)

	return &pb.GenericResponse{Result: true}, nil
}

//Start grpc server to update fingerprint-host database
func (s *TunnelServer) StartDBgRPCServer() *grpc.Server {
	log := s.log.Named("dbServer")
	port := viper.GetString("DB_GRPC_PORT")

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal("failed to listen:", zap.Error(err))
	}

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)

	pb.RegisterServicesProvidersExtentionsServiceServer(grpcServer, &DBServerAPI{TunnelServer: s})

	go func() {
		log.Info("Start DataBase gRPCServer Listening on 0.0.0.0:", zap.String("port", port), zap.Skip())
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal("failed to serve db grpc:", zap.Error(err))
		}
	}()

	// returning reference so caller can call Shutdown()
	return grpcServer
}
