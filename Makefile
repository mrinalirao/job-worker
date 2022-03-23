.PHONY: api
api:
	go build -o ./job-worker cmd/main.go

.PHONY: test
test:
	go test ./... -v

.PHONY: protofile
protofile:
	protoc  --go_out=:. --go-grpc_out=:. --go_opt=paths=source_relative proto/workerservice.proto
