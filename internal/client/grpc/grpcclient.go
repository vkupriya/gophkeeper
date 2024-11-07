package grpcclient

import (
	"context"
	"errors"
	"fmt"
	"strings"

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

func NewGRPCClient(s *Service, grpcHost string) error {
	conn, err := grpc.NewClient(grpcHost, grpc.WithTransportCredentials((insecure.NewCredentials())))
	if err != nil {
		return fmt.Errorf("failed to create GRPC client: %w", err)
	}

	s.connGRPC = conn
	s.clientGRPC = pb.NewGophKeeperClient(conn)
	return nil
}

func (s *Service) Register(user string, password string) (string, error) {
	authToken, err := s.clientGRPC.Register(context.Background(), &pb.User{
		Login:    user,
		Password: password,
	})
	if err != nil {
		return "", fmt.Errorf("failed to register: %w", err)
	}

	token := strings.Split(authToken.String(), ":")[1]
	token = strings.ReplaceAll(token, `"`, "")
	return token, nil
}

func (s *Service) Login(user string, password string) (string, error) {
	authToken, err := s.clientGRPC.Login(context.Background(), &pb.User{
		Login:    user,
		Password: password,
	})
	if err != nil {
		return "", fmt.Errorf("failed to login: %w", err)
	}
	token := strings.Split(authToken.String(), ":")[1]
	token = strings.ReplaceAll(token, `"`, "")
	return token, nil
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
			Type:    ProtoToType(secret.GetType()),
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
		Type:    TypeToProto(secret.Type),
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
			Type: TypeToProto(secret.Type),
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
		Type:    ProtoToType(resp.Secret.Type),
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

func TypeToProto(st string) pb.SecretType {
	switch st {
	case "text":
		return pb.SecretType_TEXT
	case "binary":
		return pb.SecretType_BINARY
	case "card":
		return pb.SecretType_CARD
	case "file":
		return pb.SecretType_FILE
	case "unknown":
		return pb.SecretType_UNKNOWN
	default:
		return pb.SecretType_UNKNOWN
	}
}

func ProtoToType(st pb.SecretType) string {
	switch st {
	case pb.SecretType_TEXT:
		return "text"
	case pb.SecretType_BINARY:
		return "binary"
	case pb.SecretType_CARD:
		return "card"
	case pb.SecretType_FILE:
		return "file"
	case pb.SecretType_UNKNOWN:
		return "unknown"
	default:
		return "unknown"
	}
}
