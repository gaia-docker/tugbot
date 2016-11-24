## Data we would like that "tugbot collect" will send to Result Service:
1. Container info (using docker inspect, contains data such as exit code, start time, end time, labels, etc..)
2. Console output
3. Results folder 

## Webhook or REST API?
For first implementation we prefer REST API - it is simpler to implement it that way
Later on we can add webhook support for both "collect" and "Result Service" as this is more generic and simplifies authentication from "collect" perspective.

## API Design
```
Contect-Type: "application/gzip", "application/json"
POST on http://result-service:8080/results?docker.imagename=gaia-integartion-tests:latest&mainfile=results.txt&exitcode=1&start-time=2016-05-30 14:00&end-time=2016-05-30 14:05&...
```

Query params:

None of the query params are mandatory.
* `docker.imagename` - the docker image name of the test container
* `mainfile` - the main test results file. In tugbot-result-service implementation this file will be echoed to the websocket
* `exitcode` - test exit code
* `start-time` - test start time
* `end-time` - test end time

If the content type is "application/gzip":

We think about single tar.gz that will contain all of the info (3 folders: "container_info", "console_output", "results").
For performance and simplicity tugbot "collect" will add some essential data as query params (so Result Service won't need to unzip the input everytime it need something)

If the content type is "application/json":

Example of body:
```json
{
  "ImageName": "gaiadocker/voting-e2e:latest",
  "ContainerId": "93ce780df3095f631d2a64f02a356d51dd287311488df84e84adc947a8f2e332",
  "StartedAt": "2016-07-25T18:24:20.572308911Z",
  "FinishedAt": "2016-07-25T18:24:21.549659751Z",
  "ExitCode": 0,
  "HostName": "390697d726f1",
  "TestSet": {
    "Name": "Mocha Tests",
    "Time": 0.034,
    "Tests": [
      {
        "Name": "\"before all\" hook",
        "Status": "Failed",
        "Time": 0,
        "Failure": "Cannot read property 'statusCode' of undefined\nTypeError: Cannot read property 'statusCode' of undefined\n    at Request._callback (specs/e2e/voting-test.js:15:18)\n    at self.callback (node_modules/request/request.js:187:22)\n    at Request.onRequestError (node_modules/request/request.js:813:8)\n    at Socket.socketErrorListener (_http_client.js:267:9)\n    at emitErrorNT (net.js:1269:8)"
      },
      {
        "Name": "\"after all\" hook",
        "Status": "Failed",
        "Time": 0,
        "Failure": "Cannot read property 'statusCode' of undefined\nTypeError: Cannot read property 'statusCode' of undefined\n    at Request._callback (specs/e2e/voting-test.js:53:18)\n    at self.callback (node_modules/request/request.js:187:22)\n    at Request.onRequestError (node_modules/request/request.js:813:8)\n    at Socket.socketErrorListener (_http_client.js:267:9)\n    at emitErrorNT (net.js:1269:8)"
      }
    ]
  }
}
```

## _tugbot run_ events

_tugbot run_ executes tests in response to different events & publish events' data as a webhook. Result service can index those events' data into elasticsearch.
This allows us to create a graph that correlate test failures and environment events, which we are using in order to do analysis of how environment changes, like deploy a new docker image, impacts quality.

```
Contect-Type: "application/json"
POST on http://result-service:8080/events
```

Example of body:
```json
{
  "Type": "docker container",
  "ID": "93ce780df3095f631d2a64f02a356d51dd287311488df84e84adc947a8f2e332",
  "StartedAt": "2016-07-25T18:24:20.572308911Z",
  "FinishedAt": "2016-07-25T18:24:21.549659751Z",
  "Tag": "latest",
  "From": "perl",
  "Action": "run",
  "Status": "up"
}
```
