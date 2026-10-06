VERSION := $(shell git describe --tags || echo "dev")
COMMIT  := $(shell git rev-parse --short HEAD)
DATE    := $(shell date +%F)

PATH_CMD := ./cmd/i3-snapshot

LDFLAGS := -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(DATE)"

build:
	go build $(LDFLAGS) -o i3-snapshot $(PATH_CMD)

run:
	go run $(LDFLAGS) $(PATH_CMD)
