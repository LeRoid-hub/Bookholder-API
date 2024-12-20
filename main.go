package main

import (
	"github.com/LeRoid-hub/Bookholder-API/config"
	"github.com/LeRoid-hub/Bookholder-API/database"
	"github.com/LeRoid-hub/Bookholder-API/server"
)

func main() {
	env := config.Load()

	db := database.SetEnv(env)

	server.Run(env, db)
}
