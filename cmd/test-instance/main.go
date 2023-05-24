package main

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"time"

	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/test-instance/internal/config"
	awsconf "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/rs/zerolog"
)

type cred struct {
	AccessKeyID     string    `json:"AccessKeyId"`
	SecretAccessKey string    `json:"SecretAccessKey"`
	Token           string    `json:"Token"`
	Expiration      time.Time `json:"Expiration"`
}

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Str("app", "test-instance").Logger()

	settings, err := shared.LoadConfig[config.Settings]("settings.yaml")
	if err != nil {
		logger.Fatal().Err(err).Msg("Could not load settings")
	}

	logger.Info().Msgf("Loaded settings: monitoring port %s.", settings.MonPort)

	serveMonitoring(settings.MonPort, &logger)

	ctx := context.Background()

	cfg, err := awsconf.LoadDefaultConfig(ctx)
	if err != nil {
		panic(err)
	}

	md := imds.NewFromConfig(cfg)
	mo, err := md.GetMetadata(ctx, &imds.GetMetadataInput{Path: "iam/security-credentials/eks-quickstart-ManagedNodeInstance"})
	if err != nil {
		panic(err)
	}
	defer mo.Content.Close()

	b, err := io.ReadAll(mo.Content)
	if err != nil {
		panic(err)
	}

	var c cred
	if err := json.Unmarshal(b, &c); err != nil {
		panic(err)
	}

	logger.Info().Time("expires", c.Expiration).Msg("Got credentials from metadata.")

	time.Sleep(1 * time.Hour)
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
