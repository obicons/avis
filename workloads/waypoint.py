#!/usr/bin/env python3
from pymavlink import mavutil
from target import Target
from util import *

class Waypoint(Target):
    def test(self):
        time = self.time()
        while time.tvSec < 45:
            self.step()
            time = self.time()
        self.upload_mission(
            self.takeoff_mission_items(
                20,
                -35.363261,
                149.165230,
                20
            ) +
            [
                {
                    'frame': mavutil.mavlink.MAV_FRAME_GLOBAL_RELATIVE_ALT,
                    'command': mavutil.mavlink.MAV_CMD_NAV_WAYPOINT,
                    'param1': 0,
                    'param2': 0,
                    'param3': 0,
                    'param4': 0,
                    'x': -35.362149,
                    'y': 149.165056,
                    'z': 20,
                    'mission_type': mavutil.mavlink.MAV_MISSION_TYPE_MISSION
                },
                {
                    'frame': mavutil.mavlink.MAV_FRAME_MISSION,
                    'command': mavutil.mavlink.MAV_CMD_NAV_RETURN_TO_LAUNCH,
                    'param1': 0,
                    'param2': 0,
                    'param3': 0,
                    'param4': 0,
                    'x': 0,
                    'y': 0,
                    'z': 0,
                    'mission_type': mavutil.mavlink.MAV_MISSION_TYPE_MISSION
                }
            ]
        )
        self.arm_system()
        self.enter_auto_mode()
        while self.time().tvSec < 100: self.step()
        while abs(self.position().z) > 2:
            self.step()
        self.pass_test()

if __name__ == '__main__':
    t = Waypoint(
        'udp:127.0.0.1:14550',
        get_rpc_addr(),
        get_RAL()
    )
    t.test()
