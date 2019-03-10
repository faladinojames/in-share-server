package app

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"log"
	"net/http"
)
import "../config"
import "./db"
import "./handlers"
import "./router"
type App struct {
	DB     db.DatabaseClient
	router router.Router
}

type ServerInfo struct {
	Version string`json:"version"`
	Name string`json:"name"`
	Port string`json:"port"`
}

func (a *App) Initialize(config config.Config){
	mongoClient := db.DatabaseClient{}
	mongoClient.Initialize(config)


	router := router.Router{Router: mux.NewRouter()}

	a.DB = mongoClient
	a.router = router

	router.Get("/info",  func (w http.ResponseWriter, req *http.Request){
		w.Header().Set("Content-Type", "application/json")
		info := &ServerInfo{ Version: config.Version, Name: "in-share-server", Port: config.Port}
		json.NewEncoder(w).Encode(info)
	})

	auth := handlers.Auth{Router: router, DatabaseClient:  mongoClient}
	auth.Init()

	files := handlers.Files{Router: router, DatabaseClient:  mongoClient}
	files.Init()

}



func (a *App) Run(host string){
	handler := cors.AllowAll().Handler( a.router.Router)
	log.Fatal(http.ListenAndServe(host, handler))
}

