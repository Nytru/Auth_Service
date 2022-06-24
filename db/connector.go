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

type DBmanager interface {
	Connect() (error)
	Insert(entities.User) (error)
	CheckToken(guid string) (token entities.RefreshToken, er error)
	UpdateToken(newToken entities.RefreshToken, guid string) (error)
	Replace(user entities.User) (error)
	Disconect()
}

type manager struct {
	client *mongo.Client
	logger *log.Logger
	path string
}

func (m *manager)Connect() (error) {
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

func (m *manager)Insert(user entities.User) (error) {
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

func (m *manager)CheckToken(guid string) (token entities.RefreshToken, er error) {
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

func (m *manager)UpdateToken(newToken entities.RefreshToken, guid string) (error) {
	if m.client == nil {
		return errors.New("unconected error")
	}
	var user = entities.User{Refreshtoken: newToken}
	var collection = m.client.Database(Database_Name).Collection(Collection_Name)
	var res = collection.FindOneAndUpdate(context.TODO(), bson.D{{"guid", guid}}, user)
	if res.Err() != nil {
		return res.Err()
	}
	return nil
}

func (m *manager)Replace(user entities.User) (error) {
	if m.client == nil {
		return errors.New("unconected error")
	}

	if _, er := m.CheckToken(user.GUID); er == nil {
		var collection = m.client.Database(Database_Name).Collection(Collection_Name)
		var _, er = collection.DeleteOne(context.TODO(), bson.D{{"guid", user.GUID}})
		if er != nil {
			return er
		}
	}

	var er = m.Insert(user)
	if er != nil {
		return er
	}
	return nil
}


func (m *manager)Disconect() {
	m.client.Disconnect(context.TODO())
}

func NewManager(log *log.Logger) (*manager) {
	var path = os.Getenv("DB_FULL_PASS")
	var manager = new(manager)
	if log != nil {
		manager.logger = log	
	} else {
		manager.logger = nil
	}
	manager.path = path
	return manager
}