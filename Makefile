.PHONY: pb

pb:
	protoc --go_out=. --go-grpc_out=. ./api/*.proto && \
	protoc --go_out=. --go_opt=paths=source_relative ./internal/client/models/*.proto