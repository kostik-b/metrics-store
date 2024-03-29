# Overview
`metrics-store` is a lightweight REST API server with an in-memory database, which stores the reports submitted to it.
Only two HTTP methods are supported: POST and GET. The former will accept one report in a format specified below for storage in the database. The latter will return all the reports stored to date in the database as JSON objects in an array.

### POST Requests

The JSON Schema for the POST requests can be found in the schemas folder.
An example of a JSON object for POST request is:
```
{
    "machineId": 61616,
    "stats": {
        "cpuTemp": 78,
        "fanSpeed": 500,
        "HDDSpace": 100,
        "internalTemp": 23
    },
    "lastLoggedIn": "admin/Tim",
    "sysTime": "Wed 2021-07-28 14:16:27"
}
```
An example of a response in case of a successful request is:
```
{
  "id": "c7055826-b23b-41d5-8026-951f0c424751",
  "message": "New entry added to the data store with id - c7055826-b23b-41d5-8026-951f0c424751"
}
```
### GET Requests
A GET request will return the above JSON objects as an array with the addition of an extra field - id, which is a unique id of that particular report.
The JSON Schema for GET responses can be found in the schemas folder.
An example of a GET response is:
```
[
  {
    "id": "b18ebf3c-2b22-490a-834e-bd096bb067ad",
    "machineId": 4444,
    "stats": {
      "cpuTemp": 78,
      "fanSpeed": 500,
      "HDDSpace": 100,
      "internalTemp": 23
    },
    "lastLoggedIn": "admin/Ian",
    "sysTime": "2022-04-21T19:25:43.219Z"
  }
]
```

### Directory Structure
`scripts` directory contains helper scripts to send POST and GET requests to the server, assuming the server listens on default port 4000.
`cmd` and `pkg` directories contain source code.
`schemas` directory contains schemas for GET responses and POST requests and responses.

# Compiling
In order to compile the solution, please run make in the root directory of this repository.
The binary file `metrics-store` will be placed in a newly created `bin` directory under project's root directory.

# Running
The `metrics-store` server can be run with the following parameters (these are also shown when `-h` argument is passed to the application):
```
Usage of metrics-store:
  -allow-unknown-fields
        Set to true to allow unknown fields
  -debug
        Set to true to enable debug output
  -listen-port int
        A port to listen on from 1 to 65535 (default 4000)
  -max-request-body-size int
        Maximum size of request body (default 1048576)
```

The route for `metrics-store` is `/metrics`, i.e. `http://localhost:4000/metrics`, assuming the server listens on port 4000.

# Future Work
* Patterns can be added to fetch specific metric entries, e.g. /metrics/<datetime>
* If the database grows big, compression or chunking strategies can be considered for the GET response in addition to the range selection logic.
* Some strategies need to be considered for archiving, relocating or removing data if the database gets too big.
* Potentially improve error handling of unmarshalling for handling POST request, e.g. by implementing recommendations from https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body, as currently e.g. any error occuring during umarshalling will result in "Bad Request"
* Produce swagger for the metrics-store
* Optimise concurrent access for datastore as map
* Create distinct loggers in metrics handler - for info, error and debug levels, each with its own prefix, or create different log levels
* Implement a custom type with MarshalJSON and UnmarshalJSON to parse sysTime field
* MarshalIndent can be turned on and off depending on whether we are in debug mode or not in the MetricsHandler to save bandwidth
* Comments can be improved
* Datastore as map can be moved into a separate directory under the directory it is currently in.
* An option can be added to bind to a particular interface
* Unit tests for metrics handler can be cleaned up - mocks can be moved to a separate file
* defaultMaxBodySize constant can be moved to a separate file and then used by both main.go and unit tests for metrics handler
* Wait interval during graceful shutdown, shutdownTimeout, can be made configurable
