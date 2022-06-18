package main

import (
	// "autharization/db"
	"fmt"
	"log"
	"os"
	"reflect"
	"time"

	"github.com/golang-jwt/jwt"
	"github.com/joho/godotenv"
)

type MyClaims struct{
	Subject   string `json:"sub,omitempty"`
	ExpiresAt int64  `json:"exp,omitempty"`
}

func (c MyClaims)Valid() error {
	return nil
}

// Cannot be zero lenth
var Key string

func main() {
	godotenv.Load("env/environment.env")
	Key = os.Getenv("KEY")
	if len(Key) == 0 {
		log.Fatal("empty key value")
	}	

	var token = jwt.NewWithClaims(jwt.SigningMethodHS512, MyClaims{
		ExpiresAt: time.Now().Add(time.Minute * 2).Unix(),
		Subject: "4517",
	})
	var text, e = token.SignedString([]byte("11"))
	if e != nil {
		log.Fatal(e)
	}
	
	fmt.Println(token.Header)
	fmt.Println(token.Claims)
	fmt.Println(token.Signature)
	fmt.Println(text)

	var parsedToken, err = jwt.ParseWithClaims(text, &MyClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte("11"), nil
	})

	if err != nil {
		log.Fatal(err)
	}
	if v, ok := parsedToken.Claims.(*MyClaims); ok {
		fmt.Println(v.Subject)
	}else {
		fmt.Println(reflect.TypeOf(parsedToken.Claims))
		
	}

	// http.HandleFunc("/getnewtokens", func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Printf("got new tokens request\n")
	// 	fmt.Println(r.URL.Query())
    // 	io.WriteString(w, "NEGRI PIDORASI\n")
	// })
	// http.HandleFunc("/refreshtoken", func(w http.ResponseWriter, r *http.Request) {
	// 	r.URL.Query()
	// 	fmt.Printf("got refresh request\n")
    // 	io.WriteString(w, "negri pidorasi no ne kapsom\n")
	// })
	
	// http.ListenAndServe(":8080", nil)
}
