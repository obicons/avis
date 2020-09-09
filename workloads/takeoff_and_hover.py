#!/usr/bin/env python3
from target import PX4RAL, Target

class TakeoffAndHover(Target):
    def test(self):
        while self.time().tvSec < 40:
            self.step()
        self.enter_flight_mode()
        self.arm_system()
        self.takeoff(20)
        while abs(20 - self.position().z) > 2:
            self.step()

if __name__ == '__main__':
    t = TakeoffAndHover(
        'udp:127.0.0.1:14550',
        'unix:///Users/madmax/.rmck_rpc',
        PX4RAL
    )
    t.test()
