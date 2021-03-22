gosrc := $(shell find ./ -name '*.go' | grep -v vendor)
protobufRawSrc := $(shell find ./ -name '*.proto')
protobufSrc := $(shell find ./ -name '*.proto' -exec sh -c "echo {} | sed 's/.proto/.pb.go/g'" \;)
protobufSrcGRPC := $(shell find ./ -name '*.proto' -exec sh -c "echo {} | sed 's/.proto/_grpc.pb.go/g'" \;)
pythonProtobufSrc := $(shell find ./ -name '*.proto' -exec sh -c 'echo workloads/`basename {}` | sed "s/.proto/_pb2.py/g"' \;)
pythonProtobufSrcGRPC := $(shell find ./ -name '*.proto' -exec sh -c 'echo workloads/`basename {}` | sed "s/.proto/_pb2_grpc.py/g"' \;)

./bin/avis: $(gosrc) $(protobufSrcGRPC) $(protobufSrc) $(pythonProtobufSrc) $(pythonProtobufSrcGRPC)
	go build -o bin ./cmd/avis

clean:
	go clean -testcache
	rm -f ./bin/avis
	rm -f $(protobufSrc)
	rm -f ./workloads/*pb2*.py

.PHONY: test-unit test-functional
test-unit:
	go clean -testcache
	go test -v -run=Unit ./...

test-functional:
	go clean -testcache
	go test -v -run=Functional ./...

$(protobufSrcGRPC) $(protobufSrc): $(protobufRawSrc)
	go generate ./...

$(pythonProtobufSrc) $(pythonProtobufSrcGRPC): $(protobufRawSrc)
	# TODO -- clean this up
	pipenv run python -m grpc_tools.protoc -I=./controller/ --python_out=./workloads --grpc_python_out=./workloads ./controller/*.proto
