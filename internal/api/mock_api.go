package api

import (
	"context"

	"github.com/DIMO-Network/synthetic-wallet-instance/pkg/grpc"
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/rs/zerolog"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MockServer struct {
	grpc.UnimplementedSyntheticWalletServer
	Key    *hdkeychain.ExtendedKey
	Logger *zerolog.Logger
}

func (s MockServer) GetAddress(ctx context.Context, in *grpc.GetAddressRequest) (*grpc.GetAddressResponse, error) {
	if in.ChildNumber >= hdkeychain.HardenedKeyStart {
		return nil, status.Errorf(codes.InvalidArgument, "child_number %d >= 2^31", in.ChildNumber)
	}
	s.Logger.Info().Msgf("Got address request for child %d.", in.ChildNumber)

	ck, err := s.Key.Derive(hdkeychain.HardenedKeyStart + in.ChildNumber)
	if err != nil {
		return nil, err
	}

	akh, err := ck.Address(&chaincfg.MainNetParams)
	if err != nil {
		return nil, err
	}

	return &grpc.GetAddressResponse{Address: akh.ScriptAddress()}, nil
}

func (s MockServer) SignHash(ctx context.Context, in *grpc.SignHashRequest) (*grpc.SignHashResponse, error) {
	if in.ChildNumber >= hdkeychain.HardenedKeyStart {
		return nil, status.Errorf(codes.InvalidArgument, "child_number %d >= 2^31", in.ChildNumber)
	}

	if len(in.Hash) != common.HashLength {
		return nil, status.Errorf(codes.InvalidArgument, "hash has length %d != %d", len(in.Hash), common.HashLength)
	}

	s.Logger.Info().Msgf("Got signature request for child %d.", in.ChildNumber)

	ck, err := s.Key.Derive(hdkeychain.HardenedKeyStart + in.ChildNumber)
	if err != nil {
		return nil, err
	}

	pk, err := ck.ECPrivKey()
	if err != nil {
		return nil, err
	}

	sig, err := crypto.Sign(in.Hash, pk.ToECDSA())
	if err != nil {
		return nil, err
	}

	sig[64] += 27

	return &grpc.SignHashResponse{Signature: sig}, nil
}
