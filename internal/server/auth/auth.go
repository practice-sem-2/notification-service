package auth

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/metadata"
	"os"
)

const AuthHeaderKey = "Authorization"
const UserKey = "AuthorizedUser"

var (
	ErrInvalidToken = errors.New("provided token is invalid")
	ErrBadContext   = errors.New("can't retrieve metadata from context")
	ErrNoAuth       = errors.New("token is not provided")
)

type UserClaims struct {
	jwt.RegisteredClaims
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	AvatarID  string `json:"avatar_id"`
}

type Service struct {
	privateKey *ecdsa.PrivateKey
	publicKey  ecdsa.PublicKey
}

func New(privateKeyRaw []byte) (*Service, error) {
	privateKey, err := jwt.ParseECPrivateKeyFromPEM(privateKeyRaw)

	if err != nil {
		return nil, err
	}

	return &Service{
		privateKey: privateKey,
		publicKey:  privateKey.PublicKey,
	}, nil
}

func NewFromFile(privateKeyPath string) (*Service, error) {
	privateKeyRaw, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, err
	}
	return New(privateKeyRaw)
}

func MustNewFromFile(privateKeyPath string) *Service {
	interceptor, err := NewFromFile(privateKeyPath)
	if err != nil {
		panic(err)
	}
	return interceptor
}

func (a *Service) GetUser(ctx context.Context) (*UserClaims, error) {
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, ErrBadContext
	}

	auth := meta.Get(AuthHeaderKey)
	if len(auth) == 0 {
		return nil, ErrNoAuth
	}

	user, err := a.ParseToken(auth[0])

	if err != nil {
		return nil, ErrInvalidToken
	}

	return user, nil
}

func (a *Service) ParseToken(token string) (*UserClaims, error) {

	tokenData, err := jwt.ParseWithClaims(token, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, ErrInvalidToken
		}
		return a.publicKey, nil
	})

	if err != nil {
		return nil, err
	}
	if claims, ok := tokenData.Claims.(*UserClaims); ok && tokenData.Valid {
		return claims, nil
	}
	return nil, err
}

func (a *Service) SignToken(claims *UserClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	ss, err := token.SignedString(a.privateKey)
	if err != nil {
		return "", err
	}
	return ss, nil
}
