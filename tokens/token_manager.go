package tokens

import (
	"authentication/db"
	"authentication/entities"
	"errors"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

type LiteClaims struct {
	Value     string `json:"val,omitempty"`
	GUID      string `json:"guid"`
	ExpiresAt int64  `json:"exp"`
}

func (c LiteClaims) Valid() error {
	if c.ExpiresAt >= time.Now().Unix() {
		return nil
	} else {
		return errors.New("token expired")
	}
}

type TokensProvider interface {
	Refresh() error
	Accsess() error
	GetValues() (string, string)
	Parse() (*LiteClaims, error)
}

type tokensManager struct {
	key           string
	user          entities.User
	accsess_token string
	log           *log.Logger
	dbmanager     db.DBAccessProvider
}

func NewTokenManagerWithTokens(key, accsess, refresh string, log *log.Logger) *tokensManager {
	if key == "" || accsess == "" || refresh == "" {
		return nil
	}
	var mng = new(tokensManager)
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

func NewTokenProviderWithGUID(key, value, guid string, log *log.Logger) *tokensManager {
	if key == "" || value == "" || guid == "" {
		return nil
	}

	var manager = new(tokensManager)
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

func (m *tokensManager) newAccses() (err error) {
	var accsess_duration, _ = strconv.Atoi(os.Getenv("ACCESS_DURATION")) 

	var token = jwt.NewWithClaims(jwt.SigningMethodHS512, LiteClaims{
		Value:     m.user.Value,
		ExpiresAt: time.Now().Add(time.Duration(accsess_duration)).Unix(),
		GUID:      m.user.GUID,
	})
	m.accsess_token, err = token.SignedString([]byte(m.key))
	return err
}

func (m *tokensManager) newRefresh() (err error) {
	var token, er = bcrypt.GenerateFromPassword([]byte(m.accsess_token), bcrypt.DefaultCost)
	if er != nil {
		return er
	}
	m.user.Refreshtoken.Token = string(token)
	var refresh_duration, _ = strconv.Atoi(os.Getenv("REFRESH_DURATION"))
	m.user.Refreshtoken.ExpiresAt = time.Now().Add(time.Duration(refresh_duration)).Unix()
	return nil
}

func (m *tokensManager) getNewPair() (err error) {
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

func (m *tokensManager) Refresh() (err error) {
	m.dbmanager.Connect()
	defer m.dbmanager.Disconnect()

	var claim, er = m.Parse()
	if er != nil && er.Error() != "token expired" {
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

func (m *tokensManager) Accsess() (err error) {
	m.dbmanager.Connect()
	defer m.dbmanager.Disconnect()

	err = m.getNewPair()
	if err != nil {
		return err
	}

	return m.dbmanager.Replace(m.user)
}

func (m *tokensManager) GetValues() (string, string) {
	return m.accsess_token, m.user.Refreshtoken.Token
}

func (m *tokensManager) Parse() (*LiteClaims, error) {
	var token, err = jwt.ParseWithClaims(m.accsess_token, &LiteClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(m.key), nil
	})

	if err != nil {
		if cl, ok := token.Claims.(*LiteClaims); ok {
			return cl, err
		} else {
			return nil, errors.New("claims type error")
		}
	}

	if cl, ok := token.Claims.(*LiteClaims); ok {
		return cl, nil
	}
	return nil, errors.New("claims type error")
}
