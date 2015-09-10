// Fake Mail Server specific implementation

package main

type FakeServer struct {
	Server MailServer
}

func (s *FakeServer) Send(message Message) int {
	// TODO Implement
	if Debug {
		InfoLog.Printf("sending email from %s to %s with subject %s via Fake.\n", message.From, message.To, message.Subject)
	}
	return 200
}

func (s *FakeServer) Ping() bool {
	return false
}

func (s *FakeServer) GetName() string {
	return "FakeServer"
}

func (s *FakeServer) SetKey(key string) {
	// TODO implement
	return
}