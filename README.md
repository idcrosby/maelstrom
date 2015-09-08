Maelstrom - Mail Service
==================
A simple service/application to send email. Provides an abstraction on top of multiple mail services, with automatic failover.

Password must be provided in UI to send mail.

Design
==================

GO - I chose Go as the programming language because it is both a language used at Uber and as well it is a language I am comfortable with and enjoy using. I use Go for many of my side projects (debug tooling, support applications, etc).

Javascript - I do not have a lot of experience with Javascript, but I know I could use it to build a simple, clean UI.


TO DO - Expansion
==================

- Build throttling into the API. For security against misuse as well as protection of the infrastructure.
- Add additional Mail Services. AWS, SendGrid, others...
- Advanced Email options (multiple recipients,  cc, bcc, delayed send, etc.)
- Retrieve Mail Servers / Status
- Database - to save messages, contacts, users, etc.
- Login funcionality
- OAuth integration (Facebook, etc...)
