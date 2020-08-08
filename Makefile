./bin/rmck:
	go build -o bin ./cmd/rmck 

clean:
	rm -f ./bin/rmck
