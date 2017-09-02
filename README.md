# Wamp-Redis-ChatServer
A WAMP Chat Server that uses Redis as the backend written in Go


The first piece is a authorization server that works much like oAuth. 

authServer.go 
It connects to redis and generates expiring tokens that the Wamp/(websocket) clients will pass in. 

To get redis up and running. 
docker run -d --name redisDev -p 6379:6379 redis

Then build and setup authServer. 
go build 
./authServer 

Auth Server provides the following HTTP endpoints: 
/token 
Generates a token and when timestamp for when it expires. 
/Admin 
List all valid tokens. 
/IsValidToken?TOKENSTRING  
Returns True or false

AuthServer.go Todo list: 
DDOS protection: /IsValidToken doesn't have much protection on it at the moment, could be used to hammer the database. 

Token Reset: /Admin will at somepoint in the future have the ability to invalidate tokens, and clear all tokens. 

Any Questions, send me an email. 


