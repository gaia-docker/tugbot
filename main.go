package main // import "github.com/gaia-docker/tugbot"

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/gaia-docker/tugbot-common"
	"github.com/gaia-docker/tugbot/actions"
	"github.com/gaia-docker/tugbot/container"
	"github.com/samalba/dockerclient"
)

var (
	client    container.Client
	names     []string
	wgr       sync.WaitGroup
	wgp       sync.WaitGroup
	publisher common.Publisher
)

const (
	// Release version
	Release = "v0.3.0"
)

func init() {
	log.SetLevel(log.InfoLevel)
}

func main() {
	rootCertPath := "/etc/ssl/docker"
	if os.Getenv("DOCKER_CERT_PATH") != "" {
		rootCertPath = os.Getenv("DOCKER_CERT_PATH")
	}

	app := cli.NewApp()
	app.Name = "Tugbot"
	app.Version = Release
	app.Usage = "Tugbot is a continuous testing framework for Docker based environments. Tugbot monitors changes in a runtime environment (host, os, container), runs tests (packaged into Test Containers), when event occured and collects test results."
	app.ArgsUsage = "test containers: name, list of names, or none (for all test containers)"
	app.Before = before
	app.Action = start
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "host, H",
			Usage:  "daemon socket to connect to",
			Value:  "unix:///var/run/docker.sock",
			EnvVar: "DOCKER_HOST",
		},
		cli.BoolFlag{
			Name:  "tls",
			Usage: "use TLS; implied by --tlsverify",
		},
		cli.BoolFlag{
			Name:   "tlsverify",
			Usage:  "use TLS and verify the remote",
			EnvVar: "DOCKER_TLS_VERIFY",
		},
		cli.StringFlag{
			Name:  "tlscacert",
			Usage: "trust certs signed only by this CA",
			Value: fmt.Sprintf("%s/ca.pem", rootCertPath),
		},
		cli.StringFlag{
			Name:  "tlscert",
			Usage: "client certificate for TLS authentication",
			Value: fmt.Sprintf("%s/cert.pem", rootCertPath),
		},
		cli.StringFlag{
			Name:  "tlskey",
			Usage: "client key for TLS authentication",
			Value: fmt.Sprintf("%s/key.pem", rootCertPath),
		},
		cli.BoolFlag{
			Name:  "debug",
			Usage: "enable debug mode with verbose logging",
		},
		cli.StringFlag{
			Name:   "webhooks",
			Usage:  "list of urls sperated by ';'",
			Value:  "http://result-service:8081/events",
			EnvVar: "TUGBOT_WEBHOOKS",
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func before(c *cli.Context) error {
	if c.GlobalBool("debug") {
		log.SetLevel(log.DebugLevel)
	}

	// Set-up container client
	tls, err := tlsConfig(c)
	if err != nil {
		return err
	}

	client = container.NewClient(c.GlobalString("host"), tls, !c.GlobalBool("no-pull"))

	return nil
}

func start(c *cli.Context) {
	names = c.Args()
	startMonitorEvents(c)
	log.Info("Tugbot Started OK")
	waitForInterrupt()
}

func startMonitorEvents(c *cli.Context) {
	client.StartMonitorEvents(runTestContainers)
	webhooks := c.GlobalString("webhooks")
	if webhooks != "" {
		publisher = common.NewPublisher(strings.Split(webhooks, ";"))
		client.StartMonitorEvents(publishEvent)
	}
}

func runTestContainers(e *dockerclient.Event, ec chan error, args ...interface{}) {
	log.Infof("Looking for test containers that should run on event: %+v", e)
	wgr.Add(1)
	if err := actions.Run(client, names, e); err != nil {
		log.Error(err)
	}
	wgr.Done()
}

func publishEvent(e *dockerclient.Event, ec chan error, args ...interface{}) {
	//log.Debugf("Publishing event: %+v", e)
	log.Infof("Publishing event: %+v", e)
	wgp.Add(1)
	publisher.Publish(e)
	wgp.Done()
}

func waitForInterrupt() {
	// Graceful shut-down on SIGINT/SIGTERM/SIGQUIT
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT)
	<-c
	log.Debug("Stop monitoring events ...")
	wgr.Wait()
	client.StopAllMonitorEvents()
	log.Debug("Graceful exit :-)")
	os.Exit(1)
}

// tlsConfig translates the command-line options into a tls.Config struct
func tlsConfig(c *cli.Context) (*tls.Config, error) {
	var tlsConfig *tls.Config
	var err error
	caCertFlag := c.GlobalString("tlscacert")
	certFlag := c.GlobalString("tlscert")
	keyFlag := c.GlobalString("tlskey")

	if c.GlobalBool("tls") || c.GlobalBool("tlsverify") {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: !c.GlobalBool("tlsverify"),
		}

		// Load CA cert
		if caCertFlag != "" {
			var caCert []byte

			if strings.HasPrefix(caCertFlag, "/") {
				caCert, err = ioutil.ReadFile(caCertFlag)
				if err != nil {
					return nil, err
				}
			} else {
				caCert = []byte(caCertFlag)
			}

			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)

			tlsConfig.RootCAs = caCertPool
		}

		// Load client certificate
		if certFlag != "" && keyFlag != "" {
			var cert tls.Certificate

			if strings.HasPrefix(certFlag, "/") && strings.HasPrefix(keyFlag, "/") {
				cert, err = tls.LoadX509KeyPair(certFlag, keyFlag)
				if err != nil {
					return nil, err
				}
			} else {
				cert, err = tls.X509KeyPair([]byte(certFlag), []byte(keyFlag))
				if err != nil {
					return nil, err
				}
			}
			tlsConfig.Certificates = []tls.Certificate{cert}
		}
	}

	return tlsConfig, nil
}
