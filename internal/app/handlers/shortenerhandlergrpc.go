package handlers

import (
	"context"
	"errors"

	"github.com/valinurovdenis/urlshortener/internal/app/auth"
	"github.com/valinurovdenis/urlshortener/internal/app/ipchecker"
	"github.com/valinurovdenis/urlshortener/internal/app/logger"
	"github.com/valinurovdenis/urlshortener/internal/app/service"
	"github.com/valinurovdenis/urlshortener/internal/app/urlstorage"
	pb "github.com/valinurovdenis/urlshortener/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// UsersServer поддерживает все необходимые методы сервера.
type ShortenerHandlerGrpc struct {
	pb.UnimplementedShortenerHandlerGrpcServer
	Service   service.ShortenerService
	Auth      auth.JwtAuthenticator
	IPChecker ipchecker.IPChecker
	Host      string
}

// Shortener handler contains shortener service, authenticator.
func NewShortenerHandlerGrpc(service service.ShortenerService, auth auth.JwtAuthenticator, host string, ipchecker ipchecker.IPChecker) *ShortenerHandlerGrpc {
	return &ShortenerHandlerGrpc{Service: service, Auth: auth, Host: host, IPChecker: ipchecker}
}

// Handler for redirecting to long url by short url.
func (h *ShortenerHandlerGrpc) Short2LongURL(ctx context.Context, short *pb.URL) (*pb.URL, error) {
	url, err := h.Service.GetLongURLWithContext(ctx, short.Url)
	if errors.Is(err, service.ErrDeletedURL) {
		return nil, status.Errorf(codes.OutOfRange, "Url has been deleted")
	}
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Bad url")
	}
	return &pb.URL{Url: url}, nil
}

// Handler for generating short url from long url.
func (h *ShortenerHandlerGrpc) Generate(ctx context.Context, r *pb.URL) (*pb.URL, error) {
	var userID string
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Internal, "Unexpected error")
	}
	userID = md.Get("userid")[0]

	url, err := h.Service.GenerateShortURLWithContext(ctx, r.Url, userID)

	if err == nil {
		return &pb.URL{Url: url}, nil
	} else if errors.Is(err, urlstorage.ErrConflictURL) {
		return nil, status.Errorf(codes.AlreadyExists, "Conflicting url")
	} else {
		return nil, status.Errorf(codes.InvalidArgument, "Bad url")
	}
}

// Handler for generating long urls in batch mode.
func (h *ShortenerHandlerGrpc) GenerateBatch(ctx context.Context, r *pb.GenerateBatchRequest) (*pb.GenerateBatchResponse, error) {
	var userID string
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Internal, "Unexpected error")
	}
	userID = md.Get("userid")[0]

	var response pb.GenerateBatchResponse
	var longURLs []string
	for _, v := range r.Urls {
		longURLs = append(longURLs, v.OriginalUrl)
	}
	shortURLs, err := h.Service.GenerateShortURLBatchWithContext(ctx, longURLs, userID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Bad urls")
	}
	for i, shortURL := range shortURLs {
		if shortURL != "" {
			response.Urls = append(response.Urls, &pb.GenerateBatchResponse_BatchResponse{
				CorrelationId: r.Urls[i].CorrelationId, ShortUrl: shortURL})
		}
	}
	return &response, nil
}

// Ping that service is still alive.
func (h *ShortenerHandlerGrpc) Ping(ctx context.Context, empty *emptypb.Empty) (*emptypb.Empty, error) {
	err := h.Service.Ping()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Not ok")
	} else {
		return nil, status.Errorf(codes.OK, "Ok")
	}
}

// Get all urls saved by user.
func (h *ShortenerHandlerGrpc) GetUserURLs(ctx context.Context, empty *emptypb.Empty) (*pb.URLMappings, error) {
	var userID string
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Internal, "Unexpected error")
	}
	userID = md.Get("userid")[0]

	var response pb.URLMappings
	userURLs, err := h.Service.GetUserURLs(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unexpected error")
	}

	if len(userURLs) == 0 {
		return nil, status.Errorf(codes.OutOfRange, "No urls")
	}
	for i := range userURLs {
		url := pb.URLMappings_URLMapping{ShortUrl: h.Host + userURLs[i].Short, LongUrl: userURLs[i].Long}
		response.Urls = append(response.Urls, &url)
	}
	return &response, nil
}

// Delete urls saved by user.
func (h *ShortenerHandlerGrpc) DeleteUserURLs(ctx context.Context, r *pb.UserURLs) (*emptypb.Empty, error) {
	var userID string
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Internal, "Unexpected error")
	}
	userID = md.Get("userid")[0]

	err := h.Service.DeleteUserURLs(ctx, userID, r.Urls...)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unexpected error")
	}
	return nil, nil
}

// Get service stats.
func (h ShortenerHandlerGrpc) GetStats(ctx context.Context, empty *emptypb.Empty) (*pb.StorageStats, error) {
	stats, err := h.Service.GetStats(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Unexpected error")
	}
	return &pb.StorageStats{UserCount: int64(stats.UserCount), UrlCount: int64(stats.URLCount)}, nil
}

// Defines handlers with interceptors.
func ShortenerGrpcRouter(shortenerHandler ShortenerHandlerGrpc) *grpc.Server {
	authorizationInterceptor := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler) (interface{}, error) {

		if info.FullMethod == "/shortener.ShortenerHandlerGrpc/DeleteUserURLs" {
			return shortenerHandler.Auth.OnlyWithAuthGrpc(ctx, req, info, handler)
		} else if info.FullMethod == "/shortener.ShortenerHandlerGrpc/GetStats" {
			return shortenerHandler.IPChecker.GrpcCheckFromInner(ctx, req, info, handler)
		} else {
			return shortenerHandler.Auth.CreateUserIfNeededGrpc(ctx, req, info, handler)
		}
	}

	srv := grpc.NewServer(grpc.ChainUnaryInterceptor(logger.RequestLoggerInterceptor(), authorizationInterceptor))
	pb.RegisterShortenerHandlerGrpcServer(srv, &shortenerHandler)

	return srv
}
