gosrc := $(shell find ./ -name '*.go' | grep -v vendor)

./bin/rmck: $(gosrc)
	go build -o bin ./cmd/rmck 

clean:
	go clean -testcache
	rm -f ./bin/rmck

.PHONY: test-unit test-functional
test-unit:
	go clean -testcache
	go test -v -run=Unit ./...

test-functional:
	go clean -testcache
	go test -v -run=Functional ./...
