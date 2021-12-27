package tserver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net"

	driver "github.com/arangodb/go-driver"
	pb "github.com/slntopp/nocloud-tunnel-mesh/pkg/proto"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const HOSTS_COLLECTION = "TunnelHosts"

type DBServerAPI struct {
	pb.UnimplementedDataBaseServer
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

//Add host/Fingerprint to DB
func (s *DBServerAPI) Add(ctx context.Context, in *pb.HostFingerprint) (*pb.HostFingerprintResp, error) {
	log := s.log.Named("AddFingerprintsToDB")

	for k, v := range s.fingerprints_hosts {
		if k == in.Fingerprint {
			return &pb.HostFingerprintResp{Message: in.Fingerprint + " exists!", Sucsess: false}, nil
		} else if v == in.Host {
			return &pb.HostFingerprintResp{Message: in.Host + " exists!", Sucsess: false}, nil
		}
	}

	doc := HostsFingerprintsPair{
		Fingerprint: in.Fingerprint,
		Host:        in.Host,
	}

	meta, err := s.col.CreateDocument(context.Background(), doc)
	if err != nil {
		log.Error("CreateDocument", zap.Error(err))
		return &pb.HostFingerprintResp{Message: in.Host + "CreateDocument error", Sucsess: false}, err
	}
	log.Info("Got doc with key from query", zap.String("key", meta.Key))

	s.LoadHostFingerprintsFromDB()
	// s.resetConn(in)

	return &pb.HostFingerprintResp{Message: in.Host + " added", Sucsess: true}, nil
}

//Edit host/Fingerprint to DB
func (s *DBServerAPI) Edit(ctx context.Context, in *pb.HostFingerprint) (*pb.HostFingerprintResp, error) {
	log := s.log.Named("EditFingerprintsToDB")

	old_host, ok := s.fingerprints_hosts[in.Fingerprint]
	if !ok {
		return &pb.HostFingerprintResp{Message: in.Host + " not exist!", Sucsess: false}, nil
	}

	for k, v := range s.fingerprints_hosts {
		if v == in.Host && k != in.Fingerprint {
			return &pb.HostFingerprintResp{Message: in.Host + " exists!", Sucsess: false}, nil
		}
	}

	patch := map[string]interface{}{
		"host": in.Host,
	}

	meta, err := s.col.UpdateDocument(context.Background(), in.Fingerprint, patch)
	if err != nil {
		log.Error("UpdateDocument", zap.Error(err))
		return &pb.HostFingerprintResp{Message: in.Host + "UpdateDocument error", Sucsess: false}, err
	}

	log.Info("Edit doc with key from query", zap.String("key", meta.Key))

	s.LoadHostFingerprintsFromDB()

	in = &pb.HostFingerprint{
		Fingerprint: in.Fingerprint,
		Host:        old_host,
	}
	s.resetConn(in)

	return &pb.HostFingerprintResp{Message: in.Host + " edited", Sucsess: true}, nil
}

//Delete host/Fingerprint to DB
func (s *DBServerAPI) Delete(ctx context.Context, in *pb.HostFingerprint) (*pb.HostFingerprintResp, error) {
	log := s.log.Named("DeleteFingerprintsToDB")

	old_host, ok := s.fingerprints_hosts[in.Fingerprint]
	if !ok {
		return &pb.HostFingerprintResp{Message: in.Host + " not exist!", Sucsess: false}, nil
	}

	meta, err := s.col.RemoveDocument(context.Background(), in.Fingerprint)
	if err != nil {
		log.Error("RemoveDocument", zap.Error(err))
		return &pb.HostFingerprintResp{Message: in.Host + "RemoveDocument error", Sucsess: false}, err
	}

	log.Info("Remove doc with key from query", zap.String("key", meta.Key))

	s.LoadHostFingerprintsFromDB()
	
	in = &pb.HostFingerprint{
		Fingerprint: in.Fingerprint,
		Host:        old_host,
	}
	s.resetConn(in)


	return &pb.HostFingerprintResp{Message: in.Host + " deleted", Sucsess: true}, nil
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

	pb.RegisterDataBaseServer(grpcServer, &DBServerAPI{TunnelServer: s})

	go func() {
		log.Info("Start DataBase gRPCServer Listening on 0.0.0.0:", zap.String("port", port), zap.Skip())
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatal("failed to serve db grpc:", zap.Error(err))
		}
	}()

	// returning reference so caller can call Shutdown()
	return grpcServer
}
