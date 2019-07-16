package main

import (
	"fmt"
	database "github.com/mitchfriedman/workflow/lib/db"
	"github.com/mitchfriedman/workflow/lib/rest"
	"github.com/mitchfriedman/workflow/lib/rest/webhook"
	"github.com/mitchfriedman/workflow/lib/run"
	"log"
	"net/http"
)

func main() {
	db, _ := database.Connect("master", "reader", false)
	fmt.Println("got db: ", db)

	jobStore := run.NewJobsStore()
	stepperStore := run.NewStepperStore()
	setupJob(jobStore, stepperStore)

	rr := run.NewDatabaseStorage(db)

	router := rest.NewRouter("service", jobStore, rr, &webhook.LocalParser{})
	address := ":8000"
	log.Printf("Starting Server at %s", address)
	log.Fatal(http.ListenAndServe(address, router))
}

func setupJob(js *run.JobStore, stepperStore *run.StepperStore) {
	// create jobs here.
}
