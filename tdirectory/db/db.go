package db

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"tdirectory/common"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var db *mongo.Database
var dbLock *sync.Mutex
var userColl *mongo.Collection

const (
	DB_ADDR         = "mongodb://172.17.0.2:27017"
	DB_NAME         = "T_DIRECTORY"
	USER_COLLECTION = "USERS"
)

func init() {

	client, e := mongo.Connect(context.TODO(), options.Client().ApplyURI(DB_ADDR))

	if e != nil {
		panic("NOT_ABLE_TO_CONNECT_TO_MONGO : " + e.Error())
	}

	e = client.Ping(context.TODO(), nil)
	if e != nil {
		panic("MONGO_CONNECTION : " + e.Error())
	}
	dbLock = &sync.Mutex{}
	db = client.Database(DB_NAME)
	userColl = db.Collection(USER_COLLECTION)

}

func CreateUser(usr common.User) error {
	_, e := userColl.InsertOne(context.TODO(), usr.Getbson())
	return e
}

func UpdateUser(usr common.User) error {
	u := new(common.User)
	filter := bson.M{"ID": usr.ID}
	err := userColl.FindOne(context.TODO(), filter).Decode(u)
	if err != nil && err == mongo.ErrNoDocuments {
		return errors.New("USER_DOES_NOT_EXISTS")
	} else if err != nil {
		return err
	}
	_, err = userColl.UpdateOne(context.TODO(), filter, bson.M{"$set": usr.Getbson()})
	fmt.Println("updated user", u.Email, err)
	return err
}

func GetUser(id string) (*common.User, error) {
	usr := new(common.User)
	filter := bson.M{"ID": id}
	err := userColl.FindOne(context.TODO(), filter).Decode(usr)
	if err != nil {
		return nil, err
	}
	return usr, nil
}

func DeleteUser(id string) error {
	filter := bson.M{"ID": id}
	_, err := userColl.DeleteOne(context.TODO(), filter)
	if err != nil {
		return err
	}
	fmt.Println("deleted user", id, err)
	return nil
}
