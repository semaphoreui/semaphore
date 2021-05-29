package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ansible-semaphore/semaphore/db"
	"github.com/ansible-semaphore/semaphore/db/factory"
	"github.com/gorilla/context"

	log "github.com/Sirupsen/logrus"
	"github.com/ansible-semaphore/semaphore/api"
	"github.com/ansible-semaphore/semaphore/api/sockets"
	"github.com/ansible-semaphore/semaphore/api/tasks"
	"github.com/ansible-semaphore/semaphore/cli/setup"
	"github.com/ansible-semaphore/semaphore/util"
	"github.com/gorilla/handlers"
)

func cropTrailingSlashMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			r.URL.Path = strings.TrimSuffix(r.URL.Path, "/")
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	util.ConfigInit()

	if util.InteractiveSetup {
		os.Exit(doSetup())
	}

	if util.Upgrade {
		if err := util.DoUpgrade(util.Version); err != nil {
			panic(err)
		}

		os.Exit(0)
	}

	printDebugInfo()

	store := factory.CreateStore()
	if err := store.Connect(); err != nil {
		fmt.Println("\n Have you run semaphore -setup?")
		panic(err)
	}

	defer store.Close()

	if err := store.Migrate(); err != nil {
		panic(err)
	}

	// legacy
	if util.Migration {
		fmt.Println("\n DB migrations run on startup automatically")
		return
	}

	go sockets.StartWS()
	go checkUpdates()
	go tasks.StartRunner()

	route := api.Route()

	route.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			context.Set(r, "store", store)
			next.ServeHTTP(w, r)
		})
	})

	var router http.Handler = route

	router = handlers.ProxyHeaders(router)
	http.Handle("/", router)

	fmt.Println("Server is running")

	err := http.ListenAndServe(util.Config.Interface+util.Config.Port, cropTrailingSlashMiddleware(router))
	if err != nil {
		log.Panic(err)
	}
}

func printDebugInfo() {
	fmt.Printf("Semaphore %v\n", util.Version)
	fmt.Printf("Interface %v\n", util.Config.Interface)
	fmt.Printf("Port %v\n", util.Config.Port)
	fmt.Printf("MySQL %v@%v %v\n", util.Config.MySQL.Username, util.Config.MySQL.Hostname, util.Config.MySQL.DbName)
	fmt.Printf("Tmp Path (projects home) %v\n", util.Config.TmpPath)
}

func doSetup() int {
	var config *util.ConfigType
	for {
		config = &util.ConfigType{}
		config.GenerateCookieSecrets()
		setup.InteractiveSetup(config)

		if setup.VerifyConfig(config) {
			break
		}

		fmt.Println()
	}

	configPath := setup.ScanConfigPathAndSave(config)

	// Store new config globally
	util.Config = config

	fmt.Println(" Pinging db..")
	store := factory.CreateStore()

	if err := store.Connect(); err != nil {
		fmt.Printf("\n Cannot connect to database!\n %v\n", err.Error())
		os.Exit(1)
	}

	fmt.Println("\n Running DB Migrations..")
	if err := store.Migrate(); err != nil {
		fmt.Printf("\n Database migrations failed!\n %v\n", err.Error())
		os.Exit(1)
	}

	stdin := bufio.NewReader(os.Stdin)

	var user db.UserWithPwd
	user.Username = readNewline("\n\n > Username: ", stdin)
	user.Username = strings.ToLower(user.Username)
	user.Email = readNewline(" > Email: ", stdin)
	user.Email = strings.ToLower(user.Email)

	existingUser, err := store.GetUserByLoginOrEmail(user.Username, user.Email)
	util.LogWarning(err)

	if existingUser.ID > 0 {
		// user already exists
		fmt.Printf("\n Welcome back, %v! (a user with this username/email is already set up..)\n\n", existingUser.Name)
	} else {
		user.Name = readNewline(" > Your name: ", stdin)
		user.Pwd = readNewline(" > Password: ", stdin)
		user.Admin = true

		if _, err := store.CreateUser(user); err != nil {
			fmt.Printf(" Inserting user failed. If you already have a user, you can disregard this error.\n %v\n", err.Error())
			os.Exit(1)
		}

		fmt.Printf("\n You are all setup %v!\n", user.Name)
	}

	fmt.Printf(" Re-launch this program pointing to the configuration file\n\n./semaphore -config %v\n\n", configPath)
	fmt.Printf(" To run as daemon:\n\nnohup ./semaphore -config %v &\n\n", configPath)
	fmt.Printf(" You can login with %v or %v.\n", user.Email, user.Username)

	return 0
}

func readNewline(pre string, stdin *bufio.Reader) string {
	fmt.Print(pre)

	str, err := stdin.ReadString('\n')
	util.LogWarning(err)
	str = strings.Replace(strings.Replace(str, "\n", "", -1), "\r", "", -1)

	return str
}

// checkUpdates is a goroutine that periodically checks for application updates
// does not exit on errors.
func checkUpdates() {
	handleUpdateError(util.CheckUpdate(util.Version))

	t := time.NewTicker(time.Hour * 24)

	for range t.C {
		handleUpdateError(util.CheckUpdate(util.Version))
	}
}

func handleUpdateError(err error) {
	if err != nil {
		log.Warn("Could not check for update: " + err.Error())
	}
}
