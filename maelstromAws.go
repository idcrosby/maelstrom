// AWS Mail Server specific implementation

package main

type AwsServer struct {
	Server MailServer
}

func (s *AwsServer) Send(message Message) int {
	// TODO Implement
	if Debug {
		InfoLog.Printf("sending email from %s to %s with subject %s via AWS.\n", message.From, message.To, message.Subject)
	}
	return 500
}

func (s *AwsServer) Ping() bool {
	// TODO Implement
	return false
}

func (s *AwsServer) GetName() string {
	return "AWS"
}

func (s *AwsServer) SetKey(key string) {
	// TODO implement
	return
}
