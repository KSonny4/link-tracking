package main

import (
	"io/ioutil"
	"net"
	"os"

	"google.golang.org/grpc"

	"google.golang.org/grpc/grpclog"

	"github.com/ksonny4/link-tracking/gateway"

	pb "github.com/ksonny4/link-tracking/proto"
	"github.com/ksonny4/link-tracking/server"
)



func main() {
	var ss = server.New()

	// Run HTTP server
	go func() {
		ss.StartHTTPServer()
	}()

	// Adds gRPC internal logs. This is quite verbose, so adjust as desired!
	log := grpclog.NewLoggerV2(os.Stdout, ioutil.Discard, ioutil.Discard)
	grpclog.SetLoggerV2(log)

	addr := "0.0.0.0:10000"
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln("Failed to listen:", err)
	}

	s := grpc.NewServer()
	pb.RegisterTrackerServer(s, ss)
	// Serve gRPC Server
	log.Info("Serving gRPC on http://", addr)
	go func() {
		log.Fatal(s.Serve(lis))
	}()

	err = gateway.Run("dns:///" + addr)
	log.Fatalln(err)
}
