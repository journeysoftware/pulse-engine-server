package github

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

type Delivery struct {
	UserId      string
	AccessToken string
	Repository  map[int]struct {
		Login string
		Repo  string
	}
}

func ParseDelivery(req *http.Request) (*Delivery, error) {
	payload := Delivery{}
	payload.AccessToken = req.Header.Get("Access-Token")
	payload.UserId = req.Header.Get("User-Id")
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	parseError := json.Unmarshal(body, &payload.Repository)
	if parseError != nil {
		fmt.Println("error:", parseError)
		return nil, parseError
	}

	return &payload, nil
}

// Client ...
func Client(w http.ResponseWriter, r *http.Request, auth []byte) {
	hc, err := ParseDelivery(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		log.Printf("Failed processing hook! ('%s')", err)
		io.WriteString(w, "{}")
		return
	}

	url := "https://localhost:3030"
	accessToken := [1]string{"GITHUB_TOKEN=" + hc.AccessToken}

	for _, repo := range hc.Repository {
		command := [1]string{"collect_prs " + repo.Login + " " + repo.Repo}

		m := map[string]interface{}{
			"AttachStdin":  false,
			"AttachStdout": true,
			"AttachStderr": true,
			"DetachKeys":   "ctrl-p,ctrl-q",
			"Tty":          false,
			"Cmd":          command,
			"Env":          accessToken}

		body, marshalErr := json.Marshal(m)
		if marshalErr != nil {
			fmt.Printf("Error:", marshalErr)
			return
		}
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
		req.Header.Set("X-Registry-Auth", base64.StdEncoding.EncodeToString(auth))
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Request error", err)
			return
		}
		fmt.Printf("post to collector success", resp)
		resp.Body.Close()
		return
	}
}
