package main

import (
	"context"
	"flag"
	"math/rand"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rs/zerolog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/DIMO-Network/synthetic-wallet-instance/pkg/grpc"
)

var svc = flag.String("a", "localhos:9005", "wallet gRPC address")
var runs = flag.Int("i", 100, "number of iterations")

func main() {
	rand.Seed(time.Now().UnixNano())

	flag.Parse()

	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	logger.Info().Msgf("Hitting %s for %d iterations.", *svc, *runs)

	conn, err := grpc.Dial(*svc, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Fatal().Err(err).Msgf("Couldn't dial %q.", os.Args[1])
	}

	client := pb.NewSyntheticWalletClient(conn)

	for i := 0; i < *runs; i++ {
		cn := uint32(rand.Int31())
		r1, err := client.GetAddress(context.TODO(), &pb.GetAddressRequest{ChildNumber: cn})
		if err != nil {
			panic(err)
		}

		if len(r1.Address) != common.AddressLength {
			panic(err)
		}

		addr := common.BytesToAddress(r1.Address)

		hash := make([]byte, common.HashLength)

		if _, err := rand.Read(hash); err != nil {
			panic(err)
		}

		r2, err := client.SignHash(context.TODO(), &pb.SignHashRequest{
			ChildNumber: cn,
			Hash:        hash,
		})
		if err != nil {
			panic(err)
		}

		if len(r2.Signature) != 65 {
			panic(err)
		}

		if v := r2.Signature[64]; v != 27 && v != 28 {
			panic(err)
		}

		r2.Signature[64] -= 27

		pub, err := crypto.Ecrecover(hash, r2.Signature)
		if err != nil {
			panic(err)
		}

		pk, err := crypto.UnmarshalPubkey(pub)
		if err != nil {
			panic(err)
		}

		if crypto.PubkeyToAddress(*pk) != addr {
			logger.Fatal().Msgf("recovered = %s, desired = %s", crypto.PubkeyToAddress(*pk), addr)
		}
	}
}
