proto:
	protoc --gofast_out=plugins=grpc:. internal/protocol/sql.proto
