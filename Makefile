pmbench:
	go build ./cmd/$@
test:
	go test ./...
.PHONY: pmbench test
