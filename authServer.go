package main

import (
	"crypto/rand"
	"encoding/base64"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis"
)

var addr = flag.String("addr", "localhost:8080", "http service address")
var numberOfHoursTokenIsValid = 5
var client *redis.Client

type AuthTokenString struct {
	Token          string
	AuthExpiration string
}

var AuthTokenStrings = []*AuthTokenString{}

type ValidToken struct {
	Key, Value string
}

type AuthToken struct {
	Token          string
	AuthExpiration time.Time
}

func init() {
	fmt.Println("Server init")
	client = redis.NewClient(&redis.Options{
		Addr:         ":6379",
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     10,
		PoolTimeout:  30 * time.Second,
	})

	client.FlushDB()
}

func main() {
	http.HandleFunc("/token", token)
	http.HandleFunc("/admin", adminPage)
	http.HandleFunc("/validToken", isValidToken)
	http.HandleFunc("/validtoken", isValidToken)
	http.HandleFunc("/isvalidtoken", isValidToken)
	http.HandleFunc("/isValidToken", isValidToken)
	http.HandleFunc("/isValidtoken", isValidToken)

	pong, err := client.Ping().Result()
	if err != nil {
		fmt.Println("Error: ", err)
		log.Fatal("Redis Not Up")
	}
	fmt.Println("Redis working", pong)

	iter := client.Scan(0, "token:*", 0).Iterator()
	for iter.Next() {

		fmt.Println("InterGet:\n ", client.Get(iter.Val()))

	}
	if err := iter.Err(); err != nil {
		panic(err)
	}

	log.Fatal(http.ListenAndServe(*addr, nil))
}

func isValidToken(w http.ResponseWriter, r *http.Request) {
	//Needs to do some validation here to determine if the request.
	//This potentially could lead to a DDOS against the service.

	if validToken(r.URL.RawQuery) {
		fmt.Println("Valid Token:")
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, "TRUE")

	} else {
		fmt.Println("More than one item, or invalid token:")
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, "FALSE")
	}
}

func validToken(token string) bool {
	iter := client.Scan(0, "token:*", 0).Iterator()
	for iter.Next() {
		val, err := client.Get(iter.Val()).Result()
		if val == token {
			return true
		}
		if err != nil {
			panic(err)
		}
	}
	return false
}

func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}
func GenerateRandomString(s int) (string, error) {
	b, err := GenerateRandomBytes(s)
	fmt.Println("GenerateRandomString")
	fmt.Println(base64.URLEncoding.EncodeToString(b))
	return base64.URLEncoding.EncodeToString(b), err
}

func token(w http.ResponseWriter, r *http.Request) {
	//Get a random token.
	tokenString, err := GenerateRandomString(32)
	if err == nil {
	} else {
		panic(err)
	}

	expireTime := time.Now().Add(time.Hour * time.Duration(numberOfHoursTokenIsValid))
	newToken := AuthToken{tokenString, expireTime}

	//Write the new token to the database
	//key = expiretime, token, and then the key expires after X hours.

	dbError := client.Set("token:"+expireTime.String(), newToken.Token, time.Hour*time.Duration(numberOfHoursTokenIsValid)).Err()
	fmt.Println(expireTime.String())

	if dbError != nil {
		panic(dbError)
	}
	val, err := client.Get("token:" + expireTime.String()).Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("DB OK ", val)

	//Write out the token to the client
	tmpl, err := template.New("test").Parse("{{.Token}} {{.AuthExpiration}}")
	if err != nil {
		panic(err)
	}
	err = tmpl.Execute(w, newToken)
	if err != nil {
		panic(err)
	}
}

func adminPage(w http.ResponseWriter, r *http.Request) {
	//TODO add a clear token button
	var validTokens []ValidToken

	iter := client.Scan(0, "token:*", 0).Iterator()
	for iter.Next() {
		val, err := client.Get(iter.Val()).Result()
		if err != nil {
			panic(err)
		}
		aValidToken := ValidToken{
			Key:   iter.Val(),
			Value: val,
		}
		validTokens = append(validTokens, aValidToken)
	}
	if err := iter.Err(); err != nil {
		panic(err)
	}

	t := template.New("tempTemplate")
	t = template.Must(t.Parse(`
	

	<!DOCTYPE html> 
	<head> 
	<meta charset="utf-8"> 
	<title>Token Admin </title> 
	</head> 
	<body> 
	<h1> Hi Admin </h1>
	<p> 
	{{range . }}
		<p>{{.Key}} {{.Value}}</p> 
		{{end}}
	<p> 
	</body> 
	</html> 
	`))
	t.Execute(w, validTokens)
}
