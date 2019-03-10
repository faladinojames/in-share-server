package lib

import (
	"../db"
	"fmt"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/satori/go.uuid"
	"time"
)
import "../models"

const collectionName string = "sessions"

type UserSession struct {
	UserId string `json:"userId" bson:"userId"`
	SessionToken string `json:"sessionToken" bson:"sessionToken"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt" bson:"expiresAt"`
}

type Sessions struct{
	DatabaseClient db.DatabaseClient
}
func (session *Sessions) GetUserFromToken(sessionToken string) (*models.User, error){
	user := &models.User{}
	userSession := &UserSession{}
	err := session.DatabaseClient.FindOne(collectionName, bson.M{"sessionToken": sessionToken}, userSession)

	if err != nil{
		// cannot find session
		fmt.Println("no session")
		return nil, err
	}

	fmt.Println("id",userSession.UserId )
	err = session.DatabaseClient.FindOne("users", bson.M{"id": userSession.UserId}, user)

	if err == nil{
		fmt.Println("returning user ", user)
		return user, nil
	} else {
		fmt.Println("returning error ", err)
		return nil, err
	}
}
func (session *Sessions) Create(user *models.User) (string){
	newSessionToken, _ := uuid.NewV4()

	now := time.Now()
	aYearAfter := now.AddDate(1, 0, 0)

	session.DatabaseClient.Insert(collectionName, bson.M{"userId": user.Id, "sessionToken": newSessionToken.String(), "expiresAt": aYearAfter})

	return newSessionToken.String()
}

func (session *Sessions) LogOut(sessionToken string)error{

	now := time.Now()

	return session.DatabaseClient.Update(collectionName, bson.M{"sessionToken": sessionToken}, bson.M{"expiresAt": now})

}
