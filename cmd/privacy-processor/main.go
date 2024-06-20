package main

import (
	"context"
	"os"
	"strings"

	"github.com/DIMO-Network/privacy-processor/internal/config"
	"github.com/DIMO-Network/privacy-processor/internal/processors"
	"github.com/DIMO-Network/shared"
	"github.com/IBM/sarama"
	"github.com/burdiyan/kafkautil"
	"github.com/gofiber/fiber/v2"
	"github.com/lovoo/goka"
	"github.com/rs/zerolog"
)

func serveMonitoring(port string, logger *zerolog.Logger) {
	logger.Info().Msg("Listening for health check on port " + port)

	web := fiber.New(fiber.Config{DisableStartupMessage: true})
	web.Get("/", func(_ *fiber.Ctx) error {
		return nil
	})

	if err := web.Listen(":" + port); err != nil {
		logger.Fatal().Err(err).Msg("Failed to start monitoring server on port " + port)
	}
}

func main() {
	logger := zerolog.New(os.Stdout).With().
		Timestamp().
		Str("app", "privacy-processor").
		Logger()

	settings, err := shared.LoadConfig[config.Settings]("settings.yaml")
	if err != nil {
		logger.Fatal().Err(err).Msg("could not load settings")
	}

	logLevel, err := zerolog.ParseLevel(settings.LogLevel)
	if err != nil {
		logger.Fatal().Err(err).Msgf("Couldn't parse log level %q, terminating", settings.LogLevel)
	}
	zerolog.SetGlobalLevel(logLevel)

	go serveMonitoring(settings.Port, &logger)

	gokaConfig := goka.DefaultConfig()
	gokaConfig.Version = sarama.V2_8_1_0

	goka.ReplaceGlobalConfig(gokaConfig)

	fg := processors.Privacy{
		Group:        goka.Group(settings.PrivacyProcessorConsumerGroup),
		StatusInput:  goka.Stream(settings.DeviceStatusTopic),
		FenceTable:   goka.Table(settings.PrivacyFenceTopic),
		StatusOutput: goka.Stream(settings.DeviceStatusPrivateTopic),
		Logger:       &logger,
	}

	fgg := fg.Define()

	p, err := goka.NewProcessor(strings.Split(settings.KafkaBrokers, ","), fgg, goka.WithHasher(kafkautil.MurmurHasher))
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create privacy processor")
	}

	web := fiber.New(fiber.Config{DisableStartupMessage: true})
	web.Get("/", func(_ *fiber.Ctx) error {
		return nil
	})

	logger.Info().Msg("Starting privacy processor")
	logger.Info().Msgf("Input topic %s, joining with table %s", settings.DeviceStatusTopic, settings.PrivacyFenceTopic)
	logger.Info().Msgf("Output topic %s", settings.DeviceStatusPrivateTopic)

	go func() {
		if err := p.Run(context.Background()); err != nil {
			logger.Fatal().Err(err).Msg("Failed to start privacy processor")
		}
	}()

	// V2
	fgV2 := processors.PrivacyV2{
		Group:        goka.Group(settings.PrivacyProcessorConsumerGroupV2),
		StatusInput:  goka.Stream(settings.DeviceStatusTopicV2),
		FenceTable:   goka.Table(settings.PrivacyFenceTopicV2),
		StatusOutput: goka.Stream(settings.DeviceStatusPrivateTopicV2),
		Logger:       &logger,
	}

	fggV2 := fgV2.DefineV2()
	pV2, err := goka.NewProcessor(strings.Split(settings.KafkaBrokers, ","), fggV2, goka.WithHasher(kafkautil.MurmurHasher))
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to create privacy processor")
	}

	logger.Info().Msg("Starting privacy processor V2")
	logger.Info().Msgf("Input topic %s, joining with table %s", settings.DeviceStatusTopicV2, settings.PrivacyFenceTopicV2)
	logger.Info().Msgf("Output topic %s", settings.DeviceStatusPrivateTopicV2)

	if err := pV2.Run(context.Background()); err != nil {
		logger.Fatal().Err(err).Msg("Failed to start privacy processor V2")
	}
}
