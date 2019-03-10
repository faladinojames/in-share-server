package models

import (
	"github.com/mongodb/mongo-go-driver/bson/primitive"
	"time"
)

type User struct {
	ObjectId string  //	ObjectId bson.ObjectId `bson:"_id,omitempty"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
	Username string  `json:"username"`
	Password string  `json:"-"`
	Name string `json:"name"`
	Phone string `json:"phone"`
	Email string `json:"email"`
	Admin bool `json:"admin"`
	Id string `json:"-"`
	Groups []string `json:"groups"`
	Public bool  `json:"-"`
	SessionToken string `json:"sessionToken"`
}

type File struct {
	ObjectId string
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt"`
	FileId string `json:"fileId" bson:"fileId"`
	FileDid primitive.ObjectID `json:"fileDId" bson:"fileDId"`
	UserId string `json:"userId" bson:"userId"`
	OwnerId string `json:"ownerId" bson:"ownerId"`
	FileSize int `json:"fileSize" bson:"fileSize"`
	FileName string `json:"fileName" bson:"fileName"`
	Users []string
	Groups []string
}
