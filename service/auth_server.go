package service

import (
	"context"

	"github.com/niroopreddym/interceptors-grpc-go/pb"
	"github.com/niroopreddym/interceptors-grpc-go/store"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//AuthServer auth server implements the authentication validation
type AuthServer struct {
	UserStore  store.UserStore
	jwtmanager *JWTManager
}

//NewAuthServer constructor for the new auth server
func NewAuthServer(userStore store.UserStore, jwtmanager *JWTManager) *AuthServer {
	return &AuthServer{
		jwtmanager: jwtmanager,
		UserStore:  userStore,
	}
}

//Login logs the user into the system via unary RPC
func (server *AuthServer) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	user, err := server.UserStore.Find(req.GetUsername())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot find user : %v", err)
	}

	if user == nil || !user.IsCorrectPassword(req.GetPassword()) {
		return nil, status.Errorf(codes.NotFound, "incorrect username/password")
	}

	token, err := server.jwtmanager.Generate(user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "cannot generate access token")
	}

	res := &pb.LoginResponse{
		AccessToken: token,
	}

	return res, nil
}
