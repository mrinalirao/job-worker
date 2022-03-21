api:
	go build -o ./job-worker cmd/main.go

test:
	go test ./... -v

protofile:
	protoc  --go_out=:. --go-grpc_out=:. --go_opt=paths=source_relative proto/workerservice.proto
