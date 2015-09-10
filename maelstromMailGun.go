// MailGun Mail Server specific implementation

package main

import (
	"bytes"
	"net/http"
	"net/url"
	"strconv"
)

var mailGunKey string
var mailGunPingKey string = "key-0565deb795ffa7f8e58841b43183c468"

type MailGunServer struct {
	Server MailServer
}

func (s *MailGunServer) Send(message Message) int {
	if Debug {
		InfoLog.Printf("sending email from %s to %s with subject %s via MailGun.\n", message.From, message.To, message.Subject)
	}

	data := url.Values{}
	data.Set("from", message.From)
	for _, to := range message.To {
		data.Add("to", to)
	}
	data.Set("subject", message.Subject)
	data.Set("text", message.Text)

	r, err := http.NewRequest("POST", s.Server.Url, bytes.NewBufferString(data.Encode()))
	check(err)
	r.SetBasicAuth("api", mailGunKey)
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	if Debug {
		InfoLog.Println("Sending Request " + r.URL.String())
	}
	res, err := http.DefaultClient.Do(r)
	if err != nil {
		ErrorLog.Println("Error sending mail via MailGun", err)
		return 500
	}

	if Debug {
		InfoLog.Println("Received: " + res.Status)
	}
	return res.StatusCode
}

func (s *MailGunServer) Ping() bool {

	r, err := http.NewRequest("GET", s.Server.PingUrl, nil)
	check(err)
	r.SetBasicAuth("api", mailGunPingKey)

	if Debug {
		InfoLog.Println("Sending Request " + r.URL.String())
	}
	res, err := http.DefaultClient.Do(r)
	if err != nil {
		ErrorLog.Println("Error reaching MailGun server ", err)
		return false
	}

	if Debug {
		ErrorLog.Println("Received: " + res.Status)
	}

	return res.StatusCode == 200
}

func (s *MailGunServer) GetName() string {
	return "MailGun"
}

func (s *MailGunServer) SetKey(key string) {
	mailGunKey = key
	return
}