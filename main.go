// Copyright 2018 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"google.golang.org/appengine"

	// [START imports]
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
	// [END imports]
)

var (
	indexTemplate = template.Must(template.ParseFiles("index.html"))
)

type Timestamp time.Time

type Event struct {
	Name string    `json:"event"`
	Data string    `json:"data"`
	Time time.Time `json:"published_at"`
}

type templateParams struct {
	Name string
	Data string

	// Display a notice if unable to retrieve Events
	Notice string

	Events []Event
}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/metrics", metricsHandler)
	http.HandleFunc("/_ah/health", healthCheckHandler)
	http.HandleFunc("/particle-webhook", particleWebhookHandler)

	appengine.Main()
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	ctx := appengine.NewContext(r)

	params := templateParams{}

	// Create query for Webhook events
	q := datastore.NewQuery("Event").Order("-Time").Limit(20)

	// Get all Webhook events
	if _, err := q.GetAll(ctx, &params.Events); err != nil {
		log.Errorf(ctx, "Getting events: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		params.Notice = "Couldn't get latest events. Refresh?"
		indexTemplate.Execute(w, params)
		return
	}

	if r.Method == "GET" {
		indexTemplate.Execute(w, params)
		return
	}

	return
}

func metricsHandler(w http.ResponseWriter, request *http.Request) {
	ctx := appengine.NewContext(request)

	params := templateParams{}

	// Create query for Webhook events
	q := datastore.NewQuery("Event").Order("-Time").Limit(200)

	// Get all Webhook events
	if _, err := q.GetAll(ctx, &params.Events); err != nil {
		log.Errorf(ctx, "Getting events: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		params.Notice = "Couldn't get latest events. Refresh?"
		indexTemplate.Execute(w, params)
		return
	}

	jData, err := json.Marshal(params)
	if err != nil {
		log.Errorf(ctx, "Marshalling response json for writer: %v", err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jData)
}

func particleWebhookHandler(w http.ResponseWriter, request *http.Request) {
	ctx := appengine.NewContext(request)

	decoder := json.NewDecoder(request.Body)

	// Handle POST request
	var event Event
	err := decoder.Decode(&event)

	if err != nil {
		log.Errorf(ctx, "datastore.Put: %v", err)
	}

	fmt.Println(event)

	log.Infof(ctx, "Recieved event: %v", event)
	// Ignore GET requests to this endpoint
	if request.Method == "GET" {
		return
	}

	// Save to database
	key := datastore.NewIncompleteKey(ctx, "Event", nil)

	if _, err := datastore.Put(ctx, key, &event); err != nil {
		log.Errorf(ctx, "datastore.Put: %v", err)

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	return
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "ok")
}
