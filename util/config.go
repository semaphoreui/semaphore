package util

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"golang.org/x/crypto/bcrypt"

	"github.com/bugsnag/bugsnag-go"
	"github.com/gin-gonic/gin"
	"github.com/mattbaird/gochimp"
)

var mandrillAPI *gochimp.MandrillAPI
var Migration bool
var InteractiveSetup bool

type mySQLConfig struct {
	Hostname string `json:"host"`
	Username string `json:"user"`
	Password string `json:"pass"`
	DbName   string `json:"name"`
}

type mandrillConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type configType struct {
	MySQL mySQLConfig `json:"mysql"`
	// Format as is with net.Dial
	SessionDb string         `json:"session_db"`
	Mandrill  mandrillConfig `json:"mandrill"`
	// Format `:port_num` eg, :3000
	Port       string `json:"port"`
	BugsnagKey string `json:"bugsnag_key"`

	// semaphore stores projects here
	TmpPath string `json:"tmp_path"`
}

var Config configType

func init() {
	flag.BoolVar(&InteractiveSetup, "setup", false, "perform interactive setup")
	flag.BoolVar(&Migration, "migrate", false, "execute migrations")
	path := flag.String("config", "", "config path")

	var pwd string
	flag.StringVar(&pwd, "hash", "", "generate hash of given password")

	var printConfig bool
	flag.BoolVar(&printConfig, "printConfig", false, "print example configuration")

	flag.Parse()

	if printConfig {
		b, _ := json.MarshalIndent(&configType{
			MySQL: mySQLConfig{
				Hostname: "127.0.0.1:3306",
				Username: "root",
				DbName:   "semaphore",
			},
			SessionDb: "127.0.0.1:6379",
			Port:      ":3000",
			TmpPath:   "/tmp/semaphore",
		}, "", "\t")
		fmt.Println(string(b))

		os.Exit(0)
	}

	if len(pwd) > 0 {
		password, _ := bcrypt.GenerateFromPassword([]byte(pwd), 11)
		fmt.Println("Generated password: ", string(password))

		os.Exit(0)
	}

	if path != nil && len(*path) > 0 {
		// load
		file, err := os.Open(*path)
		if err != nil {
			panic(err)
		}

		if err := json.NewDecoder(file).Decode(&Config); err != nil {
			fmt.Println("Could not decode configuration!")
			panic(err)
		}
	} else {
		configFile, err := Asset("config.json")
		if err != nil {
			fmt.Println("Cannot Find configuration.")
			os.Exit(1)
		}

		if err := json.Unmarshal(configFile, &Config); err != nil {
			fmt.Println("Could not decode configuration!")
			panic(err)
		}
	}

	if len(os.Getenv("PORT")) > 0 {
		Config.Port = ":" + os.Getenv("PORT")
	}
	if len(Config.Port) == 0 {
		Config.Port = ":3000"
	}

	if len(Config.Mandrill.Password) > 0 {
		api, _ := gochimp.NewMandrill(Config.Mandrill.Password)
		mandrillAPI = api
	}

	if len(Config.TmpPath) == 0 {
		Config.TmpPath = "/tmp/semaphore"
	}

	stage := ""
	if gin.Mode() == "release" {
		stage = "production"
	} else {
		stage = "development"
	}
	bugsnag.Configure(bugsnag.Configuration{
		APIKey:              Config.BugsnagKey,
		ReleaseStage:        stage,
		NotifyReleaseStages: []string{"production"},
		AppVersion:          Version,
		ProjectPackages:     []string{"github.com/ansible-semaphore/semaphore/**"},
	})
}

// encapsulate mandrill providing some defaults

func MandrillMessage(important bool) gochimp.Message {
	return gochimp.Message{
		AutoText:  true,
		InlineCss: true,
		Important: important,
		FromName:  "Semaphore Daemon",
		FromEmail: "noreply@semaphore.local",
	}
}

func MandrillRecipient(name string, email string) gochimp.Recipient {
	return gochimp.Recipient{
		Email: email,
		Name:  name,
		Type:  "to",
	}
}

func MandrillSend(message gochimp.Message) ([]gochimp.SendResponse, error) {
	return mandrillAPI.MessageSend(message, false)
}

func ScanSetup() configType {
	var conf configType

	fmt.Print("DB Hostname (example 127.0.0.1:3306): ")
	fmt.Scanln(&conf.MySQL.Hostname)

	fmt.Print("DB User (example root): ")
	fmt.Scanln(&conf.MySQL.Username)

	fmt.Print("DB Password: ")
	fmt.Scanln(&conf.MySQL.Password)

	fmt.Print("DB Name: ")
	fmt.Scanln(&conf.MySQL.DbName)

	fmt.Print("Redis Connection (example 127.0.0.1:6379): ")
	fmt.Scanln(&conf.SessionDb)

	fmt.Print("Playbook path (will be auto-created if does not exist): ")
	fmt.Scanln(&conf.TmpPath)

	return conf
}
