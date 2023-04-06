# System for Processing Runtime Metrics
The system for processing runtime metrics helps monitor the state of a remote machine and consists of two parts - an agent (client) and a server.

# Agent
Collects and sends runtime metrics of the system at a specified frequency. You can add a key for encrypting data using SHA256.
## Agent Settings
* command line flag `a` or environment variable `ADDRESS` to specify the server address, `127.0.0.1:8080` by default
* command line flag `p` or environment variable `POLL_INTERVAL` to specify intervals between metric measurements, 2 seconds by default
* command line flag `r` or environment variable `REPORT_INTERVAL` to specify intervals between sending metrics, 5 seconds by default
* command line flag `k` or environment variable `KEY` to specify the encryption key

# Server
Accepts and processes metrics. Interacts with the PostgreSQL database at the specified address. If not available, uses internal memory. Additionally, there is an option to save data to a file.
## Features
* receive a metric for saving
* receive a group of metrics for saving
* return a metric value
* return a list of known metrics
* check database availability
## Server Settings
* command line flag `a` or environment variable `ADDRESS` to specify the address, `127.0.0.1:8080` by default
* command line flag `i` or environment variable `STORE_INTERVAL` to specify intervals between creating file backup when using internal memory, 300 seconds by default
* command line flag `r` or environment variable `RESTORE` to specify whether to load metrics from the file when using internal memory, `true` by default
* command line flag `f` or environment variable `STORE_FILE` to specify the file for backup when using internal memory, `/tmp/devops-metrics-db.json` by default
* command line flag `k` or environment variable `KEY` to specify the encryption key
* command line flag `d` or environment variable `DATABASE_DSN` to specify the PostgreSQL database DSN
## Usage
The server accepts `POST` and `GET` requests with content-type application/json.
### Receive a metric for saving
#### Request
`POST` to `/update` in the format

    {
        "id" : "Alloc",
        "type": "gauge",
        "value": 1.23,
        "hash": "someHash"
    }
#### Responses
* `200 OK` on successful saving of the metric
* `400 Bad Request` on parsing value error or incorrect hash
* `501 Status Not Implemented` when attempting to save a metric with an unknown type
* `404 Not Found` on other errors
### Receive multiple metrics for saving
#### Request
`POST` to `/updates` in the format

    [
        {
            "id" : "Alloc",
            "type": "gauge",
            "value": 1.23,
            "hash": "someHash"
        }
    ]
#### Responses
* `200 OK` on successful saving of metrics
* `400 Bad Request` on parsing value errors or incorrect hash
* `501 Status Not Implemented` when attempting to save a metric with an unknown type
* `404 Not Found` on other errors
### Return metric value
#### Request
`POST` to `/value` in the format

    {
        "id" : "Alloc",
        "type": "gauge",
    }
#### Responses
* `200 OK` and the metric value
* `404 Not Found` when attempting to get an unknown metric
* `501 Status Not Implemented` when attempting to get a metric with an unknown type
* `400 Bad Request` on other errors 
### Return all known metrics
#### Request
`GET` on `/`
#### Response
* `200 OK` and the list of metrics
### Check database availability
#### Request
`GET` on `/ping`
#### Responses
* `200 OK` if the database is available
* `500 Status Internal Server Error` if the database is unavailable
## Planned improvements
* Move metric types and available metrics to server settings
* Add the ability to change available metric types using requests to the server
* Improve error handling 
