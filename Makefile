VERSION := $(shell git describe --tags || echo "dev")
COMMIT  := $(shell git rev-parse --short HEAD)
DATE    := $(shell date +%F)

PATH_CMD := cmd/i3-snapshot
MODULE := $(shell go list -m)

LDFLAGS := -ldflags "-X $(MODULE)/$(PATH_CMD)/main.version=$(VERSION) -X  $(MODULE)/$(PATH_CMD)/main.commit=$(COMMIT) -X  $(MODULE)/$(PATH_CMD)/main.date=$(DATE)"

build:
	go build $(LDFLAGS) -o i3-snapshot $(PATH_CMD)/*.go

run:
	go run $(LDFLAGS) $(PATH_CMD)/*.go
