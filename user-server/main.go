package main

import (
	"context"
	pb "github.com/ergildo/go-lang-grpc-gateway-mysql-example/proto/user"
	"github.com/ergildo/go-lang-grpc-gateway-mysql-example/user-server/model"
	"github.com/ergildo/go-lang-grpc-gateway-mysql-example/user-server/service"
	"github.com/ergildo/go-lang-grpc-gateway-mysql-example/user-server/setup"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
)

const (
	srvAddr = "127.0.0.1:50051"
	gwAddr  = "127.0.0.1:8080"
)

type server struct {
	pb.UnimplementedUserServiceBPServer
}

func (s *server) CreateUser(ctx context.Context, in *pb.NewUserRequest) (*pb.UserResponse, error) {
	user := service.Save(model.User{Name: in.GetName()})
	return &pb.UserResponse{Id: user.Id, Name: user.Name}, nil
}
func (s *server) UpdateUser(ctx context.Context, in *pb.UpdateUserRequest) (*pb.UserResponse, error) {
	user := service.Update(model.User{Id: in.Id, Name: in.Name})
	return &pb.UserResponse{
		Id: user.Id, Name: user.Name,
	}, nil
}
func (s *server) FindUserById(ctx context.Context, in *pb.UserRequest) (*pb.UserResponse, error) {
	user := service.FindById(in.Id)
	return &pb.UserResponse{
		Id: user.Id, Name: user.Name,
	}, nil
}
func (s *server) ListAllUsers(ctx context.Context, in *pb.Void) (*pb.ListAllUsersResponse, error) {
	users := service.ListAll()
	var userResponses []*pb.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, &pb.UserResponse{
			Id: user.Id, Name: user.Name,
		})
	}
	return &pb.ListAllUsersResponse{UserResponse: userResponses}, nil
}
func (s *server) DeleteUser(ctx context.Context, in *pb.UserRequest) (*pb.Void, error) {
	service.Delete(in.Id)
	return &pb.Void{}, nil
}

func main() {
	setup.SetUpDB()
	startGrpcServer()
}

func startGrpcServer() {

	lis, err := net.Listen("tcp", srvAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterUserServiceBPServer(s, &server{})
	log.Println("gRPC server is running at:", lis.Addr())
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalln("could not start gRPC server:", err)
		}
	}()

	conn, err := grpc.DialContext(context.Background(), srvAddr, grpc.WithBlock(), grpc.WithInsecure())
	if err != nil {
		log.Fatalln("could not connect gRPC server:", err)
	}

	gwMux := runtime.NewServeMux()

	err = pb.RegisterUserServiceBPHandler(context.Background(), gwMux, conn)

	if err != nil {
		log.Fatalln("could register gw mux handler:", err)
	}

	gwServer := &http.Server{Addr: gwAddr, Handler: gwMux}

	log.Println("gRPC gateway is running at:", gwServer.Addr)

	if err := gwServer.ListenAndServe(); err != nil {
		log.Fatalln("could not start gRPC Gateway:", err)
	}
}
