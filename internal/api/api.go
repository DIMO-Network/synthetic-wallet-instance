package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/DIMO-Network/synthetic-wallet-instance/pkg/grpc"
	awsconf "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/rs/zerolog"
	"golang.org/x/sys/unix"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	grpc.UnimplementedSyntheticWalletServer
	CID           uint32
	Port          uint32
	EncryptedSeed string
	Logger        *zerolog.Logger
}

type cred struct {
	AccessKeyID     string `json:"AccessKeyId"`
	SecretAccessKey string `json:"SecretAccessKey"`
	Token           string `json:"Token"`
}

type Request[A any] struct {
	Credentials   AWSCredentials `json:"credentials"`
	EncryptedSeed string         `json:"encryptedSeed"`
	Type          string         `json:"type"`
	Data          A              `json:"data"`
}

type AWSCredentials struct {
	AccessKeyID     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
	Token           string `json:"token"`
}

type AddrReqData struct {
	ChildNumber uint32 `json:"childNumber"`
}

type SignReqData struct {
	ChildNumber uint32      `json:"childNumber"`
	Hash        common.Hash `json:"hash"`
}

type AddrResData struct {
	Address common.Address `json:"address"`
}

type SignResData struct {
	Signature hexutil.Bytes `json:"signature"`
}

type ErrData struct {
	Message string `json:"message"`
}

type Response[A any] struct {
	Code int `json:"code"`
	Data A   `json:"data"`
}

const bufferSize = 4096

func (s Server) GetAddress(ctx context.Context, in *grpc.GetAddressRequest) (*grpc.GetAddressResponse, error) {
	if in.ChildNumber >= hdkeychain.HardenedKeyStart {
		return nil, status.Errorf(codes.InvalidArgument, "child_number %d >= 2^31", in.ChildNumber)
	}

	s.Logger.Info().Msgf("Got address request, child number %d.", in.ChildNumber)

	cfg, err := awsconf.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	md := imds.NewFromConfig(cfg)
	mo, err := md.GetMetadata(ctx, &imds.GetMetadataInput{Path: "iam/security-credentials/eks-quickstart-ManagedNodeInstance"})
	if err != nil {
		return nil, err
	}
	defer mo.Content.Close()

	s.Logger.Debug().Msg("Got EC2 metadata.")

	b, err := io.ReadAll(mo.Content)
	if err != nil {
		return nil, err
	}

	var c cred
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, err
	}

	fd, err := unix.Socket(unix.AF_VSOCK, unix.SOCK_STREAM, 0)
	if err != nil {
		return nil, err
	}

	s.Logger.Debug().Msgf("Created socket %d.", fd)

	sa := &unix.SockaddrVM{CID: s.CID, Port: s.Port}

	if err := unix.Connect(fd, sa); err != nil {
		return nil, err
	}

	s.Logger.Debug().Msgf("Connected socket to CID %d, port %d.", s.CID, s.Port)

	m := Request[AddrReqData]{
		Credentials:   AWSCredentials(c),
		EncryptedSeed: s.EncryptedSeed,
		Type:          "GetAddress",
		Data: AddrReqData{
			ChildNumber: in.ChildNumber,
		},
	}

	b, _ = json.Marshal(m)

	s.Logger.Debug().Msgf("Sending request: %q", string(b))

	if err := unix.Send(fd, b, 0); err != nil {
		return nil, err
	}

	s.Logger.Debug().Msg("Request sent.")

	buf := make([]byte, bufferSize)

	n, err := unix.Read(fd, buf)
	if err != nil {
		return nil, err
	}

	s.Logger.Debug().Msgf("Got response: %q", string(buf[:n]))

	var r Response[json.RawMessage]
	if err := json.Unmarshal(buf[:n], &r); err != nil {
		return nil, err
	}

	if r.Code != 0 {
		var e ErrData
		if err := json.Unmarshal(r.Data, &e); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("error from enclave: %s", e.Message)
	}

	var ad AddrResData
	if err := json.Unmarshal(r.Data, &ad); err != nil {
		return nil, err
	}

	return &grpc.GetAddressResponse{Address: ad.Address.Bytes()}, nil
}

func (s Server) SignHash(ctx context.Context, in *grpc.SignHashRequest) (*grpc.SignHashResponse, error) {
	if in.ChildNumber >= hdkeychain.HardenedKeyStart {
		return nil, status.Errorf(codes.InvalidArgument, "child_number %d >= 2^31", in.ChildNumber)
	}

	if len(in.Hash) != common.HashLength {
		return nil, status.Errorf(codes.InvalidArgument, "hash has length %d != %d", len(in.Hash), common.HashLength)
	}

	s.Logger.Info().Msgf("Got signature request, child number %d, hash %d.", in.ChildNumber, common.BytesToHash(in.Hash))

	cfg, err := awsconf.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	md := imds.NewFromConfig(cfg)
	mo, err := md.GetMetadata(ctx, &imds.GetMetadataInput{Path: "iam/security-credentials/eks-quickstart-ManagedNodeInstance"})
	if err != nil {
		return nil, err
	}
	defer mo.Content.Close()

	s.Logger.Debug().Msg("Got EC2 metadata.")

	b, err := io.ReadAll(mo.Content)
	if err != nil {
		return nil, err
	}

	var c cred
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, err
	}

	fd, err := unix.Socket(unix.AF_VSOCK, unix.SOCK_STREAM, 0)
	if err != nil {
		return nil, err
	}

	s.Logger.Debug().Msgf("Created socket %d.", fd)

	sa := &unix.SockaddrVM{CID: s.CID, Port: s.Port}

	if err := unix.Connect(fd, sa); err != nil {
		return nil, err
	}

	s.Logger.Debug().Msgf("Connected to CID %d, port %d.", s.CID, s.Port)

	m := Request[SignReqData]{
		Credentials:   AWSCredentials(c),
		EncryptedSeed: s.EncryptedSeed,
		Type:          "SignHash",
		Data: SignReqData{
			ChildNumber: in.ChildNumber,
			Hash:        common.BytesToHash(in.Hash),
		},
	}

	b, _ = json.Marshal(m)

	s.Logger.Debug().Msgf("Sending request %q.", string(b))

	if err := unix.Send(fd, b, 0); err != nil {
		return nil, err
	}

	s.Logger.Debug().Msg("Request sent.")

	buf := make([]byte, bufferSize)

	n, err := unix.Read(fd, buf)
	if err != nil {
		return nil, err
	}

	s.Logger.Debug().Msgf("Got response: %s", string(buf[:n]))

	var r Response[json.RawMessage]
	if err := json.Unmarshal(buf[:n], &r); err != nil {
		return nil, err
	}

	if r.Code != 0 {
		var e ErrData
		if err := json.Unmarshal(r.Data, &e); err != nil {
			return nil, err
		}
		return nil, fmt.Errorf("error from enclave: %s", e.Message)
	}

	var sr SignResData
	if err := json.Unmarshal(r.Data, &sr); err != nil {
		return nil, err
	}

	return &grpc.SignHashResponse{Signature: sr.Signature}, nil
}
