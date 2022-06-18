package tokens

import (
	"autharization/entities"
	"github.com/golang-jwt/jwt"
)

type Manager interface {
	GetTokens() (jwt.Token, RefreshToken)
}

type token_manager struct {
	user entities.User
}

func (m *token_manager) GetTokens() (jwt.Token, RefreshToken) {
	return jwt.Token{}, RefreshToken{}
}

func NewTokenManager(User entities.User) *token_manager {
	return &token_manager{user: User}
}