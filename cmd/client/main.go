package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/niroopreddym/interceptors-grpc-go/client"
	"github.com/niroopreddym/interceptors-grpc-go/pb"
	"github.com/niroopreddym/interceptors-grpc-go/sample"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	username        = "admin1"
	password        = "secret"
	refreshDuration = 30 * time.Second
)

func authMethods() map[string]bool {
	const laptopServicePath = " /pb.LaptopService"
	return map[string]bool{
		laptopServicePath + "CreateLaptop": true,
		laptopServicePath + "UploadImage":  true,
		laptopServicePath + "RateLaptop":   true,
	}
}

func loadTLSCredentials() (credentials.TransportCredentials, error) {
	//this below code is for mutual authentication
	//load server certificate and private key
	clientCert, err := tls.LoadX509KeyPair("cert/client-cert.pem", "cert/client-key.pem")
	if err != nil {
		return nil, err
	}

	//load certificate of the CA who signed server's certificate
	pemServerCA, err := ioutil.ReadFile("cert/ca-cert.pem")
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}

	//create the creds and return it
	config := &tls.Config{
		RootCAs:      certPool,
		Certificates: []tls.Certificate{clientCert},
	}

	return credentials.NewTLS(config), nil
}

func main() {
	serverAddress := flag.String("address", "", "the server address")
	flag.Parse()
	log.Printf("dial server %s", *serverAddress)

	tlsCredentials, err := loadTLSCredentials()
	if err != nil {
		log.Fatal("cannot  load TLS credentials: ", err)
	}

	cc1, err := grpc.Dial(*serverAddress, grpc.WithTransportCredentials(tlsCredentials))
	if err != nil {
		log.Fatal("cannot dial server: ", err)
	}

	// use the above connection to ocnnect to auth client
	authClient := client.NewAuthClient(cc1, username, password)
	interceptor, err := client.NewAuthInterceptor(authClient, authMethods(), refreshDuration)

	if err != nil {
		log.Fatal("cannot create auth interceptor: ", err)
	}

	cc2, err := grpc.Dial(*serverAddress, grpc.WithTransportCredentials(tlsCredentials), grpc.WithUnaryInterceptor(interceptor.Unary()), grpc.WithStreamInterceptor(interceptor.Stream()))
	if err != nil {
		log.Fatal("cannot dial server: ", err)
	}

	laptopClient := client.NewLaptopClient(cc2)
	// testSearchLaptop(laptopClient)
	// testUploadImage(laptopClient)
	testRateLaptop(laptopClient)
}

func testCreateLaptop(laptopClient client.LaptopClient) {
	laptopClient.CreateLaptop(sample.NewLaptop())
}

func testSearchLaptop(laptopClient client.LaptopClient) {
	for i := 0; i < 10; i++ {
		laptopClient.CreateLaptop(sample.NewLaptop())
	}
	filter := &pb.Filter{
		MaxPriceUsd: 30000,
		MinCpuCores: 1,
		MinCpuGhz:   1.0,
		MinRam: &pb.Memory{
			Value: 1,
			Unit:  pb.Unit_GIGABTE,
		},
	}

	laptopClient.SearchLaptop(filter)
}

func testUploadImage(laptopClient client.LaptopClient) {
	laptop := sample.NewLaptop()
	laptopClient.CreateLaptop(sample.NewLaptop())
	laptopClient.UploadImage(laptop.GetId(), "C:/Users/maneti.n/go/src/github.com/niroopreddym/interceptors-grpc-go/tmp/laptop.png")
}

func testRateLaptop(laptopClient *client.LaptopClient) {
	n := 3
	laptopIDs := make([]string, n)

	for i := 0; i < 3; i++ {
		laptop := sample.NewLaptop()
		laptopIDs[i] = laptop.GetId()
		laptopClient.CreateLaptop(sample.NewLaptop())
	}

	scores := make([]float64, n)
	for {
		fmt.Print("rate laptop (y/n)?")
		var answer string
		fmt.Scan(&answer)
		if strings.ToLower(answer) != "y" {
			break
		}

		for i := 0; i < n; i++ {
			scores[i] = sample.RandomLaptopScore()
		}

		err := laptopClient.RateLaptop(laptopIDs, scores)
		if err != nil {
			log.Fatal(err)
		}

	}
}
