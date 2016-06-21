package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	// wonaming "github.com/wothing/wonaming/consul"
	wonaming "github.com/wothing/wonaming/etcd"
	"github.com/wothing/wonaming/example/pb"
)

var (
	serv = flag.String("service", "hello_service", "service name")
	port = flag.Int("port", 1701, "listening port")
	// reg  = flag.String("reg", "127.0.0.1:8500", "register address")
	reg  = flag.String("reg", "http://127.0.0.1:2379", "register address")
)

func main() {
	flag.Parse()

	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", *port))
	if err != nil {
		panic(err)
	}

	err = wonaming.Register(*serv, "127.0.0.1", *port, *reg, time.Second*3, 5)
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
