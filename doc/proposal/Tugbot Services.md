# Tagbot Proposal Draft

Everything below is subject to change, see it as a starting point only.

## Tugbot Framework

**Tagbot** is an Integration Testing Framework for Docker based production/staging/testing environment. The **Tagbot** Framework performs two core tasks:
1. Orchestrate *test containers* execution, triggered by some *event* (see below)
2. Continuously collect test results and upload them to some **Result Service** (more about it later)

> **Tagbot** should not change the way users deploy and run Docker container.

**Tagbot** is not responsible to **first** run of *test container*! User should use same automation tool (Chef, Ansible, etc) or Docker scheduler (Kubernetes, Swarm/Compose, etc.), s/he is using regularly to deploy and run application containers. **Tugbot** does not force user to learn a new paradigm for deployment and execution of Docker containers. Use whatever tool you are already using!

**Tugbot** performs **subsequent** *test container* execution(s), triggered by subscribed *event*. The event can be timer event (every 20 min), Docker event (create, run, others...) or some other events (in the future), like package update, kernel update, configuration change, etc. **Tagbot** automatically handles collection of test results. It will try to do so even for first *test container* run.

### Result Service

The **Result Service** is a web service, that implements [Result Service API](#TODO). We will provide an open source **Result Service** implementation that echo test results, aggregated from multiple hosts, for every connected client. We should consider another **Result Service** implementation, running on AWS and storing all sent test results into S3, for further processing (example: Gaia events) and notifications (example: Slack)

### Tugbot Packaging

**Tugbot** framework consists from several services, currently two: `watch` and `run`. For the first version and in sake of simplicity, we package both services into a single binary file. **Tagbot** binary can be deployed as a native application or as a Docker image.
The **Tugbot** Docker image is based on Alpine Linux and contains the above single binary: `tugbot`

## Test Container

*Test container* is a regular Docker container. We use Docker `LABEL` to discover *test container* and **Tugbot** related test metadata. These labels can be part of image or can be specified at runtime, using `--label` `docker run` option.
**Tugbot** will trigger a sequential *test container* execution on *event* (see `tagbot.event.*` labels).

### Tagbot labels
All **Tugbot** labels must be prefixed with `tagbot.` to avoid potential conflict with other labels.

- `tugbot.test` - this is a *test container* marker label; without it, **Tugbot** will not recognize this container a a *test container*
- `tugbot.results` - directory, where *test container* reports test results; default to `/var/tests/results`
- `tugbot.event.timer` - recurrent time interval; can use time suffix ("s", "m", "h") for readability
- `tugbot.event.docker` - list of comma separated Docker events

## Watch Service

```dockerfile
tugbot watch ...
```

`watch` - setup a watch monitor all test results directories from **test containers** and upload all new/modified files to **Result Service**.

- `--result-service, -r`    - result service URL for uploading test results
- `--api-key, -k`           - result service API key
- `--help, -h`              - Show documentation about the supported flags for `watch` command

```dockerfile
tugbot watch -v /var/run/docker:/var/run/docker --result-service http://nga.hp.com --api-key ACSD34SSD85DF
```

**Implementation Tip:** For directory monitor we should use `inotify` as a Linux tool, or wrapper on top of it. There are built-in golang wrapper library and some more high level libraries on top of it too.
The basic idea it to get file system modification events, such as "directory created", "close on write" and others, and upload modified files to analytics service.

**Tugbot** `watch` service should automatically discover all *test containers* on host and inspect their metadata to find out where *test result* directory from container is mounted. This approach requires less configuration and is more secure (you reduce chance to upload wrong files to cloud), but it also requires more sophisticated code. `watch` can also extract test results from already "exited" containers, using `docker cp` command.
**Implementation Tip:** `watch` service should "subscribe" to Docker events to catch different states of container lifecycle.

## Test Orchestrate Service

```dockerfile
tugbot orchestrate ... [container, container ...]
```

`orchestrate` - orchestrate execution of specified *test containers*, based on further configuration. If no container specified, **Tagbot** will orchestrate tests for all automatically discovered *test containers* on Docker host, where it runs.

By default, **Tagbot** inspects *test container* configuration: to know when to run *test container* and where to look for test results. This configuration can be also specified at runtime and can overwrite "default" configuration.

### Orchestrate Options

* `--container, -c` - Use this option to specify/overwrite test settings for specified *test container*. The valid value contains semicolon separated list of test labels. For example: `name=selenium_tests;results=/var/log/results;event.docker=create` or `name=docker_bench;results=/var/bench/results;event.timer=4h30m`.
* `--host, -h` Docker daemon socket to connect to. Defaults to "unix:///var/run/docker.sock" but can be pointed at a remote Docker host by specifying a TCP endpoint as "tcp://hostname:port". The host value can also be provided by setting the `DOCKER_HOST` environment variable.
* `--tls` Use TLS when connecting to the Docker socket but do NOT verify the server's certificate. If you are connecting a TCP Docker socket protected by TLS you'll need to use either this flag or the `--tlsverify` flag (described below). The `--tlsverify` flag is preferred as it will cause the server's certificate to be verified before a connection is made.
* `--tlsverify` Use TLS when connecting to the Docker socket and verify the server's certificate. If you are connecting a TCP Docker socket protected by TLS you'll need to use either this flag or the `--tls` flag (describe above).
* `--tlscacert` Trust only certificates signed by this CA. Used in conjunction with the `--tlsverify` flag to identify the CA certificate which should be used to verify the identity of the server. The value for this flag can be either the fully-qualified path to the *.pem* file containing the CA certificate or a string containing the CA certificate itself. Defaults to "/etc/ssl/docker/ca.pem".
* `--tlscert` Client certificate for TLS authentication. Used in conjunction with the `--tls` or `--tlsverify` flags to identify the certificate to use for client authentication. The value for this flag can be either the fully-qualified path to the *.pem* file containing the client certificate or a string containing the certificate itself. Defaults to "/etc/ssl/docker/cert.pem".
* `--tlskey` Client key for TLS authentication. Used in conjunction with the `--tls` or `--tlsverify` flags to identify the key to use for client authentication. The value for this flag can be either the fully-qualified path to the *.pem* file containing the client key or a string containing the key itself. Defaults to "/etc/ssl/docker/key.pem".
* `--debug` Enable debug mode. When this option is specified you'll see more verbose logging in the **Tugbot** log file.
* `--help` Show documentation about the supported flags.
* `--version, -v` Print **Tagbot** version
