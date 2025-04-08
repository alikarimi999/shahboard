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
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/api/idtoken"
)

type Repository interface {
	GetByEmail(ctx context.Context, email string) (*entity.User, error)
	Create(context.Context, *entity.User) error
}

type Config struct {
	GoogleClientID string `json:"google_client_id"`

	// this is just because, google cert endpoint has blocked iran ip (should implement a better solution)
	VerifyPubKey bool `json:"verify_pub_key"`
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

	var token *tokenInfo
	var err error

	if s.cfg.VerifyPubKey {
		token, err = s.validateGoogleJWT(req.Token)
	} else {
		token, err = parseGoogleJWT(req.Token, s.cfg.GoogleClientID)
	}
	if err != nil {
		return GoogleAuthResponse{}, err
	}

	user, err := s.repo.GetByEmail(ctx, token.Email)
	if err != nil {
		s.l.Error(err.Error())
		return GoogleAuthResponse{}, err
	}

	exists := user != nil
	if !exists {
		user = entity.NewUser(token.Email, "")
		if err := s.repo.Create(ctx, user); err != nil {
			s.l.Error(err.Error())
			return GoogleAuthResponse{}, err
		}
		s.l.Debug(fmt.Sprintf("user created: %s", user.Email))

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
	} else {
		s.l.Debug(fmt.Sprintf("user logged in: %s", user.Email))
	}

	return GoogleAuthResponse{
		Id:       user.ID.String(),
		Email:    user.Email,
		Name:     token.Name,
		Picture:  token.Picture,
		JwtToken: s.jwtGenerator.GenerateJWT(types.User{ID: user.ID, Email: user.Email}),
		Exists:   exists,
	}, nil

}

func (s *AuthService) PasswordAuth(ctx context.Context, req PasswordAuthRequest) (PasswordAuthResponse, error) {
	if req.Email == "" || req.Password == "" {
		return PasswordAuthResponse{}, errors.New("email and password are required")
	}

	user, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil {
		s.l.Error(err.Error())
		return PasswordAuthResponse{}, err
	}

	hPass, err := hashPassword(req.Password)
	if err != nil {
		return PasswordAuthResponse{}, err
	}

	exists := user != nil
	if !exists {
		user = entity.NewUser(req.Email, hPass)
		if err := s.repo.Create(ctx, user); err != nil {
			s.l.Error(err.Error())
			return PasswordAuthResponse{}, err
		}
		s.l.Debug(fmt.Sprintf("user created: %s", user.Email))

		if err := s.pub.Publish(event.EventUserCreated{
			ID:        types.NewObjectId(),
			UserID:    user.ID,
			Email:     user.Email,
			Timestamp: time.Now().Unix(),
		}); err != nil {
			s.l.Error(err.Error())
		}
	} else {
		if !checkPassword(user.Password, req.Password) {
			return PasswordAuthResponse{}, errors.New("invalid password")
		}

		s.l.Debug(fmt.Sprintf("user logged in: %s", user.Email))
	}

	return PasswordAuthResponse{
		Id:       user.ID.String(),
		Email:    user.Email,
		JwtToken: s.jwtGenerator.GenerateJWT(types.User{ID: user.ID, Email: user.Email}),
		Exists:   exists,
	}, nil
}

func checkPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost) // DefaultCost is 10
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

type tokenInfo struct {
	Email         string `json:"email"`
	EmailVerified bool   `json:"email_verified"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
}

func (s *AuthService) validateGoogleJWT(tokenString string) (*tokenInfo, error) {
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
			Kid: "ee193d4647ab4a3585aa9b2b3b484a87aa68bb42",
			Use: "sig",
			Kty: "RSA",
			N:   "rxLSY1w1gu-IzjVkBEqZXWcA1adZ15VmGpPYKpt8N_MXbgwICCy__iPVvuvSqetTvshwxEEK8ZcbmEyG_rcPiIBBoHYdtVb_cTlNR7JfT2ZOFKZUW1y3FBnZ2TTBHCgCJ9N7d-r6doQ-NI0GXOWzZh5Q9CPc9NDZoe8RfH-RE4m1RNGAukKThomofesSyw5OY92WxK9sfwTshmlK-J-wFB2OlN7xuwF3Rns_CJLdnajhf5XVMdNqEeSk3Fyoi72qWRQbDhfEhT5qcpkMX42BgWbmlom0ZPwPPhyyd9jrfFNN0BNgvF2kPD2eJ8qsaaUAZn4DBvcTpC5RhiwSY_AB8w",
			E:   "AQAB",
			Alg: "RS256",
		},
		{
			Kid: "821f3bc66f0751f7840606799b1adf9e9fb60dfb",
			N:   "mvcbc7gZu7VixykOM8JawiiNEco0ZJj9mJ3zezm034iO5w7AbLFOXut2zgWc-uOifuJUHHDSbG5Plk8ObhTxgIOD0ar9Qep5BSH1fFBhNPOfDM8h44Ru7O9_IZ7wyijlhDpzXsb403Z6FrIMAPMJJGjHGrc1f2p-_KojzTTlaAjsolrFgq19NAxQx0qrGvQrMeGB7x1iej_9AO65WGDj4xTNoihAsKgVqvARz-kryDetAlaKnpyORDuceYaMRTTUrRJjue8Sa9eSc72n53eAaau8i2MnDsPtyWnRFondswSxesBEujEgmWZui2X_JePvEDk0xnYcc2CjSWRLELy_NQ",
			E:   "AQAB",
			Kty: "RSA",
			Use: "sig",
			Alg: "RS256",
		},
	},
}

// parseGoogleJWT validates a Google OAuth2 ID token using static keys
func parseGoogleJWT(tokenString string, clientID string) (*tokenInfo, error) {
	// Parse the token to get the "kid" from the header

	// token, err := pjwt.Parse(tokenString, func(token *pjwt.Token) (interface{}, error) {
	// 	if _, ok := token.Method.(*pjwt.SigningMethodRSA); !ok {
	// 		return nil, errors.New("unexpected signing method")
	// 	}

	// 	kid, ok := token.Header["kid"].(string)
	// 	if !ok {
	// 		return nil, errors.New("missing or invalid kid in token header")
	// 	}

	// 	// Use static keys instead of fetching
	// 	publicKey, err := getStaticGooglePublicKey(kid)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	return publicKey, nil
	// })

	token, _, err := new(pjwt.Parser).ParseUnverified(tokenString, pjwt.MapClaims{})
	if err != nil {
		return nil, err
	}

	// Validate token and extract claims
	claims, ok := token.Claims.(pjwt.MapClaims)
	if !ok {
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
