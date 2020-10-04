#!/usr/bin/env python3
import target
from sys import argv

RALS = {
    'ardupilot': target.ArduPilotRAL,
    'px4': target.PX4RAL,
}

def get_RAL() -> type:
    if len(argv) < 2:
        raise Exception('error: missing platform in program arguments')
    platform = argv[1].lower()
    if not platform in RALS:
        raise Exception(f'error: unknown platform {platform} in command arguments')
    return RALS[platform]
