#!/usr/bin/env python3
from target import Target
from util import *

class TakeoffAndHover(Target):
    def test(self):
        time = self.time()
        while time.tvSec < 40:
            self.step()
            time = self.time()
        self.enter_flight_mode()
        self.arm_system()
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
        get_rpc_addr(),
        get_RAL()
    )
    t.test()
