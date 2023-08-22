package common

import (
	"sync"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
)

func (u User) Getbson() bson.M {
	return bson.M{"ID": u.ID, "Name": u.Name, "Email": u.Email, "Address": u.Address, "Phone": u.Phone}
}

type User struct {
	ID      string `json:"ID"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
}

type Server struct {
	Name         string
	Address      string
	Router       *mux.Router
	readtimeout  int
	writetimeout int
	Lock         *sync.Mutex
}
