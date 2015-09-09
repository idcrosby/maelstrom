// Mandrill Mail Server specific implementation

package main

import (
	"bytes"
	"encoding/json"
	// "io/ioutil"
	"net/http"
)

type MandrillServer struct {
	Server MailServer
}

func (s *MandrillServer) Send(message Message) int {
	mail := MandrillMail{Key: s.Server.ApiKey}
	mail.Message.Text = message.Text
	mail.Message.Subject = message.Subject
	mail.Message.From = message.From
	mail.Message.To = make([]MandrillTo, len(message.To))
	for i, to := range message.To {
		mail.Message.To[i] = MandrillTo{Email: to}
	}
	jsonBuff, err := json.Marshal(mail)
	check(err)

	r, err := http.NewRequest("POST", s.Server.Url, bytes.NewBuffer(jsonBuff))
	check(err)
	r.Header.Add("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(r)
	if err != nil {
		ErrorLog.Println("Error sending mail via Mandrill ", err)
		return 500
	}

	if Debug {
		InfoLog.Println("Received: " + res.Status)
	}
	return res.StatusCode
}

func (s *MandrillServer) Ping() bool {

	var jsonStr = []byte(`{"key":"` + s.Server.ApiKey + `"}`)
	res, err := http.Post(s.Server.PingUrl, "application/json", bytes.NewBuffer(jsonStr))
	if err != nil {
		if Debug {
			ErrorLog.Println("Mandrill ping failed: ", err)
		}
		return false
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if string(body) != `"PONG!"` {
		if Debug {
			ErrorLog.Println("Mandrill Ping failed with response: " + string(body))
		}
		return false
	}
	return true
}

func (s *MandrillServer) GetName() string {
	return "Mandrill"
}

type MandrillMail struct {
	Key     string `json:"key"`
	Message struct {
		Text    string       `json:"text"`
		Subject string       `json:"subject"`
		From    string       `json:"from_email"`
		To      []MandrillTo `json:"to"`
	} `json:"message"`
}

type MandrillTo struct {
	Email string `json:"email"`
}
