all: pmbench pmc pmd
pmbench:
	go build ./cmd/$@
pmc:
	go build ./cmd/$@
pmd:
	go build ./cmd/$@
test:
	go test ./...
.PHONY: pmbench pmc pmd test
