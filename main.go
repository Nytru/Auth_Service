package main

import (
	"autharization/entities"
	"autharization/tokens"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

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
	// DBpath = os.Getenv("DB_FULL_PASS")
	var file, err = os.OpenFile("logs.txt", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}

	logger = log.New(file, "Debug: ", log.Flags())
}

func HandleFuncNew(w http.ResponseWriter, r *http.Request) {
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
	
	var tok = tokens.NewTokenManagerWithGUID(Key, user.Value, user.GUID, logger)
	if tok == nil {
		w.WriteHeader(http.StatusForbidden)
	}
	if e := tok.Accsess(); e != nil {
		w.WriteHeader(404)
	}

	var access, refresh = tok.GetValues()
	cookie := http.Cookie{
		Name: "accsess",
		Value: access,
	}

	http.SetCookie(w, &cookie)
	cookie = http.Cookie{
		Name: "refresh",
		Value: refresh,
	}
	http.SetCookie(w, &cookie)

	logger.Println("new tokens: \naccess: ", access, "\nrefresh: ", refresh)
	logger.Println("succses new request, method: ", r.Method)
}

func HandleFuncRefresh(w http.ResponseWriter, r *http.Request) {
	var access, refresh string
	var header = r.Header
	for k, v := range header {
		if k == "Cookie" {
			var arr = strings.Split(v[0], ";")
			access = strings.Split(arr[0], "=")[1]
			refresh = strings.Split(arr[1], "=")[1]
			logger.Println(access, "\t", refresh)
		}
	}

	if access == "" || refresh == "" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	var manager = tokens.NewTokenManagerWithTokens(Key, access, refresh, logger)
	var err = manager.Refresh()
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(err.Error()))
		logger.Println(err)
		return
	}

	access, refresh = manager.GetValues()
	cookie := http.Cookie{
		Name: "accsess",
		Value: access,
	}

	http.SetCookie(w, &cookie)
	cookie = http.Cookie{
		Name: "refresh",
		Value: refresh,
	}
	http.SetCookie(w, &cookie)

	logger.Println("new tokens: \naccess: ", access, "\nrefresh: ", refresh)
	logger.Println("succses refresh request, method: ", r.Method)
}

func main() {
	http.HandleFunc("/getnewtokens", HandleFuncNew)
	http.HandleFunc("/refreshtoken", HandleFuncRefresh)
	logger.Println("Server started")
	http.ListenAndServe(":8080", nil)
	logger.Println("Server stopped")
}
