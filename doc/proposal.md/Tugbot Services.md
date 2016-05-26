# Tagbot Proposal Draft

Everything below is subject to change, see it as a starting point only.

## Tugbot Framework

**Tagbot** is an Integration Testing Framework for Docker based production/staging/testing environment. The **Tagbot** Framework performs two core tasks:
1. Executes *test containers* on some *event* (see below)
2. Continuously collects test results and uploads them to some **Result Service** (more about it later)

> **Tagbot** is not resposible to **first** run of *test container*! 

User may use same automation tool (Chef, Ansible, etc) or Docker scheduler (Kubernetes, Swarm/Compose, etc.), s/he is using regullary to deploy and run application containers. The **Tugbot** does not force user to learn a new paradigm for deplyment and execition of Docker containers. Use whatever tool you are already using!

**Tugbot** permorms **next** *test container* execution(s) on specified *event*. This can be timer event (every 20 min), Docker event (create, run, others...) or other events (in the future) like package update, kernel update, configuration change, etc. **Tagbot** also collects test results, and it tries to do so also for first *test container* run too.

### Result Service

The **Result Service** is a web service, that implements [Result Service API](#TODO). We will provide an open source **Result Service** implementation, that can be run on AWS Cloud, using API Gateway, Lambda and S3.

### Tugbot Packaging

**Tugbot** framework consists from several services, currently two: `watch` and `run`. For the first version and in sake of simplicity, we package both services into a single binary file. **Tagbot** binary can be deployed as a native application or as a Docker image.
The **Tugbot** Docker image is based on Alpine Linux and contains the above single binary: `tugbot`

## Test Container

*Test container* is a regular Docker container. We use Docker `LABEL` to discover *test container* and **Tugbot** related test metadata. These labels can be part of image or can be specified at runtime, using `--label` `docker run` option.
**Tugbot** will trigger a sequential *test container* execution on *event* (see `tagbot.event.*` labels).

### Tagbot labels
All **Tugbot** lables must be prefixed with `tagbot.` to avoid potential conflict with other labels. 

- `tugbot.test` - this is a *test container* marker label; without it, **Tugbot** will not recognize this container a a *test container*
- `tugbot.results` - directory, where *test container* reports test results; deafult to `/var/tests/results`
- `tugbot.event.timer` - recurrency time interval; can use time suffix ("s", "m", "h") for readability
- `tugbot.event.docker` - list of comma separated Docker events

## Watch Service

```dockerfile
tugbot watch ...
```

`watch` - setup a watch monitor for specified directories (watching multiple directories is possible); upload all new files and modified files (tails?) to analytics service; use API key

- ` --dir, -d [ --dir ...]` - directory to watch fore new test results; it’s possible to watch after multiple directories
- `--result-service, -r`    - result service URL for uploading test results
- `--api-key, -k`           - result service API key
- `--help, -h`              - Show documentation about the supported flags for `watch` command

```dockerfile
tugbot watch --dir /var/tests -dir /home/misc/tests --result-service http://nga.hp.com --api-key ACSD34SSD85DF 
```

For directory monitor we shoud use `inotify` as a Linux tool, or wrapper on top of it. There are built-in golang wrapper library and some more high level libraries on top of it too.
The basic idea it to get file system modification events, such as "directory created", "close on write" and others, and upload modified files to analytics service.

Maybe the `--dir` option is obsolete. **Tugbot** `watch` service can discover all *test containers* on host and inspect their metadata to find out where *test result* directory from container is mounted. This approach requires less configuration and is more secure (you reduce chance to upload wrong files to cloud), but it also requires more sofisticated code.

Currently, for **Result Service**, I suggest to use API Gateway wrapper on top of S3. We can always pickup stored results later and pass them further. This also might make the `--result-service` URL an optional parameter. 

## Test Run Service

```dockerfile
tugbot run ... [container, container ...]
```

`run` - run specified *test containers*, based on further configuration. If no container specified, **Tagbot** will run all *test containers* on current Docker host.

By default, **Tagbot** inspects *test container* configuration: to know when to run *test container* and where to look for test results. This configuration can be also specified at runtime and can overwrite "default" configuration.

### Run Options

* `--container, -c` - Use this option to specify/overwrite test settings for specified *test container*. The valid value contains semicolon separated list of test labels. For example: `name=selenium_tests;results=/var/log/results;event.docker=create` or `name=docker_bench;results=/var/bench/results;event.timer=4h30m`.
* `--host, -h` Docker daemon socket to connect to. Defaults to "unix:///var/run/docker.sock" but can be pointed at a remote Docker host by specifying a TCP endpoint as "tcp://hostname:port". The host value can also be provided by setting the `DOCKER_HOST` environment variable.
* `--interval, -i` Poll interval (in seconds). This value controls how frequently **Tugbot** will poll for new images. Defaults to 300 seconds (5 minutes).
* `--no-pull` Do not pull new images. When this flag is specified, **Tugbot** will not attempt to pull new images from the registry. Instead it will only monitor the local image cache for changes. Use this option if you are building new images directly on the Docker host without pushing them to a registry.
* `--cleanup` Remove old images after updating. When this flag is specified, **Tugbot** will remove the old image after restarting a container with a new image. Use this option to prevent the accumulation of orphaned images on your system as containers are updated.
* `--tls` Use TLS when connecting to the Docker socket but do NOT verify the server's certificate. If you are connecting a TCP Docker socket protected by TLS you'll need to use either this flag or the `--tlsverify` flag (described below). The `--tlsverify` flag is preferred as it will cause the server's certificate to be verified before a connection is made.
* `--tlsverify` Use TLS when connecting to the Docker socket and verify the server's certificate. If you are connecting a TCP Docker socket protected by TLS you'll need to use either this flag or the `--tls` flag (describe above).  
* `--tlscacert` Trust only certificates signed by this CA. Used in conjunction with the `--tlsverify` flag to identify the CA certificate which should be used to verify the identity of the server. The value for this flag can be either the fully-qualified path to the *.pem* file containing the CA certificate or a string containing the CA certificate itself. Defaults to "/etc/ssl/docker/ca.pem".
* `--tlscert` Client certificate for TLS authentication. Used in conjunction with the `--tls` or `--tlsverify` flags to identify the certificate to use for client authentication. The value for this flag can be either the fully-qualified path to the *.pem* file containing the client certificate or a string containing the certificate itself. Defaults to "/etc/ssl/docker/cert.pem".
* `--tlskey` Client key for TLS authentication. Used in conjunction with the `--tls` or `--tlsverify` flags to identify the key to use for client authentication. The value for this flag can be either the fully-qualified path to the *.pem* file containing the client key or a string containing the key itself. Defaults to "/etc/ssl/docker/key.pem".
* `--debug` Enable debug mode. When this option is specified you'll see more verbose logging in the **Tugbot** log file.
* `--help` Show documentation about the supported flags.
* `--version, -v` Print **Tagbot** version