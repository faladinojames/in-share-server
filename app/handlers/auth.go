package handlers

import (
	"../crypto"
	"../db"
	"../lib"
	"../utils"
	"../models"
	"../router"
	"context"
	"encoding/json"
	"fmt"
	"github.com/mongodb/mongo-go-driver/bson"
	"github.com/rs/xid"
	"log"
	"net/http"
)

type Auth struct {
	DatabaseClient db.DatabaseClient
	Router         router.Router
	session        lib.Sessions
}

type LoginFields struct {
	Username string
	Password string
}

func (auth *Auth) Init() {

	auth.session = lib.Sessions{DatabaseClient: auth.DatabaseClient}
	auth.Router.Router.Use(auth.checkAuth)
	auth.Router.Post("/login", auth.login)
	auth.Router.Post("/register", auth.register)



}

// Middleware function, which will be called for each request
func (auth *Auth) checkAuth(next http.Handler) http.Handler {


	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		checkToken := func(token string) {
			fmt.Println("checking auth", token)

			if user, _ := auth.session.GetUserFromToken(token); user != nil {
				log.Printf("Authenticated user %s\n", user)
				log.Printf("Authenticated user %s\n", user.Username)

				ctx := context.WithValue(r.Context(), "user", user)
				// Pass down the request to the next middleware (or final handler
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				// Write an error and stop the handler chain
				http.Error(w, "Forbidden", http.StatusForbidden)
			}
		}

		token := r.Header.Get("X-Session-Token")

		var noAuthPaths = []string{"/login", "/register", "/files", "/info"}


		if  utils.IncludePrefix(noAuthPaths, r.RequestURI) {

			if (token != ""){
				checkToken(token)
				return
			}
			user := &models.User{}

			user.Public = true

			ctx := context.WithValue(r.Context(), "user", user)

			next.ServeHTTP(w,  r.WithContext(ctx))
			return
		}


		checkToken(token)



	})
}

func (auth *Auth) login (w http.ResponseWriter, req *http.Request) {

	decoder := json.NewDecoder(req.Body)
	crypto := crypto.Crypto{}
	var fields LoginFields
	err := decoder.Decode(&fields)
	if err != nil {
		panic(err) //TODO error res
	}

	filter := bson.M{"username": fields.Username}

	var user = &models.User{}

	err = auth.DatabaseClient.FindOne("users", filter, user)

	if err!= nil{
		log.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid Username/Password")
		return
	}

	if crypto.CheckPassword(fields.Password, user.Password){
		fmt.Println(user)



		token := auth.session.Create(user)

		user.SessionToken = token

		w.Header().Set("Content-Type", "application/json")

		w.WriteHeader(http.StatusOK)

		json.NewEncoder(w).Encode(user)

		//fmt.Fprintf(w, token)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprintf(w, "Invalid Username/Password")
	}

}

func (auth *Auth) register (w http.ResponseWriter, req *http.Request){

	decoder := json.NewDecoder(req.Body)

	var fields models.User
	err := decoder.Decode(&fields)
	if err != nil {
		panic(err)
	}

	//TODO username/email already exists?
	crypto := crypto.Crypto{}
	fields.Password = crypto.EncryptPassword(fields.Password)

	guid := xid.New()


	id := auth.DatabaseClient.Insert("users", bson.M{"id": guid.String(), "username": fields.Username, "password": fields.Password, "email": fields.Email, "phone": fields.Phone, "name": fields.Name})

	fmt.Fprintf(w, id)
}
