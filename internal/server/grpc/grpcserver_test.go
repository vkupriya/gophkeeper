package grpcserver

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"testing"
	"time"

	pb "github.com/vkupriya/gophkeeper/internal/proto"
	ic "github.com/vkupriya/gophkeeper/internal/server/grpc/interceptors"
	"github.com/vkupriya/gophkeeper/internal/server/models"
	"github.com/vkupriya/gophkeeper/internal/server/storage"
	"go.uber.org/zap"

	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/resolver"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

func ServerGRPC(ctx context.Context) (pb.GophKeeperClient, func()) {
	logConfig := zap.NewDevelopmentConfig()
	logger, err := logConfig.Build()
	if err != nil {
		log.Panic(fmt.Errorf("failed to initialize Logger: %w", err))
	}

	dsn := "postgres://postgres:postgres@localhost:5432/gophkeeper?sslmode=disable"
	cfg := &models.Config{
		Logger:         logger,
		Address:        ":3200",
		PostgresDSN:    dsn,
		JWTKey:         "vcwYCYkum_2Fsukk",
		JWTTokenTTL:    3600 * time.Second,
		ContextTimeout: 3 * time.Second,
	}

	buffer := 101024 * 1024
	lis := bufconn.Listen(buffer)

	s, err := storage.NewPostgresDB(cfg.PostgresDSN)
	if err != nil {
		log.Panic(fmt.Errorf("failed initializing PostgresDB: %w", err))
	}

	srv := grpc.NewServer(
		grpc.UnaryInterceptor(ic.AuthInterceptor(cfg)),
	)

	pb.RegisterGophKeeperServer(srv, &GophKeeperServer{
		Store:  s,
		config: cfg,
	})

	go func() {
		if err := srv.Serve(lis); err != nil {
			log.Printf("error serving server: %v", err)
		}
	}()

	resolver.SetDefaultScheme("passthrough")
	conn, err := grpc.NewClient("bufnet", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("error connecting to server: %v", err)
	}

	closer := func() {
		err := lis.Close()
		if err != nil {
			log.Printf("error closing listener: %v", err)
		}
		srv.Stop()
		s.Close()
	}
	client := pb.NewGophKeeperClient(conn)

	return client, closer
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func TestUserRegister(t *testing.T) {
	ctx := context.Background()

	client, closer := ServerGRPC(ctx)
	defer closer()

	type expectation struct {
		out  *pb.UserAuthToken
		code codes.Code
	}

	tests := map[string]struct {
		in       *pb.User
		expected expectation
	}{
		"Register_Success": {
			in: &pb.User{
				Login:    "user01",
				Password: "pass",
			},
			expected: expectation{
				out:  &pb.UserAuthToken{},
				code: codes.Code(code.Code_OK),
			},
		},
		"Register_Fail_InvalidArgument": {
			in: &pb.User{
				Login:    "user01",
				Password: "",
			},
			expected: expectation{
				out:  &pb.UserAuthToken{},
				code: codes.Code(code.Code_INVALID_ARGUMENT),
			},
		},
	}

	for test, tt := range tests {
		t.Run(test, func(t *testing.T) {
			out, err := client.Register(ctx, tt.in)
			if err != nil {
				status, _ := status.FromError(err)
				if tt.expected.code != status.Code() {
					t.Errorf("Err -> \nWant: %q\nGot: %q\n", tt.expected.code, status.Code())
				}
			} else if out.Token == "" {
				t.Errorf("Out -> \nWant: %q\nGot : %q", tt.expected.out, out)
			}
		})
	}
}

func TestUserLogin(t *testing.T) {
	ctx := context.Background()

	client, closer := ServerGRPC(ctx)
	defer closer()

	type expectation struct {
		out  *pb.UserAuthToken
		code codes.Code
	}

	tests := map[string]struct {
		in       *pb.User
		expected expectation
	}{
		"Login_Success": {
			in: &pb.User{
				Login:    "user01",
				Password: "pass",
			},
			expected: expectation{
				out:  &pb.UserAuthToken{},
				code: codes.Code(code.Code_OK),
			},
		},
		"Login_PermissionDenied": {
			in: &pb.User{
				Login:    "user01",
				Password: "passss",
			},
			expected: expectation{
				out:  &pb.UserAuthToken{},
				code: codes.PermissionDenied,
			},
		},
		"Login_UserNotFound": {
			in: &pb.User{
				Login:    "user99",
				Password: "passss",
			},
			expected: expectation{
				out:  &pb.UserAuthToken{},
				code: codes.NotFound,
			},
		},
	}

	for test, tt := range tests {
		t.Run(test, func(t *testing.T) {
			out, err := client.Login(ctx, tt.in)
			if err != nil {
				status, _ := status.FromError(err)
				if tt.expected.code != status.Code() {
					t.Errorf("Err -> \nWant: %q\nGot: %q\n", tt.expected.code, status.Code())
				}
			} else if out.Token == "" {
				t.Errorf("Out -> \nWant: %q\nGot : %q", tt.expected.out, out)
			}
		})
	}
}

func TestSecretAdd(t *testing.T) {
	ctx := context.Background()

	client, closer := ServerGRPC(ctx)
	defer closer()

	out, _ := client.Login(ctx, &pb.User{
		Login:    "user01",
		Password: "pass",
	})
	token := out.Token

	md := metadata.New(map[string]string{"authorization": token})
	md.Append("secretkey", "myencryptionsecret")
	ctxWithAuth := metadata.NewOutgoingContext(context.Background(), md)

	type expectation struct {
		out  *pb.Empty
		code codes.Code
	}

	tests := map[string]struct {
		in       *pb.AddSecretRequest
		expected expectation
	}{
		"SecretAdd_Success": {
			in: &pb.AddSecretRequest{
				Secret: &pb.Secret{
					Name:    "secret01",
					Type:    pb.SecretType_TEXT,
					Meta:    "very important",
					Data:    []byte("secret"),
					Version: 1,
				},
			},
			expected: expectation{
				out:  &pb.Empty{},
				code: codes.Code(code.Code_OK),
			},
		},
	}

	for test, tt := range tests {
		t.Run(test, func(t *testing.T) {
			_, err := client.AddSecret(ctxWithAuth, tt.in)
			if err != nil {
				status, _ := status.FromError(err)
				if tt.expected.code != status.Code() {
					t.Errorf("Err -> \nWant: %q\nGot: %q\n", tt.expected.code, status.Code())
				}
			}
		})
	}
}

func TestSecretGet(t *testing.T) {
	ctx := context.Background()

	client, closer := ServerGRPC(ctx)
	defer closer()

	out, _ := client.Login(ctx, &pb.User{
		Login:    "user01",
		Password: "pass",
	})
	token := out.Token

	md := metadata.New(map[string]string{"authorization": token})
	md.Append("secretkey", "myencryptionsecret")
	ctxWithAuth := metadata.NewOutgoingContext(context.Background(), md)

	type expectation struct {
		out  *pb.GetSecretResponse
		code codes.Code
	}

	tests := map[string]struct {
		in       *pb.GetSecretRequest
		expected expectation
	}{
		"SecretGet_Success": {
			in: &pb.GetSecretRequest{
				Name: "secret01",
			},
			expected: expectation{
				out: &pb.GetSecretResponse{
					Secret: &pb.Secret{
						Name:    "secret01",
						Type:    pb.SecretType_TEXT,
						Meta:    "very important",
						Data:    []byte("secret"),
						Version: 1,
					},
				},
				code: codes.Code(code.Code_OK),
			},
		},
		"SecretGet_Fail_NotFound": {
			in: &pb.GetSecretRequest{
				Name: "secret99",
			},
			expected: expectation{
				out:  &pb.GetSecretResponse{},
				code: codes.NotFound,
			},
		},
	}

	for test, tt := range tests {
		t.Run(test, func(t *testing.T) {
			out, err := client.GetSecret(ctxWithAuth, tt.in)
			fmt.Println("out: ", out)
			if err != nil {
				status, _ := status.FromError(err)
				if tt.expected.code != status.Code() {
					t.Errorf("Err -> \nWant: %q\nGot: %q\n", tt.expected.code, status.Code())
				}
			} else if !bytes.Equal(tt.expected.out.Secret.GetData(), out.Secret.GetData()) {
				t.Errorf("Out -> \nWant: %q\nGot : %q", tt.expected.out, out)
			}
		})
	}
}

func TestSecretUpdate(t *testing.T) {
	ctx := context.Background()

	client, closer := ServerGRPC(ctx)
	defer closer()

	out, _ := client.Login(ctx, &pb.User{
		Login:    "user01",
		Password: "pass",
	})
	token := out.Token

	md := metadata.New(map[string]string{"authorization": token})
	md.Append("secretkey", "myencryptionsecret")
	ctxWithAuth := metadata.NewOutgoingContext(context.Background(), md)

	type expectation struct {
		out  *pb.Empty
		code codes.Code
	}

	tests := map[string]struct {
		in       *pb.UpdateSecretRequest
		expected expectation
	}{
		"SecretUpdate_Success": {
			in: &pb.UpdateSecretRequest{
				Secret: &pb.Secret{
					Name:    "secret01",
					Type:    pb.SecretType_TEXT,
					Meta:    "very important secret",
					Data:    []byte("secret"),
					Version: 0,
				},
			},
			expected: expectation{
				out:  &pb.Empty{},
				code: codes.Code(code.Code_OK),
			},
		},
		"SecretUpdate_Fail_NotFound": {
			in: &pb.UpdateSecretRequest{
				Secret: &pb.Secret{
					Name:    "secret99",
					Type:    pb.SecretType_TEXT,
					Meta:    "very important secret",
					Data:    []byte("secret"),
					Version: 0,
				},
			},
			expected: expectation{
				out:  &pb.Empty{},
				code: codes.NotFound,
			},
		},
	}

	for test, tt := range tests {
		t.Run(test, func(t *testing.T) {
			out, err := client.UpdateSecret(ctxWithAuth, tt.in)
			fmt.Println("out: ", out)
			if err != nil {
				status, _ := status.FromError(err)
				if tt.expected.code != status.Code() {
					t.Errorf("Err -> \nWant: %q\nGot: %q\n", tt.expected.code, status.Code())
				}
			}
		})
	}
}

func TestSecretList(t *testing.T) {
	ctx := context.Background()

	client, closer := ServerGRPC(ctx)
	defer closer()

	out, _ := client.Login(ctx, &pb.User{
		Login:    "user01",
		Password: "pass",
	})
	token := out.Token

	md := metadata.New(map[string]string{"authorization": token})
	md.Append("secretkey", "myencryptionsecret")
	ctxWithAuth := metadata.NewOutgoingContext(context.Background(), md)

	type expectation struct {
		out  *pb.ListSecretsResponse
		code codes.Code
	}

	tests := map[string]struct {
		in       *pb.Empty
		expected expectation
	}{
		"SecretGetList_Success": {
			in: &pb.Empty{},
			expected: expectation{
				out:  &pb.ListSecretsResponse{},
				code: codes.Code(code.Code_OK),
			},
		},
	}

	for test, tt := range tests {
		t.Run(test, func(t *testing.T) {
			out, err := client.ListSecrets(ctxWithAuth, tt.in)
			if err != nil {
				status, _ := status.FromError(err)
				if tt.expected.code != status.Code() {
					t.Errorf("Err -> \nWant: %q\nGot: %q\n", tt.expected.code, status.Code())
				} else if len(out.Items) == 0 {
					t.Error("empty list of secrets.")
				}
			}
		})
	}
}

func TestSecretDelete(t *testing.T) {
	ctx := context.Background()

	client, closer := ServerGRPC(ctx)
	defer closer()

	out, _ := client.Login(ctx, &pb.User{
		Login:    "user01",
		Password: "pass",
	})
	token := out.Token

	md := metadata.New(map[string]string{"authorization": token})
	md.Append("secretkey", "myencryptionsecret")
	ctxWithAuth := metadata.NewOutgoingContext(context.Background(), md)

	type expectation struct {
		out  *pb.Empty
		code codes.Code
	}

	tests := map[string]struct {
		in       *pb.DeleteSecretRequest
		expected expectation
	}{
		"SecretDelete_Success": {
			in: &pb.DeleteSecretRequest{
				Name: "secret01",
			},
			expected: expectation{
				out:  &pb.Empty{},
				code: codes.Code(code.Code_OK),
			},
		},
		"SecretDelete_Fail_NotFound": {
			in: &pb.DeleteSecretRequest{
				Name: "secret99",
			},
			expected: expectation{
				out:  &pb.Empty{},
				code: codes.NotFound,
			},
		},
	}

	for test, tt := range tests {
		t.Run(test, func(t *testing.T) {
			_, err := client.DeleteSecret(ctxWithAuth, tt.in)
			if err != nil {
				status, _ := status.FromError(err)
				if tt.expected.code != status.Code() {
					t.Errorf("Err -> \nWant: %q\nGot: %q\n", tt.expected.code, status.Code())
				}
			}
		})
	}
}
