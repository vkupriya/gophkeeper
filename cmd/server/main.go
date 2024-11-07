package main

import (
	"fmt"
	"log"

	"go.uber.org/zap"

	"github.com/vkupriya/gophkeeper/internal/server"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {
	logConfig := zap.NewDevelopmentConfig()
	logger, err := logConfig.Build()
	if err != nil {
		log.Panic(fmt.Errorf("failed to initialize Logger: %w", err))
	}

	logger.Sugar().Infof("Build version: %s", buildVersion)
	logger.Sugar().Infof("Build date: %s", buildDate)
	logger.Sugar().Infof("Build commit: %s", buildCommit)

	if err := server.Start(logger); err != nil {
		logger.Sugar().Errorf("server has been terminated with error: %w", err)
	}
}
