protoc --go_out=pkg/plugin/proto/src/golang --go_opt=paths=source_relative \
    --go-grpc_out=pkg/plugin/proto/src/golang --go-grpc_opt=paths=source_relative \
    pkg/plugin/proto/*.proto
mv pkg/plugin/proto/src/golang/pkg/plugin/proto/* pkg/plugin/proto/src/golang/
rm -rf pkg/plugin/proto/src/golang/pkg