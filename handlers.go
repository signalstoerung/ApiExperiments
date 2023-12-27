package main

import (
	"encoding/json"
	"fmt"
	"log"
	"main/internal/breaking"
	"main/internal/theguardian"
	"net/http"
	"time"
)

const (
	sinceDefault time.Duration = -1 * time.Hour
)

func breakingNewsAPIHandler(w http.ResponseWriter, r *http.Request) {
	// http.Error(w, "Not implemented", http.StatusNotImplemented)
	if theguardian.Stats.PauseTimeElapsed() {
		log.Printf("Retrieving latest updates...")
		var since time.Time
		var err error
		sinceReq := r.FormValue("since")
		if since, err = time.Parse(time.RFC3339, sinceReq); err != nil {
			since = time.Now().Add(sinceDefault)
		}
		theguardian.GetLiveblogUpdates(since)
	}
	latestFromDB(w, r)
}

func latestFromDB(w http.ResponseWriter, r *http.Request) {
	var stories []breaking.DevelopingStory
	var since time.Time
	var err error
	sinceReq := r.FormValue("since")
	if since, err = time.Parse(time.RFC3339, sinceReq); err != nil {
		since = time.Now().Add(sinceDefault)
	}
	log.Printf("Getting all stories updated since %v", since.Format("15:04"))
	stories, err = breaking.StoriesSince(since)
	if err != nil {
		http.Error(w, fmt.Sprintf("An error occured: %v. We're sorry.", err), http.StatusInternalServerError)
		return
	}
	enc := json.NewEncoder(w)
	enc.Encode(&stories)
}
