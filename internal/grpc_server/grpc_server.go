package grpcserver

import (
	"context"
	"errors"
	"fmt"

	"github.com/sinfirst/URL-Cutter/internal/handlers"
	pb "github.com/sinfirst/URL-Cutter/proto/url_cutter"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type URLCutterServer struct {
	pb.UnimplementedURLCutterServer
	logger  zap.SugaredLogger
	handler handlers.Handler
	userId  int
}

func NewURLCutterServer(logger zap.SugaredLogger, handler handlers.Handler) pb.URLCutterServer {
	return &URLCutterServer{logger: logger, handler: handler, userId: 0}

}

func (s *URLCutterServer) BatchShortenURL(ctx context.Context, in *pb.BatchShortenURLRequest) (*pb.BatchShortenURLResponse, error) {
	var responces []*pb.ShortenResponseForBatch
	for _, req := range in.ShortenRequests {
		shortURL, err := s.handler.Shortener(ctx, req.OriginalURL, 0)
		if err != nil {
			s.logger.Errorw("Problem with set in storage", err)
			return nil, status.Error(codes.Internal, "Server problem")
		}
		responces = append(responces, &pb.ShortenResponseForBatch{
			CorrelationID: req.CorrelationID,
			ShortURL:      shortURL,
		})
	}
	return &pb.BatchShortenURLResponse{ShortenResponses: responces}, status.Error(codes.OK, "OK")
}

func (s *URLCutterServer) GetHandler(ctx context.Context, in *pb.GetHandlerRequest) (*pb.GetHandlerResponse, error) {
	url, err := s.handler.GetHandler(ctx, in.Id)
	if err != nil {
		return nil, status.Error(codes.NotFound, "URL not found")
	}
	return &pb.GetHandlerResponse{OrigURL: url}, status.Error(codes.OK, "OK")
}
func (s *URLCutterServer) PostHandler(ctx context.Context, in *pb.PostHandlerRequest) (*pb.PostHandlerResponse, error) {
	var id int
	if in.UserID == -1 {
		id = s.userId + 1
		s.userId++
	} else {
		id = int(in.UserID)
	}

	shortURL, err := s.handler.Shortener(ctx, in.Url, id)
	if errors.Is(err, fmt.Errorf("conflict")) {
		return &pb.PostHandlerResponse{ShortUrl: shortURL, UserID: int64(id)}, status.Error(codes.AlreadyExists, "URL already exist")
	} else if err != nil {
		s.logger.Errorw("problem with set in storage", err)
		return nil, status.Error(codes.Internal, "Server problem")
	}
	return &pb.PostHandlerResponse{ShortUrl: shortURL, UserID: int64(id)}, status.Error(codes.OK, "OK")
}

func (s *URLCutterServer) JSONPostHandler(ctx context.Context, in *pb.JSONPostHandlerRequest) (*pb.JSONPostHandlerResponse, error) {
	return &pb.JSONPostHandlerResponse{}, nil
	//не вижу логического смысла так как в proto нет json структур, соответственно придется передавать либо через string либо byte, выглядеть будет сомнительно

}

func (s *URLCutterServer) DBPing(ctx context.Context, in *emptypb.Empty) (*pb.DBPingResponse, error) {
	err := s.handler.DBPing()
	if err != nil {
		return nil, status.Error(codes.Internal, "Server problem")
	}
	return &pb.DBPingResponse{}, status.Error(codes.OK, "OK")
}

func (s *URLCutterServer) GetUserUrls(ctx context.Context, in *pb.GetUserUrlsRequest) (*pb.GetUserUrlsResponse, error) {
	var urlsForResp []*pb.ShortenOrigURLs
	urls, err := s.handler.GetUserUrls(ctx, int(in.UserID))
	if err != nil {
		return nil, status.Error(codes.Internal, "Server problem")
	}
	for _, i := range urls {
		urlsForResp = append(urlsForResp, &pb.ShortenOrigURLs{
			ShortURL:    i.ShortURL,
			OriginalURL: i.OriginalURL,
		})
	}
	resp := pb.GetUserUrlsResponse{
		ShortenResponse: urlsForResp,
	}
	return &resp, status.Error(codes.OK, "OK")
}

func (s *URLCutterServer) DeleteUrls(ctx context.Context, in *pb.DeleteUrlsRequest) (*pb.DeleteUrlsResponse, error) {
	s.handler.DeleteUrls(in.Urls)
	return &pb.DeleteUrlsResponse{}, status.Error(codes.OK, "OK")
}

func (s *URLCutterServer) GetStats(ctx context.Context, in *emptypb.Empty) (*pb.GetStatsResponse, error) {
	var ip string
	if pr, ok := peer.FromContext(ctx); ok {
		ip = pr.Addr.String()
	} else {
		return nil, status.Error(codes.Internal, "Server problem")
	}
	urls, users, err := s.handler.GetStats(ctx, ip)
	if err != nil {
		return nil, status.Error(codes.Internal, "Server problem")
	}
	return &pb.GetStatsResponse{Urls: int64(urls), Users: int64(users)}, status.Error(codes.OK, "OK")
}
