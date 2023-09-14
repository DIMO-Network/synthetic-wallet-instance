package main

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/DIMO-Network/synthetic-wallet-instance/internal/api"
	"github.com/DIMO-Network/synthetic-wallet-instance/internal/config"
	pb "github.com/DIMO-Network/synthetic-wallet-instance/pkg/grpc"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/ethereum/go-ethereum/common"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
)

func InterceptorLogger(l zerolog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l := l.With().Fields(fields).Logger()

		switch lvl {
		case logging.LevelDebug:
			l.Debug().Msg(msg)
		case logging.LevelInfo:
			l.Info().Msg(msg)
		case logging.LevelWarn:
			l.Warn().Msg(msg)
		case logging.LevelError:
			l.Error().Msg(msg)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}

func startGRPCServer(settings *config.Settings, logger *zerolog.Logger) {
	opts := []logging.Option{
		logging.WithLogOnEvents(logging.StartCall, logging.FinishCall),
		// Add any other option (check functions starting with logging.With).
	}

	lis, err := net.Listen("tcp", ":"+settings.GRPCPort)
	if err != nil {
		logger.Fatal().Err(err).Msgf("Couldn't listen on gRPC port %s.", settings.GRPCPort)
	}

	logger.Info().Msgf("Starting gRPC server on port %s.", settings.GRPCPort)
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logging.UnaryServerInterceptor(InterceptorLogger(*logger), opts...),
		),
	)

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

		cid, _ := strconv.Atoi(settings.EnclaveCID)
		encPort, _ := strconv.Atoi(settings.EnclavePort)
		wal = api.Server{
			CID:           uint32(cid),
			Port:          uint32(encPort),
			EncryptedSeed: settings.BIP32Seed,
			Logger:        logger,
		}
	}

	pb.RegisterSyntheticWalletServer(server, wal)

	if err := server.Serve(lis); err != nil {
		logger.Fatal().Err(err).Msg("gRPC server terminated unexpectedly")
	}
}
