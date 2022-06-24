package tokens

import (
	"autharization/db"
	"autharization/entities"
	"errors"
	"log"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

var accsess_duration = time.Minute * 2
var refresh_duration = time.Minute * 7

type MyClaims struct {
	Value     string `json:"val,omitempty"`
	GUID      string `json:"guid"`
	ExpiresAt int64  `json:"exp"`
}

func (c MyClaims) Valid() error {
	if c.ExpiresAt >= time.Now().Unix() {
		return nil
	} else {
		return errors.New("token expired")
	}
}

type Manager interface {
	Refresh() error
	Accsess() error
	GetValues() (string, string)
	Parse() (*MyClaims, error)
}

type token_manager struct {
	key           string
	user          entities.User
	accsess_token string
	log           *log.Logger
	dbmanager     db.DBmanager
}

func NewTokenManagerWithTokens(key, accsess, refresh string, log *log.Logger) *token_manager {
	if key == "" || accsess == "" || refresh == "" {
		return nil
	}
	var mng = new(token_manager)
	if log != nil {
		mng.log = log
		mng.dbmanager = db.NewManager(log)
	} else {
		mng.log = nil
		mng.dbmanager = db.NewManager(nil)
	}
	mng.key = key
	mng.accsess_token = accsess
	mng.user.Refreshtoken.Token = refresh

	return mng
}

func NewTokenManagerWithGUID(key, value, guid string, log *log.Logger) *token_manager {
	if key == "" || value == "" || guid == "" {
		return nil
	}

	var manager = new(token_manager)
	manager.key = key
	manager.user = entities.User{GUID: guid, Value: value}

	if log != nil {
		manager.log = log
		manager.dbmanager = db.NewManager(log)
	} else {
		manager.log = nil
		manager.dbmanager = db.NewManager(nil)
	}

	return manager
}

func (m *token_manager) newAccses() (err error) {
	var token = jwt.NewWithClaims(jwt.SigningMethodHS512, MyClaims{
		Value:     m.user.Value,
		ExpiresAt: time.Now().Add(accsess_duration).Unix(),
		GUID:      m.user.GUID,
	})
	m.accsess_token, err = token.SignedString([]byte(m.key))
	return err
}

func (m *token_manager) newRefresh() (err error) {
	var token, er = bcrypt.GenerateFromPassword([]byte(m.accsess_token), bcrypt.DefaultCost)
	if er != nil {
		return er
	}
	m.user.Refreshtoken.Token = string(token)
	m.user.Refreshtoken.ExpiresAt = time.Now().Add(refresh_duration).Unix()
	return nil
}

func (m *token_manager) getNewPair() (err error) {
	err = m.newAccses()
	if err != nil {
		return err
	}

	err = m.newRefresh()
	if err != nil {
		return err
	}

	return nil
}

func (m *token_manager) Refresh() (err error) {
	m.dbmanager.Connect()
	defer m.dbmanager.Disconect()

	var claim, er = m.Parse()
	if er != nil && er.Error() != "token expired"{
		return er
	}

	m.user.GUID = claim.GUID
	m.user.Value = claim.Value

	var dbToken, e = m.dbmanager.CheckToken(m.user.GUID)
	if e != nil {
		return e
	}
	// is token correct and unmodifiled
	err = bcrypt.CompareHashAndPassword([]byte(dbToken.Token), []byte(m.accsess_token))

	switch {
	case err != nil:
		return nil
	case dbToken.Token != m.user.Refreshtoken.Token:
		return errors.New("modifiled token")
	case dbToken.ExpiresAt < time.Now().Unix():
		return errors.New("expired refresh token")
	}
	err = m.getNewPair()
	m.dbmanager.Replace(m.user)
	return err
}

func (m *token_manager) Accsess() (err error) {
	m.dbmanager.Connect()
	defer m.dbmanager.Disconect()

	err = m.getNewPair()
	if err != nil {
		return err
	}

	return m.dbmanager.Replace(m.user)
}

func (m *token_manager) GetValues() (string, string) {
	return m.accsess_token, m.user.Refreshtoken.Token
}

func (m *token_manager) Parse() (*MyClaims, error) {
	var token, err = jwt.ParseWithClaims(m.accsess_token, &MyClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(m.key), nil
	})

	if err != nil {
		if cl, ok := token.Claims.(*MyClaims); ok{
			return cl, err
		} else {
			return nil, errors.New("claims type error")
		}
	}

	if cl, ok := token.Claims.(*MyClaims); ok {
		return cl, nil
	}
	return nil, errors.New("claims type error")
}
