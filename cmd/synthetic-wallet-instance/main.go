package main

import (
	"os"

	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/synthetic-wallet-instance/internal/config"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", "synthetic-wallet-instance").Logger()

	settings, err := shared.LoadConfig[config.Settings]("settings.yaml")
	if err != nil {
		logger.Fatal().Err(err).Msg("Could not load settings")
	}

	if ll, err := zerolog.ParseLevel(settings.LogLevel); err != nil {
		logger.Fatal().Err(err).Msgf("Couldn't parse log level setting %q.", settings.LogLevel)
	} else {
		logger = logger.Level(ll)
	}

	logger.Info().Msgf("Loaded settings: CID %s, port %s.", settings.EnclaveCID, settings.EnclavePort)

	serveMonitoring(settings.MonPort, &logger)

	startGRPCServer(&settings, &logger)
}

func serveMonitoring(port string, logger *zerolog.Logger) *fiber.App {
	monApp := fiber.New(fiber.Config{DisableStartupMessage: true})

	// Health check.
	monApp.Get("/", func(c *fiber.Ctx) error { return nil })
	monApp.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	go func() {
		if err := monApp.Listen(":" + port); err != nil {
			logger.Fatal().Err(err).Str("port", port).Msg("Failed to start monitoring web server.")
		}
	}()

	logger.Info().Msgf("Started monitoring web server on port %s.", port)

	return monApp
}
