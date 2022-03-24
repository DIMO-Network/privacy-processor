package main

import (
	"context"
	"os"
	"strings"

	"github.com/DIMO-Network/privacy-processor/internal/config"
	"github.com/DIMO-Network/privacy-processor/internal/processors"
	"github.com/Shopify/sarama"
	"github.com/burdiyan/kafkautil"

	"github.com/gofiber/fiber/v2"
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

	goka.ReplaceGlobalConfig(gokaConfig)

	fg := processors.Privacy{
		Group:        goka.Group(settings.PrivacyProcessorConsumerGroup),
		StatusInput:  goka.Stream(settings.DeviceStatusTopic),
		FenceTable:   goka.Table(settings.PrivacyFenceTopic),
		StatusOutput: goka.Stream(settings.DeviceStatusPrivateTopic),
		Logger:       &log,
	}

	fgg := fg.Define()

	p, err := goka.NewProcessor(strings.Split(settings.KafkaBrokers, ","), fgg, goka.WithHasher(kafkautil.MurmurHasher))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create privacy processor")
	}

	web := fiber.New(fiber.Config{DisableStartupMessage: true})
	web.Get("/", func(ctx *fiber.Ctx) error {
		return nil
	})

	go func() {
		log.Info().Msg("Listening for health check on port 4195")
		if err := web.Listen(":4195"); err != nil {
			log.Fatal().Err(err).Msg("Failed to start web server")
		}
	}()

	log.Info().Msg("Starting privacy processor")
	log.Info().Msgf("Input topic %s, joining with table %s", settings.DeviceStatusTopic, settings.PrivacyFenceTopic)
	log.Info().Msgf("Output topic %s", settings.DeviceStatusPrivateTopic)

	if err := p.Run(context.Background()); err != nil {
		log.Fatal().Err(err).Msg("Failed to start privacy processor")
	}
}
