package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"google.golang.org/cloud/compute/metadata"
	"gopkg.in/mgo.v2/bson"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

var config Config
var Debug bool
var gce bool
var Password string
var quit chan struct{}
var Servers map[MailSender]bool
var emailRegex string = "\\b[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,4}\\b"
var indexHtml = "resources/html/index.html"
var throttle chan int
var datastore Datastore
var InfoLog *log.Logger
var ErrorLog *log.Logger

func init() {

	// Read Config file
	config = Config{}
	file, err := os.Open("conf.json")
	if err != nil {
		fmt.Println("No config file found. Using Defaults")
		config.PingPeriod = 60
		config.EmailThrottle = 5
	} else {
		decoder := json.NewDecoder(file)
		err = decoder.Decode(&config)
		check(err)
	}

	// Check if running on GCE
	if metadata.OnGCE() {
		fmt.Println("Running on GCE. Pulling attributes.")
		gce = true
	} else {
		fmt.Println("Not running on GCE.")
		gce = false
	}
	// init loggers
	var writer io.Writer
	if len(config.LogFileName) > 0 {
		// Create Directory
		// os.MkdirAll(, 0777)
		logFile, err := os.OpenFile(config.LogFileName, os.O_CREATE | os.O_WRONLY | os.O_APPEND, 0666)
		if err != nil {
			fmt.Println("Error opening log file: ", err)
			writer = os.Stdout
		} else {
			defer logFile.Close()
			writer = logFile
		}
	} else {
		writer = os.Stdout
	}

	InfoLog = log.New(writer, "INFO: ", log.LstdFlags)
	ErrorLog = log.New(writer, "ERROR: ", log.LstdFlags)
}

func main() {

	// Parse command line args
	flag.BoolVar(&Debug, "debug", true, "Turn on debug logging.")
	flag.StringVar(&Password, "password", "", "Password needed by users to send emails.")
	flag.Parse()

	// Check GCE for Password
	if gce {
		pw, _ := metadata.InstanceAttributeValue("emailPW")
		if len(pw) > 0 {
			Password = pw
		}
	}

	// Initiate throttle
	throttle = make(chan int, config.EmailThrottle)

	buildServersMap()

	initiatePing()

	// Create Database
	datastore = &MongoDatastore{}
	if datastore.Ping() {
		InfoLog.Println("MongoDB running.")
	} else {
		ErrorLog.Println("MongoDB connection unsuccessful.")
	}

	http.HandleFunc("/", errorHandler(rootHandler))
	http.HandleFunc("/messages/", errorHandler(messageHandler))
	http.HandleFunc("/status", errorHandler(statusHandler))
	http.HandleFunc("/contacts/", errorHandler(contactsHandler))

	// To Serve CSS and JS files
	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("resources"))))

	// Read Port from Env
	port := os.Getenv("PORT")
	if port == "" {
		port = "8123"
	}

	if Debug {
		InfoLog.Println("Server running on Port:", port)
	}

	http.ListenAndServe(":"+port, nil)
}

// Generic Interface for a Mail Server
type MailSender interface {
	Send(Message) int
	Ping() bool
	GetName() string
	SetKey(string)
}

// Select a MailServer which is currently 'up'
// TODO allow specifying of Server?
func chooseMailSender() MailSender {

	// Weighted Ranking? Random?
	for serv, status := range Servers {
		if status {
			if Debug {
				InfoLog.Printf("Selected Mail Server: %s\n", serv.GetName())
			}
			return serv
		}
	}

	if Debug {
		ErrorLog.Println("No MailServers are currently available")
	}
	return nil
}

// Build list of Servers as defined in the Configuration
func buildServersMap() {
	Servers = make(map[MailSender]bool)
	for _, conf := range config.MailServers {
		if Debug {
			InfoLog.Println("Adding Server: " + conf.Name)
		}
		var server MailSender
		if conf.Name == "MailGun" {
			server = &MailGunServer{conf}
		} else if conf.Name == "SendGrid" {
			server = &SendGridServer{conf}
		} else if conf.Name == "Mandrill" {
			server = &MandrillServer{conf}
		} else if conf.Name == "AWS" {
			server = &AwsServer{conf}
		} else {
			if Debug {
				ErrorLog.Println("Unknown MailServer: " + conf.Name)
			}
			continue
		}
		var apiKey string
		if gce {
			apiKey, _ = metadata.InstanceAttributeValue(conf.Name)
			fmt.Println("Setting api key from GCE " + apiKey)
		} else {
			apiKey = conf.ApiKey
			fmt.Println("Setting api key from Config " + apiKey)
		}
		server.SetKey(apiKey)
		Servers[server] = false
	}
	checkServers()
}

// Starts a periodic Ping for the Mail Servers
func initiatePing() {
	pinger := time.NewTicker(time.Duration(config.PingPeriod) * time.Second)
	quit = make(chan struct{})
	go func() {
		for {
			select {
			case <-pinger.C:
				checkServers()
			case <-quit:
				pinger.Stop()
				return
			}
		}
	}()
}

// Check and update the status for all Mail Servers
func checkServers() {
	for server, _ := range Servers {
		status := server.Ping()
		if Debug {
			InfoLog.Printf("Mail Server: %s status: %t\n", server.GetName(), status)
		}
		Servers[server] = status
	}
}

// Enforce Throttling by requiring a 'slot' to send
func requestSlot() bool {
	
	if len(throttle) >= cap(throttle) {
		if Debug {
			InfoLog.Println("Request blocked. No open slots.")
		}
		return false
	}
	throttle <- 1
	slotTimer := time.NewTimer(time.Second * 1)
	go func() {
		<- slotTimer.C
		_ = <-throttle
	}()
	return true
}

// Standard error check function
func check(err error) {
	if err != nil {
		ErrorLog.Println("Panicking: ", err)
		panic(err)
	}
}

type Datastore interface {
	Status() bool
	StoreContact(Contact) Contact
	DeleteContact(string) bool
	UpdateContact(Contact) Contact
	RetrieveContactsBy(string, string) []Contact
	Ping() bool
}

type Contact struct {
	Id bson.ObjectId    `json:"id" bson:"_id,omitempty"`
	Email string        `json:"email"`
	Name string         `json:"name"`
	Tags []string       `json:"tags"`
}

// Generic Message object
type Message struct {
	Id      int      `json:"id"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	From    string   `json:"from"`
	Text    string   `json:"text"`
}

// Generic Mail Server Configuration
type MailServer struct {
	Name    string
	Url     string
	PingUrl string
	ApiKey  string
	PingKey string
}

// Structure for Applications Configuration
type Config struct {
	MailServers []MailServer
	PingPeriod  int
	EmailThrottle int
	LogFileName string
}

// Structure of Server Status
type Status struct {
	ServerStatus []struct {
		Name string
		Status bool
	}
}
