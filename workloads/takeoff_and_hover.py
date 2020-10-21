#!/usr/bin/env python3
from target import Target
from util import *

class TakeoffAndHover(Target):
    def test(self):
        time = self.time()
        print('running for 40 seconds', flush=True)
        while time.tvSec < 45:
            self.step()
            time = self.time()
        print('entering flight mode', flush=True)
        self.enter_flight_mode()
        print('arming system', flush=True)
        self.arm_system()
        print('taking off', flush=True)
        self.takeoff(20)
        while abs(20 - self.position().z) > 2:
            self.step()
        self.land()
        while abs(self.position().z) > 2:
            self.step()
        self.pass_test()

if __name__ == '__main__':
    t = TakeoffAndHover(
        'udp:127.0.0.1:14550',
        'unix:///Users/madmax/.rmck_rpc',
        get_RAL()
    )
    t.test()
