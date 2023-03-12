package auth

import (
	"context"
	"crypto/ecdsa"
	"crypto/x509"
	"errors"
	"github.com/golang-jwt/jwt/v5"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"os"
)

const AuthHeaderKey = "Authorization"
const UserKey = "AuthorizedUser"

var (
	ErrInvalidToken = status.Error(codes.Unauthenticated, "provided token is invalid")
)

type UserClaims struct {
	jwt.RegisteredClaims
	Username  string `json:"username"`
	Email     string `json:"email"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	AvatarID  string `json:"avatar_id"`
}

type AuthService struct {
	privateKey *ecdsa.PrivateKey
	publicKey  ecdsa.PublicKey
}

func NewFromFile(privateKeyPath string) (*Interceptor, error) {
	privateKeyRaw, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, err
	}

	key, err := x509.ParsePKCS8PrivateKey(privateKeyRaw)

	if err != nil {
		return nil, err
	}

	privateKey, ok := key.(*ecdsa.PrivateKey)

	if !ok {
		return nil, errors.New("key has invalid format")
	}

	return &Interceptor{
		privateKey: privateKey,
		publicKey:  privateKey.PublicKey,
	}, nil
}

func MustNewFromFile(privateKeyPath string) *Interceptor {
	interceptor, err := NewFromFile(privateKeyPath)
	if err != nil {
		panic(err)
	}
	return interceptor
}

func (a *Interceptor) Callback(ctx context.Context) error {
	meta, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return status.Error(codes.InvalidArgument, "can't retrieve metadata")
	}

	auth := meta.Get(AuthHeaderKey)
	if len(auth) == 0 {
		return nil
	}
	user, err := ParseToken(auth[0], a.publicKey)
	context.Set
	if err != nil {
		return ErrInvalidToken
	}

	return nil
}

func ParseToken(token string, signingKey ecdsa.PublicKey) (*UserClaims, error) {

	tokenData, err := jwt.ParseWithClaims(token, &UserClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodECDSA); !ok {
			return nil, ErrInvalidToken
		}
		return signingKey, nil
	})

	if err != nil {
		return nil, err
	}
	if claims, ok := tokenData.Claims.(*UserClaims); ok && tokenData.Valid {
		return claims, nil
	}
	return nil, err
}

func GetUserFromContext() {

}
