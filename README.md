![tugbot logo](https://hpe-tugbot.github.io/images/Tugbot_green.png "tugbot logo")


[![CircleCI](https://circleci.com/gh/gaia-docker/tugbot.svg?style=shield)](https://circleci.com/gh/gaia-docker/tugbot)
[![codecov](https://codecov.io/gh/gaia-docker/tugbot/branch/master/graph/badge.svg)](https://codecov.io/gh/gaia-docker/tugbot)
[![Go Report Card](https://goreportcard.com/badge/github.com/gaia-docker/tugbot)](https://goreportcard.com/report/github.com/gaia-docker/tugbot)
[![Docker](https://img.shields.io/docker/pulls/gaiadocker/tugbot.svg)](https://hub.docker.com/r/gaiadocker/tugbot/)
[![Docker Image Layers](https://imagelayers.io/badge/gaiadocker/tugbot:latest.svg)](https://imagelayers.io/?images=gaiadocker/tugbot:latest)

## What is Tugbot?

It's widely accepted to run tests during CI build flow, but there are also some problems with this default approach. Here is a non complete list of such problems:

- It's not easy to create realistic test environment (close to production) with underlying infrastructure and different configurations.
- Integration tests might require access to "non-exposed" services and also highly depend on infrastructure and configuration.
- Some tests takes too long to run during CI build, for example: performance/stress tests, security tests, scans, background job tests, etc. Selecting appropriate tests for CI job is always balance between speed and simplicity of execution and "safety net" (how much tests need to be run to feel safe).
- Move to microservice architecture (tens and hundreds of micro services) and to Continuous Deployment (CD). Now every service has a separate CI pipeline for build/test/deploy and teams can achieve multiple deployments per day. Running too many tests or integration tests for each commit is a huge overhead and can become a team productivity "bottle neck".

The idea behind **Continuous Testing** is to execute tests, that require access to underlying infrastructure and to internal services or takes too much time to run, or need to be run on very specific infrastructure, inside real Docker cluster. Such tests should be run 24x7 with test execution triggered by timer or change event (service update, host OS update, configuration change, etc.)

**Tugbot** is in-cluster Continuous Testing Framework for Docker based runtime environments: like testing, staging, production.
**Tugbot** extends testing into deployment environments. It executes tests in response to different **change events** or periodically, collects test results and upload collected results to Test Analytics service.

**Check out the demo flow (including video) [here](https://github.com/gaia-docker/example-voting-app/blob/master/DEMO-FLOW.md)**

## How does Tugbot execute tests?

**Tugbot** can automatically detect and execute tests, packaged into Docker container. We call these containers - **Test Containers**. *Test container* is a regular Docker container. **Tugbot** uses Docker `LABEL` to discover *test container* and some **Tugbot** related test metadata. These labels can be part of image or can be specified at runtime, using `--label` option of `docker run` command.

> **Tugbot** is not responsible for **deployment** and **first** run of *test container*

**Tugbot** does not specify how to deploy and run application and *test* containers: user ay use an automation tool (Chef, Ansible, ...) or Docker scheduler (Kubernetes, Swarm, Nomad, ...). **Tugbot** will trigger a sequential *test container* execution on specified *events*.

### Tugbot Labels

All **Tugbot** labels must be prefixed with `tugbot.` to avoid potential conflict with other labels.

- `tugbot.test` - this is a *test container* discovery label; without it, **Tugbot** will not recognize this container as a *test container*
- `tugbot.results.dir` - directory, where *test container* reports test results; default to `/var/tests/results`
- `tugbot.event.docker` - marker label (no value is required) to subscribe **test container** to Docker events
- `tugbot.event.docker.filter.type` - Docker event type filter; can be one of `container, image, daemon, network, plugin, volume`
- `tugbot.event.docker.filter.action` - Docker event action (event type specific); multiple actions can be defined (comma separated)
- - `container` event type actions: `attach, commit, copy, create, destroy, detach, die, exec_create, exec_detach, exec_start, export, health_status, kill, oom, pause, rename, resize, restart, start, stop, top, unpause, update`
- - `image` event type actions: `delete, import, load, pull, push, save, tag, untag`
- - `plugin` event type actions: `install, enable, disable, remove`
- - `volume` event type actions: `create, mount, unmount, destroy`
- - `network` event type actions: `create, connect, disconnect, destroy`
- - `daemon` event type action: `reload`
- `tugbot.event.docker.filter.container` - container name, comma separated list of names or [RE2 regexp](https://github.com/google/re2/wiki/Syntax) (use `re2:` prefix); use this label to trigger test execution for events coming from these containers.
- `tugbot.event.docker.filter.image` - image name, comma separated list of names or [RE2 regexp](https://github.com/google/re2/wiki/Syntax) (use `re2:` prefix); use this filter to limit events coming from Docker images or containers created from these images
- `tugbot.event.docker.filter.label` - filter events coming from resource (container, image, volume, network), that has specified labels (and optionally values); this can be comma separated list of `key=value` pairs.

#####Example (Dockerfile):
```
...
# this is Tugbot Test Container
LABEL tugbot.test

# test results are saved into `/var/tests/results`
LABEL tugbot.results.dir=/var/tests/results

# subscribe to Docker events
LABEL tugbot.event.docker

# filter by event type == `container`
LABEL tugbot.event.docker.filter.type=container

# filter by event action: either `start` or `stop`
LABEL tugbot.event.docker.filter.action=start,stop

# subscribe to events from all containers with name prefix `hp...`
LABEL tugbot.event.docker.filter.container=re2:^hp
...
```

## Tugbot Run Service

```
$ tugbot help

NAME:
   Tugbot - Tugbot is a continuous testing framework for Docker based environments. Tugbot monitors changes in a runtime environment (host, os, container), runs tests (packaged into Test Containers), when event occurred and collects test results.

USAGE:
   tugbot [global options] command [command options] test containers: name, list of names, or none (for all test containers)

VERSION:
   v0.2.0

COMMANDS:
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --host value, -H value  daemon socket to connect to (default: "unix:///var/run/docker.sock") [$DOCKER_HOST]
   --tls                   use TLS; implied by --tlsverify
   --tlsverify             use TLS and verify the remote [$DOCKER_TLS_VERIFY]
   --tlscacert value       trust certs signed only by this CA (default: "/etc/ssl/docker/ca.pem")
   --tlscert value         client certificate for TLS authentication (default: "/etc/ssl/docker/cert.pem")
   --tlskey value          client key for TLS authentication (default: "/etc/ssl/docker/key.pem")
   --debug                 enable debug mode with verbose logging
   --help, -h              show help
   --version, -v           print the version
```

## Running Tugbot inside a Docker container

```
$ docker run -d --name tugbot-run --log-driver=json-file -v /var/run/docker.sock:/var/run/docker.sock gaiadocker/tugbot:master
```
