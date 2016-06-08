package main

import (
	"flag"
	"fmt"
	"time"

	wonaming "github.com/wothing/wonaming/consul"
	"github.com/wothing/wonaming/example/pb"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

var (
	serv   = flag.String("service", "hello_service", "service name")
	consul = flag.String("consul", "127.0.0.1:8500", "consul address")
)

func main() {
	r := &wonaming.ConsulResolver{ServiceName: *serv}
	b := grpc.RoundRobin(r)

	conn, err := grpc.Dial(*consul, grpc.WithInsecure(), grpc.WithBalancer(b))
	if err != nil {
		panic(err)
	}

	ticker := time.NewTicker(2 * time.Second)
	for t := range ticker.C {
		client := pb.NewHelloServiceClient(conn)
		resp, err := client.SayHello(context.Background(), &pb.HelloRequest{Greeting: "world"})
		if err != nil {
			panic(err)
		}
		fmt.Printf("%v: Reply is %s\n", t, resp.Reply)
	}
}
