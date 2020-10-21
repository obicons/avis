#!/usr/bin/env python3
import json
import matplotlib.pyplot as plt
import numpy as np
import pandas as pd
import seaborn as s
from sklearn.linear_model import LinearRegression
from sklearn.metrics import mean_squared_error
from sklearn.model_selection import train_test_split
from scipy.stats import norm

GPS_FILENAME = 'gps.json'
ACCEL_FILENAME = 'accel.json'
GYRO_FILENAME = 'gyro.json'
BARO_FILENAME = 'baro.json'
COMPASS_FILENAME = 'compass.json'
POS_FILENAME = 'position.json'

INJECTION_TICKS = 1000

ACCEL_X_PARENTS = [
    'GPS.VelocityNorth',
    'GPS.VelocityEast',
    'GPS.VelocityDown',
    'GPS.Velocity',
    'Accel.X',
    'Gyroscope.X',
]

ACCEL_Y_PARENTS = [
    'GPS.VelocityNorth',
    'GPS.VelocityEast',
    'GPS.VelocityDown',
    'GPS.Velocity',
    'Accel.Y',
    'Gyroscope.Y',
]

ACCEL_Z_PARENTS = [
    'GPS.VelocityNorth',
    'GPS.VelocityEast',
    'GPS.VelocityDown',
    'GPS.Velocity',
    'Accel.Z',
    'Gyroscope.Z',
]

def read_file(filename):
    with open(filename) as fd:
        return json.load(fd)

def max_iteration_no(json_data):
    return max(map(lambda it: int(it), json_data.keys()))

def gps_json_to_pd(json_data, total_iterations):
    np_array = np.zeros((total_iterations + 1, 10), dtype=np.float64)
    presence_array = np.zeros(total_iterations + 1, dtype=np.bool)
    for iteration_no_str, gps_data in json_data.items():
        iteration_no = int(iteration_no_str)
        np_array[iteration_no][0] = gps_data['Latitude']
        np_array[iteration_no][1] = gps_data['Longitude']
        np_array[iteration_no][2] = gps_data['Altitude']
        np_array[iteration_no][3] = gps_data['EPH']
        np_array[iteration_no][4] = gps_data['EPV']
        np_array[iteration_no][5] = gps_data['Velocity']
        np_array[iteration_no][6] = gps_data['VelocityNorth']
        np_array[iteration_no][7] = gps_data['VelocityEast']
        np_array[iteration_no][8] = gps_data['VelocityDown']
        np_array[iteration_no][9] = gps_data['CourseOverGround']
        presence_array[iteration_no] = True
    for iteration_no, has_data in enumerate(presence_array):
        if (not has_data) and iteration_no > 0:
            np_array[iteration_no] = np_array[iteration_no - 1]
    return pd.DataFrame(
        data=np_array,
        columns=[
            'GPS.Latitude',
            'GPS.Longitude',
            'GPS.Altitude',
            'GPS.EPH', # what's this again?
            'GPS.EPV', # what's this again?
            'GPS.Velocity',
            'GPS.VelocityNorth',
            'GPS.VelocityEast',
            'GPS.VelocityDown',
            'GPS.CourseOverGround',
        ],
    )

def accel_json_to_pd(json_data, total_iterations):
    np_array = np.zeros((total_iterations + 1, 3), dtype=np.float64)
    presence_array = np.zeros(total_iterations + 1, dtype=np.bool)
    for iteration_no_str, accel_data in json_data.items():
        iteration_no = int(iteration_no_str)
        np_array[iteration_no][0] = accel_data['AccelerationX']
        np_array[iteration_no][1] = accel_data['AccelerationY']
        np_array[iteration_no][2] = accel_data['AccelerationZ']
        presence_array[iteration_no] = True
    for iteration_no, has_data in enumerate(presence_array):
        if (not has_data) and iteration_no > 0:
            np_array[iteration_no] = np_array[iteration_no - 1]
    return pd.DataFrame(
        data=np_array,
        columns=[
            'Accel.X',
            'Accel.Y',
            'Accel.Z',
        ],
    )

def gyro_json_to_pd(json_data, total_iterations):
    np_array = np.zeros((total_iterations + 1, 3), dtype=np.float64)
    presence_array = np.zeros(total_iterations + 1, dtype=np.bool)
    for iteration_no_str, gyro_data in json_data.items():
        iteration_no = int(iteration_no_str)
        np_array[iteration_no][0] = gyro_data['X']
        np_array[iteration_no][1] = gyro_data['Y']
        np_array[iteration_no][2] = gyro_data['Y']
        presence_array[iteration_no] = True
    for iteration_no, has_data in enumerate(presence_array):
        if (not has_data) and iteration_no > 0:
            np_array[iteration_no] = np_array[iteration_no - 1]
    return pd.DataFrame(
        data=np_array,
        columns=[
            'Gyroscope.X',
            'Gyroscope.Y',
            'Gyroscope.Z',
        ],
    )

def baro_json_to_pd(json_data, total_iterations):
    np_array = np.zeros((total_iterations + 1, 2), dtype=np.float64)
    presence_array = np.zeros(total_iterations + 1, dtype=np.bool)
    for iteration_no_str, baro_data in json_data.items():
        iteration_no = int(iteration_no_str)
        np_array[iteration_no][0] = baro_data['Pressure']
        np_array[iteration_no][1] = baro_data['Temperature']
        presence_array[iteration_no] = True
    for iteration_no, has_data in enumerate(presence_array):
        if (not has_data) and iteration_no > 0:
            np_array[iteration_no] = np_array[iteration_no - 1]
    return pd.DataFrame(
        data=np_array,
        columns=[
            'Barometer.Pressure',
            'Barometer.Temperature',
        ],
    )

def compass_json_to_pd(json_data, total_iterations):
    np_array = np.zeros((total_iterations + 1, 3), dtype=np.float64)
    presence_array = np.zeros(total_iterations + 1, dtype=np.bool)
    for iteration_no_str, compass_data in json_data.items():
        iteration_no = int(iteration_no_str)
        np_array[iteration_no][0] = compass_data['Mag0']
        np_array[iteration_no][1] = compass_data['Mag1']
        np_array[iteration_no][2] = compass_data['Mag2']
        presence_array[iteration_no] = True
    for iteration_no, has_data in enumerate(presence_array):
        if (not has_data) and iteration_no > 0:
            np_array[iteration_no] = np_array[iteration_no - 1]
    return pd.DataFrame(
        data=np_array,
        columns=[
            'Compass.Mag0',
            'Compass.Mag1',
            'Compass.Mag2',
        ],
    )

def pos_json_to_pd(json_data, total_iterations):
    np_array = np.zeros((total_iterations + 1, 3), dtype=np.float64)    
    presence_array = np.zeros(total_iterations + 1, dtype=np.bool)
    for iteration_no_str, pos_data in json_data.items():
        iteration_no = int(iteration_no_str)
        np_array[iteration_no][0] = pos_data['X']
        np_array[iteration_no][1] = pos_data['Y']
        np_array[iteration_no][2] = pos_data['Z']
        presence_array[iteration_no] = True
    for iteration_no, has_data in enumerate(presence_array):
        if (not has_data) and iteration_no > 0:
            np_array[iteration_no] = np_array[iteration_no - 1]
    return pd.DataFrame(
        data=np_array,
        columns=[
            'Pos.X',
            'Pos.Y',
            'Pos.Z',
        ],
    )

def velocity(pos_pd):
    np_array = np.zeros(pos_pd.shape)
    arr = pos_pd.to_numpy()
    for i in range(1, pos_pd.shape[0]):
        np_array[i] = arr[i] - arr[i-1]
    return pd.DataFrame(
        data=np_array,
        columns=[
            'Velocity.X',
            'Velocity.Y',
            'Velocity.Z',
        ],
    )

class Gaussian(object):
    def __init__(self, parent_nodes):
        self.parent_nodes = parent_nodes

    def fit(self, training_inputs, training_outputs):
        self.linear_regression = LinearRegression().fit(training_inputs, training_outputs)
        self.stddev = training_outputs.std()

    def generate(self):
        '''Generates data'''
        return np.random.multivariate_normal(self.mean, self.cov)

    def probability(self, x, parent_nodes):
        '''Returns P(x | parent_nodes)'''
        pred = self.linear_regression.predict(parent_nodes)
        gaussian = norm(scale=self.stddev, loc=pred)
        return gaussian(x)

    def sample(self, parent_nodes):
        '''Returns a random sample from the Gaussian (given the parent's nodes)'''
        pred = self.linear_regression.predict(parent_nodes)
        return np.random.normal(scale=self.stddev, loc=pred)

def inject_faults(all_data_pd,
                  start_time,
                  end_time,
                  fault_node,
                  gaussian_x,
                  gaussian_y,
                  gaussian_z):
    new_data = all_data_pd.copy()
    new_data[fault_node][start_time:end_time] = 0
    samples = np.zeros((end_time - start_time, 3))
    for t in range(end_time - start_time):
        samples[t][0] = gaussian_x.sample(
            new_data[gaussian_x.parent_nodes].iloc[t].to_numpy().reshape(1, -1),
        )
        samples[t][1] = gaussian_y.sample(
            new_data[gaussian_y.parent_nodes].iloc[t].to_numpy().reshape(1, -1),
        )
        samples[t][2] = gaussian_z.sample(
            new_data[gaussian_z.parent_nodes].iloc[t].to_numpy().reshape(1, -1),
        )
    return samples
        
if __name__ == '__main__':
    gps_data = read_file(GPS_FILENAME)
    accel_data = read_file(ACCEL_FILENAME)
    gyro_data = read_file(GYRO_FILENAME)
    baro_data = read_file(BARO_FILENAME)
    compass_data = read_file(COMPASS_FILENAME)
    pos_data = read_file(POS_FILENAME)
    total_iterations = max(
        max_iteration_no(gps_data),
        max_iteration_no(accel_data),
        max_iteration_no(gyro_data),
        max_iteration_no(baro_data),
        max_iteration_no(compass_data),
        max_iteration_no(pos_data),
    )
    gps_pd = gps_json_to_pd(gps_data, total_iterations)
    accel_pd = accel_json_to_pd(accel_data, total_iterations)
    gyro_pd = gyro_json_to_pd(gyro_data, total_iterations)
    baro_pd = baro_json_to_pd(baro_data, total_iterations)
    compass_pd = compass_json_to_pd(compass_data, total_iterations)
    pos_pd = pos_json_to_pd(pos_data, total_iterations)
    velocity_pd = velocity(pos_pd)
    accel_actual_pd = velocity(velocity_pd)
    all_inputs = pd.concat(
        (
            gps_pd,
            accel_pd,
            gyro_pd,
            baro_pd,
            compass_pd,
        ),
        axis=1,
    )

    train_sensors, test_sensors, train_actuators, test_actuators = train_test_split(
        all_inputs,
        velocity_pd,
        # accel_actual_pd,
        test_size=.4,
        random_state=42,
    )

    accel_x_node = Gaussian(ACCEL_X_PARENTS)
    accel_x_node.fit(
        train_sensors[ACCEL_X_PARENTS],
        train_actuators['Velocity.X']
    )
    accel_x_pred = accel_x_node.linear_regression.predict(
        test_sensors[ACCEL_X_PARENTS],
    )
    print(f'Accel.X model performance: {mean_squared_error(accel_x_pred, test_actuators["Velocity.X"])}')

    accel_y_node = Gaussian(ACCEL_Y_PARENTS)
    accel_y_node.fit(
        train_sensors[ACCEL_Y_PARENTS],
        train_actuators['Velocity.Y']
    )
    accel_y_pred = accel_y_node.linear_regression.predict(
        test_sensors[ACCEL_Y_PARENTS],
    )
    print(f'Accel.Y model performance: {mean_squared_error(accel_y_pred, test_actuators["Velocity.Y"])}')

    accel_z_node = Gaussian(ACCEL_Z_PARENTS)
    accel_z_node.fit(
        train_sensors[ACCEL_Z_PARENTS],
        train_actuators['Velocity.Z']
    )
    accel_z_pred = accel_z_node.linear_regression.predict(
        test_sensors[ACCEL_Z_PARENTS],
    )
    print(f'Accel.Z model performance: {mean_squared_error(accel_z_pred, test_actuators["Velocity.Z"])}')

    scenarios_considered = 0
    fault_nodes = set(ACCEL_X_PARENTS + ACCEL_Y_PARENTS + ACCEL_Z_PARENTS)
    for injection_start in range(total_iterations - INJECTION_TICKS + 1):
        for node_to_fail in fault_nodes:
            scenarios_considered += 1
            inject_faults(
                all_inputs,
                injection_start,
                injection_start + INJECTION_TICKS,
                node_to_fail,
                accel_x_node,
                accel_y_node,
                accel_z_node,
            )
            print(f'considered {scenarios_considered}')
