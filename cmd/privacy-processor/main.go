package main

import (
	"context"
	"os"
	"strings"

	"github.com/DIMO-Network/privacy-processor/internal/config"
	"github.com/DIMO-Network/privacy-processor/internal/processors"
	"github.com/Jeffail/benthos/v3/lib/util/hash/murmur2"
	"github.com/Shopify/sarama"

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

	gokaConfig := goka.DefaultConfig()
	gokaConfig.Version = sarama.V2_8_1_0
	gokaConfig.Producer.Partitioner = sarama.NewCustomPartitioner(
		sarama.WithAbsFirst(),
		sarama.WithCustomHashFunction(murmur2.New32),
	)

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

	if err := p.Run(context.Background()); err != nil {
		log.Fatal().Err(err).Msg("Failed to start privacy processor")
	}
}
