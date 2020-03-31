package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"text/template"
)

// Response ...
type Response struct {
	ID              string   `json:"id"`
	ActionID        string   `json:"action_id"`
	Amount          int      `json:"amount"`
	Currency        string   `json:"currency"`
	Approved        bool     `json:"approved"`
	Status          string   `json:"status"`
	AuthCode        string   `json:"auth_code"`
	ResponseCode    string   `json:"response_code"`
	ResponseSummary string   `json:"response_summary"`
	ProcessedOn     string   `json:"processed_on"`
	Risk            Risk     `json:"risk"`
	Source          Source   `json:"source"`
	Customer        Customer `json:"customer"`
	Links           Links    `json:"_links"`
}

// Risk ...
type Risk struct {
	Flagged bool `json:"flagged"`
}

// Source ...
type Source struct {
	ID            string `json:"id"`
	Type          string `json:"type"`
	ExpiryMonth   int    `json:"expiry_month"`
	ExpiryYear    int    `json:"expiry_year"`
	Scheme        string `json:"scheme"`
	Last4         string `json:"last4"`
	Fingerprint   string `json:"fingerprint"`
	Bin           string `json:"bin"`
	CardType      string `json:"card_type"`
	CardCategory  string `json:"card_category"`
	Issuer        string `json:"issuer"`
	IssuerCountry string `json:"issuer_country"`
	ProductID     string `json:"product_id"`
	ProductType   string `json:"product_type"`
}

// Customer ...
type Customer struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  int    `json:"name"`
}

// Links ...
type Links struct {
	Current     URL `json:"self"`
	RedirectURL URL `json:"redirect"`
}

// URL ...
type URL struct {
	URLString string `json:"href"`
}

func main() {

	mainPage := template.Must(template.ParseFiles("template/main.html"))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/pages/", http.StripPrefix("/pages/", http.FileServer(http.Dir("pages"))))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		r.ParseForm()
		if r.Method != http.MethodPost {
			mainPage.Execute(w, nil)
			return
		}
		requestPayment(r.FormValue("cko-card-token"))
		mainPage.Execute(w, struct{ Success bool }{true})
	})
	http.HandleFunc("/favicon.ico", doNothing)
	fmt.Println("Listening")
	successPage := template.Must(template.ParseFiles("pages/success.html"))
	http.HandleFunc("/pages/success.html", func(w http.ResponseWriter, r *http.Request) {

		keys, ok := r.URL.Query()["cko-session-id"]

		if !ok || len(keys[0]) < 1 {
			log.Println("Url Param 'key' is missing")
			return
		}
		key := keys[0]
		log.Println("cko-session-id: " + string(key))
		successPage.Execute(w, struct{ Success bool }{true})
	})
	http.ListenAndServe(":8080", nil)
}

func doNothing(w http.ResponseWriter, r *http.Request) {}

func requestPayment(token string) {

	if len(token) < 0 {
		log.Printf("Empty Token")
		return
	}

	url := "https://api.sandbox.checkout.com/payments"
	httpMethod := "POST"
	contentType := "application/json"
	sKey := "sk_test_b4bc5b1d-e509-4df3-9690-073fa5cb0fe9"
	body := map[string]interface{}{
		"source": map[string]string{
			"type":  "token",
			"token": token,
		},
		"amount":    "2500",
		"currency":  "GBP",
		"reference": "Test Order",
		"3ds": map[string]bool{
			"enabled":     true,
			"attempt_n3d": true,
		},
		"customer": map[string]string{
			"email": "shiuhyaw.phang@checkout.com",
			"name":  "Yaw",
		},
	}
	bytesRepresentation, err := json.Marshal(body)
	if err != nil {
		log.Fatalln(err)
	}

	req, err := http.NewRequest(httpMethod, url, bytes.NewBuffer(bytesRepresentation))
	if err != nil {
		log.Printf("http.NewRequest() error: %v\n", err)
		return
	}
	req.Header.Add("Content-Type", contentType)
	req.Header.Add("Authorization", sKey)
	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		fmt.Printf("http.Do() error: %v\n", err)
		return
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("ioutil.ReadAll() error: %v\n", err)
		return
	}

	var response Response
	err = json.Unmarshal(data, &response)
	fmt.Println(response)
	switch response.Status {
	case "Pending":
		fmt.Println(response.ID)
		fmt.Println(response.Status)
		fmt.Println(response.Approved)
		fmt.Println(response.Links.RedirectURL.URLString)
		open(response.Links.RedirectURL.URLString)
	case "Authorized":
		fmt.Println(response.ID)
		fmt.Println(response.Status)
		fmt.Println(response.Approved)
	default:
		fmt.Println(response.ID)
	}
}

func open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}