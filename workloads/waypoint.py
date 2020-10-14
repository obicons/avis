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
        print('Uploading mission')
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
                    'x': -35.363012,
                    'y': 149.165209,
                    'z': 20,
                    'mission_type': mavutil.mavlink.MAV_MISSION_TYPE_MISSION
                },
                {
                    'frame': mavutil.mavlink.MAV_FRAME_GLOBAL_RELATIVE_ALT,
                    'command': mavutil.mavlink.MAV_CMD_NAV_WAYPOINT,
                    'param1': 0,
                    'param2': 0,
                    'param3': 0,
                    'param4': 0,
                    'x': -35.363261,
                    'y': 149.165230,
                    'z': 20,
                    'mission_type': mavutil.mavlink.MAV_MISSION_TYPE_MISSION
                },
                {
                    'frame': mavutil.mavlink.MAV_FRAME_GLOBAL_RELATIVE_ALT,
                    'command': mavutil.mavlink.MAV_CMD_NAV_WAYPOINT,
                    'param1': 0,
                    'param2': 0,
                    'param3': 0,
                    'param4': 0,
                    'x': -35.363253,
                    'y': 149.165328,
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
        print('arming system', flush=True)
        self.arm_system()
        print('entering auto mode', flush=True)
        self.enter_auto_mode()
        print('stepping until 140', flush=True)

        time = self.time()
        while self.time().tvSec < 59:
            self.step()
            time = self.time()

        print('waiting on altitude', flush=True)
        while abs(self.position().z) > 2:
            self.step()
        self.pass_test()

if __name__ == '__main__':
    t = Waypoint(
        'udp:127.0.0.1:14550',
        'unix:///Users/madmax/.rmck_rpc',
        get_RAL()
    )
    t.test()
