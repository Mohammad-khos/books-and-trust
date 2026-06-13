package auth

import (
	"books-and-trust/services/user-service/internal/domain"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTAuthenticator struct {
	secret string
	aud    string
	iss    string
	exp time.Duration
}

func NewJWTAuthenticator(secret string, aud string, iss string , exp time.Duration) *JWTAuthenticator {
	return &JWTAuthenticator{
		secret: secret,
		aud:    aud,
		iss:    iss,
		exp: exp,
	}
}

func (ja *JWTAuthenticator) GenerateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"sub": userID,
		"exp": time.Now().Add(ja.exp).Unix(),
		"iss": ja.iss,
		"aud": ja.aud,
		"iat": time.Now().Unix(),
		"nbf": time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(ja.secret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (ja *JWTAuthenticator) VerifyToken(token string) (string, error) {
	jwtToken, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		_, ok := t.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(ja.secret), nil

	},
		jwt.WithExpirationRequired(),
		jwt.WithAudience(ja.aud),
		jwt.WithIssuer(ja.iss),
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}),
	)
	if err != nil {
		return "", err
	}

	if claims, ok := jwtToken.Claims.(jwt.MapClaims); ok && jwtToken.Valid  {
		userID , ok := claims["sub"].(string)
	
		if !ok {
			return "" , domain.ErrInvalidToken
		}
	
		return userID , nil
	
	}
	return "" , domain.ErrInvalidToken
}
