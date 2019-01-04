// Copyright 2018 Google Inc. All rights reserved.
// Use of this source code is governed by the Apache 2.0
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine/log"
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
	err := godotenv.Load()
	if err != nil {
		log.Errorf(context.Background(), "Error loading .env file")
	}

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

	if r.Method == "GET" {
		indexTemplate.Execute(w, "")
		return
	}

	return
}

func metricsHandler(w http.ResponseWriter, request *http.Request) {
	ctx := appengine.NewContext(request)

	params := templateParams{}

	queryLimit, err := strconv.Atoi(request.URL.Query().Get("limit"))
	if err != nil { // No limit param was passed
		queryLimit = 200
		//return
	}

	// Create query for metrics
	q := datastore.NewQuery("Event").Order("-Time").Limit(queryLimit)

	// Restructure query if time was queried in request
	layout := time.RFC3339
	queryTime, err := time.Parse(layout, request.URL.Query().Get("time"))
	if err != nil {
		log.Errorf(ctx, "Error parsing time query param: %v", err)
	} else {
		log.Infof(ctx, "Recieved Date: %v", queryTime)

		// Query should look like:
		// SELECT * FROM `Event` WHERE `Time` <= "2019-01-03T09:30:20.00002-05:00"
		// Default for GetAll limit is 1k
		q = q.Filter("Time >= ", queryTime).Limit(1000)
	}

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

	webhookAuth := os.Getenv("PARTICLE_WEBHOOK_AUTH")

	auth := request.Header.Get("Authorization")
	if auth != webhookAuth {
		log.Errorf(ctx, "Webhook doesn't have valid Authorizatoin header")

		w.WriteHeader(http.StatusInternalServerError)
		return
	}

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
