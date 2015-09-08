// MailGun Mail Server specific implementation

package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type MailGunServer struct {
	Server MailServer
}

func (s *MailGunServer) Send(message Message) int {
	if Debug {
		fmt.Printf("sending email from %s to %s with subject %s via MailGun.\n", message.From, message.To, message.Subject)
	}

	data := url.Values{}
	data.Set("from", message.From)
	data.Add("to", message.To)
	data.Set("subject", message.Subject)
	data.Set("text", message.Text)

	r, err := http.NewRequest("POST", s.Server.Url, bytes.NewBufferString(data.Encode()))
	check(err)
	r.SetBasicAuth("api", s.Server.ApiKey)
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	if Debug {
		fmt.Println("Sending Request " + r.URL.String())
	}
	res, err := http.DefaultClient.Do(r)
	if err != nil {
		fmt.Println("Error sending mail via MailGun", err)
		return 500
	}

	if Debug {
		fmt.Println("Received: " + res.Status)
	}
	return res.StatusCode
}

func (s *MailGunServer) Ping() bool {

	r, err := http.NewRequest("GET", s.Server.PingUrl, nil)
	check(err)
	r.SetBasicAuth("api", s.Server.PingKey)

	if Debug {
		fmt.Println("Sending Request " + r.URL.String())
	}
	res, err := http.DefaultClient.Do(r)
	if err != nil {
		fmt.Println("Error reaching MailGun server ", err)
		return false
	}

	if Debug {
		fmt.Println("Received: " + res.Status)
	}

	return res.StatusCode == 200
}

func (s *MailGunServer) GetName() string {
	return "MailGun"
}
