package main

import (
	"fmt"
	"net"

	"github.com/DIMO-Network/synthetic-wallet-instance/internal/api"
	"github.com/DIMO-Network/synthetic-wallet-instance/internal/config"
	pb "github.com/DIMO-Network/synthetic-wallet-instance/pkg/grpc"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/ethereum/go-ethereum/common"
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

	var wal pb.SyntheticWalletServer

	if settings.MockEnclave {
		logger.Warn().Msg("Mocking out the enclave. Do not do this in production.")

		fmt.Println(settings.MockSeed)

		seed := common.FromHex(settings.MockSeed)

		if len(seed) != hdkeychain.RecommendedSeedLen {
			logger.Fatal().Msgf("Seed must be %d bytes.", hdkeychain.RecommendedSeedLen)
		}

		key, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
		if err != nil {
			logger.Fatal().Err(err).Msg("Couldn't load seed for mock enclave.")
		}

		wal = api.MockServer{
			Key:    key,
			Logger: logger,
		}
	} else {
		wal = api.Server{
			CID:           uint32(settings.EnclaveCID),
			Port:          uint32(settings.EnclavePort),
			EncryptedSeed: settings.BIP32Seed,
			Logger:        logger,
		}
	}

	pb.RegisterSyntheticWalletServer(server, wal)

	if err := server.Serve(lis); err != nil {
		logger.Fatal().Err(err).Msg("gRPC server terminated unexpectedly")
	}
}
