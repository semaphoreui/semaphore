package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"github.com/ansible-semaphore/semaphore/database"
	"github.com/ansible-semaphore/semaphore/migration"
	"github.com/ansible-semaphore/semaphore/models"
	"github.com/ansible-semaphore/semaphore/routes"
	"github.com/ansible-semaphore/semaphore/routes/sockets"
	"github.com/ansible-semaphore/semaphore/routes/tasks"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/bugsnag/bugsnag-go"
	"github.com/gin-gonic/gin"
)

func main() {
	if util.InteractiveSetup {
		os.Exit(doSetup())
	}

	fmt.Printf("Semaphore %v\n", util.Version)
	fmt.Printf("Port %v\n", util.Config.Port)
	fmt.Printf("MySQL %v@%v %v\n", util.Config.MySQL.Username, util.Config.MySQL.Hostname, util.Config.MySQL.DbName)
	fmt.Printf("Redis %v\n", util.Config.SessionDb)
	fmt.Printf("Tmp Path (projects home) %v\n", util.Config.TmpPath)

	if err := database.Connect(); err != nil {
		panic(err)
	}

	models.SetupDBLink()

	defer database.Mysql.Db.Close()
	database.RedisPing()

	if util.Migration {
		fmt.Println("\n Running DB Migrations")
		if err := migration.MigrateAll(); err != nil {
			panic(err)
		}

		return
	}

	go sockets.StartWS()
	r := gin.New()
	r.Use(gin.Recovery(), recovery, gin.Logger())

	routes.Route(r)

	go tasks.StartRunner()
	r.Run(util.Config.Port)
}

func recovery(c *gin.Context) {
	defer bugsnag.AutoNotify()
	c.Next()
}

func doSetup() int {
	fmt.Print(`
Hello, you will now be guided through a setup to:

- Set up configuration for a MySQL/MariaDB database
- Set up redis for session storage
- Set up a path for your playbooks
- Run DB Migrations
- Set up your user and password

`)

	var b []byte
	setup := util.ScanSetup()
	for true {
		var err error
		b, err = json.MarshalIndent(&setup, "", "\t")
		if err != nil {
			panic(err)
		}

		fmt.Printf("Config:\n%v\n\n", string(b))
		fmt.Print("Is this correct? (yes/no): ")

		var answer string
		fmt.Scanln(&answer)

		if !(answer == "yes" || answer == "y") {
			fmt.Println()
			setup = util.ScanSetup()

			continue
		}

		break
	}

	fmt.Print("Configuration written to /tmp/semaphore_config.json\n")
	if err := ioutil.WriteFile("/tmp/semaphore_config.json", b, 0644); err != nil {
		panic(err)
	}

	fmt.Println("\nPinging database...")
	util.Config = setup

	if err := database.Connect(); err != nil {
		fmt.Println("Connection to database unsuccessful.")
		panic(err)
	}

	fmt.Println("Pinging redis...")
	database.RedisPing()

	fmt.Println("\nRunning DB Migrations")
	if err := migration.MigrateAll(); err != nil {
		panic(err)
	}

	var user models.User
	fmt.Print("\n\nYour name: ")
	fmt.Scanln(&user.Name)

	fmt.Print("Username: ")
	fmt.Scanln(&user.Username)

	fmt.Print("Email: ")
	fmt.Scanln(&user.Email)

	fmt.Print("Password: ")
	fmt.Scanln(&user.Password)

	pwdHash, _ := bcrypt.GenerateFromPassword([]byte(user.Password), 11)
	user.Username = strings.ToLower(user.Username)
	user.Email = strings.ToLower(user.Email)

	if _, err := database.Mysql.Exec("insert into user set name=?, username=?, email=?, password=?, created=NOW()", user.Name, user.Username, user.Email, pwdHash); err != nil {
		panic(err)
	}

	fmt.Printf("\nYou are all setup %v\n", user.Name)
	fmt.Println("Re-launch this program pointing to the configuration file\n./semaphore -config /tmp/semaphore_config.json")
	fmt.Println("Your login is %v or %v.", user.Email, user.Username)

	return 0
}
