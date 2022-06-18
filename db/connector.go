package db

import (
	"autharization/entities"
	"context"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const connectParams = "mongodb://yan:good@localhost:27015"

func Connect() {

	var client, err = mongo.NewClient(options.Client().ApplyURI(connectParams))
	if err != nil {
		log.Fatal(err)
	}

	// Create connect
	err = client.Connect(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB!")

	var user =  entities.User{Name: "Sanya", GUID: "1886"}

	collection := client.Database("Authentication").Collection("users")
	var cur, er = collection.Find(context.TODO(), "")
	if er != nil {
		log.Fatal(er)
	}
	var users = make([]entities.User, 0)
	err = cur.All(context.TODO(), users)
	if err != nil {
		log.Fatal(err)
	}
	insertResult, err := collection.InsertOne(context.TODO(), user)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted a single document: ", insertResult.InsertedID)
	client.Disconnect(context.TODO())
}
