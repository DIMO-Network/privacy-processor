package main

import (
	"context"
	"os"
	"strings"

	"github.com/DIMO-Network/privacy-processor/internal/config"
	"github.com/DIMO-Network/privacy-processor/internal/processors"
	"github.com/lovoo/goka"
	"github.com/rs/zerolog"
)

func main() {
	log := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("app", "privacy-processor").
		Logger()

	settings, err := config.LoadConfig("settings.yaml")
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config file")
	}

	fg := processors.Privacy{
		Group:        goka.Group(settings.PrivacyProcessorConsumerGroup),
		StatusInput:  goka.Stream(settings.DeviceStatusTopic),
		FenceTable:   goka.Table(settings.PrivacyFenceTopic),
		StatusOutput: goka.Stream(settings.DeviceStatusPrivateTopic),
		Logger:       &log,
	}

	fgg := fg.Define()

	p, err := goka.NewProcessor(strings.Split(settings.KafkaBrokers, ","), fgg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create privacy processor")
	}
	err = p.Run(context.Background())
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to start privacy processor")
	}
}
