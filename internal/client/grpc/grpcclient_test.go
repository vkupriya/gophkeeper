package grpcclient

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/vkupriya/gophkeeper/internal/client/models"
	pb "github.com/vkupriya/gophkeeper/internal/proto"

	mocks "github.com/vkupriya/gophkeeper/internal/proto/mocks"
)

func TestRegister(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockGophKeeperClient(ctrl)

	value := "ksjdfksjldkfjsldkfjsdlkf"
	m.EXPECT().Register(gomock.Any(), gomock.Any()).Return(&pb.UserAuthToken{
		Token: value,
	}, nil)

	svc := NewService()
	svc.clientGRPC = m

	token, err := svc.Register("user", "password")
	require.NoError(t, err)
	require.Equal(t, token, value)
}

func TestLogin(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockGophKeeperClient(ctrl)

	value := "ksjdfksjldkfjsldkfjsdlkf"
	m.EXPECT().Login(gomock.Any(), gomock.Any()).Return(&pb.UserAuthToken{
		Token: value,
	}, nil)

	svc := NewService()
	svc.clientGRPC = m

	token, err := svc.Login("user", "password")
	require.NoError(t, err)
	require.Equal(t, token, value)
}

func TestListSecrets(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockGophKeeperClient(ctrl)

	m.EXPECT().ListSecrets(gomock.Any(), gomock.Any()).Return(&pb.ListSecretsResponse{
		Items: []*pb.SecretItem{
			{
				Name:    "secret01",
				Type:    pb.SecretType_TEXT,
				Version: 1,
			},
		},
	}, nil)

	secretsExpected := []*models.SecretItem{
		{
			Name:    "secret01",
			Type:    "text",
			Version: 1,
		},
	}
	svc := NewService()
	svc.clientGRPC = m

	secrets, err := svc.ListSecrets("user")
	require.NoError(t, err)
	require.Equal(t, secrets, secretsExpected)
}

func TestAddSecret(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockGophKeeperClient(ctrl)

	m.EXPECT().AddSecret(gomock.Any(), gomock.Any()).Return(&pb.Empty{}, nil)
	card := "{'card':'0123456789876','name': 'Super Agent', 'expiry': '02/25', 'cvv':'555'}"
	secret := &models.Secret{
		Name:    "card01",
		Type:    "card",
		Meta:    "metadata",
		Data:    []byte(card),
		Version: 1,
	}
	svc := NewService()
	svc.clientGRPC = m

	err := svc.AddSecret("token", "encryptionkey", secret)
	require.NoError(t, err)
}

func TestGetSecret(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	card := "{'card':'0123456789876','name': 'Super Agent', 'expiry': '02/25', 'cvv':'555'}"

	m := mocks.NewMockGophKeeperClient(ctrl)

	m.EXPECT().GetSecret(gomock.Any(), gomock.Any()).Return(&pb.GetSecretResponse{
		Secret: &pb.Secret{
			Name:    "card01",
			Type:    pb.SecretType_CARD,
			Meta:    "metadata",
			Data:    []byte(card),
			Version: 1,
		},
	}, nil)

	secretExpected := &models.Secret{
		Name:    "card01",
		Type:    "card",
		Meta:    "metadata",
		Data:    []byte(card),
		Version: 1,
	}
	svc := NewService()
	svc.clientGRPC = m

	secret, err := svc.GetSecret("token", "encryptionkey", "card01")
	require.NoError(t, err)
	require.Equal(t, secret, secretExpected)
}

func TestUpdateSecret(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	card := "{'card':'0123456789876','name': 'Super Agent', 'expiry': '02/25', 'cvv':'555'}"

	m := mocks.NewMockGophKeeperClient(ctrl)

	m.EXPECT().UpdateSecret(gomock.Any(), gomock.Any()).Return(&pb.Empty{}, nil)

	secret := &models.Secret{
		Name:    "card01",
		Type:    "card",
		Meta:    "metadata",
		Data:    []byte(card),
		Version: 1,
	}

	svc := NewService()
	svc.clientGRPC = m

	err := svc.UpdateSecret("token", "encryptionkey", secret)
	require.NoError(t, err)
}

func TestDeleteSecret(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := mocks.NewMockGophKeeperClient(ctrl)

	m.EXPECT().DeleteSecret(gomock.Any(), gomock.Any()).Return(&pb.Empty{}, nil)

	svc := NewService()
	svc.clientGRPC = m

	err := svc.DeleteSecret("token", "card01")
	require.NoError(t, err)
}
