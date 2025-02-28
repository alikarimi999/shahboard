package service

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"time"

	"github.com/alikarimi999/shahboard/authservice/entity"
	"github.com/alikarimi999/shahboard/event"
	"github.com/alikarimi999/shahboard/pkg/jwt"
	"github.com/alikarimi999/shahboard/pkg/log"
	"github.com/alikarimi999/shahboard/types"
	pjwt "github.com/golang-jwt/jwt"
	"google.golang.org/api/idtoken"
)

type Repository interface {
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	Create(context.Context, *entity.User) error
}

type Config struct {
	GoogleClientID string `json:"google_client_id"`
}

type AuthService struct {
	cfg          Config
	repo         Repository
	jwtGenerator *jwt.Generator
	pub          event.Publisher
	l            log.Logger
}

func NewAuthService(cfg Config, repo Repository, jwtGenerator *jwt.Generator, pub event.Publisher, l log.Logger) *AuthService {
	return &AuthService{
		cfg:          cfg,
		repo:         repo,
		jwtGenerator: jwtGenerator,
		pub:          pub,
		l:            l,
	}
}

func (s *AuthService) GoogleAuth(ctx context.Context, req GoogleAuthRequest) (GoogleAuthResponse, error) {
	// token, err := s.parseGoogleToken(req.Token)
	// if err != nil {
	// 	return GoogleAuthResponse{}, err
	// }

	token, err := ValidateGoogleJWT(req.Token, s.cfg.GoogleClientID)

	user, err := s.repo.GetByEmail(ctx, token.Email)
	if err != nil {
		s.l.Error(err.Error())
		return GoogleAuthResponse{}, err
	}

	exists := user != nil
	if !exists {
		user = entity.NewUser(token.Email)
		if err := s.repo.Create(ctx, user); err != nil {
			s.l.Error(err.Error())
			return GoogleAuthResponse{}, err
		}
	}

	go func() {
		if exists {
			if err := s.pub.Publish(event.EventUserLoggedIn{
				ID:        types.NewObjectId(),
				UserID:    user.ID,
				Email:     user.Email,
				Timestamp: time.Now().Unix(),
			}); err != nil {
				s.l.Error(err.Error())
			}
			return
		}

		if err := s.pub.Publish(event.EventUserCreated{
			ID:        types.NewObjectId(),
			UserID:    user.ID,
			Email:     user.Email,
			Name:      token.Name,
			Picture:   token.Picture,
			Timestamp: time.Now().Unix(),
		}); err != nil {
			s.l.Error(err.Error())
		}

	}()

	return GoogleAuthResponse{
		Id:       user.ID.String(),
		Email:    user.Email,
		Name:     token.Name,
		Picture:  token.Picture,
		JwtToken: s.jwtGenerator.GenerateJWT(types.User{ID: user.ID, Email: user.Email}),
		Exists:   exists,
	}, nil

}

type tokenInfo struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

func (s *AuthService) parseGoogleToken(tokenString string) (*tokenInfo, error) {
	p, err := idtoken.Validate(context.Background(), tokenString, s.cfg.GoogleClientID)
	if err != nil {
		s.l.Error(err.Error())
		return nil, err
	}

	if !p.Claims["email_verified"].(bool) {
		return nil, errors.New("email not verified")
	}

	return &tokenInfo{
		Email:         p.Claims["email"].(string),
		EmailVerified: p.Claims["email_verified"].(bool),
		Name:          p.Claims["name"].(string),
		Picture:       p.Claims["picture"].(string),
	}, nil
}

// GoogleKey represents a single key in the Google certs JSON
type GoogleKey struct {
	Kid string `json:"kid"`
	Use string `json:"use"`
	Kty string `json:"kty"`
	N   string `json:"n"` // Modulus
	E   string `json:"e"` // Exponent
	Alg string `json:"alg"`
}

// GoogleCerts holds the list of keys
type GoogleCerts struct {
	Keys []GoogleKey `json:"keys"`
}

// Static Google public keys (from your JSON)
var staticGoogleCerts = GoogleCerts{
	Keys: []GoogleKey{
		{
			Kid: "5d12ab782cb6096285f69e48aea99079bb59cb86",
			Use: "sig",
			Kty: "RSA",
			N:   "uac7NRcojCutcceWq1nrpLGJjQ7ywvgWsUcb1DWMKJ3KNNHiRzh9jshoi9tmq1zlarJ_h7GQg8iU1qD7SgpVYJmjlKG1MNVRAtuNrNMC0UAnNfG7mBBNorHFndfp-9cLTiMjXSXRzhNqiMvTVKeolRdMB2lH9RzJnwlpXtvUbD7M1pXOlPlMaOy1zxUnHn0uszU5mPRQk79i03BNrAdhwrAUB-ZuMnqpjaUcb9VU3KIwuZNPtsVenLN12sRYpaZ6WBw8Q9q7fAoaJUovM0Go8deC9pJYyxJuHdVo9HP0osyzg3g_rOYi14wmvMBuiDf3F4pTnudAfFyl3d0Mn_i4ZQ",
			E:   "AQAB",
			Alg: "RS256",
		},
		{
			Kid: "763f7c4cd26a1eb2b1b39a88f4434d1f4d9a368b",
			N:   "y8TPCPz2Fp0OhBxsxu6d_7erT9f9XJ7mx7ZJPkkeZRxhdnKtg327D4IGYsC4fLAfpkC8qN58sZGkwRTNs-i7yaoD5_8nupq1tPYvnt38ddVghG9vws-2MvxfPQ9m2uxBEdRHmels8prEYGCH6oFKcuWVsNOt4l_OPoJRl4uiuiwd6trZik2GqDD_M6bn21_w6AD_jmbzN4mh8Od4vkA1Z9lKb3Qesksxdog-LWHsljN8ieiz1NhbG7M-GsIlzu-typJfud3tSJ1QHb-E_dEfoZ1iYK7pMcojb5ylMkaCj5QySRdJESq9ngqVRDjF4nX8DK5RQUS7AkrpHiwqyW0Csw",
			E:   "AQAB",
			Kty: "RSA",
			Use: "sig",
			Alg: "RS256",
		},
	},
}

// ValidateGoogleJWT validates a Google OAuth2 ID token using static keys
func ValidateGoogleJWT(tokenString string, clientID string) (*tokenInfo, error) {
	// Parse the token to get the "kid" from the header
	token, err := pjwt.Parse(tokenString, func(token *pjwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*pjwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method")
		}

		kid, ok := token.Header["kid"].(string)
		if !ok {
			return nil, errors.New("missing or invalid kid in token header")
		}

		// Use static keys instead of fetching
		publicKey, err := getStaticGooglePublicKey(kid)
		if err != nil {
			return nil, err
		}

		return publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	// Validate token and extract claims
	claims, ok := token.Claims.(pjwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Verify standard JWT claims
	exp, ok := claims["exp"].(float64)
	if !ok || time.Now().Unix() > int64(exp) {
		return nil, errors.New("token has expired or missing exp claim")
	}

	iss, ok := claims["iss"].(string)
	if !ok || (iss != "https://accounts.google.com" && iss != "accounts.google.com") {
		return nil, errors.New("invalid issuer")
	}

	aud, ok := claims["aud"].(string)
	if !ok || aud != clientID {
		return nil, errors.New("invalid audience")
	}

	email, ok := claims["email"].(string)
	if !ok {
		return nil, errors.New("missing or invalid email claim")
	}

	return &tokenInfo{
		Email:         email,
		EmailVerified: claims["email_verified"].(bool),
		Name:          claims["name"].(string),
		Picture:       claims["picture"].(string),
	}, nil
}

// getStaticGooglePublicKey retrieves the RSA public key from the static data
func getStaticGooglePublicKey(kid string) (*rsa.PublicKey, error) {
	for _, key := range staticGoogleCerts.Keys {
		if key.Kid == kid {
			// Decode the modulus (n) from base64url
			nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
			if err != nil {
				return nil, fmt.Errorf("failed to decode modulus: %v", err)
			}

			// Decode the exponent (e) from base64url
			eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
			if err != nil {
				return nil, fmt.Errorf("failed to decode exponent: %v", err)
			}

			// Construct the RSA public key manually
			pubKey := &rsa.PublicKey{
				N: new(big.Int).SetBytes(nBytes),
				E: int(new(big.Int).SetBytes(eBytes).Int64()),
			}

			return pubKey, nil
		}
	}

	return nil, errors.New("no matching public key found for kid: " + kid)
}
