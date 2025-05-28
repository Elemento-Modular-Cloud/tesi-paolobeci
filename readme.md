# Appunti tesi magistrale Paolo Beci
In questa repository verrà sviluppato: "Elemento Cloud Client Libraries for Go"

## Link agli appunti su Notion
[Notion Site](https://glimmer-slip-7ec.notion.site/appunti-tesi-paolo?pvs=4)

## Test Elemento Cloud Go Client locally
```bash
# First import the env variables
export ELEMENTO_CLOUD_API_KEY=your_api_key
# ...

# Then run the tests
go test -v ./ecloud
```

## Test kOps locally
From inisde the kOps source directory, run the following command to start a Docker container with the Go environment set up:
```bash
docker run -it \
  -v "$(pwd)":/go/src/k8s.io/kops \
  -w /go/src/k8s.io/kops \
  golang \
  bash

make
```

Or just run it locally:
```bash
make

# for the debugging
DEBUGGING=true make
```