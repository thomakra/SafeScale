GO?=go

.PHONY: generate clean

all:    generate vet

vet:
	@$(GO) vet ./...

DIRECTORIES := $(sort $(dir $(wildcard */)))

generate:
	@$(GO) generate ./...

clean:
	@(for d in $(DIRECTORIES); do (cd $$d; $(RM) *_string.go || true); done)
