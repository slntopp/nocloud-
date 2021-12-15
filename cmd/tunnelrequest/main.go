package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/spf13/viper"
)

func init() {
	viper.AutomaticEnv()
	viper.SetDefault("TUNNEL_HOST", "localhost:8080")
	viper.SetDefault("SECURE", true)
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
// 		opts = append(opts, grpc.WithInsecure())
// 	}

// 	opts = append(opts, grpc.WithBlock())

// 	conn, err := grpc.Dial(host, opts...)
// 	if err != nil {
// 		log.Fatal("fail to dial:", err)
// 	}
// 	defer conn.Close()

// 	log.Println("Connected to server", host)

// 	client := pb.NewTunnelClient(conn)

// 	stdreader := bufio.NewReader(os.Stdin)

// 	for {
// 		fmt.Print("c2s > ")
// 		note, _ := stdreader.ReadString('\n')

// 		// if err := stream.Send(&pb.StreamData{Message: note}); err != nil {
// 		// 	lg.Fatal("Failed to send a note:", zap.Error(err))

// 		r, err := client.SendData(context.Background(), &pb.SendDataRequest{Host: "ClientZero", Message: note})
// 		if err != nil {
// 			log.Printf("could not SendData: %v", err)
// 		}
// 		log.Printf("Greeting c: %v", r.GetResult())

// 	}

// }
func stripRegex(in string) string {
	reg, _ := regexp.Compile("[^a-zA-Z0-9 <>()]+")
	return reg.ReplaceAllString(in, "")
}

func restClient() {

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
		},
		Timeout: time.Second * 30,
	}

	netClient.Transport.(*http.Transport).DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		// if addr == "google.com:443" {
		//     addr = "216.58.198.206:443"
		addr = "localhost"
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
	req, err := http.NewRequest("GET", strings.Replace("https://httpbin.org/delay/9", "https://", "http://", 1), nil)

	if err != nil {
		fmt.Println("Failed to get http", err)
		return
	}

	req.Header = http.Header{
		// "Host": []string{"www.host.com"},
		"Content-Type": []string{"application/json"},
		// "Authorization": []string{"Bearer Token"},
	}

	fmt.Println("Send long get")
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

func restClient2() {

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
		},
		Timeout: time.Second * 30,
	}

	netClient.Transport.(*http.Transport).DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		// if addr == "google.com:443" {
		//     addr = "216.58.198.206:443"
		addr = "localhost"
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
	// 	strings.Replace("https://httpbin.org/put", "https://", "http://", 1),
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

	req, err := http.NewRequest("GET", strings.Replace("https://httpbin.org/status/404", "https://", "http://", 1), nil)
	// req, err := http.NewRequest("GET", strings.Replace("https://httpbin.org/delay/9", "https://", "http://", 1), nil)

	if err != nil {
		fmt.Println("Failed to get http", err)
		return
	}

	req.Header = http.Header{
		// "Host": []string{"www.host.com"},
		"Content-Type": []string{"application/json"},
		// "Authorization": []string{"Bearer Token"},
	}

	fmt.Println("Send short get")
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
	//Asynch Requests
	//Function add/delete hosts
	fmt.Println("Short response.Status", response.Status)

	body1, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Failed to read responce")
		return
	}
	fmt.Printf("Short response: %v\n", string(body1))

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
	// grpcClient()

	go restClient()
	time.Sleep(1 * time.Second)
	// <-time.After(3 * time.Second)
	restClient2()
	time.Sleep(11 * time.Second)
		

	// httpJSONClient()
}

// func httpJSONClient() {

// 	// req, err := http.NewRequest("GET", "https://ione-cloud.net/?ggg=fff", nil)
// 	req, err := http.NewRequest("POST", "https://ione-cloud.net/?ggg=fff", bytes.NewBuffer([]byte("Hello!")))
// 	if err != nil {
// 		fmt.Println("http.NewRequest", err)
// 		//Handle Error
// 	}

// 	req.Header = http.Header{
// 		"Host":          []string{"www.host.com"}, //not use!
// 		"Content-Type":  []string{"application/json"},
// 		"Authorization": []string{"Bearer Token"},
// 	}

// 	// req.Method
// 	// req.Host
// 	// req.URL
// 	// req.Header
// 	// req.Body
// 	body, _ := io.ReadAll(req.Body)
// 	req_struct := pb.HttpData{
// 		Method: req.Method,
// 		URL:    *req.URL,
// 		Header: req.Header,
// 		Body:   body,
// 	}

// 	jsonValue, _ := json.Marshal(req_struct)

// 	var h pb.HttpData
// 	err = json.Unmarshal(jsonValue, &h)
// 	// err := json.NewDecoder(jsonValue).Decode(&h)
// 	if err != nil {
// 		fmt.Println("json.Unmarshal", err)
// 	}

// 	fmt.Println(h.URL, h.Body)

// 	req1, err := http.NewRequest(req.Method, req.URL.String(), req.Body)
// 	if err != nil {
// 		fmt.Println("http.NewRequest", err)
// 	}

// 	netClient := &http.Client{
// 		Timeout: time.Second * 10,
// 	}

// 	response, err := netClient.Do(req1)
// 	if err != nil {
// 		fmt.Println("nc.Do", err)
// 	}

// 	bodyS, err := ioutil.ReadAll(response.Body)
// 	if err != nil {
// 		fmt.Println("http.NewRequest", err)
// 	}

// 	fmt.Println("response.Status", response.Status)
// 	i, err := strconv.Atoi(response.Status[0:3])
// 	if err != nil {
// 		// handle error
// 		fmt.Println(err)
// 	}

// 	resp_struct := pb.HttpData{
// 		Status: i,
// 		Method: response.Request.Method,
// 		URL:    *response.Request.URL,
// 		Header: response.Header,
// 		Body:   bodyS,
// 	}

// 	jsonResp, _ := json.Marshal(resp_struct)

// 	var resp pb.HttpData
// 	err = json.Unmarshal(jsonResp, &resp)
// 	// err := json.NewDecoder(jsonValue).Decode(&h)
// 	if err != nil {
// 		fmt.Println("json.Unmarshal", err)
// 	}

// 	fmt.Println("http.NewRequest", resp_struct.Status, string(resp.Body[:40]))

// 	// req1, err := http.NewRequest("GET", "https://ione-cloud.net/", nil)
// 	// if err != nil {
// 	// 	fmt.Println("http.NewRequest", err)
// 	// 	//Handle Error
// 	// }

// 	// jsonValue, _ := json.Marshal(req)

// 	// // netClient.httpRequest1()

// 	// fmt.Printf("response: %v\n", string(jsonValue))

// 	// stdreader := bufio.NewScanner(os.Stdin)
// 	// fmt.Print("c2s > ")
// 	// for stdreader.Scan() {

// 	// 	// netClient := &http.Client{
// 	// 	// 	Timeout: time.Second * 30,
// 	// 	// }

// 	// 	req, err := http.NewRequest("GET", "https://ione-cloud.net/", nil)
// 	// 	if err != nil {
// 	// 		fmt.Println("http.NewRequest", err)
// 	// 		//Handle Error
// 	// 	}

// 	// 	req.Header = http.Header{
// 	// 		"Host":          []string{"www.host.com"},
// 	// 		"Content-Type":  []string{"application/json"},
// 	// 		"Authorization": []string{"Bearer Token"},
// 	// 	}

// 	// 	jsonValue, _ := json.Marshal(req)

// 	// 	// netClient.httpRequest1()

// 	// 	fmt.Printf("response: %v\n", string(jsonValue))
// 	// 	fmt.Print("c2s > ")

// 	// }
// }
