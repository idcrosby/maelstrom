package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"os"
	"time"
)

var config Config
var Debug bool
var Password string
var quit chan struct{}
var Servers map[MailSender]bool

var indexHtml = "resources/html/index.html"

func main() {

	// Parse command line args
	flag.BoolVar(&Debug, "debug", true, "Turn on debug logging.")
	flag.StringVar(&Password, "password", "", "Password needed by users to send emails.")
	flag.Parse()

	// Read Config file
	file, err := os.Open("conf.json")
	check(err)
	decoder := json.NewDecoder(file)
	config = Config{}
	err = decoder.Decode(&config)
	check(err)

	buildServersMap()

	initiatePing()

	http.HandleFunc("/", errorHandler(rootHandler))
	http.HandleFunc("/messages/", errorHandler(messageHandler))

	// To Serve CSS and JS files
	http.Handle("/resources/", http.StripPrefix("/resources/", http.FileServer(http.Dir("resources"))))

	// Read Port from Env
	port := os.Getenv("PORT")
	if port == "" {
		port = "8123"
	}

	if Debug {
		fmt.Println("Server running on Port:", port)
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

		sender := chooseMailSender()
		if sender == nil {
			http.Error(w, "No Mail Server Available.", 500)
			return
		}
		status := sender.Send(email)
		w.WriteHeader(status)
		return
	}

	// Other methods not supported
	w.WriteHeader(405)
}

// Error Handler Wrapper
func errorHandler(fn http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if Debug {
			fmt.Println(time.Now().String())
			reqDump, _ := httputil.DumpRequest(req, true)
			fmt.Printf("Request: %s\n\n", reqDump)
		}
		defer func() {
			if e, ok := recover().(error); ok {
				w.WriteHeader(500)
				fmt.Print("error: ")
				fmt.Println(e)
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
}

// Select a MailServer which is currently 'up'
// TODO allow specifying of Server
func chooseMailSender() MailSender {

	// Weighted Ranking? Random?
	for serv, status := range Servers {
		if status {
			if Debug {
				fmt.Printf("Selected Mail Server: %s\n", serv.GetName())
			}
			return serv
		}
	}

	// TODO try anyway?
	if Debug {
		fmt.Println("No MailServers are currently available")
	}
	return nil
}

// Build list of Servers as defined in the Configuration
func buildServersMap() {
	Servers = make(map[MailSender]bool)
	for _, conf := range config.MailServers {
		if Debug {
			fmt.Println("Adding Server: " + conf.Name)
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
				fmt.Println("Unknown MailServer: " + conf.Name)
			}
			continue
		}
		Servers[server] = false
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
			fmt.Printf("Mail Server: %s status: %t\n", server.GetName(), status)
		}
		Servers[server] = status
	}
}

// Standard error check function
func check(err error) {
	if err != nil {
		fmt.Println("Panicking: ", err)
		panic(err)
	}
}

// Generic Message object
type Message struct {
	To      string `json:"to"`
	Subject string `json:"subject"`
	From    string `json:"from"`
	Text    string `json:"text"`
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
}
