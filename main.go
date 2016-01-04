package main

import (
	"fmt"
	"github.com/bugsnag/bugsnag-go"
	"github.com/castawaylabs/semaphore/database"
	"github.com/castawaylabs/semaphore/migration"
	"github.com/castawaylabs/semaphore/routes"
	"github.com/castawaylabs/semaphore/util"
	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Printf("Semaphore %v\n", util.Version)
	fmt.Printf("Port %v\n", util.Config.Port)
	fmt.Printf("MySQL %v@%v\n", util.Config.MySQL.Username, util.Config.MySQL.Hostname)
	fmt.Printf("Redis %v\n", util.Config.SessionDb)

	defer database.Mysql.Db.Close()
	database.RedisPing()

	if util.Migration {
		fmt.Println("\n Running DB Migrations")
		if err := migration.MigrateAll(); err != nil {
			panic(err)
		}

		return
	}

	r := gin.New()
	r.Use(gin.Recovery(), recovery, gin.Logger())

	routes.Route(r)

	r.Run(util.Config.Port)
}

func recovery(c *gin.Context) {
	defer bugsnag.AutoNotify()
	c.Next()
}
