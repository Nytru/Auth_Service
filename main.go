package main

import (
	"authentication/entities"
	"authentication/tokens"
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Logger
var logger *log.Logger

func init() {
	godotenv.Load("env/.env")

	var file, err = os.OpenFile("logs.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}

	logger = log.New(file, "Debug: ", log.Flags())
}

func GetNewTokensHandler(w http.ResponseWriter, r *http.Request) {
	var user entities.User
	if len(r.URL.Query()) > 2 {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, "invalid querry request")
		logger.Println("invalid querry request")
	}

	for key, que := range r.URL.Query() {
		switch key {
		case "value":
			user.Value = que[0]
		case "guid":
			user.GUID = que[0]
		default:
			w.WriteHeader(http.StatusBadRequest)
			io.WriteString(w, "invalid querry request")
			logger.Println("invalid querry request")
			return
		}
	}

	var tok = tokens.NewTokenProviderWithGUID(os.Getenv("KEY"), user.Value, user.GUID, logger)
	if tok == nil {
		w.WriteHeader(http.StatusForbidden)
	}
	if e := tok.Accsess(); e != nil {
		w.WriteHeader(404)
	}

	var duration, _ = strconv.Atoi(os.Getenv("REFRESH_DURATION"))
	var access, refresh = tok.GetValues()
	cookie := http.Cookie{
		Name: "accsess",
		Value: access,
		Expires: time.Now().Add(time.Duration(duration)),
	}
	http.SetCookie(w, &cookie)
	cookie = http.Cookie{
		Name: "refresh",
		Value: base64.URLEncoding.EncodeToString([]byte(refresh)),
		Expires: time.Now().Add(time.Duration(duration)),
	}
	http.SetCookie(w, &cookie)

	logger.Println("new tokens: \naccess: ", access, "\nrefresh: ", refresh)
	logger.Println("succses new request, method: ", r.Method)
}

func RefreshTokensHandler(w http.ResponseWriter, r *http.Request) {
	var access, refresh string
	var header = r.Header
	for k, v := range header {
		if k == "Cookie" {
			var arr = strings.Split(v[0], ";")
			access = strings.Split(arr[0], "=")[1]
			refresh = strings.Split(arr[1], "=")[1]
			logger.Println("cookie from refresh request", access, "\t", refresh)
		}
	}

	if access == "" || refresh == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	var txt, err = base64.URLEncoding.DecodeString(refresh)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		logger.Println(err)
	}
	refresh = string(txt)

	var manager = tokens.NewTokenManagerWithTokens(os.Getenv("KEY"), access, refresh, logger)
	err = manager.Refresh()
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.Error()))
		logger.Println(err)
		return
	}

	var duration, _ = strconv.Atoi(os.Getenv("REFRESH_DURATION"))
	access, refresh = manager.GetValues()
	cookie := http.Cookie{
		Name:  "accsess",
		Value: access,
		Expires: time.Now().Add(time.Duration(duration)),
	}
	http.SetCookie(w, &cookie)
	cookie = http.Cookie{
		Name:  "refresh",
		Value: base64.URLEncoding.EncodeToString([]byte(refresh)),
		Expires: time.Now().Add(time.Duration(duration)),
	}
	http.SetCookie(w, &cookie)

	logger.Println("new tokens: \naccess: ", access, "\nrefresh: ", refresh)
	logger.Println("succses refresh request, method: ", r.Method)
}

func main() {
	http.HandleFunc("/getnewtokens", GetNewTokensHandler)
	http.HandleFunc("/refreshtoken", RefreshTokensHandler)
	logger.Println("Server started")
	http.ListenAndServe(":8080", nil)
	logger.Println("Server stopped")
}
