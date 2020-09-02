# RMCK

RMCK is the robotic model checker.

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

## TODO
- HINJ support:
  - Read message types
  - Start hardware server
- Implementing API gateway (e.g. over unix sockets)
- Porting Python drivers to use this new API gateway
- Reimplementing the core checker
