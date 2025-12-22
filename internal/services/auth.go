package services

import (
	"bookwise/internal/config"
	"bookwise/internal/models"
	e "bookwise/utils/errors"
	"bookwise/utils/validator"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	user   UserService
	config config.Config
}

type AuthServiceInterface interface {
	Login(v *validator.Validator, email, password string) (string, error)
	ExtractUsername(tokenString string) (string, error)
}

func NewAuthService(userService UserService, config config.Config) *AuthService {
	return &AuthService{
		user:   userService,
		config: config,
	}
}

func (s *AuthService) Login(
	v *validator.Validator,
	email,
	password string,
) (string, error) {
	models.ValidateEmail(v, email)
	models.ValidatePasswordPlaintext(v, password)

	if !v.Valid() {
		return "", e.ErrInvalidData
	}

	user, err := s.user.GetUserByEmail(email, v)
	if err != nil {
		switch {
		case errors.Is(err, e.ErrRecordNotFound):
			return "", e.ErrInvalidCredentials
		default:
			return "", err
		}
	}

	if !user.Activated {
		return "", e.ErrInactiveAccount
	}

	match, err := user.Password.Matches(password)
	if err != nil {
		return "", err
	}

	if !match {
		return "", e.ErrInvalidCredentials
	}

	token, err := s.createToken(user.Email)
	if err != nil {
		return "", err
	}

	return token, nil
}

func (s *AuthService) createToken(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"username": username,
			"exp":      time.Now().Add(time.Hour * 24).Unix(),
		})
	tokenStr, err := token.SignedString([]byte(s.config.Security.SecretKey))

	if err != nil {
		return "", err
	}

	return tokenStr, nil
}

func (s *AuthService) ExtractUsername(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return []byte(s.config.Security.SecretKey), nil
	})

	if err != nil {
		return "", err
	}

	if !token.Valid {
		return "", nil
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", nil
	}

	username, ok := claims["username"].(string)
	if !ok {
		return "", nil
	}

	return username, nil
}
