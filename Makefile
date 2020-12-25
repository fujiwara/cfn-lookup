.PHONY: test install

test:
	go test ./...

install:
	go install github.com/fujiwara/cfn-lookup/cmd/cfn-lookup
