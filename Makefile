.PHONY: api
api:
	go build -o ./job-worker cmd/main.go

.PHONY: client
client:
	go build -o ./client cli/client/userclient.go

.PHONY: adminclient
adminclient:
	go build -o ./adminclient cli/adminclient/adminclient.go

.PHONY: test
test:
	go test ./... -v

.PHONY: protofile
protofile:
	protoc  --go_out=:. --go-grpc_out=:. --go_opt=paths=source_relative proto/workerservice.proto
