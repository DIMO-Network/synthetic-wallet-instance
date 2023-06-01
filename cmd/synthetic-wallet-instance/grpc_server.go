package main

import (
	"net"

	"github.com/DIMO-Network/synthetic-wallet-instance/internal/api"
	"github.com/DIMO-Network/synthetic-wallet-instance/internal/config"
	pb "github.com/DIMO-Network/synthetic-wallet-instance/pkg/grpc"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

func startGRPCServer(settings *config.Settings, logger *zerolog.Logger) {
	lis, err := net.Listen("tcp", ":"+settings.GRPCPort)
	if err != nil {
		logger.Fatal().Err(err).Msgf("Couldn't listen on gRPC port %s.", settings.GRPCPort)
	}

	logger.Info().Msgf("Starting gRPC server on port %s.", settings.GRPCPort)
	server := grpc.NewServer()
	wal := api.Server{
		CID:           uint32(settings.EnclaveCID),
		Port:          uint32(settings.EnclavePort),
		EncryptedSeed: settings.BIP32Seed,
		Logger:        logger,
	}

	pb.RegisterSyntheticWalletServer(server, wal)

	if err := server.Serve(lis); err != nil {
		logger.Fatal().Err(err).Msg("gRPC server terminated unexpectedly")
	}
}
