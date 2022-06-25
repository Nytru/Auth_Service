package db

import (
	"autharization/entities"
	"context"
	"errors"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const Database_Name = "users"
const Collection_Name = "user"

type Manager interface {
	Connect() error
	Insert(entities.User) error
	CheckToken(guid string) (token entities.RefreshToken, er error)
	Replace(user entities.User) error
	Disconect()
}

type dbmanager struct {
	client *mongo.Client
	logger *log.Logger
	path   string
}

func (m *dbmanager) Connect() error {
	var path = m.path
	if m.client != nil {
		return errors.New("exsisted connection")
	}

	var client, err = mongo.NewClient(options.Client().ApplyURI(path))
	if err != nil {
		return err
	}
	m.client = client

	// Create connect
	err = m.client.Connect(context.TODO())
	if err != nil {
		return err
	}

	// Check the connection
	err = m.client.Ping(context.TODO(), nil)
	if err != nil {
		return err
	}
	if m.logger != nil {
		m.logger.Println("Connected to db")
	}
	return nil
}

func (m *dbmanager) Insert(user entities.User) error {
	if m.client == nil {
		return errors.New("unconected error")
	}

	var ins, err = m.client.Database(Database_Name).Collection(Collection_Name).InsertOne(context.TODO(), user)
	if err != nil {
		return err
	}
	if m.logger != nil {
		m.logger.Println("Inserted with id: ", ins.InsertedID)
	}
	return nil
}

func (m *dbmanager) CheckToken(guid string) (token entities.RefreshToken, er error) {
	if m.client == nil {
		return entities.RefreshToken{}, errors.New("unconected error")
	}

	var collection = m.client.Database(Database_Name).Collection(Collection_Name)
	var res = new(entities.User)
	var result = collection.FindOne(context.TODO(), bson.D{{"guid", guid}})
	if er = result.Err(); er != nil {
		return entities.RefreshToken{}, er
	}
	e := result.Decode(res)
	if e != nil {
		return entities.RefreshToken{}, e
	}
	return res.Refreshtoken, nil
}

func (m *dbmanager) Replace(user entities.User) error {
	if m.client == nil {
		return errors.New("unconected error")
	}

	var collect = m.client.Database(Database_Name).Collection(Collection_Name)
	var res, err = collect.ReplaceOne(context.TODO(), bson.D{{"guid", user.GUID}}, user)
	if err != nil || res.MatchedCount > 12 {
		var er = m.Insert(user)
		if er != nil {
			return er
		}
	}

	return nil
}

func (m *dbmanager) Disconect() {
	m.client.Disconnect(context.TODO())
}

func NewManager(log *log.Logger) *dbmanager {
	var path = os.Getenv("DB_FULL_PASS")
	var manager = new(dbmanager)
	if log != nil {
		manager.logger = log
	} else {
		manager.logger = nil
	}
	manager.path = path
	return manager
}
