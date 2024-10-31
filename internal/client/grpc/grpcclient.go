package grpcclient

import (
	"context"
	"errors"
	"fmt"

	"github.com/vkupriya/gophkeeper/internal/client/models"
	pb "github.com/vkupriya/gophkeeper/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var (
	ErrServerUnavailable = errors.New("server not available")
)

type Service struct {
	connGRPC   *grpc.ClientConn
	clientGRPC pb.GophKeeperClient
}

func NewService() *Service {
	return &Service{}
}

func NewGRPCClient(s *Service) error {
	grpcHost := "127.0.0.1:3200"

	conn, err := grpc.NewClient(grpcHost, grpc.WithTransportCredentials((insecure.NewCredentials())))
	if err != nil {
		return fmt.Errorf("failed to create GRPC client: %w", err)
	}

	s.connGRPC = conn
	s.clientGRPC = pb.NewGophKeeperClient(conn)
	return nil
}

func (s *Service) Register(user string, password string) (string, error) {
	token, err := s.clientGRPC.Register(context.Background(), &pb.User{
		Login:    user,
		Password: password,
	})
	if err != nil {
		return "", fmt.Errorf("failed to register: %w", err)
	}
	return token.String(), nil
}

func (s *Service) Login(user string, password string) (string, error) {
	token, err := s.clientGRPC.Login(context.Background(), &pb.User{
		Login:    user,
		Password: password,
	})
	if err != nil {
		return "", fmt.Errorf("failed to login: %w", err)
	}
	return token.String(), nil
}

func (s *Service) ListSecrets(t string) ([]*models.SecretItem, error) {
	md := metadata.New(map[string]string{"authorization": t})
	ctxWithAuth := metadata.NewOutgoingContext(context.Background(), md)
	secrets, err := s.clientGRPC.ListSecrets(ctxWithAuth, &pb.Empty{})
	if err != nil {
		status, ok := status.FromError(err)
		if !ok {
			return nil, fmt.Errorf("error in getting list of secrets: %w", err)
		}

		switch status.Code() {
		case codes.Unavailable:
			return nil, ErrServerUnavailable
		default:
			return nil, fmt.Errorf("error in getting list of secrets: %w", err)
		}
	}

	resultItems := make([]*models.SecretItem, 0, len(secrets.Items))
	for _, secret := range secrets.Items {
		resultItems = append(resultItems, &models.SecretItem{
			Name:    secret.GetName(),
			Type:    "text",
			Version: secret.GetVersion(),
		})
	}
	return resultItems, nil
}

// AddSecret - function adding secret to gophkeeper server, it takes
// login token, encryption key and secret struct.
func (s *Service) AddSecret(t string, key string, secret *models.Secret) error {
	md := metadata.New(map[string]string{"authorization": t})
	md.Append("secretkey", key)
	pbSecret := &pb.Secret{
		Name:    secret.Name,
		Meta:    secret.Meta,
		Data:    secret.Data,
		Type:    pb.SecretType(models.TypeToProto(secret.Type)),
		Version: secret.Version,
	}

	ctxWithAuth := metadata.NewOutgoingContext(context.Background(), md)
	_, err := s.clientGRPC.AddSecret(ctxWithAuth, &pb.AddSecretRequest{
		Secret: pbSecret,
	})

	if err != nil {
		return fmt.Errorf("failed to add secret: %w", err)
	}
	return nil
}

// UpdateSecret - function uptading named secret on gophkeeper server, it takes
// login token, encryption key and secret struct.
func (s *Service) UpdateSecret(t string, key string, secret *models.Secret) error {
	md := metadata.New(map[string]string{"authorization": t})
	md.Append("secretkey", key)
	ctxWithAuth := metadata.NewOutgoingContext(context.Background(), md)
	_, err := s.clientGRPC.UpdateSecret(ctxWithAuth, &pb.UpdateSecretRequest{
		Secret: &pb.Secret{
			Name: secret.Name,
			Meta: secret.Meta,
			Data: secret.Data,
			Type: 1,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to update secret: %w", err)
	}
	return nil
}

// GetSecret - function getting named secret from gophkeeper server, it takes
// login token, encryption key and secret name.
func (s *Service) GetSecret(t string, key string, name string) (*models.Secret, error) {
	md := metadata.New(map[string]string{"authorization": t})
	md.Append("secretkey", key)
	ctxWithAuth := metadata.NewOutgoingContext(context.Background(), md)
	resp, err := s.clientGRPC.GetSecret(ctxWithAuth, &pb.GetSecretRequest{
		Name: name,
	})
	if err != nil {
		status, ok := status.FromError(err)
		if !ok {
			return nil, fmt.Errorf("error in getting a secret: %w", err)
		}

		switch status.Code() {
		case codes.Unavailable:
			return nil, ErrServerUnavailable
		default:
			return nil, fmt.Errorf("error in getting a secret: %w", err)
		}
	}

	secret := models.Secret{
		Name:    resp.Secret.GetName(),
		Type:    models.ProtoToType(int32(resp.Secret.Type)),
		Meta:    resp.Secret.GetMeta(),
		Data:    resp.Secret.GetData(),
		Version: resp.Secret.GetVersion(),
	}

	return &secret, nil
}

func (s *Service) DeleteSecret(t string, name string) error {
	md := metadata.New(map[string]string{"authorization": t})
	ctxWithAuth := metadata.NewOutgoingContext(context.Background(), md)
	_, err := s.clientGRPC.DeleteSecret(ctxWithAuth, &pb.DeleteSecretRequest{
		Name: name,
	})
	if err != nil {
		return fmt.Errorf("failed to delete secret: %w", err)
	}

	return nil
}
