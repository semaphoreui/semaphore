package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/ansible-semaphore/semaphore/database"
	"github.com/ansible-semaphore/semaphore/migration"
	"github.com/ansible-semaphore/semaphore/models"
	"github.com/ansible-semaphore/semaphore/routes"
	"github.com/ansible-semaphore/semaphore/routes/sockets"
	"github.com/ansible-semaphore/semaphore/routes/tasks"
	"github.com/ansible-semaphore/semaphore/upgrade"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/bugsnag/bugsnag-go"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	if util.InteractiveSetup {
		os.Exit(doSetup())
	}

	if util.Upgrade {
		if err := upgrade.Upgrade(util.Version); err != nil {
			panic(err)
		}

		os.Exit(0)
	}

	fmt.Printf("Semaphore %v\n", util.Version)
	fmt.Printf("Port %v\n", util.Config.Port)
	fmt.Printf("MySQL %v@%v %v\n", util.Config.MySQL.Username, util.Config.MySQL.Hostname, util.Config.MySQL.DbName)
	fmt.Printf("Tmp Path (projects home) %v\n", util.Config.TmpPath)

	if err := database.Connect(); err != nil {
		panic(err)
	}

	models.SetupDBLink()

	defer database.Mysql.Db.Close()

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

	go upgrade.CheckUpdate(util.Version)
	go tasks.StartRunner()
	r.Run(util.Config.Port)
}

func recovery(c *gin.Context) {
	defer bugsnag.AutoNotify()
	c.Next()
}

func doSetup() int {
	fmt.Print(`
 Hello! You will now be guided through a setup to:

 1. Set up configuration for a MySQL/MariaDB database
 2. Set up a path for your playbooks (auto-created)
 3. Run database Migrations
 4. Set up initial seamphore user & password

`)

	var b []byte
	setup := util.NewConfig()
	for true {
		setup.Scan()

		var err error
		b, err = json.MarshalIndent(&setup, " ", "\t")
		if err != nil {
			panic(err)
		}

		fmt.Printf("\n Generated configuration:\n %v\n\n", string(b))
		fmt.Print(" > Is this correct? (yes/no): ")

		var answer string
		fmt.Scanln(&answer)
		if answer == "yes" || answer == "y" {
			break
		}

		fmt.Println()
		setup = util.NewConfig()
	}

	setup.GenerateCookieSecrets()

	fmt.Printf(" Running: mkdir -p %v..\n", setup.TmpPath)
	os.MkdirAll(setup.TmpPath, 0755)

	configPath := path.Join(setup.TmpPath, "/semaphore_config.json")
	fmt.Printf(" Configuration written to %v..\n", setup.TmpPath)
	if err := ioutil.WriteFile(configPath, b, 0644); err != nil {
		panic(err)
	}

	fmt.Println(" Pinging database..")
	util.Config = setup

	if err := database.Connect(); err != nil {
		fmt.Printf("\n Cannot connect to database!\n %v\n", err.Error())
		os.Exit(1)
	}

	fmt.Println("\n Running DB Migrations..")
	if err := migration.MigrateAll(); err != nil {
		fmt.Printf("\n Database migrations failed!\n %v\n", err.Error())
		os.Exit(1)
	}

	var user models.User
	fmt.Print("\n\n > Your name: ")
	fmt.Scanln(&user.Name)

	fmt.Print(" > Username: ")
	fmt.Scanln(&user.Username)

	fmt.Print(" > Email: ")
	fmt.Scanln(&user.Email)

	fmt.Print(" > Password: ")
	fmt.Scanln(&user.Password)

	pwdHash, _ := bcrypt.GenerateFromPassword([]byte(user.Password), 11)
	user.Username = strings.ToLower(user.Username)
	user.Email = strings.ToLower(user.Email)

	if _, err := database.Mysql.Exec("insert into user set name=?, username=?, email=?, password=?, created=NOW()", user.Name, user.Username, user.Email, pwdHash); err != nil {
		fmt.Printf(" Inserting user failed. If you already have a user, you can disregard this error.\n %v\n", err.Error())
		os.Exit(1)
	}

	fmt.Printf("\n You are all setup %v!\n", user.Name)
	fmt.Printf(" Re-launch this program pointing to the configuration file\n\n./semaphore -config %v\n\n", configPath)
	fmt.Printf(" To run as daemon:\n\nnohup ./semaphore -config %v &\n\n", configPath)
	fmt.Println(" Your login is %v or %v.", user.Email, user.Username)

	return 0
}
