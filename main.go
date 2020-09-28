package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

var (
	tmpl         = template.Must(template.ParseGlob("templates/*.html"))
	clientID     = "7607677"
	clientSecret = "XVC9zJmiYs6AVM83f3er"
	redirectURI  = "http://localhost:8080/me"
	scope        = []string{"account"}
	state        = "12345"
)

func main() {
	http.HandleFunc("/", index)
	http.HandleFunc("/me", me)
	log.Println("-> Server has started")
	log.Print(http.ListenAndServe(":8080", nil))
	log.Println("-> Server has stopped")
}

func index(w http.ResponseWriter, r *http.Request) {
	scopeTemp := strings.Join(scope, "+")
	url := fmt.Sprintf("https://oauth.vk.com/authorize?response_type=code&client_id=%s&redirect_uri=%s&scope=%s&state=%s", clientID, redirectURI, scopeTemp, state)
	err := tmpl.ExecuteTemplate(w, "index.html", url)
	if err != nil {
		log.Printf("error = %s", err)
	}
}

func me(w http.ResponseWriter, r *http.Request) {
	stateTemp := r.URL.Query().Get("state")
	if stateTemp[len(stateTemp)-1] == '}' {
		stateTemp = stateTemp[:len(stateTemp)-1]
	}
	if stateTemp == "" {
		respErr(w, fmt.Errorf("state query param is not provided"))
		return
	} else if stateTemp != state {
		respErr(w, fmt.Errorf("state query param do not match original one, got=%s", stateTemp))
		return
	}
	code := r.URL.Query().Get("code")
	if code == "" {
		respErr(w, fmt.Errorf("code query param is not provided"))
		return
	}
	url := fmt.Sprintf("https://oauth.vk.com/access_token?grant_type=authorization_code&code=%s&redirect_uri=%s&client_id=%s&client_secret=%s",
		code, redirectURI, clientID, clientSecret)
	req, _ := http.NewRequest("POST", url, nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		respErr(w, err)
		return
	}
	defer resp.Body.Close()
	token := struct {
		AccessToken string `json:"access_token"`
	}{}
	bytes, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(bytes, &token)
	url = fmt.Sprintf("https://api.vk.com/method/%s?v=5.124&access_token=%s", "users.get", token.AccessToken)
	req, err = http.NewRequest("GET", url, nil)
	if err != nil {
		respErr(w, err)
		return
	}
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		respErr(w, err)
		return
	}
	defer resp.Body.Close()
	bytes, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		respErr(w, err)
	}
	err = tmpl.ExecuteTemplate(w, "me.html", string(bytes))
	if err != nil {
		log.Printf("error = %s", err)
	}
}

func respErr(w http.ResponseWriter, err error) {
	_, er := io.WriteString(w, err.Error())
	if er != nil {
		log.Println(err)
	}
}
