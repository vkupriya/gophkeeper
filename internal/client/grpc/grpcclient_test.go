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
