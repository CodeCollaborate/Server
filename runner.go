package main

import (
	"flag"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"

	"github.com/CodeCollaborate/Server/modules/config"
	"github.com/CodeCollaborate/Server/modules/dbfs"
	"github.com/CodeCollaborate/Server/modules/handlers"
	"github.com/CodeCollaborate/Server/modules/rabbitmq"
	"github.com/CodeCollaborate/Server/utils"
)

/**
 * Runner.go starts the server. It initializes processes and begins listening for websocket requests.
 */

var logDir = flag.String("log_dir", "./data/logs/", "log file location")

func main() {
	flag.Parse()

	config.EnableLoggingToFile(*logDir)
	err := config.LoadConfig()
	if err != nil {
		utils.LogFatal("Failed to load configuration", err, nil)
	}
	cfg := config.GetConfig()

	// Get working directory
	dir, err := os.Getwd()
	utils.LogFatal("Could not get working directory", err, nil)

	utils.LogInfo("Working directory initalized", utils.LogFields{
		"Working Directory": dir,
	})

	// Creates a NewControl block for multithreading control
	AMQPControl := utils.NewControl(1)

	// RabbitMQ uses "Exchanges" as containers for Queues, and ours is initialized here.
	rabbitmq.SetupRabbitExchange(
		&rabbitmq.AMQPConnCfg{
			ConnCfg: cfg.ConnectionConfig["RabbitMQ"],
			Exchanges: []rabbitmq.AMQPExchCfg{
				{
					ExchangeName: cfg.ServerConfig.Name,
					Durable:      true,
				},
			},
			Control: AMQPControl,
		},
	)

	dbfsImpl := new(dbfs.DatabaseImpl)
	handlers.StartWorker(dbfsImpl)

	// FIXME: separate logging and ProjectFiles locations
	// FIXME: point fs at shared directory

	// FIXME: check config to see if we even need to serve http
	http.HandleFunc("/ws/", handlers.NewWSConn)

	addr := fmt.Sprintf("0.0.0.0:%d", cfg.ServerConfig.Port)

	_, certErr := os.Stat("config/TLS/cert.pem")
	_, keyErr := os.Stat("config/TLS/key.pem")

	useTLS := certErr == nil && keyErr == nil
	utils.LogInfo("Starting server", utils.LogFields{
		"Address": addr,
		"TLS":     useTLS,
	})

	go func() {
		addr := fmt.Sprintf("0.0.0.0:%d", cfg.ServerConfig.Port+1)
		err := http.ListenAndServe(addr, nil)
		if err != nil {
			utils.LogError("Failed to start pprof", err, utils.LogFields{
				"Address": addr,
			})
		}
	}()

	if useTLS {
		err = http.ListenAndServeTLS(addr, "config/TLS/cert.pem", "config/TLS/key.pem", nil)
	} else {
		utils.LogWarn("No Cert/Key pair found; starting without TLS", nil)
		err = http.ListenAndServe(addr, nil)
	}

	utils.LogError("Could not bind to port", err, nil)

	// Kill the SetupRabbitExchange thread (Multithreading control)
	defer func() {
		AMQPControl.Exit <- true
	}()
}
