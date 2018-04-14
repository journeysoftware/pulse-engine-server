package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"pulseengine/github"
	"time"
)

type AuthForDocker struct {
	Username      string `json:"username"`
	Password      string `json:"password"`
	Email         string `json:"email"`
	ServerAddress string `json:"serveraddress"`
}

func collectorHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	collectors := map[string]func(http.ResponseWriter, *http.Request, []byte){
		"github": github.Client,
		// "jira": jira
	}

	localImport, err := ioutil.ReadFile("./local.json")
	if err != nil {
		fmt.Println(err.Error())
		return
	}

	authOptions := AuthForDocker{}
	parseError := json.Unmarshal(localImport, &authOptions)

	if parseError != nil {
		fmt.Println("error:", parseError)
		return
	}

	registryAuth, marshalErr := json.Marshal(authOptions)
	if marshalErr != nil {
		fmt.Printf("Error: ", marshalErr)
		return
	}
	collectors[vars["source"]](w, r, registryAuth)
}

func main() {
	var wait time.Duration
	flag.DurationVar(&wait, "graceful-timeout", time.Second*15, "the duration for which the server gracefully wait for existing connections to finish - e.g. 15s or 1m")
	flag.Parse()

	r := mux.NewRouter()
	r.HandleFunc("/collector/{source}", collectorHandler).
		Methods("POST")

	srv := &http.Server{
		Addr: "0.0.0.0:3000",
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r, // Pass our instance of gorilla/mux in.
	}

	// Run our server in a goroutine so that it doesn't block.
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			fmt.Println(err)
		}
	}()

	c := make(chan os.Signal, 1)
	// We'll accept graceful shutdowns when quit via SIGINT (Ctrl+C)
	// SIGKILL, SIGQUIT or SIGTERM (Ctrl+/) will not be caught.
	signal.Notify(c, os.Interrupt)

	// Block until we receive our signal.
	<-c

	// Create a deadline to wait for.
	ctx, cancel := context.WithTimeout(context.Background(), wait)
	defer cancel()
	// Doesn't block if no connections, but will otherwise wait
	// until the timeout deadline.
	srv.Shutdown(ctx)
	// Optionally, you could run srv.Shutdown in a goroutine and block on
	// <-ctx.Done() if your application should wait for other services
	// to finalize based on context cancellation.
	fmt.Println("shutting down")
	os.Exit(0)
}
