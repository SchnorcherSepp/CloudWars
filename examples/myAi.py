#!/usr/bin/env python

"""
This file contains an example in Python for an AI controlled client.
Use this example to program your own AI in Python.
"""

import socket
import math
import time
import json

# CONFIG
TCP_IP = '127.0.0.1'
TCP_PORT = 3333
CLOUD_NAME = "AlexAI"
CLOUD_COLOR = "red"

# TCP connection
conn = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
conn.connect((TCP_IP, TCP_PORT))


# ------ Helper ------------------------------------------------------------------------------------------------------ #

def cmd_write_read(cmd):
    # remove protocol break
    cmd = cmd.replace('\n', '')
    cmd = cmd.replace('\r', '')

    # send command
    conn.send(bytes(cmd, 'utf8') + b'\n')
    print("SEND", cmd)  # DEBUG !!!

    # read response
    resp = conn.makefile().readline()
    print("RESP", resp)  # DEBUG !!!

    # return
    return resp


# ------ Commands ---------------------------------------------------------------------------------------------------- #


# Close disconnects from the server.
# The controlled cloud remains unchanged (use Kill() before this call).
# Returns the server response (OK or ERR) as a string.
def close():
    resp = cmd_write_read("quit")
    conn.close()
    return resp


# Stat returns the world status as a json string.
# Use core.World FromJson() to parse the string.
def stat():
    return cmd_write_read("list")


# Name set the player name
# Use this before calling Play()
# Returns the server response (OK or ERR) as a string.
def name(n):
    cmd = "name" + n
    return cmd_write_read(cmd)


# Color set the player color.
# 'blue', 'gray', 'orange', 'purple' or 'red'
# Use this before calling Play()
# Returns the server response (OK or ERR) as a string.
def color(c):
    cmd = "type" + c
    return cmd_write_read(cmd)


# Play creates a new player cloud.
# The attributes of Name() and Color() are used.
# Returns the server response (OK or ERR) as a string.
def play():
    return cmd_write_read("play")


# Move sends a move command for your player cloud to the server.
# Returns the server response (OK or ERR) as a string.
def move(x, y):
    cmd = "move" + str(x) + ";" + str(y)
    return cmd_write_read(cmd)


# MoveByAngle sends a move command for your player cloud to the server.
# Returns the server response (OK or ERR) as a string.
def move_by_angle(angle, strength):
    x = math.cos(math.pi / 180 * angle) * strength * (-1)
    y = math.sin(math.pi / 180 * angle) * strength * (-1)
    return move(x, y)


# Kill blasts the controlled cloud and removes it from the game.
# Returns the server response (OK or ERR) as a string.
def kill():
    return cmd_write_read("kill")


# ----- Alex TotalCloudWar AI ---------------------------------------------------------------------------------------- #

if __name__ == '__main__':
    # set name, color and start the game
    name(CLOUD_NAME)
    color(CLOUD_COLOR)
    play()

    # get world status
    json_str = stat()
    world = json.loads(json_str)

    # parse some world stats
    world_width = world['Width']  # game board size (DEFAULT: 2048)
    world_height = world['Height']  # game board size (DEFAULT: 1152)
    world_game_speed = world['GameSpeed']  # updates per second (DEFAULT: 60)
    world_iteration = world['Iteration']  # increases with every server update
    world_vapor = world['WorldVapor']  # vapor of all clouds together
    world_alive = world['Alive']  # active clouds
    world_clouds = world['Clouds']  # cloud list

    # cloud list
    me = None  # your controlled cloud (find in list)
    for cloud in world_clouds:
        cloud_name = cloud['Player']  # only player controlled clouds have names
        cloud_color = cloud['Color']  # cloud color
        cloud_vapor = cloud['Vapor']  # cloud vapor (mass)
        cloud_pos_x = cloud['Pos']['X']  # x position
        cloud_pos_y = cloud['Pos']['Y']  # y position
        cloud_vel_x = cloud['Vel']['X']  # x velocity (speed)
        cloud_vel_y = cloud['Vel']['Y']  # y velocity (speed)
        if cloud_name == CLOUD_NAME:
            me = cloud  # set 'me'

    # make some decisions
    # move to the center
    if me['Pos']['X'] < world_width:
        time.sleep(2)
        move_by_angle(180, 33)  # move right
    else:
        time.sleep(2)
        move_by_angle(0, 33)  # move left

    # move around
    time.sleep(2)
    move_by_angle(0, 10)  # move left
    time.sleep(2)
    move_by_angle(90, 10)  # move up
    time.sleep(2)
    move_by_angle(180, 10)  # move right
    time.sleep(2)
    move_by_angle(270, 10)  # move down
    time.sleep(2)

    # it makes no sense
    kill()
    close()
