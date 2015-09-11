package main

import (
	"fmt"
	"testing"
	"time"
)

func TestPinger(t *testing.T) {
	fmt.Println("Running Test: TestPinger")

	// Setup
	config = buildTestConfig()
	buildServersMap()
	Debug = true

	initiatePing()

	time.Sleep(3 * time.Second)
	close(quit)

	fmt.Println("Test Complete.")
}

func TestBuildServersMap(t *testing.T) {
	fmt.Println("Running Test: TestBuildServersMap")

	testServer := buildTestServer()
	badServer := buildTestServer()
	badServer.Name = "Unknown"

	// Setup
	config = Config{}
	config.MailServers = make([]MailServer, 2, 2)
	config.MailServers[0] = testServer
	config.MailServers[1] = badServer

	buildServersMap()

	if len(Servers) != 1 {
		t.Errorf("buildServersMap created map of length %d should be 1.", +len(Servers))
	}

	fmt.Println("Test Complete.")
}

func TestChooseMailSender(t *testing.T) {
	fmt.Println("Running Test: TestChooseMailSenderEmpty")

	// Setup
	config = buildTestConfig()
	buildServersMap()
	initiatePing()

	// Add Mock
	mockServer := &MockServer{}
	Servers[mockServer] = true

	time.Sleep(1 * time.Second)

	s := chooseMailSender()
	if s == nil {
		t.Errorf("chooseMailSender returned nil should be ")
	}
	fmt.Println("Test Complete.")
}

func TestChooseMailSenderEmpty(t *testing.T) {
	fmt.Println("Running Test: TestChooseMailSenderEmpty")

	Servers = make(map[MailSender]bool)
	s := chooseMailSender()
	if s != nil {
		t.Errorf("chooseMailSender returned %s should be nil.", s.GetName())
	}
	fmt.Println("Test Complete.")
}

// Helper functions
func buildTestServer() MailServer {
	s := MailServer{}
	s.Name = "AWS"
	s.Url = "testUrl"
	s.PingUrl = "http://example.com"
	s.ApiKey = "testApiKey"
	s.PingKey = "testPingKey"
	return s
}

func buildTestConfig() Config {
	s1 := buildTestServer()
	s2 := buildTestServer()
	s2.Name = "MailGun"
	c := Config{}
	c.PingPeriod = 1
	c.EmailThrottle = 1
	c.MailServers = make([]MailServer, 2, 2)
	c.MailServers[0] = s1
	c.MailServers[1] = s2

	return c
}

// Mocks

type MockServer struct {
	Server MailServer
}

func (s *MockServer) Send(message Message) int {
	// TODO Implement
	if Debug {
		InfoLog.Printf("sending email from %s to %s with subject %s via Mock.\n", message.From, message.To, message.Subject)
	}
	return 200
}

func (s *MockServer) Ping() bool {
	return true
}

func (s *MockServer) GetName() string {
	return "MockServer"
}

func (s *MockServer) SetKey(key string) {
	return
}
