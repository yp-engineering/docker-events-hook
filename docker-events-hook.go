package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"

	"github.com/fsouza/go-dockerclient"
	"github.com/natefinch/pie"
	"gopkg.in/yaml.v2"
)

type DockerConfig struct {
	Endpoint string
}

type Config struct {
	Plugins []string
	Docker  DockerConfig
}

var defaultConfig = `
plugins:
  - echo
docker:
  endpoint: unix:///var/run/docker.sock
`

var (
	pluginPath = "/etc/docker-events-hook/plugins"
	configFile = ""
)

func init() {
	configFileMessage := "override default config path of " + configFile
	flag.StringVar(&configFile, "config", configFile, configFileMessage)

	pluginPathMessage := "override default plugin path of " + pluginPath
	flag.StringVar(&pluginPath, "plugin-path", pluginPath, pluginPathMessage)

	flag.Parse()
}

func assert(err error) {
	if err != nil {
		log.Fatal("fatal: ", err)
	}
}

func parseConfig() {
	var toLoad []byte
	if configFile != "" {
		var err error
		toLoad, err = ioutil.ReadFile(configFile)
		assert(err)
	} else {
		toLoad = []byte(defaultConfig)
	}

	config := Config{}
	assert(yaml.Unmarshal(toLoad, &config))

	log.Printf("--- config:\n%v\n\n", config)
	log.Printf("Plugins: %v", config.Plugins)
	log.Printf("Docker Endpoint: %v", config.Docker.Endpoint)
}

func newDockerClient() *docker.Client {
	client, err := docker.NewClient("unix:///var/run/docker.sock")
	if nil != err {
		log.Fatal(err)
		os.Exit(1)
	}
	return client
}

// Decision switch for what type of events we receive from daemon
func handleEvent(event *docker.APIEvents) {
	switch event.Status {
	case "start":
	}
}

func main() {
	parseConfig()
	return
	dockerClient := newDockerClient()
	pie.NewProvider()

	eventChannel := make(chan *docker.APIEvents)
	dockerClient.AddEventListener(eventChannel)

	// Channel blocks until input.
	for {
		handleEvent(<-eventChannel)
	}
}
