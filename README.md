# RMCK

RMCK is the robotic model checker.

## Building
Just run `make`.

## Running
Run `bin/rmck`. The following environment variables must be set:
- `ARDUPILOT_SRC_PATH` - the path to the source code of ArduPilot
- `ARDUPILOT_GZ_PATH` - the path to the source code of ArduPilot's Gazebo plugin

Optionally, `DEBUG` can be set to any value to enable verbose output.

## TODO
- Provisioning instances of PX4, ArduPilot (see `platforms/`)
- Implementing API gateway (e.g. over unix sockets)
- Porting Python drivers to use this new API gateway
- Reimplementing the core checker
