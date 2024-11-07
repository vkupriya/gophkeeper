package grpcserver

import (
	// ...

	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"

	pb "github.com/vkupriya/gophkeeper/internal/proto"
	ic "github.com/vkupriya/gophkeeper/internal/server/grpc/interceptors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"golang.org/x/crypto/bcrypt"

	"github.com/vkupriya/gophkeeper/internal/server/helpers"
	"github.com/vkupriya/gophkeeper/internal/server/models"
	"github.com/vkupriya/gophkeeper/internal/server/storage"
)

type Storage interface {
	UserAdd(c *models.Config, u models.User) error
	UserGet(c *models.Config, userid string) (models.User, error)
	SecretGet(c *models.Config, userid string, name string) (*models.Secret, error)
	SecretList(c *models.Config, userid string) (*models.SecretList, error)
	SecretAdd(c *models.Config, userid string, secret *models.Secret) error
	SecretUpdate(c *models.Config, userid string, secret *models.Secret) error
	SecretDelete(c *models.Config, userid string, name string) error
}

const (
	errFormat                     = "error: %w"
	msgUserCredentialsBadRequest  = "invalid credentials"
	msgUserFailedToCreate         = "failed to create user"
	msgUserAlreadyExists          = "user already exists"
	msgUserFailedToCreateToken    = "failed to create user token"
	msgUserTokenError             = "user token error"
	msgUserInvalidLoginOrPassword = "user invalid login or password"
	msgUserNotFound               = "user not found"
	msgUserFailedToLogin          = "failed to login"
	msgSecretsNotFound            = "secrets not found"
	msgSecretsFailedToGet         = "failed to get secrets"
	msgSecretBadRequest           = "invalid secret"
	msgSecretFailedToCreate       = "failed to create secret"
	msgSecretAlreadyExists        = "secret already exists"
	msgSecretNotFound             = "secret not found"
	msgSecretFailedToDelete       = "failed to delete secret"
	msgSecretFailedToUpdate       = "failed to update secret"
	msgMetadataNotFound           = "grpc metadata not found"
)

type GophKeeperServer struct {
	pb.UnimplementedGophKeeperServer
	Store  Storage
	config *models.Config
}

func (g *GophKeeperServer) Register(ctx context.Context, in *pb.User) (*pb.UserAuthToken, error) {
	logger := g.config.Logger
	var response pb.UserAuthToken
	user := models.User{
		UserID:   in.GetLogin(),
		Password: in.GetPassword(),
	}

	if user.UserID == "" || user.Password == "" {
		logger.Sugar().Errorf("invalid credentials for user %s", user.UserID)
		return nil, fmt.Errorf(errFormat, status.Error(codes.InvalidArgument, msgUserCredentialsBadRequest))
	}
	password, err := helpers.HashPassword(user.Password)
	if err != nil {
		logger.Sugar().Errorf("failed to hash password for user %s: %v", err)
		return nil, fmt.Errorf(errFormat, status.Error(codes.Internal, msgUserFailedToCreate))
	}
	user.Password = password

	err = g.Store.UserAdd(g.config, user)
	if err != nil {
		if errors.Is(err, storage.ErrUserAlreadyExists) {
			logger.Sugar().Errorf("failed to create user %s: already exists", user.UserID)
			return nil, fmt.Errorf(errFormat, status.Error(codes.AlreadyExists, msgUserAlreadyExists))
		}
		logger.Sugar().Errorf("failed adding user %s into db: %v", user.UserID, err)
		return nil, fmt.Errorf(errFormat, status.Error(codes.Internal, msgUserFailedToCreate))
	}

	token, err := helpers.CreateJWTString(g.config, user.UserID)
	if err != nil {
		logger.Sugar().Errorf("Error creating JWT token: %v", err)
		return nil, fmt.Errorf(errFormat, status.Error(codes.Internal, msgUserFailedToCreateToken))
	}
	response.Token = token

	return &response, nil
}

func (g *GophKeeperServer) Login(ctx context.Context, in *pb.User) (*pb.UserAuthToken, error) {
	logger := g.config.Logger
	var response pb.UserAuthToken

	user, err := g.Store.UserGet(g.config, in.GetLogin())
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			logger.Sugar().Errorf("login error for user %s: not found", user.UserID)
			return nil, fmt.Errorf(errFormat, status.Error(codes.NotFound, msgUserNotFound))
		}
		logger.Sugar().Error("failed to login for user %s: %v", user.UserID, err)
		return nil, fmt.Errorf(errFormat, status.Error(codes.Internal, msgUserFailedToLogin))
	}

	ok := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.GetPassword()))
	if ok != nil {
		logger.Sugar().Errorf("login error for user %s: wrong credentials", user.UserID)
		return nil, fmt.Errorf(errFormat, status.Error(codes.PermissionDenied, msgUserInvalidLoginOrPassword))
	}

	token, err := helpers.CreateJWTString(g.config, user.UserID)
	if err != nil {
		logger.Sugar().Errorf("Error creating JWT token: %v", err)
		return nil, fmt.Errorf(errFormat, status.Error(codes.Internal, msgUserFailedToCreateToken))
	}
	response.Token = token

	return &response, nil
}

func (g *GophKeeperServer) ListSecrets(ctx context.Context, in *pb.Empty) (*pb.ListSecretsResponse, error) {
	logger := g.config.Logger
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		logger.Sugar().Error(msgMetadataNotFound)
		return nil, fmt.Errorf(errFormat, status.Error(codes.NotFound, msgMetadataNotFound))
	}
	userid := md["userid"][0]

	secretsDB, err := g.Store.SecretList(g.config, userid)
	if err != nil {
		if errors.Is(err, storage.ErrNoSecrets) {
			return nil, fmt.Errorf(errFormat, status.Error(codes.NotFound, msgSecretsNotFound))
		}
		logger.Sugar().Errorf("failed to get list of secrets: %v", err)
		return nil, fmt.Errorf(errFormat, status.Error(codes.Internal, msgSecretsFailedToGet))
	}

	response := &pb.ListSecretsResponse{Items: make([]*pb.SecretItem, 0, len(*secretsDB))}
	for _, dbItem := range *secretsDB {
		response.Items = append(response.Items, &pb.SecretItem{
			Name:    dbItem.Name,
			Type:    TypeToProto(dbItem.Type),
			Version: dbItem.Version,
		})
	}

	return response, nil
}

func (g *GophKeeperServer) AddSecret(ctx context.Context, in *pb.AddSecretRequest) (*pb.Empty, error) {
	logger := g.config.Logger
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		logger.Sugar().Error(msgMetadataNotFound)
		return nil, fmt.Errorf(errFormat, status.Error(codes.NotFound, msgMetadataNotFound))
	}
	userid := md["userid"][0]
	key := md["secretkey"][0]
	data := in.Secret.GetData()
	dataEncrypted, err := helpers.Encrypt(key, &data)
	if err != nil {
		logger.Sugar().Errorf("error encrypting data: %v", err)
		return nil, fmt.Errorf(errFormat, status.Error(codes.Internal, msgSecretFailedToCreate))
	}
	secret := &models.Secret{
		Name:    in.Secret.GetName(),
		Meta:    in.Secret.GetMeta(),
		Type:    ProtoToType(in.Secret.Type),
		Data:    dataEncrypted,
		Version: 1,
	}
	err = g.Store.SecretAdd(g.config, userid, secret)
	if err != nil {
		if errors.Is(err, storage.ErrSecretAlreadyExists) {
			logger.Sugar().Errorf("failed creating secret for user %s:  already exists", userid)
			return nil, fmt.Errorf(errFormat, status.Error(codes.AlreadyExists, msgSecretAlreadyExists))
		}
		logger.Sugar().Errorf("error adding secret into DB: %v", err)
		return nil, fmt.Errorf(errFormat, status.Error(codes.Internal, msgSecretFailedToCreate))
	}

	return &pb.Empty{}, nil
}

func (g *GophKeeperServer) UpdateSecret(
	ctx context.Context,
	in *pb.UpdateSecretRequest) (*pb.Empty, error) {
	logger := g.config.Logger
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		logger.Sugar().Error(msgMetadataNotFound)
		return nil, fmt.Errorf(errFormat, status.Error(codes.NotFound, msgMetadataNotFound))
	}
	userid := md["userid"][0]
	key := md["secretkey"][0]
	data := in.Secret.GetData()
	dataEncrypted, err := helpers.Encrypt(key, &data)
	if err != nil {
		logger.Sugar().Errorf("error encrypting data: %v", err)
		return nil, fmt.Errorf(errFormat, status.Error(codes.Internal, msgSecretFailedToUpdate))
	}
	secret := &models.Secret{
		Name: in.Secret.GetName(),
		Meta: in.Secret.GetMeta(),
		Type: ProtoToType(in.Secret.Type),
		Data: dataEncrypted,
	}
	err = g.Store.SecretUpdate(g.config, userid, secret)
	if err != nil {
		logger.Sugar().Errorf("error updating secret: %v", err)
		return nil, fmt.Errorf(errFormat, status.Error(codes.Internal, msgSecretFailedToUpdate))
	}

	return &pb.Empty{}, nil
}

func (g *GophKeeperServer) GetSecret(ctx context.Context, in *pb.GetSecretRequest) (*pb.GetSecretResponse, error) {
	logger := g.config.Logger
	var response pb.GetSecretResponse

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		logger.Sugar().Error(msgMetadataNotFound)
		return nil, fmt.Errorf(errFormat, status.Error(codes.NotFound, msgMetadataNotFound))
	}
	userid := md["userid"][0]
	key := md["secretkey"][0]

	s, err := g.Store.SecretGet(g.config, userid, in.GetName())
	if err != nil {
		if errors.Is(err, storage.ErrSecretNotFound) {
			logger.Sugar().Errorf("secret %s not found for user %s", in.Name, userid)
			return nil, fmt.Errorf(errFormat, status.Error(codes.NotFound, msgSecretNotFound))
		}
		logger.Sugar().Errorf("error getting secret from DB: %v", err)
		return &response, fmt.Errorf(errFormat, status.Error(codes.Internal, msgSecretsFailedToGet))
	}

	data, err := helpers.Decrypt(key, s.Data)
	if err != nil {
		logger.Sugar().Errorf("error decrypting secret data: %v", err)
		return &response, fmt.Errorf(errFormat, status.Error(codes.Internal, msgSecretsFailedToGet))
	}

	response = pb.GetSecretResponse{
		Secret: &pb.Secret{
			Name:    s.Name,
			Type:    TypeToProto(s.Type),
			Meta:    s.Meta,
			Data:    *data,
			Version: s.Version,
		},
	}
	return &response, nil
}

func (g *GophKeeperServer) DeleteSecret(ctx context.Context,
	in *pb.DeleteSecretRequest) (*pb.Empty, error) {
	logger := g.config.Logger
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		logger.Sugar().Error(msgMetadataNotFound)
		return nil, fmt.Errorf(errFormat, status.Error(codes.NotFound, msgMetadataNotFound))
	}
	userid := md["userid"][0]

	err := g.Store.SecretDelete(g.config, userid, in.GetName())
	if err != nil {
		if errors.Is(err, storage.ErrSecretNotFound) {
			logger.Sugar().Errorf("secret %s not found for user %s", in.Name, userid)
			return nil, fmt.Errorf(errFormat, status.Error(codes.NotFound, msgSecretNotFound))
		}
		logger.Sugar().Errorf("failed to delete secret for user %s: %v", userid, err)
		return nil, fmt.Errorf(errFormat, status.Error(codes.Internal, msgSecretFailedToDelete))
	}

	return &pb.Empty{}, nil
}

func Run(ctx context.Context, s Storage, c *models.Config) error {
	const MaxSizeBytes = 10 * 1024 * 1024
	logger := c.Logger
	hostport := strings.Replace(c.Address, "http://", "", 1)
	grpcHost := strings.Split(hostport, ":")[0]
	if grpcHost == "" || grpcHost == "localhost" {
		grpcHost = "127.0.0.1"
	}
	grpcHost += ":3200"
	listen, err := net.Listen("tcp", grpcHost)
	if err != nil {
		return fmt.Errorf("failed to set up listener on port 3200: %w", err)
	}

	srv := grpc.NewServer(
		grpc.UnaryInterceptor(ic.AuthInterceptor(c)),
		grpc.MaxRecvMsgSize(MaxSizeBytes),
		grpc.MaxSendMsgSize(MaxSizeBytes),
	)

	pb.RegisterGophKeeperServer(srv, &GophKeeperServer{
		Store:  s,
		config: c,
	})

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		<-ctx.Done()

		log.Printf("got signal %v, attempting graceful shutdown", s)

		srv.GracefulStop()

		wg.Done()
	}()

	logger.Sugar().Infow("gRPC server is starting", "Address", grpcHost)

	if err := srv.Serve(listen); err != nil {
		logger.Sugar().Fatal(err)
		return fmt.Errorf("failed to run grpc server: %w", err)
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
