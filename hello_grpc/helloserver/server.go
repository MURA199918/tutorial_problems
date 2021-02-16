package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"../hellopb"
	"google.golang.org/grpc"
)

type server struct {
}

func (*server) Hello(ctx context.Context, request *hellopb.HelloRequest) (*hellopb.HelloResponse, error) {
	name := request.Name
	response := &hellopb.HelloResponse{
		Greeting: "Hello " + name,
	}
	return response, nil
}

func main() {
	address := "0.0.0.0:50051"
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("Error ", err)
	}
	fmt.Printf("Server is listening on ", address)

	s := grpc.NewServer()
	hellopb.RegisterHelloServiceServer(s, &server{})
	s.Serve(lis)
}
