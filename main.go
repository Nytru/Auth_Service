package main

import (
	"autharization/db"
	_ "autharization/entities"
	"fmt"

	// "fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "go.mongodb.org/mongo-driver/mongo"
	_ "go.mongodb.org/mongo-driver/mongo/options"
)

// Cannot be zero lenth
var Key string // EncryptingKey
var DbName string // mongodb name
var DbPassword string // mongodb password
var DBpath string // if exsit can be used for connection

func main() {
	fmt.Println("Programm start")
	godotenv.Load("env/environment.env")
	Key = os.Getenv("KEY")
	if len(Key) == 0 {
		log.Fatal("empty env value")
	}	
	
	if str, ok := os.LookupEnv("DB_FULL_PASS"); ok {
		DBpath = str
	}
	var manager = db.NewManager()
	er := manager.Connect(DBpath)
	if er != nil {
		log.Fatal(er)
	}
	defer manager.Disconect()
	// manager.Insert(entities.User{Name: "serega"})
	manager.Pick()

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
