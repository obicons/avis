#!/usr/bin/env python3
import json
import matplotlib.pyplot as plt
import numpy as np
import pandas as pd
import seaborn as sn

GPS_FILENAME = 'gps.json'
ACCEL_FILENAME = 'accel.json'
GYRO_FILENAME = 'gyro.json'
BARO_FILENAME = 'baro.json'
COMPASS_FILENAME = 'compass.json'

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
            'GPS.CourseOverGround'
        ]
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
        ]
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
        ]
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
        ]
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
        ]
    )

if __name__ == '__main__':
    gps_data = read_file(GPS_FILENAME)
    accel_data = read_file(ACCEL_FILENAME)
    gyro_data = read_file(GYRO_FILENAME)
    baro_data = read_file(BARO_FILENAME)
    compass_data = read_file(COMPASS_FILENAME)

    total_iterations = max(
        max_iteration_no(gps_data),
        max_iteration_no(accel_data),
        max_iteration_no(gyro_data),
        max_iteration_no(baro_data),
        max_iteration_no(compass_data),
    )

    gps_pd = gps_json_to_pd(gps_data, total_iterations)
    accel_pd = accel_json_to_pd(accel_data, total_iterations)
    gyro_pd = gyro_json_to_pd(gyro_data, total_iterations)
    baro_pd = baro_json_to_pd(baro_data, total_iterations)
    compass_pd = compass_json_to_pd(compass_data, total_iterations)

    all_var_pd = pd.concat(
        [
            gps_pd,
            accel_pd,
            gyro_pd,
            baro_pd,
            compass_pd,
        ],
        axis=1
    )

    sn.heatmap(all_var_pd.corr(), annot=True)
    plt.show()
