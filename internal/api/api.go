package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/DIMO-Network/test-instance/pkg/grpc"
	awsconf "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/btcsuite/btcutil/hdkeychain"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"golang.org/x/sys/unix"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	grpc.UnimplementedVirtualDeviceWalletServer
	CID           uint32
	Port          uint32
	EncryptedSeed string
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

type AddrReqData struct {
	ChildNumber uint32 `json:"childNumber"`
}

type SignReqData struct {
	ChildNumber uint32      `json:"childNumber"`
	Hash        common.Hash `json:"hash"`
}

type AWSCredentials struct {
	AccessKeyID     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
	Token           string `json:"token"`
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

	log.Printf("Got EC2 metadata.")

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

	sa := &unix.SockaddrVM{CID: s.CID, Port: s.Port}

	if err := unix.Connect(fd, sa); err != nil {
		return nil, err
	}

	log.Printf("Connected to socket.")

	m := Request[AddrReqData]{
		Credentials:   AWSCredentials(c),
		EncryptedSeed: s.EncryptedSeed,
		Type:          "GetAddress",
		Data: AddrReqData{
			ChildNumber: in.ChildNumber,
		},
	}

	b, _ = json.Marshal(m)

	if err := unix.Send(fd, b, 0); err != nil {
		return nil, err
	}

	log.Printf("Request sent.")

	buf := make([]byte, bufferSize)

	n, err := unix.Read(fd, buf)
	if err != nil {
		return nil, err
	}

	log.Printf("Got response: %s", string(buf[:n]))

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

	log.Printf("Got EC2 metadata.")

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

	sa := &unix.SockaddrVM{CID: s.CID, Port: s.Port}

	if err := unix.Connect(fd, sa); err != nil {
		return nil, err
	}

	log.Printf("Connected to socket.")

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

	if err := unix.Send(fd, b, 0); err != nil {
		return nil, err
	}

	log.Printf("Request sent.")

	buf := make([]byte, bufferSize)

	n, err := unix.Read(fd, buf)
	if err != nil {
		return nil, err
	}

	log.Printf("Got response: %s", string(buf[:n]))

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
