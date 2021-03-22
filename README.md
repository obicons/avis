# Avis
Avis is the aerial vehicle in situ model checker.

## Building
Just run `make`.

## Running
Run `bin/rmck`. The following environment variables must be set:
- `ARDUPILOT_SRC_PATH` - the path to the source code of ArduPilot
- `ARDUPILOT_GZ_PATH` - the path to the source code of ArduPilot's Gazebo plugin
- `PX4_PATH` - the path to the source code of PX4

Optionally, `RMCK_DEBUG` can be set to any value to enable verbose output.

## Testing

### Unit Tests
Run `make test-unit`.

### Functional Tests
Run `make test-functional`.

### GRPC 
GRPC is used for the workload to interact with RMCK. To obtain the needed programs, run:
```
$ go get -u google.golang.org/protobuf/cmd/protoc-gen-go
$ go install google.golang.org/protobuf/cmd/protoc-gen-go
$ go get google.golang.org/grpc/cmd/protoc-gen-go-grpc
$ go install google.golang.org/grpc/cmd/protoc-gen-go-grpc
```

## TODO
- Metrics server for interesting stats (e.g. # of synthetic readings, etc)
