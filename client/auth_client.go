package client

import (
	"context"
	"time"

	"github.com/niroopreddym/interceptors-grpc-go/pb"
	"google.golang.org/grpc"
)

//AuthClient to authenticate
type AuthClient struct {
	service  pb.AuthServiceClient
	username string
	password string
}

//NewAuthClient is the ctor
func NewAuthClient(cc *grpc.ClientConn, username string, password string) *AuthClient {
	service := pb.NewAuthServiceClient(cc)
	return &AuthClient{
		service:  service,
		username: username,
		password: password,
	}
}

//Login logs in and returns the access token
func (client *AuthClient) Login() (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req := &pb.LoginRequest{
		Username: client.username,
		Password: client.password,
	}

	res, err := client.service.Login(ctx, req)
	if err != nil {
		return "", err
	}

	return res.GetAccessToken(), nil
}
