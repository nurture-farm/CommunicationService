/*
 *  Copyright 2023 NURTURE AGTECH PVT LTD
 *
 *    Licensed under the Apache License, Version 2.0 (the "License");
 *    you may not use this file except in compliance with the License.
 *    You may obtain a copy of the License at
 *
 *        http://www.apache.org/licenses/LICENSE-2.0
 *
 *    Unless required by applicable law or agreed to in writing, software
 *    distributed under the License is distributed on an "AS IS" BASIS,
 *    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *    See the License for the specific language governing permissions and
 *    limitations under the License.
 */

package main

import (
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/nurture-farm/CommunicationService/kafka"
	log "github.com/nurture-farm/CommunicationService/log"
	"github.com/nurture-farm/CommunicationService/server"
	commEngine "github.com/nurture-farm/Contracts/CommunicationEngine/Gen/GoCommunicationEngine"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	zap "go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
	"net/http"
	"os"
)

var logger = log.GetLogger()

func getWorkflowClient() client.Client {
	logger.Info("Temporal Worker config", zap.String("temporal.worker.namespace", viper.GetString("temporal.worker.namespace")),
		zap.String("temporal.worker.address", viper.GetString("temporal.worker.address")))

	c, err := client.NewClient(client.Options{
		Namespace: viper.GetString("temporal.worker.namespace"),
		HostPort:  viper.GetString("temporal.worker.address"),
	})
	if err != nil {
		logger.Fatal("Unable to create temporal workflow client", zap.Error(err))
	}
	return c
}

func runMonitoring(grpcServer *grpc.Server) {
	// register prometheus
	grpc_prometheus.Register(grpcServer)
	http.Handle("/metrics", promhttp.Handler())

	port := viper.GetInt("server.prometheus.port")
	err := http.ListenAndServe(fmt.Sprintf("%s:%d", "0.0.0.0", port), nil)
	if err != nil {
		logger.Error("Unable to attach listener to service", zap.Int("port", port), zap.Error(err))
	}
}

func registerAsWorker() client.Client {
	workflowClient := getWorkflowClient()
	w := worker.New(workflowClient, "CEWorker", worker.Options{})
	w.RegisterActivity(server.CommEngine)

	logger.Info("Starting CEWorker", zap.Any("worker", w))
	workerErr := w.Run(worker.InterruptCh())
	if workerErr != nil {
		logger.Panic("Unable to start activity worker", zap.Error(workerErr))
	}

	return workflowClient
}

func init() {
	configDir := os.Getenv("CONFIG_DIR")
	if configDir == "" {
		configDir = "config"
	}
	viper.AddConfigPath(configDir)
	viper.SetConfigType("json")
	viper.SetConfigFile(configDir + "/communication_service.json")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		logger.Panic("Error reading config file", zap.Error(err))
	}
	kafka.InitProducer()
}

func main() {
	logger.Info("Starting communication Engine")

	port := viper.GetInt("server.port")
	lis, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", port))
	if err != nil {
		logger.Fatal("Unable to listen on port", zap.Int("port", port), zap.Error(err))
	}

	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor))
	commEngine.RegisterCommunicationEngineServer(grpcServer, server.CommEngine)

	logger.Info("Registered server",
		zap.Any("grpcServer", grpcServer), zap.Any("listener", lis), zap.Int("port", port))

	// on GRPC services
	go runMonitoring(grpcServer)

	// register worker
	go func() {
		c := registerAsWorker()
		defer c.Close()
	}()

	// Start server
	err = grpcServer.Serve(lis)
	if err != nil {
		logger.Fatal("Unable to attach listener to service", zap.Int("port", port), zap.Error(err))
	}

}
