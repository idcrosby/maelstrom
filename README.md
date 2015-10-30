Maelstrom - Mail Service
==================
A simple service/application to send email. Provides an abstraction on top of multiple mail services, with automatic failover.

Password must be provided in UI to send mail.

Design
==================

GO - I chose Go as the programming language because it is both a language used commonly and as well it is a language I am comfortable with and enjoy using. I use Go for many of my side projects (debug tooling, support applications, etc).

Javascript - I do not have a lot of experience with Javascript, but I know I could use it to build a simple, clean UI.

Docker - This was my first real experience using Docker. I thought it would be interesting to try and convenient for deployments and running locally.

Google Compute Engine - This was my first use of Google Compute Engine. I have previously used Google App Engine for deploying/hosting. I went with GCE because I had decided to use Docker and also wanted to host my own (Mongo) database.

MongoDB - First real use of Mongo. For this actual usage, a relational database would have been sufficient but I wanted to use a NoSQL datastore.


Exposed API
==================

/messages/ - POST method to send an email. Email contained in Request body.

/status - GET returns current status of the available Mail Servers

/contacts/ (Not exposed via UI) - CRUD operations for email contacts. GET can be performed on id, name, or tag via query parameters


TO DO - Expansion
==================

- Use third party routing library
- Add additional Mail Services. AWS, SendGrid, etc...
- Advanced Email options (multiple recipients,  cc, bcc, delayed send, etc.)
- Login funcionality
- OAuth integration (Facebook, etc...)


