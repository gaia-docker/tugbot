## Data we would like that "tugbot collect" will send to Result Service:
1. container info (using docker inspect, contains data such as exit code, start time, end time, labels, etc..)
2. Console output
3. Results folder 

## Webhook or REST API?
For first implementation we prefer REST API - it is simpler to implement it that way
Later on we can add webhook support for both "collect" and "Result Service" as this is more generic and simplifies authentication from "collect" perspective.

## API Design

```
Contect-Type: "application/gzip"
POST on http://result-service:8080/results?mainfile=results.txt&exitcode=1&start-time=2016-05-30 14:00&end-time=2016-05-30 14:05&...
```

Query params:

Non of the query params are mandatory.
* `mainfile` - the main test results file. In tugbot-result-service implementation this file will be echoed to the websocket
* `exitcode` - test exit code
* `starttime` - test start time
* `endtime` - test end time

We think about single tar.gz that will contain all of the info (3 folders: "container_info", "console_output", "results").
For performance and simplicity tugbot "collect" will add some essential data as query params (so Result Service won't need to unzip the input everytime it need something)
