package jwt

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"os"
	"time"

	"github.com/alikarimi999/shahboard/types"
	"github.com/golang-jwt/jwt"
)

type GeneratorConfig struct {
	PrivateKeyPath string `json:"private_key_path"`
	Expiration     uint   `json:"expiration_in_seconds"`
}

type Generator struct {
	privateKey *rsa.PrivateKey
	expiration time.Duration
}

func NewGenerator(cfg GeneratorConfig) (*Generator, error) {
	if cfg.Expiration == 0 {
		return nil, errors.New("expiration must be greater than 0")
	}

	privateKey, err := loadPrivateKey(cfg.PrivateKeyPath)
	if err != nil {
		return nil, err
	}

	return &Generator{
		privateKey: privateKey,
		expiration: time.Duration(cfg.Expiration) * time.Second,
	}, nil
}

func (g *Generator) GenerateJWT(u types.User) string {
	t := time.Now()

	claims := jwt.MapClaims{
		"id":       u.ID.String(),
		"email":    u.Email,
		"is_guest": u.IsGuest,
		"exp":      t.Add(g.expiration).Unix(),
		"iat":      t.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)

	signedToken, _ := token.SignedString(g.privateKey)

	return signedToken
}

type ValidatorConfig struct {
	PublicKeyPath string `json:"public_key_path"`
}

type Validator struct {
	publicKey *rsa.PublicKey
}

func NewValidator(cfg ValidatorConfig) (*Validator, error) {
	publicKey, err := loadPublicKey(cfg.PublicKeyPath)
	if err != nil {
		return nil, err
	}

	return &Validator{
		publicKey: publicKey,
	}, nil
}

func (v *Validator) ValidateJWT(tokenString string) (types.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return v.publicKey, nil
	})

	if err != nil {
		return types.User{}, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return types.User{}, errors.New("invalid token")
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return types.User{}, errors.New("exp claim missing or invalid")
	}
	if time.Now().Unix() > int64(exp) {
		return types.User{}, errors.New("token has expired")
	}

	id, idOk := claims["id"].(string)
	email, emailOk := claims["email"].(string)
	if !idOk || !emailOk {
		return types.User{}, errors.New("missing or invalid id or email claim")
	}

	var isGuest bool
	if _, ok := claims["is_guest"]; ok {
		isGuest = claims["is_guest"].(bool)
	}

	return types.User{
		ID:      types.ObjectId(id),
		Email:   email,
		IsGuest: isGuest,
	}, nil
}

func loadPrivateKey(privateKeyPath string) (*rsa.PrivateKey, error) {

	keyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyBytes)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, errors.New("invalid private key format")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}

func loadPublicKey(publicKeyPath string) (*rsa.PublicKey, error) {
	keyBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(keyBytes)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, errors.New("invalid public key format")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	rsaPubKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not a valid RSA public key")
	}

	return rsaPubKey, nil
}
