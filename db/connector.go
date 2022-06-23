package db

import (
	"autharization/entities"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DBmanager interface {
	Connect(string)
	Insert(entities.User)
	Pick() (entities.User)
	Disconect()
}

type Manager struct {
	Client *mongo.Client
	collection mongo.Collection
}

func (m *Manager)Connect(path string) (error) {
	if m.Client != nil {
		return errors.New("exsisted connection")
	}

	var client, err = mongo.NewClient(options.Client().ApplyURI(path))
	if err != nil {
		return err
	}
	m.Client = client

	// Create connect
	err = m.Client.Connect(context.TODO())
	if err != nil {
		return err
	}

	// Check the connection
	err = m.Client.Ping(context.TODO(), nil)
	if err != nil {
		return err
	}
	log.Println("Connected to DB")
	return nil
}

func (m *Manager)Insert(user entities.User) (error) {
	var ins, err = m.Client.Database("users").Collection("user").InsertOne(context.TODO(), user)
	if err != nil {
		return err
	}
 	log.Println("Inserted with: ", ins.InsertedID)
	return nil
}

func (m *Manager)All() (*[]entities.User, error) {
	var ans = new([]entities.User)
	var collection = m.Client.Database("users").Collection("user")
	var cur, err = collection.Find(context.TODO(), bson.D{})
	if err != nil {
		return nil, err
	}
	err = cur.All(context.TODO(), ans)
	if err != nil {
		return nil, err
	}
	return ans, nil
}

func (m *Manager)Pick() (*entities.User, error) {
	var collection = m.Client.Database("users").Collection("user")
	var res bson.A
	var err = collection.FindOne(context.TODO(), bson.D{{"name", "serega"}}).Decode(&res)
	if err == mongo.ErrNoDocuments {
		fmt.Printf("No document was found with the title %s\n", "serega")
		return nil, nil
	}
	if err != nil {
		panic(err)
	}
	jsonData, err := json.MarshalIndent(res, "", "    ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", jsonData)
	 
	return nil, nil
}

func (m *Manager)Disconect() {
	m.Client.Disconnect(context.TODO())
}

func NewManager() (*Manager) {
	return &Manager{}
}