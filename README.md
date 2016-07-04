# Tugbot

[![CircleCI](https://circleci.com/gh/gaia-docker/tugbot.svg?style=shield)](https://circleci.com/gh/gaia-docker/tugbot)
[![codecov](https://codecov.io/gh/gaia-docker/tugbot/branch/master/graph/badge.svg)](https://codecov.io/gh/gaia-docker/tugbot)
[![Go Report Card](https://goreportcard.com/badge/github.com/gaia-docker/tugbot)](https://goreportcard.com/report/github.com/gaia-docker/tugbot)
[![Docker badge](https://img.shields.io/docker/pulls/gaiadocker/tugbot.svg)](https://hub.docker.com/r/gaiadocker/tugbot/)


**Tugbot** is an Continuous Testing Framework for Docker based production/staging/testing environment. **Tugbot** executes *test containers* on some *event*.

User may use same automation tool (Chef, Ansible, etc) or Docker scheduler (Kubernetes, Swarm/Compose, etc.), s/he is using regularly to deploy and run application containers. The **Tugbot** does not force user to learn a new paradigm for deployment and execution of Docker containers. Use whatever tool you are already using!

> **Tugbot** is not responsible for **first** run of *test container*

**Tugbot** performs **subsequent** *test container* execution(s) on specified Docker *event*(create, run, others...).

## Test Container

*Test container* is a regular Docker container. We use Docker `LABEL` to discover *test container* and **Tugbot** related test metadata. These labels can be part of image or can be specified at runtime, using `--label` `docker run` option.
**Tugbot** will trigger a sequential *test container* execution on *event* (see `Tugbot.event.*` labels).

### Tugbot labels

All **Tugbot** labels must be prefixed with `tugbot.` to avoid potential conflict with other labels.

- `tugbot.test` - this is a *test container* marker label; without it, **Tugbot** will not recognize this container a a *test container*
- `tugbot.results` - directory, where *test container* reports test results; default to `/var/tests/results`
- `tugbot.event.docker` - list of comma separated Docker events

#####Example (Dockerfile):
```
LABEL tugbot.test=true
LABEL tugbot.results=/var/tests/results
LABEL tugbot.event.docker=start
```

## Tugbot Run Service

```
tugbot ... [container, container ...]
```

`run` - run specified *test containers*, based on further configuration. If no container specified, **Tugbot** will run all *test containers* on current Docker host.

By default, **Tugbot** inspects *test container* configuration: to know when to run *test container* and where to look for test results.

### Run Options

- `--host, -h`             Docker daemon socket to connect to. Defaults to "unix:///var/run/docker.sock" but can be pointed at a remote Docker host by specifying a TCP endpoint as "tcp://hostname:port". The host value can also be provided by setting the `DOCKER_HOST` environment variable.
- `--tls`                  Use TLS when connecting to the Docker socket but do NOT verify the server's certificate. If you are connecting a TCP Docker socket protected by TLS you'll need to use either this flag or the `--tlsverify` flag (described below). The `--tlsverify` flag is preferred as it will cause the server's certificate to be verified before a connection is made.
- `--tlsverify`            Use TLS when connecting to the Docker socket and verify the server's certificate. If you are connecting a TCP Docker socket protected by TLS you'll need to use either this flag or the `--tls` flag (describe above).
- `--tlscacert`            Trust only certificates signed by this CA. Used in conjunction with the `--tlsverify` flag to identify the CA certificate which should be used to verify the identity of the server. The value for this flag can be either the fully-qualified path to the *.pem* file containing the CA certificate or a string containing the CA certificate itself. Defaults to "/etc/ssl/docker/ca.pem".
- `--tlscert`              Client certificate for TLS authentication. Used in conjunction with the `--tls` or `--tlsverify` flags to identify the certificate to use for client authentication. The value for this flag can be either the fully-qualified path to the *.pem* file containing the client certificate or a string containing the certificate itself. Defaults to "/etc/ssl/docker/cert.pem".
- `--tlskey`               Client key for TLS authentication. Used in conjunction with the `--tls` or `--tlsverify` flags to identify the key to use for client authentication. The value for this flag can be either the fully-qualified path to the *.pem* file containing the client key or a string containing the key itself. Defaults to "/etc/ssl/docker/key.pem".
- `--debug`                Enable debug mode. When this option is specified you'll see more verbose logging in the **Tugbot** log file.
* `--help`                 Show documentation about the supported flags.
* `--version, -v`          Print `tugbot` version

## Running Tugbot inside a Docker container

```bash
docker run -d --name tugbot --log-driver=json-file -v /var/run/docker.sock:/var/run/docker.sock gaiadocker/tugbot
```