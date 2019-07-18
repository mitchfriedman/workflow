package main

import (
	"context"
	"fmt"
	"log"
	"time"

	database "github.com/mitchfriedman/workflow/lib/db"
	"github.com/mitchfriedman/workflow/lib/run"
	"github.com/mitchfriedman/workflow/lib/worker"
)

func main() {
	dbURL := "postgres://localhost:5432/workflow?sslmode=disable"
	db, _ := database.Connect(dbURL, dbURL, false)
	fmt.Println("got db: ", db)

	jobStore := run.NewJobsStore()
	stepperStore := run.NewStepperStore()
	setupJob(jobStore, stepperStore)

	//rr := run.NewDatabaseStorage(db)

	//router := rest.NewRouter("service", jobStore, rr, &webhook.LocalParser{})
	//address := ":8000"
	//log.Printf("Starting Server at %s", address)
	//log.Fatal(http.ListenAndServe(address, router))

	w := worker.NewWorker()
	wr := worker.NewDatabaseStorage(db)
	if err := wr.Register(context.Background(), w); err != nil {
		log.Fatal("failed to register worker")
	}
	if err := wr.RenewLease(context.Background(), w, 10*time.Second); err != nil {
		log.Fatal("failed to register worker")
	}

	fmt.Printf("%+v", w)
}

func setupJob(js *run.JobStore, stepperStore *run.StepperStore) {
	// create jobs here.
}
