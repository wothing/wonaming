package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	wonaming "github.com/wothing/wonaming/consul"
	"github.com/wothing/wonaming/example/pb"
)

var (
	serv = flag.String("service", "hello_service", "service name")
	port = flag.Int("port", 1701, "listening port")
	cons = flag.String("consul", "127.0.0.1:8500", "consul address")
)

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", *port))
	if err != nil {
		panic(err)
	}

	err = wonaming.Register(*serv, *port, *cons, time.Second*3, "5s")
	if err != nil {
		panic(err)
	}

	log.Printf("starting hello service at %d", *port)
	s := grpc.NewServer()
	pb.RegisterHelloServiceServer(s, &helloServer{})
	s.Serve(lis)
}

type helloServer struct {
}

func (helloServer) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	log.Printf("getting request from client.\n")
	return &pb.HelloResponse{Reply: "Hello, " + req.Greeting}, nil
}
