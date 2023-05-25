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
	"github.com/ethereum/go-ethereum/common"
	"golang.org/x/sys/unix"
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

type Request struct {
	Credentials   AWSCredentials `json:"credentials"`
	EncryptedSeed string         `json:"encryptedSeed"`
	ChildNumber   uint32         `json:"childNumber"`
}

type AWSCredentials struct {
	AccessKeyID     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
	Token           string `json:"token"`
}

type AddrData struct {
	Address common.Address `json:"address"`
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
	log.Printf("Child request: %d, CID: %d, Port: %d, Encrypted: %s", in.ChildNumber, s.CID, s.Port, s.EncryptedSeed)
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

	m := Request{
		Credentials:   AWSCredentials(c),
		EncryptedSeed: s.EncryptedSeed,
		ChildNumber:   in.ChildNumber,
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

	var ad AddrData
	if err := json.Unmarshal(r.Data, &ad); err != nil {
		return nil, err
	}

	return &grpc.GetAddressResponse{Address: ad.Address.Bytes()}, nil
}
