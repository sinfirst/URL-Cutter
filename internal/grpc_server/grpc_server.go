package grpc_server

import (
	"context"
	"fmt"

	pb "github.com/sinfirst/URL-Cutter/proto/url_cutter"
)

type URLCutterServer struct {
	pb.UnimplementedURLCutterServer
}

func NewURLCutterServer() pb.URLCutterServer {
	return &URLCutterServer{}

}
func (s *URLCutterServer) PostHandler(ctx context.Context, in *pb.PostHandlerRequest) (*pb.PostHandlerResponse, error) {
	fmt.Println(1)
	return nil, nil
}
