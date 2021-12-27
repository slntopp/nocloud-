package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/structpb"

	sppb "github.com/slntopp/nocloud/pkg/services_providers/proto"
	"github.com/spf13/viper"
)

func init() {
	viper.AutomaticEnv()
	// viper.SetDefault("DB_HOST", "localhost:8080")
	viper.SetDefault("DB_GRPC_PORT", "8000")
	viper.SetDefault("SECURE", true)
}
func bdgrpcClient() {

	host := viper.GetString("DB_GRPC_PORT")

	var opts []grpc.DialOption

	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	// opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())

	log.Printf("Connect to server %s...\n", host)
	// conn, err := grpc.Dial("127.0.0.1:"+host, opts...)
	conn, err := grpc.Dial("localhost:"+host, opts...)
	if err != nil {
		log.Fatal("fail to dial:", err)
	}
	defer conn.Close()

	log.Println("Connected to server", host)

	client := sppb.NewServicesProvidersExtentionsServiceClient(conn)
	fp := structpb.NewStringValue("419d3335b2b533526d4e7f6f1041b3c492d086cad0f5876739800ffd51659545")
	h := structpb.NewStringValue("demo.ione.local")

	data, _ := structpb.NewStruct(map[string]interface{}{"hostname": h, "fingerprint": fp})

	req := sppb.ServicesProvidersExtentionData{
		Uuid: "-", Data: data,
	}

	//res, err := client.Register(context.Background(), &req)
	res, err := client.Update(context.Background(), &req)
	// res, err := client.Unregister(context.Background(), &req)
	if err != nil {
		log.Printf("could not SendData: %v", err)
	}
	fmt.Println(res)
}

// func grpcClient() {

// 	host := viper.GetString("TUNNEL_HOST")

// 	var opts []grpc.DialOption
// 	// opts = append(opts, grpc.WithInsecure())
// 	// if viper.GetBool("SECURE") {
// 	// 	cred := credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
// 	// 	opts[0] = grpc.WithTransportCredentials(cred)
// 	// }

// 	if viper.GetBool("SECURE") {
// 		// Load client cert
// 		//cert, err := tls.LoadX509KeyPair("cert/0client.crt", "cert/0client.key")
// 		cert, err := tls.LoadX509KeyPair("cert/1client.crt", "cert/1client.key")
// 		if err != nil {
// 			log.Fatal("fail to LoadX509KeyPair:", err)
// 		}

// 		// // Load CA cert
// 		// //Certification authority, CA
// 		// //A CA certificate is a digital certificate issued by a certificate authority (CA), so SSL clients (such as web browsers) can use it to verify the SSL certificates sign by this CA.
// 		// caCert, err := ioutil.ReadFile("../cert/cacerts.cer")
// 		// if err != nil {
// 		// 	log.Fatal(err)
// 		// }
// 		// caCertPool := x509.NewCertPool()
// 		// caCertPool.AppendCertsFromPEM(caCert)

// 		// Setup HTTPS client
// 		config := &tls.Config{
// 			Certificates: []tls.Certificate{cert},
// 			// RootCAs:            caCertPool,
// 			// InsecureSkipVerify: false,
// 			InsecureSkipVerify: true,
// 		}
// 		// config.BuildNameToCertificate()
// 		cred := credentials.NewTLS(config)

// 		// cred := credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
// 		// opts[0] = grpc.WithTransportCredentials(cred)
// 		opts = append(opts, grpc.WithTransportCredentials(cred))
// 	} else {
// 	}

// 	opts = append(opts, grpc.WithInsecure())
// 	opts = append(opts, grpc.WithBlock())

// 	conn, err := grpc.Dial(host, opts...)
// 	if err != nil {
// 		log.Fatal("fail to dial:", err)
// 	}
// 	defer conn.Close()

// 	log.Println("Connected to server", host)

// 	client := pb.DataBaseClient()

// 	stdreader := bufio.NewReader(os.Stdin)

// 	for {
// 		fmt.Print("c2s > ")
// 		note, _ := stdreader.ReadString('\n')

// 		// if err := stream.Send(&pb.StreamData{Message: note}); err != nil {
// 		// 	lg.Fatal("Failed to send a note:", zap.Error(err))

// 		// req, err := client.ScalarSendData(context.Background(), &pb.HttpReQuest2Loc{})
// 		// if err != nil {
// 		// 	log.Printf("could not SendData: %v", err)
// 		// }
// 		// log.Printf("Greeting c: %v", req.GetJson())

// 		req, err := client.Add(context.Background(), &pb.HostFingerprint{Host: note, Fingerprint: note})
// 		if err != nil {
// 			log.Printf("could not SendData: %v", err)
// 		}
// 		log.Printf("Greeting c: %v", req.Sucsess)
// 	}

// }

func restClient(url string, message string) {

	// stdreader := bufio.NewScanner(os.Stdin)
	// fmt.Print("c2s > ")
	// for stdreader.Scan() {

	// // localAddr, err := net.ResolveIPAddr("ip", "localhost")
	// localAddr, err := net.ResolveIPAddr("ip", "127.0.0.1")
	// if err != nil {
	// 	fmt.Println("Failed to read responce", err)
	// 	return
	// }

	// //=========================
	// http.DefaultTransport.(*http.Transport).DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
	// 	dialer := &net.Dialer{
	// 		// Resolver: &net.Resolver{ //Resolver not work!
	// 		// 	PreferGo: true,
	// 		// 	Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
	// 		// 		d := net.Dialer{
	// 		// 			Timeout: time.Duration(5) * time.Second,
	// 		// 		}
	// 		// 		return d.DialContext(ctx, "tcp", "localhost")
	// 		// 	},
	// 		// },
	// 	}

	// 	return dialer.DialContext(ctx, network, "localhost")
	// }
	// netClient := &http.Client{}

	//=========================
	netClient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			// DialContext: (&net.Dialer{
			// 		LocalAddr: &net.TCPAddr{
			// 			// IP: net.ParseIP("127.0.0.1"), //localhost
			// 			// IP: net.ParseIP("localhost"),
			// 			IP:localAddr.IP,
			// 			Port: 8000,
			// 		},
			// 	Timeout:   30 * time.Second,
			// 	KeepAlive: 30 * time.Second,
			// 	DualStack: true,
			// }).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			// DisableCompression:    true,
		},
		Timeout: time.Second * 30,
	}

	netClient.Transport.(*http.Transport).DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		// if addr == "google.com:443" {
		//     addr = "216.58.198.206:443"
		addr = "localhost:80" //missing port in address
		// }
		dialer := &net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
			DualStack: true,
		}
		return dialer.DialContext(ctx, network, addr)
	}

	// req, err := http.NewRequest("GET", "http://node.com/sometestpass/"+stdreader.Text(), nil)
	// req, err := http.NewRequest("GET", strings.Replace("https://reqbin.com/echo", "https://", "http://", 1), nil)

	// req, err := http.NewRequest("POST",
	// 	strings.Replace("https://reqbin.com/echo/post/json", "https://", "http://", 1),
	// 	bytes.NewBuffer([]byte(`{
	// 	"Id": 12345,
	// 	"Customer": "John Smith",
	// 	"Quantity": 1,
	// 	"Price": 10.00
	//   }`)))

	// req, err := http.NewRequest("PUT",
	// 	strings.Replace("https://reqbin.com/echo/put/json", "https://", "http://", 1),
	// 	bytes.NewBuffer([]byte(`{
	// 	  "Id": 12345,
	// 	  "Customer": "John Smithkkkk",
	// 	  "Quantity": 1,
	// 	  "Price": 10.00
	// 	}`)))

	// req, err := http.NewRequest("DELETE",
	// 	strings.Replace("https://reqbin.com/sample/delete/json", "https://", "http://", 1),
	// 	bytes.NewBuffer([]byte(`{
	// 	  "Id": 12345,
	// 	  "Customer": "John Smithkkkk",
	// 	  "Quantity": 1,
	// 	  "Price": 10.00
	// 	}`)))

	// req, err := http.NewRequest("GET", strings.Replace("https://httpbin.org/status/500", "https://", "http://", 1), nil)
	// req, err := http.NewRequest("GET", strings.Replace("https://httpbin.org/delay/9", "https://", "http://", 1), nil)
	req, err := http.NewRequest("GET", strings.Replace(url, "https://", "http://", 1), nil)

	if err != nil {
		fmt.Println("Failed to get http", err)
		return
	}

	req.Header = http.Header{
		// "Host": []string{"www.host.com"},
		"Content-Type": []string{"application/json"},
		// "Authorization": []string{"Bearer Token"},
	}

	fmt.Println(message)
	response, err := netClient.Do(req)

	//=========================

	//TODO Resolver
	// fmt.Print("pppp2s > ")
	// response, err := netClient.Get("http://ione-cloud.net/")
	// response, err := netClient.Get("http://zero.client.net/sometestpass/" + stdreader.Text())
	// response, err := http.Get("http://localhost/sometestpass/" + stdreader.Text())
	if err != nil {
		fmt.Println("Failed to get http", err)
		return
	}

	//тестирование на несколько клиентов
	//Тестирование POST, REST запросов
	//Статусы клиента
	//install Redis
	fmt.Println("Long response.Status", response.Status)

	body1, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Failed to read responce")
		return
	}
	fmt.Printf("Long response: %v\n", string(body1))

	// sb := stripRegex(string(body1))
	// // fmt.Println(sb, in.Message)
	// index1 := strings.Index(sb, stdreader.Text())
	// if 0 < index1 {
	// 	sb = sb[index1 : index1+20]
	// } else {
	// 	sb = "Text not found!"
	// }

	// fmt.Printf("response: %v\n", string(sb))

	// fmt.Print("c2s > ")

	// }
}

func main() {
	bdgrpcClient()
	// // grpcClient()
	// url := "https://httpbin.org/delay/9"
	// go restClient(url,"Send long get")
	// time.Sleep(1 * time.Second)
	// // <-time.After(3 * time.Second)
	// url1 := "https://httpbin.org/status/500"
	// restClient(url1,"Send short get")

	// time.Sleep(11 * time.Second)

	// // httpJSONClient()

}
