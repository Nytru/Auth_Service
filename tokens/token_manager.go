package tokens

import (
	"autharization/entities"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/golang-jwt/jwt"
)

type MyClaims struct{
	Subject   string `json:"sub,omitempty"`
	ExpiresAt int64  `json:"exp,omitempty"`
}

func (c MyClaims)Valid() error {
	if c.ExpiresAt >= time.Now().Unix() {
		return nil
	}else {
		return errors.New("token expired")
	}
}


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

func Give(Key string)  {
	var token = jwt.NewWithClaims(jwt.SigningMethodHS512, MyClaims{
		ExpiresAt: time.Now().Add(time.Second * 20).Unix(),
		Subject: "4517",
	})
	var text, e = token.SignedString([]byte(Key))
	if e != nil {
		log.Fatal(e)
	}
	
	fmt.Println(token.Header)
	fmt.Println(token.Claims)
	fmt.Println(token.Signature)
	fmt.Println(text)
	
	var parsedToken, err = jwt.ParseWithClaims(text, &MyClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(Key), nil
	})

	if err != nil {
		log.Fatal(err)
	}
	if v, ok := parsedToken.Claims.(*MyClaims); ok {
		fmt.Println(v.Subject)
	}
}