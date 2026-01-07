package services

import (
	"encoding/json"
	"net/http"
	"fmt"
	
)

type Credentials struct {
	username string `json:"username"`
	password string	`json:"password"`
}


func auditCredentials(username string, password string) bool {

	return true
}




func P() {
	fmt.Println("hi")
}

func Login(writer http.ResponseWriter, request *http.Request) {
	var creds Credentials
	json.NewDecoder(request.Body).Decode(&creds)


	if auditCredentials(creds.username, creds.password) {

	}
		
}