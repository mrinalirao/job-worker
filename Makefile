.PHONY: api
api:
	@-$(MAKE) protofile
	go build -o ./job-worker cmd/api/main.go

.PHONY: client
client:
	@-$(MAKE) protofile
	go build -o ./client cmd/client/main.go

.PHONY: test
test:
	@-$(MAKE) protofile
	go test ./... -v

.PHONY: protofile
protofile:
	protoc  --go_out=:. --go-grpc_out=:. --go_opt=paths=source_relative proto/workerservice.proto
