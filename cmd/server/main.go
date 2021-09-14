package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"time"

	"github.com/niroopreddym/interceptors-grpc-go/pb"
	"github.com/niroopreddym/interceptors-grpc-go/service"
	"github.com/niroopreddym/interceptors-grpc-go/store"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

const (
	secretKey     = "secret"
	tokenDuration = 15 * time.Minute
)

func seedUsers(userStore store.UserStore) error {
	err := createUser(userStore, "admin1", "secret", "admin")
	if err != nil {
		return err
	}

	return createUser(userStore, "user1", "secret", "user")
}

func createUser(userStore store.UserStore, userName string, password string, role string) error {
	user, err := store.NewUser(userName, password, role)
	if err != nil {
		return err
	}

	return userStore.Save(user)
}

func accessibleRoles() map[string][]string {
	const laptopServicePath = " /pb.LaptopService"
	return map[string][]string{
		laptopServicePath + "CreateLaptop": {"admin"},
		laptopServicePath + "UploadImage":  {"admin"},
		laptopServicePath + "RateLaptop":   {"admin", "user"},
	}
}

func loadTLSCredentials() (credentials.TransportCredentials, error) {
	//load server certificate and private key
	serverCert, err := tls.LoadX509KeyPair("cert/server-cert.pem", "cert/server-key.pem")
	if err != nil {
		return nil, err
	}

	//load certificate of the CA who signed clients's certificate
	pemClientCA, err := ioutil.ReadFile("cert/ca-cert.pem")
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemClientCA) {
		return nil, fmt.Errorf("failed to add client CA's certificate")
	}

	//create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{
			serverCert,
		},
		// //this line is for server side tls alone
		// ClientAuth: tls.NoClientCert,
		ClientAuth: tls.RequireAndVerifyClientCert,
		ClientCAs:  certPool,
	}

	return credentials.NewTLS(config), nil
}

func main() {
	port := flag.Int("port", 0, "the server port")
	flag.Parse()
	log.Printf("satrted the server on port %d", *port)

	userStore := store.NewInMemoryUserStore()
	err := seedUsers(userStore)
	if err != nil {
		log.Fatal("cannot seed users")
	}

	jwtManager := service.NewJWTManager(secretKey, tokenDuration)
	authServer := service.NewAuthServer(userStore, jwtManager)

	imageStore := store.NewDiskImageStore("C:/Users/maneti.n/go/src/github.com/niroopreddym/interceptors-grpc-go/tmp")
	ratingStore := store.NewInMemoryRatingStore()
	laptopServer := service.NewLaptopServer(store.NewInMemoryLaptopStore(), &imageStore, ratingStore)

	tlsCredentials, err := loadTLSCredentials()
	if err != nil {
		log.Fatal("cannot load tls credentials: ", err)
	}

	interceptor := service.NewAuthInterceptor(jwtManager, accessibleRoles())

	grpcServer := grpc.NewServer(
		grpc.Creds(tlsCredentials),
		grpc.UnaryInterceptor(interceptor.Unary()),
		grpc.StreamInterceptor(interceptor.Stream()),
	)

	pb.RegisterAuthServiceServer(grpcServer, authServer)
	pb.RegisterLaptopServiceServer(grpcServer, laptopServer)

	reflection.Register(grpcServer)

	address := fmt.Sprintf("127.0.0.1:%d", *port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("cannot start the tcp listner: %w", err)
	}

	err = grpcServer.Serve(listener)
	if err != nil {
		log.Fatal("cannot start server: ", err)
	}
}
