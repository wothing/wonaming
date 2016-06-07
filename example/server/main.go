package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	consul "github.com/hashicorp/consul/api"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

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

	// self-registe service to consul
	conf := &consul.Config{Scheme: "http", Address: *cons}
	client, err := consul.NewClient(conf)
	if err != nil {
		log.Fatalf("connecting to consul '%s': %s", *cons, err)
	}
	//generate id
	regis := &consul.AgentServiceRegistration{
		ID: fmt.Sprintf("%s-127.0.0.1-%d", *serv, *port)
		Name: *serv,
		Address: "127.0.0.1",
		Port: *port,
	}
	err = client.Agent().ServiceRegister(regis)
	if err != nil {
		log.Fatalf("registering to consul error: %s", err)
	}

	log.Printf("starting hello service at %d", *port)
	s := grpc.NewServer()
	pb.RegisterHelloServiceServer(s, &helloServer{})
	s.Serve(lis)
}

type helloServer struct {
}

func (helloServer) SayHello(ctx context.Context, req *pb.HelloRequest) (*pb.HelloResponse, error) {
	fmt.Printf("getting request from client.\n")
	return &pb.HelloResponse{Reply: "Hello, " + req.Greeting}, nil
}
