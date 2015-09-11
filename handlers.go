package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strings"
	"time"
)

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
		if !requestSlot() {
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

func contactsHandler(w http.ResponseWriter, req *http.Request) {

	values := req.URL.Query()
	id := values.Get("id")
	tag := values.Get("tag")
	name := values.Get("name")

	path := req.URL.Path
	pieces := strings.Split(path, "/")

	// If ID is in Path it takes precedence
	if len(pieces) > 2 && len(pieces[2]) > 0 {
		id = pieces[2]
	}

	if len(id) > 0 {
		if !bson.IsObjectIdHex(id) {
			w.WriteHeader(404)
			return
		}
	}

	switch req.Method {
	case "GET":
		if Debug {
			InfoLog.Println("Get Contact")
		}
		var contacts []Contact
		if len(id) > 0 {
			contacts = datastore.RetrieveContactsBy("id", id)
		} else if len(tag) > 0 {
			contacts = datastore.RetrieveContactsBy("tag", tag)
		} else if len(name) > 0 {
			contacts = datastore.RetrieveContactsBy("name", name)
		} else {
			// Fetch All
		}

		jsonContacts, err := json.Marshal(contacts)
		if err != nil {
			ErrorLog.Println("Error marshalling Contacts.")
			w.WriteHeader(400)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		fmt.Fprintf(w, "%s", jsonContacts)
	case "POST":
		if Debug {
			InfoLog.Println("Create Contact")
		}
		bytes, err := ioutil.ReadAll(req.Body)
		check(err)
		var contact Contact
		err = json.Unmarshal(bytes, &contact)
		if err != nil {
			http.Error(w, "Invalid JSON", 400)
			return
		}
		result := datastore.StoreContact(contact)
		jsonContact, _ := json.Marshal(result)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		fmt.Fprintf(w, "%s", jsonContact)
	case "PUT":
		if Debug {
			InfoLog.Println("Update Contact")
		}
		bytes, err := ioutil.ReadAll(req.Body)
		check(err)
		var contact Contact
		err = json.Unmarshal(bytes, &contact)
		if err != nil {
			http.Error(w, "Invalid JSON", 400)
			return
		}
		result := datastore.UpdateContact(contact)
		jsonContact, _ := json.Marshal(result)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		fmt.Fprintf(w, "%s", jsonContact)
	case "DELETE":
		if Debug {
			InfoLog.Println("Delete Contact")
		}
		if datastore.DeleteContact(id) {
			w.WriteHeader(200)
		} else {
			if Debug {
				InfoLog.Println("Could not delete contact:  " + id)
			}
			w.WriteHeader(404)
		}
	default:
		w.WriteHeader(405)
	}
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
