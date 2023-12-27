package main

import (
	"log"
	"main/internal/openai"
	"main/internal/theguardian"
	"net/http"
	"os"
	"time"
)

func main() {
	// check and set API keys
	theguardian.Stats.ApiKey = os.Getenv("GUARDIAN_API")
	if theguardian.Stats.ApiKey == "" {
		log.Panic("Guardian API key missing")
	}
	// set poll interval to 20 minutes
	theguardian.Stats.PauseBetweenCalls = 20 * time.Minute

	if openai.Stats.ApiKey = os.Getenv("OPENAI_API"); openai.Stats.ApiKey == "" {
		log.Panic("OpenAI API key missing")
	}

	// create server mux
	mux := http.NewServeMux()

	mux.HandleFunc("/breaking/", breakingNewsAPIHandler)
	mux.Handle("/", http.FileServer(http.Dir("./web/")))

	err := http.ListenAndServe(":8988", mux)
	log.Fatal(err)
}
