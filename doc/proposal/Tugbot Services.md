# Tugbot Proposal Draft

Everything below is subject to change, see it as a starting point only.

## Tugbot Framework

**Tugbot** is an Integration Testing Framework for Docker based production/staging/testing environment. The **Tugbot** Framework performs two core tasks:
1. Executes *test containers* on some *event* (see below)
2. Continuously collects test results and uploads them to some **Result Service** (more about it later)

> **Tugbot** is not responsible to **first** run of *test container*!

User may use same automation tool (Chef, Ansible, etc) or Docker scheduler (Kubernetes, Swarm/Compose, etc.), s/he is using regularly to deploy and run application containers. The **Tugbot** does not force user to learn a new paradigm for deployment and execution of Docker containers. Use whatever tool you are already using!

**Tugbot** performs **subsequent** *test container* execution(s) on specified *event*. This can be timer event (every 20 min), Docker event (create, run, others...) or other events (in the future) like package update, kernel update, configuration change, etc. **Tugbot** also collects test results, and it tries to do so also for first *test container* run too.

### Result Service

The **Result Service** is a web service, that implements [Result Service API](Result Service API.md). We will provide an open source **Result Service** implementation, that can be run on AWS Cloud, using API Gateway, Lambda and S3.

### Tugbot Packaging

**Tugbot** framework consists from several services, currently two: `tagbot-watch` and `tugbot-run`. Each service is avaiable as single binary and can be deployed as a native application or as a Docker image.
The Docker image for any **Tugbot** service should be based on Alpine Linux and contain single service binary. This binary should not run as a `root` user.

## Test Container

*Test container* is a regular Docker container. We use Docker `LABEL` to discover *test container* and **Tugbot** related test metadata. These labels can be part of image or can be specified at runtime, using `--label` `docker run` option.
**Tugbot** will trigger a sequential *test container* execution on *event* (see `Tugbot.event.*` labels).

### Tugbot labels
All **Tugbot** labels must be prefixed with `Tugbot.` to avoid potential conflict with other labels.

- `tugbot.test` - this is a *test container* marker label; without it, **Tugbot** will not recognize this container a a *test container*
- `tugbot.results` - directory, where *test container* reports test results; default to `/var/tests/results`
- `tugbot.event.timer` - recurrent time interval; can use time suffix ("s", "m", "h") for readability
- `tugbot.event.docker` - list of comma separated Docker events

## Tugbot Watch Service


`tugbot-watch` - automatically discovers all *test containers*, collects test results from "exited" *test containers* (either from mounted volume or using `docker cp` command) and uploads all test result files to specified **Result Service**; use API key

- `--result-service, -r`  Result service URL for uploading test results.
- `--api-key, -k`         Result service API key.
- `--host, -h`            Docker daemon socket to connect to. Defaults to "unix:///var/run/docker.sock" but can be pointed at a remote Docker host by specifying a TCP endpoint as "tcp://hostname:port". The host value can also be provided by setting the `DOCKER_HOST` environment variable.
- `--cleanup`             Remove "exited" container, after collecting and successfully uploading all test results to specified Result Service.
- `--tls`                 Use TLS when connecting to the Docker socket but do NOT verify the server's certificate. If you are connecting a TCP Docker socket protected by TLS you'll need to use either this flag or the `--tlsverify` flag (described below). The `--tlsverify` flag is preferred as it will cause the server's certificate to be verified before a connection is made.
- `--tlsverify`           Use TLS when connecting to the Docker socket and verify the server's certificate. If you are connecting a TCP Docker socket protected by TLS you'll need to use either this flag or the `--tls` flag (describe above).
- `--tlscacert`           Trust only certificates signed by this CA. Used in conjunction with the `--tlsverify` flag to identify the CA certificate which should be used to verify the identity of the server. The value for this flag can be either the fully-qualified path to the *.pem* file containing the CA certificate or a string containing the CA certificate itself. Defaults to "/etc/ssl/docker/ca.pem".
- `--tlscert`             Client certificate for TLS authentication. Used in conjunction with the `--tls` or `--tlsverify` flags to identify the certificate to use for client authentication. The value for this flag can be either the fully-qualified path to the *.pem* file containing the client certificate or a string containing the certificate itself. Defaults to "/etc/ssl/docker/cert.pem".
- `--tlskey`              Client key for TLS authentication. Used in conjunction with the `--tls` or `--tlsverify` flags to identify the key to use for client authentication. The value for this flag can be either the fully-qualified path to the *.pem* file containing the client key or a string containing the key itself. Defaults to "/etc/ssl/docker/key.pem".
- `--debug`               Enable debug mode. When this option is specified you'll see more verbose logging in the **Tugbot** log file.
- `--help`                Show documentation about the supported flags.
- `--version, -v`         Print `tugbot-watch` service version.

```
tugbot-watch --result-service http://nga.hp.com --api-key ACSD34SSD85DF
```

## Tugbot Run Service

```
tugbot-run ... [container, container ...]
```

`run` - run specified *test containers*, based on further configuration. If no container specified, **Tugbot** will run all *test containers* on current Docker host.

By default, **Tugbot** inspects *test container* configuration: to know when to run *test container* and where to look for test results. This configuration can be also specified at runtime and can overwrite "default" configuration.

### Run Options

- `--container, -c`        Use this option to specify/overwrite test settings for specified *test container*. The valid value contains semicolon separated list of test labels. For example: `name=selenium_tests;results=/var/log/results;event.docker=create` or `name=docker_bench;results=/var/bench/results;event.timer=4h30m`.
- `--host, -h`             Docker daemon socket to connect to. Defaults to "unix:///var/run/docker.sock" but can be pointed at a remote Docker host by specifying a TCP endpoint as "tcp://hostname:port". The host value can also be provided by setting the `DOCKER_HOST` environment variable.
- `--tls`                  Use TLS when connecting to the Docker socket but do NOT verify the server's certificate. If you are connecting a TCP Docker socket protected by TLS you'll need to use either this flag or the `--tlsverify` flag (described below). The `--tlsverify` flag is preferred as it will cause the server's certificate to be verified before a connection is made.
- `--tlsverify`            Use TLS when connecting to the Docker socket and verify the server's certificate. If you are connecting a TCP Docker socket protected by TLS you'll need to use either this flag or the `--tls` flag (describe above).
- `--tlscacert`            Trust only certificates signed by this CA. Used in conjunction with the `--tlsverify` flag to identify the CA certificate which should be used to verify the identity of the server. The value for this flag can be either the fully-qualified path to the *.pem* file containing the CA certificate or a string containing the CA certificate itself. Defaults to "/etc/ssl/docker/ca.pem".
- `--tlscert`              Client certificate for TLS authentication. Used in conjunction with the `--tls` or `--tlsverify` flags to identify the certificate to use for client authentication. The value for this flag can be either the fully-qualified path to the *.pem* file containing the client certificate or a string containing the certificate itself. Defaults to "/etc/ssl/docker/cert.pem".
- `--tlskey`               Client key for TLS authentication. Used in conjunction with the `--tls` or `--tlsverify` flags to identify the key to use for client authentication. The value for this flag can be either the fully-qualified path to the *.pem* file containing the client key or a string containing the key itself. Defaults to "/etc/ssl/docker/key.pem".
- `--debug`                Enable debug mode. When this option is specified you'll see more verbose logging in the **Tugbot** log file.
* `--help`                 Show documentation about the supported flags.
* `--version, -v`          Print `tugbot-run` version
