package auth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

var (
	privateKey, _ = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	x509Bytes, _  = x509.MarshalPKCS8PrivateKey(privateKey)
	pemEncoded    = pem.EncodeToMemory(&pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: x509Bytes,
	})
)

func TestService_New(t *testing.T) {
	service, err := New(pemEncoded)
	assert.NoErrorf(t, err, "should not return any errors")
	assert.Equal(t, privateKey, service.privateKey, "must be equal")
}

func TestService_SignToken(t *testing.T) {
	service, _ := New(pemEncoded)
	issuedAt := time.Now().UTC()
	expiresAt := issuedAt.Add(3 * time.Hour)
	claims := &UserClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
		Username:  "johndoe",
		Email:     "doe@example.com",
		FirstName: "John",
		LastName:  "John",
		AvatarID:  "1cbbb567-6e35-4267-abe5-41cf2208fbe8",
	}
	tokenString, err := service.SignToken(claims)

	assert.NoError(t, err, "should not return any errors")
	parsedClaims, err := service.ParseToken(tokenString)
	assert.NoError(t, err, "should not return any errors")
	assert.True(t, assert.ObjectsAreEqualValues(claims, parsedClaims), "should have same values")
}
