all: pmbench pmc
pmbench:
	go build ./cmd/$@
pmc:
	go build ./cmd/$@
test:
	go test ./...
.PHONY: pmbench pmc test
