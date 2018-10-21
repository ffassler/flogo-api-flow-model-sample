//go:generate go run $GOPATH/src/github.com/TIBCOSoftware/flogo-lib/flogo/gen/gen.go $GOPATH
package main

import (
	"encoding/json"
	"fmt"
	"github.com/TIBCOSoftware/flogo-lib/app"
	"github.com/TIBCOSoftware/flogo-lib/engine"
	"github.com/TIBCOSoftware/flogo-lib/logger"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/TIBCOSoftware/flogo-contrib/action/flow"
	_ "github.com/TIBCOSoftware/flogo-contrib/activity/log"
	_ "github.com/TIBCOSoftware/flogo-contrib/trigger/rest"
	_ "github.com/debovema/flogo-opentracing-listener"
)

var log = logger.GetLogger("main-engine")

const flogoJSON string = `{
  "name": "flogo-opentracing-sample",
  "type": "flogo:app",
  "version": "0.0.1",
  "appModel": "1.0.0",
  "triggers": [
    {
      "id": "receive_http_message",
      "ref": "github.com/TIBCOSoftware/flogo-contrib/trigger/rest",
      "name": "Receive HTTP Message",
      "description": "Simple REST Trigger",
      "settings": {
        "port": 9233
      },
      "handlers": [
        {
          "action": {
            "ref": "github.com/TIBCOSoftware/flogo-contrib/action/flow",
            "data": {
              "flowURI": "res://flow:sample_flow"
            }
          },
          "settings": {
            "method": "GET",
            "path": "/test"
          }
        }
      ]
    }
  ],
  "resources": [
    {
      "id": "flow:sample_flow",
      "data": {
        "name": "SampleFlow",
        "tasks": [
          {
            "id": "log_1",
            "name": "Log Message",
            "description": "Simple Log Activity",
            "activity": {
              "ref": "github.com/TIBCOSoftware/flogo-contrib/activity/log",
              "input": {
                "message": "first log",
                "flowInfo": "false",
                "addToFlow": "false"
              }
            }
          }
        ],
        "links": [
        ]
      }
    }
  ]
}`

func main() {

	config := &app.Config{}
	err := json.Unmarshal([]byte(flogoJSON), config)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)	}

	e, err := engine.New(config)
	if err != nil {
		log.Errorf("Failed to create engine instance due to error: %s", err.Error())
		os.Exit(1)
	}

	err = e.Start()
	if err != nil {
		log.Errorf("Failed to start engine due to error: %s", err.Error())
		os.Exit(1)
	}

	exitChan := setupSignalHandling()

	code := <-exitChan

	e.Stop()

	os.Exit(code)
}

func setupSignalHandling() chan int {

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	exitChan := make(chan int, 1)
	select {
	case s := <-signalChan:
		switch s {
		case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			exitChan <- 0
		default:
			logger.Debug("Unknown signal.")
			exitChan <- 1
		}
	}
	return exitChan
}
