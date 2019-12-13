# Workflow
Workflow is a workflow engine that supports building and executing workflows
as a Directed Acyclic Graph (DAG).

It supports persistence and runs in a stateless fashion. It's only dependency
is a database (currently Postgres is the only supported database).

Workflow is at a very early stage in it's lifecycle. Bugs are being fixed as they are found and new features as they are proposed.

Use at your own risk.

## Concepts

There are 2 basic building blocks of workflows. There are [Steppers](https://github.com/mitchfriedman/workflow/blob/master/lib/run/stepper.go#L9-L13) that
are an individual unit of work. Steppers can execute a new stepper after it completes and can branch based on `onSuccess` or `onFailure`.

Steppers combine together to form [Jobs](https://github.com/mitchfriedman/workflow/blob/master/lib/run/job.go#L155-L160) that are a specific
ordering of Steppers.

Steppers and Jobs are stored in store (in memory) that can be accesed by the engine as needed.

Workflows are started when the engine receives the correct Trigger. You can create Triggers however you'd like by registering a [`Parser`](https://github.com/mitchfriedman/workflow/blob/master/lib/rest/router.go#L15-L17) with the provided [`Router`](https://github.com/mitchfriedman/workflow/blob/master/lib/rest/router.go#L20).

Workers are goroutines that are capable of executing your Workflows. They are continually polling the database which is acting as a queue for
the Workflows to execute.

There is a [`Watchdog`](https://github.com/mitchfriedman/workflow/blob/master/lib/engine/watchdog.go#L18) that watches for dead/inactive `Workers` and unassigns them from
jobs so another worker can pick up that job. This allows the system to run in stateless environments where all state is persisted in the database and is
reslient to worker failures.

Workers indicate they are still actively performing work by using a [`Heartbeat`](https://github.com/mitchfriedman/workflow/blob/master/lib/worker/heartbeats.go#L21).

## Usage

First, create a [logger](https://github.com/mitchfriedman/workflow/blob/master/lib/logging/logger.go#L17):
```go
logger := logging.New(cfg.Logger.AppName, os.Stderr)
```

Then, create your connection to the database with [Connect](https://github.com/mitchfriedman/workflow/blob/master/lib/db/db.go#L44):
```go

db, err := database.Connect(cfg.Postgres.MasterURL, cfg.Postgres.ReaderURL, true, logger)
if err != nil {
    logger.Fatalf("failed to connect to db: %+v, %+v", err, cfg.Postgres)
}
```

Next, you'll setup your repository access to the `Runs` and `Workers`:
```go
rr := run.NewDatabaseStorage(db)
wr := worker.NewDatabaseStorage(db)

jobStore := run.NewJobsStore()
stepperStore := run.NewStepperStore()
```

Then, you can create [Steppers](https://github.com/mitchfriedman/workflow/blob/master/lib/run/stepper.go#L9-L13) and register them in the `stepperStore` with
```go
stepperStore.Register(myStep)
```

and [Jobs](https://github.com/mitchfriedman/workflow/blob/master/lib/run/job.go#L155-L160) can be registered in the `jobStore` with:
```go
jobStore.Register(myJob)
```

Now, you'll want to setup your [`Router`](https://github.com/mitchfriedman/workflow/blob/master/lib/rest/router.go#L20) and [`Parser`](https://github.com/mitchfriedman/workflow/blob/master/lib/rest/router.go#L15-L17) with:
```go
parsers := []rest.Parser{webhook.NewGithubParser(workflows, logger, statsClient, rr)}
router := rest.NewRouter(cfg.Environment.ServiceName, jobStore, rr, parsers, logger)
```

Initialize your worker with:
```go
w := worker.NewWorker()
if err := wr.Register(context.Background(), w); err != nil {
	logger.Fatalf("failed to register worker: %v", err)
}
if err := wr.RenewLease(context.Background(), w, 45*time.Second); err != nil {
	logger.Fatalf("failed to renew workers first lease: %v", err)
}
```

Finally, you can run the `Engine`, `Router`, `Watchdog`, and `HeartbeatProcessor`:
```go
server := &http.Server{Addr: fmt.Sprintf(":%s", cfg.Server.Port), Handler: router}
hb := make(chan worker.Heartbeat, 1)
hbp := worker.NewHeartbeatProcessor(hb, wr, logger)
e := engine.NewEngine(w, stepperStore, rr, wr, hb, logger, stats)

go e.Start(context.Background()) // start the engine

go hbProcessor.Start(ctx) // start the heartbeat processor

go server.ListenAndServe() // start your favorite http server with the registered router

go engine.Watch(ctx, logger, wr, rr, time.Hour) // start the watchdog
```

## Contributing

Contributions are very welcome to Workflow. Workflow is in an early alpha phase while features are being proposed and use cases are being determined.

