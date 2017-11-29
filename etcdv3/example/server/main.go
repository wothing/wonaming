/**
 * Copyright 2015-2017, Wothing Co., Ltd.
 * All rights reserved.
 *
 * Created by elvizlai on 2017/11/28 11:47.
 */

package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"../../../etcdv3"
	"../pb"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const svcName = "project/test"

var addr = "127.0.0.1:50051"

func main() {
	flag.StringVar(&addr, "addr", addr, "addr to lis")
	flag.Parse()

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %s", err)
	}
	defer lis.Close()

	s := grpc.NewServer()
	defer s.GracefulStop()

	pb.RegisterHelloServiceServer(s, &hello{})

	go etcdv3.Register(":2379", svcName, addr, 5)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)
	go func() {
		s := <-ch
		etcdv3.UnRegister(svcName, addr)

		if i, ok := s.(syscall.Signal); ok {
			os.Exit(int(i))
		} else {
			os.Exit(0)
		}

	}()

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}
}

type hello struct {
}

func (*hello) Echo(ctx context.Context, req *pb.Payload) (*pb.Payload, error) {
	req.Data = req.Data + ", from:" + addr
	return req, nil
}
