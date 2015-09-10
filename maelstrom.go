package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"google.golang.org/cloud/compute/metadata"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"regexp"
	"time"
)

var config Config
var Debug bool
var Password string
var quit chan struct{}
var Servers map[MailSender]bool
var emailRegex string = "\\b[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,4}\\b"
var indexHtml = "resources/html/index.html"
var throttle chan int
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
		fmt.Println("Running on GCE")

		for s, _ := range Servers {
			apiKey, _ := metadata.InstanceAttributeValue(s.GetName())
			s.SetKey(apiKey)
		}
	} else {
		fmt.Println("Not running on GCE.")
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

	// Initiate throttle
	throttle = make(chan int, config.EmailThrottle)

	buildServersMap()

	initiatePing()

	http.HandleFunc("/", errorHandler(rootHandler))
	http.HandleFunc("/messages/", errorHandler(messageHandler))
	http.HandleFunc("/status", errorHandler(statusHandler))

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

// ***********
// Handlers
// ***********

// Handler for root resource, returns web page
func rootHandler(w http.ResponseWriter, req *http.Request) {
	var rootTemplate, err = template.ParseFiles(indexHtml)
	check(err)
	rootTemplate.Execute(w, nil)
}

// Handler for Messages resource. Verfies received data and sends email
func messageHandler(w http.ResponseWriter, req *http.Request) {

	// Parse Data
	if req.Method == "POST" {

		values := req.URL.Query()
		password := values.Get("password")
		if password != Password {
			w.WriteHeader(403)
			return
		}

		// Build Message
		bytes, err := ioutil.ReadAll(req.Body)
		check(err)
		var email Message
		err = json.Unmarshal(bytes, &email)
		if err != nil {
			http.Error(w, "Invalid JSON", 400)
			return
		}

		// Validate Fields
		for _, to := range email.To {
			matchTo, _ := regexp.MatchString(emailRegex, to)
			if !matchTo {
				if Debug {
					ErrorLog.Println("To address not valid email: " + to)
				}
				http.Error(w, "Invalid 'To' Email Address.", 400)
				return
			}
		}
		matchFrom, _ := regexp.MatchString(emailRegex, email.From)
		if !matchFrom {
			if Debug {
				ErrorLog.Println("From address not valid email.")
			}
			http.Error(w, "Invalid 'From' Email Address.", 400)
			return
		}

		sender := chooseMailSender()
		if sender == nil {
			http.Error(w, "No Mail Server Available.", 500)
			return
		}
		if ! requestSlot() {
			http.Error(w, "Over throttle limit.", 403)
			return
		}
		status := sender.Send(email)
		w.WriteHeader(status)
		return
	}

	// Other methods not supported
	w.WriteHeader(405)
}

// Handler to return status of MailServers
func statusHandler(w http.ResponseWriter, req *http.Request) {

	result := make(map[string]bool)
	for s, b := range Servers {
		result[s.GetName()] = b
	}
	statusJson, err := json.Marshal(result)
	check(err)

	fmt.Fprintf(w, string(statusJson))
}

// Error Handler Wrapper
func errorHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if Debug {
			InfoLog.Println(time.Now().String())
			reqDump, _ := httputil.DumpRequest(req, true)
			InfoLog.Printf("Request: %s\n\n", reqDump)
		}
		defer func() {
			if e, ok := recover().(error); ok {
				w.WriteHeader(500)
				ErrorLog.Print("error: ")
				ErrorLog.Println(e)
			}
		}()
		fn(w, req)
	}
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
		} else if conf.Name == "Mock" {
			server = &FakeServer{conf}
		} else {
			if Debug {
				ErrorLog.Println("Unknown MailServer: " + conf.Name)
			}
			continue
		}
		Servers[server] = false
		checkServers()
	}
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

// Generic Message object
type Message struct {
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
