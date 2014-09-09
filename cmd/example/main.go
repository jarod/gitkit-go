package main

import (
	"fmt"
	"github.com/jarod/gitkit-go/gitkit"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
)

var (
	client    *gitkit.Client
	indexTpl  *template.Template
	widgetTpl []byte
)

func main() {
	var err error

	if indexTpl, err = template.ParseFiles("templates/index.html"); err != nil {
		log.Fatalln(err)
	}
	if widgetTpl, err = ioutil.ReadFile("templates/gitkit-widget.html"); err != nil {
		log.Fatalln(err)
	}

	client, err = gitkit.NewClientFromJSON("gitkit-server-config.json")
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/", login)
	http.HandleFunc("/login", login)
	http.HandleFunc("/gitkit", widget)
	log.Fatal(http.ListenAndServe(":4567", nil))
}

func widget(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write(widgetTpl)
	if err != nil {
		w.Write([]byte(err.Error()))
	}
}

func login(w http.ResponseWriter, r *http.Request) {
	user, err := client.ValidateTokenInRequest(r)

	var text string
	if err != nil {
		text = "You are not logged in"
		log.Println(err)
	} else {
		text = fmt.Sprintf("Welcome %s! Your user info is: %v", user.Email, *user)
	}
	err = indexTpl.Execute(w, text)
	if err != nil {
		w.Write([]byte(err.Error()))
	}
}
