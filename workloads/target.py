#!/usr/bin/env python3
import grpc
import simulator_controller_pb2
import simulator_controller_pb2_grpc
from abc import ABC, abstractmethod
from pymavlink import mavutil
from time import sleep

# The robot abstraction layer
class RAL(ABC):
    def __init__(self, mav_addr, rpc_addr):
        self.address = mav_addr
        self.channel = grpc.insecure_channel(rpc_addr)
        self.stub = simulator_controller_pb2_grpc.SimulatorControllerStub(self.channel)
        self.mav = mavutil.mavlink_connection(mav_addr)

    @property
    @abstractmethod
    def enter_flight_mode(self):
        raise AttributeError('enter_flight_mode() must be implemented')

    @property
    @abstractmethod
    def takeoff(self, altitude: float, pitch: float=-1,
                yaw_angle: float=float('nan'), latitude: float=float('nan'), longitude: float=float('nan')):
        raise AttributeError('takeoff() must be implemented')

    def time(self):
        '''Returns the current time of simulation'''
        return self.stub.Time(simulator_controller_pb2.TimeRequest())

    def step(self):
        '''Advances the simulation 1 tick'''
        return self.stub.Step(simulator_controller_pb2.StepRequest())

    def position(self):
        '''Returns the current position'''
        return self.stub.Position(simulator_controller_pb2.PositionRequest())

    def pass_test(self):
        return self.stub.Terminate(simulator_controller_pb2.TerminateRequest())

    def change_mode(self, mode_no):
        return self.stub.ModeChange(simulator_controller_pb2.ModeChangeRequest(nextMode=mode_no))

    def arm_system(self):
        '''Arms the system for takeoff'''
        verified = False
        m = True
        while not verified:
            self.mav.arducopter_arm()
            self.step()
            m = self.mav.recv_match(type='COMMAND_ACK', blocking=True, timeout=0.1)
            verified = m is not None \
                and m.command == mavutil.mavlink.MAV_CMD_COMPONENT_ARM_DISARM \
                and m.result == mavutil.mavlink.MAV_RESULT_ACCEPTED
        self.change_mode(0)

    def recv_heartbeat_and_step(self):
        message = None
        while message is None:
            message = self.mav.recv_match(
                type='HEARTBEAT',
                blocking=True,
                timeout=0.001
            )
            self.step()
        return message

    def really_send_command(self, command, p1, p2, p3, p4, p5, p6, p7):
        '''sends command repeatedly over connection until a confirmation is sent back'''
        received_confirmation = False
        m = False
        while not received_confirmation:
            self.mav.mav.command_long_send(
                self.mav.target_system,
                self.mav.target_component,
                command,
                0, # confirmation
                p1,
                p2,
                p3,
                p4,
                p5,
                p6,
                p7
            )
            self.step()
            sleep(0.01)
            m = self.mav.recv_match(
                type='COMMAND_ACK',
                blocking=True,
                timeout=0.01
            )
            received_confirmation = m is not None \
                and m.command == command \
                and m.result == mavutil.mavlink.MAV_RESULT_ACCEPTED

    def reset_connection(self):
        '''Resets the MAVLink connection'''
        mav = self.mav
        self.mav = None
        mav.close()

        while self.mav == None:
            try:
                self.mav = mavutil.mavlink_connection(self.address)
            except socket.error:
                sleep(0)

    def land(self,
             abort_alt=0,
             precision_land_mode=0,
             yaw_angle=float('nan'),
             latitude=0,
            longitude=0,
             ground_altitude=0):
        '''lands the vehicle at the current location'''
        self.really_send_command(
            mavutil.mavlink.MAV_CMD_NAV_LAND,
            abort_alt,
            precision_land_mode,
            0, # empty
            yaw_angle,
            latitude,
            longitude,
            ground_altitude
        )

class PX4RAL(RAL):
    def enter_flight_mode(self):
        confirmed = False
        while not confirmed:
            self.mav.set_mode('STABILIZED', 'OFFBOARD')
            m = self.mav.recv_match(type='COMMAND_ACK', blocking=True, timeout=0.1)
            confirmed = m is not None and \
                m.command == mavutil.mavlink.MAV_CMD_DO_SET_MODE and \
                m.result == mavutil.mavlink.MAV_RESULT_ACCEPTED
            self.step()
        self.change_mode(1)

    def takeoff(self, altitude: float, pitch: float=-1,
                yaw_angle: float=float('nan'), latitude: float=float('nan'), longitude: float=float('nan')):
        # This works around a bug we discovered in PX4
        self.mav.param_set_send('MIS_TAKEOFF_ALT', altitude)
        self.really_send_command(
            mavutil.mavlink.MAV_CMD_NAV_TAKEOFF,
            pitch,
            0, 0, # empties
            yaw_angle,
            latitude,
            longitude,
            altitude
        )
        self.change_mode(2)

class ArduPilotRAL(RAL):
    def enter_flight_mode(self):
        while True:
            self.recv_heartbeat_and_step()
            self.mav.set_mode('GUIDED')
            self.reset_connection()
            self.recv_heartbeat_and_step()
            if self.mav.flightmode == 'GUIDED':
                break
            sleep(0.1)
        self.change_mode(1)

    def takeoff(self, altitude, pitch=-1,
            yaw_angle=float('nan'), latitude=float('nan'), longitude=float('nan')):
        self.recv_heartbeat_and_step()
        mav_autopilot = self.mav.field('HEARTBEAT', 'autopilot', None)
        self.really_send_command(
            mavutil.mavlink.MAV_CMD_NAV_TAKEOFF,
            pitch,
            0, 0, # empties
            yaw_angle,
            latitude,
            longitude,
            altitude
        )
        self.change_mode(2)

class Target(object):
    '''Target is extended to create new workloads'''
    def __init__(self, mav_addr: str, rpc_addr: str, ral_class: type):
        self.ral = ral_class(mav_addr, rpc_addr)

    def __getattr__(self, attr):
        return getattr(self.ral, attr)

    def test(self):
        '''Conducts a test'''
        raise AttributeError('error: test() must be reified')
