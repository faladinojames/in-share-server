package main

import (
	"in-share-server/app"
	"in-share-server/config"
	"runtime"
)
import _ "github.com/joho/godotenv/autoload"

func main() {

	runtime.GOMAXPROCS(1)

	config := config.GetConfig()
	app := &app.App{}

	app.Initialize(config)

	app.Run(":" + config.Port)

}
