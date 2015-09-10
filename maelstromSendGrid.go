// SendGrid Mail Server specific implementation

package main

type SendGridServer struct {
	Server MailServer
}

func (s *SendGridServer) Send(message Message) int {
	// TODO implement
	if Debug {
		InfoLog.Printf("sending email from %s to %s with subject %s via SendGrid.\n", message.From, message.To, message.Subject)
	}
	return 500
}

func (s *SendGridServer) Ping() bool {
	// TODO implement
	return false
}

func (s *SendGridServer) GetName() string {
	return "SendGrid"
}

func (s *SendGridServer) SetKey(key string) {
	// TODO implement
	return
}