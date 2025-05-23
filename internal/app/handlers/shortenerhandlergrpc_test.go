package handlers_test

import (
	"context"
	"log"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/valinurovdenis/urlshortener/internal/app/auth"
	"github.com/valinurovdenis/urlshortener/internal/app/handlers"
	"github.com/valinurovdenis/urlshortener/internal/app/ipchecker"
	"github.com/valinurovdenis/urlshortener/internal/app/service"
	"github.com/valinurovdenis/urlshortener/internal/app/shortcutgenerator"
	"github.com/valinurovdenis/urlshortener/internal/app/urlstorage"
	"github.com/valinurovdenis/urlshortener/internal/app/userstorage"
	pb "github.com/valinurovdenis/urlshortener/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024

var lis *bufconn.Listener

func init() {
	lis = bufconn.Listen(bufSize)
	storage := urlstorage.NewSimpleMapLockStorage()
	userStorage := userstorage.NewSimpleUserStorage()
	generator := shortcutgenerator.NewRandBase64Generator(8)
	service := service.NewShortenerService(storage, storage, generator)
	auth := auth.NewAuthenticator("secret_key", userStorage)
	ipchecker := ipchecker.NewIPChecker("192.168.0.0/24")
	grpcHandler := handlers.NewShortenerHandlerGrpc(*service, *auth, "/", *ipchecker)
	srv := handlers.ShortenerGrpcRouter(*grpcHandler)
	go func() {
		if err := srv.Serve(lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
}

func getGrpcConn(t *testing.T) *grpc.ClientConn {
	conn, err := grpc.NewClient("passthrough:///bufnet", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	return conn
}

func getStatusFromGrpcError(t *testing.T, err error) codes.Code {
	if err == nil {
		return codes.OK
	}
	if e, ok := status.FromError(err); ok {
		return e.Code()
	} else {
		t.Fatalf("Не получилось распарсить ошибку %v", err)
		return codes.Internal
	}
}

func TestShortenerHandlerGrpc_TestGenerate(t *testing.T) {
	conn := getGrpcConn(t)
	defer conn.Close()
	grpcClient := pb.NewShortenerHandlerGrpcClient(conn)

	resp, err := grpcClient.Generate(context.Background(), &pb.URL{Url: "url1"})
	require.Equal(t, codes.OK, getStatusFromGrpcError(t, err))

	resp, err = grpcClient.Short2LongURL(context.Background(), &pb.URL{Url: resp.Url})
	require.Equal(t, codes.OK, getStatusFromGrpcError(t, err))
	require.Equal(t, "http://url1", resp.Url)
}

func TestShortenerHandlerGrpc_TestGenerateBatch(t *testing.T) {
	conn := getGrpcConn(t)
	defer conn.Close()
	grpcClient := pb.NewShortenerHandlerGrpcClient(conn)

	resp, err := grpcClient.GenerateBatch(context.Background(),
		&pb.GenerateBatchRequest{Urls: []*pb.GenerateBatchRequest_BatchRequest{{OriginalUrl: "url2", CorrelationId: "1"}, {OriginalUrl: "url3", CorrelationId: "2"}}})
	require.Equal(t, codes.OK, getStatusFromGrpcError(t, err))

	getShort, err := grpcClient.Short2LongURL(context.Background(), &pb.URL{Url: resp.Urls[0].ShortUrl})
	require.Equal(t, codes.OK, getStatusFromGrpcError(t, err))
	require.Contains(t, []string{"http://url1", "http://url2"}, getShort.Url)
}

func TestShortenerHandlerGrpc_GetUserUrls(t *testing.T) {
	conn := getGrpcConn(t)
	defer conn.Close()
	grpcClient := pb.NewShortenerHandlerGrpcClient(conn)

	var header metadata.MD
	_, err := grpcClient.Generate(context.Background(), &pb.URL{Url: "url4"}, grpc.Header(&header))
	require.Equal(t, codes.OK, getStatusFromGrpcError(t, err))

	ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{"Authorization": header.Get("Authorization")[0]}))
	grpcClient.GetUserURLs(ctx, nil)
}

func TestShortenerHandlerGrpc_DeleteUserUrls(t *testing.T) {
	conn := getGrpcConn(t)
	defer conn.Close()
	grpcClient := pb.NewShortenerHandlerGrpcClient(conn)

	var header metadata.MD
	_, err := grpcClient.Generate(context.Background(), &pb.URL{Url: "url5"}, grpc.Header(&header))
	require.Equal(t, codes.OK, getStatusFromGrpcError(t, err))

	_, err = grpcClient.DeleteUserURLs(context.Background(), nil)
	require.Equal(t, codes.Unauthenticated, getStatusFromGrpcError(t, err))

	ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{"Authorization": header.Get("Authorization")[0]}))
	_, err = grpcClient.DeleteUserURLs(ctx, nil)
	require.NotEqual(t, codes.Unauthenticated, getStatusFromGrpcError(t, err))
}

func TestShortenerHandlerGrpc_Ping(t *testing.T) {
	conn := getGrpcConn(t)
	defer conn.Close()
	grpcClient := pb.NewShortenerHandlerGrpcClient(conn)

	_, err := grpcClient.Ping(context.Background(), nil)
	require.Equal(t, codes.OK, getStatusFromGrpcError(t, err))
}

func TestShortenerHandlerGrpc_GetStats(t *testing.T) {
	conn := getGrpcConn(t)
	defer conn.Close()
	grpcClient := pb.NewShortenerHandlerGrpcClient(conn)

	_, err := grpcClient.GetStats(context.Background(), nil)
	require.Equal(t, codes.PermissionDenied, getStatusFromGrpcError(t, err))

	ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{"X-Real-IP": "192.168.1.1"}))
	_, err = grpcClient.GetStats(ctx, nil)
	require.Equal(t, codes.PermissionDenied, getStatusFromGrpcError(t, err))

	ctx = metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{"X-Real-IP": "192.168.0.1"}))
	_, err = grpcClient.GetStats(ctx, nil)
	require.NotEqual(t, codes.PermissionDenied, getStatusFromGrpcError(t, err))
}
