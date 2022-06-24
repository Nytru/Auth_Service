package main

import (
	// "autharization/db"
	"autharization/entities"
	"autharization/tokens"
	"errors"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	_"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
)

// Cannot be zero lenth
var Key string // EncryptingKey
var DbName string // mongodb name
var DbPassword string // mongodb password
var DBpath string // if exsit can be used for connection

// Logger
var logger *log.Logger

func init() {
	godotenv.Load("env/.env")
	var ok bool	
	if Key, ok = os.LookupEnv("KEY"); !ok {
		panic("Empty env err")
	}
	DBpath = os.Getenv("DB_FULL_PASS")
	var file, err = os.OpenFile("logs.txt", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}

	logger = log.New(file, "Debug: ", log.Flags())
}

func HandleFuncNew(w http.ResponseWriter, r *http.Request) {
	var user entities.User
	// var ans string
	for key, que := range r.URL.Query() {
		switch key {
		case "name":
			user.Value = que[0]
		case "guid":
			user.GUID = que[0]
		default:
			io.WriteString(w, "invalid querry request")
			logger.Println("invalid querry request")
			return
		}
	}
	
	var tok = tokens.NewTokenManagerWithGUID(Key, user.Value, user.GUID, logger)
	if tok == nil {
		w.WriteHeader(http.StatusForbidden)
	}
	if e := tok.Accsess(); e != nil {
		w.WriteHeader(404)
	}
	var access, refresh = tok.GetValues()
	w.Header().Add("accsess", access)
	w.Header().Add("refresh", refresh)


	logger.Println("succses request, method: ", r.Method)
	
}

func HandleFuncRefresh(w http.ResponseWriter, r *http.Request) {
	var user entities.User
	for key, que := range r.URL.Query() {
		switch key {
		case "name":
			user.Value = que[0]
		case "guid":
			user.GUID = que[0]
		default:
			w.WriteHeader(http.StatusForbidden)
			logger.Println("invalid querry request")
			return
		}
	}
	
	logger.Println("succses request, method: ", r.Method)
}

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

func main() {
	// var tok = jwt.NewWithClaims(jwt.SigningMethodHS512, MyClaims{Value: "q", GUID: "12", ExpiresAt: time.Now().Add(time.Minute).Unix()})
	// var token, er = tok.SignedString([]byte("12"))
	// if er != nil {
	// 	panic(er)
	// }

	// parsed, err := jwt.ParseWithClaims(token, &MyClaims{}, func(t *jwt.Token) (interface{}, error) {return []byte("12"), nil})
	// if err != nil {
	// 	panic(err)
	// }
	// if mc, ok := parsed.Claims.(*MyClaims); ok {
	// 	fmt.Println(mc.Value, "\n", mc.GUID)
	// } else {
	// 	fmt.Println("lox")
	// }
	var tok1 = tokens.NewTokenManagerWithGUID("45", "value", "43etrgdgfv", logger)
	tok1.Accsess()
	var a, b = tok1.GetValues()
	var tok2 = tokens.NewTokenManagerWithTokens("45", a, b, logger)
	tok2.Refresh()
	return

	http.HandleFunc("/getnewtokens", HandleFuncNew)
	http.HandleFunc("/refreshtoken", HandleFuncRefresh)
	logger.Println("Server started")
	http.ListenAndServe(":8080", nil)
	logger.Println("Server stopped")
}
