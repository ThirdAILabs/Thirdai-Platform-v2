# Running Model Bazaar Locally

1. Update the variables in the `.env.example` file. This should be very similar to that in the env variables from the old backend. The variables in the `.env.example` that you will likely have to change are marked as TODO, through some of these, particularly the db uris, are likely the same as the example. 
2. There are two options to run the backend. You can compile and run using `go build cmd/model_bazaar/main.go` followed by `./main <args>`. The second is to just use `go run cmd/model_bazaar/main.go <args>`.
3. There are a few options that can be passed to the backend, all of these are optional: 
```
  -env string
    	File to load env variables from, if not specified will just load them from the environment variables already defined.
  -port int
    	Port to run server on (default 8000).
  -skip_all
    	If specified will not restart llm-cache, llm-dispatch, and telemetry jobs.
  -skip_cache
    	If specified will not restart llm-cache job.
  -skip_dispatch
    	If specified will not restart llm-dispatch job.
  -skip_telemetry
    	If specified will not restart telemetry job.
```
4. Thus to run locally the simplest way is `go run cmd/model_bazaar/main.go --env <path to .env file>`. It is also useful to pass `--skip_all` if you are restarting frequently, since this will avoid having ot restart the telemetry, cache, and dispatch jobs for each run. 