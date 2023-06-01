# DIMO Wallet Key Signing

## Table of contents

- [Developing locally](#developing-locally)
  - [Authentication](#authenticating)
  - [Linting](#linting)
- [Mocks](#mocks)
- [API](#api)
  - [Generating Swagger / OpenAPI spec](#generating-swagger--openapi-spec)
- [gRPC](#gRPC)

## Developing locally

**TL;DR**

```bash
cp settings.sample.yaml settings.yaml
go run ./cmd/synthetic-wallet-instance
```

### To run the project

`go run ./cmd/synthetic-wallet-instance`

1. Create a settings file by copying the sample

   ```sh
   cp settings.sample.yaml settings.yaml
   ```

   Adjust these as necessary. The sample file should have what you need for local development. (Make sure you do this step each time you run `git pull` in case there have been any changes to the sample settings file.)

2. You are now ready to run the application:
   ````sh
   go run ./cmd/synthetic-wallet-instance
    ```
   > If you get a port conflict, you can find the existing process using the port with, e.g., `lsof -i :<Your_Port_Number_Here>` or simply kill whatever is on the port with `npx kill-port --port <Port_Number>`.
   ````

### Authenticating

One of the variables set in `settings.yaml` is `JWT_KEY_SET_URL`. By default this is set to `http://127.0.0.1:5556/dex/keys`. To make use of this, clone the DIMO Dex fork:

```sh
git clone git@github.com:DIMO-Network/dex.git
cd dex
make build examples
./bin/dex serve examples/config-dev.yaml
```

This will start up the Dex identity server on port 5556. Next, start up the example interface by running

```sh
./bin/example-app
```

You can reach this on port 5555. The "Log in with Example" option is probably the easiest. This will give you an ID token you can provide to the [API](#api).

### Linting

`brew install golangci-lint`

`golangci-lint run`

This should use the settings from `.golangci.yml`, which you can override.

## Mocks

To regenerate a mock, you can use go gen since the files that are mocked have a `//go:generate mockgen ...` at the top. For example:
`nhtsa_api_service.go`

## API

## Protocol buffers

If you make changes to any of the `.proto` files in `pkg/grpc`, regenerate the Go code by running `make gen-proto`.
