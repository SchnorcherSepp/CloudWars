#!/usr/bin/env python

import socket
import math
import time
import json
import random
import copy
import datetime
import statistics

# CONFIG
TCP_IP = '127.0.0.1'
TCP_PORT = 3333
CLOUD_NAME = "von Richthofen " + str(random.randint(1, 100))
CLOUD_COLOR = "red"

# TCP connection
conn = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
conn.connect((TCP_IP, TCP_PORT))

alive = True

# ------ Helper ------------------------------------------------------------------------------------------------------ #

def cmd_write_read(cmd):
    # remove protocol break
    cmd = cmd.replace('\n', '')
    cmd = cmd.replace('\r', '')

    # send command
    conn.send(bytes(cmd, 'utf8') + b'\n')
    
    # print("SEND", cmd)  # DEBUG !!!

    # read response
    resp = conn.makefile().readline()
    # print("RESP", resp)  # DEBUG !!!

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
    x = math.cos(angle) * strength * (-1)
    y = math.sin(angle) * strength * (-1)
    return move(x, y)

def move_by_angle_cost_xy(angle, strength):
    x = math.cos(angle) * strength * (-1)
    y = math.sin(angle) * strength * (-1)
    return x, y 

# MoveByAngle sends a move command for your player cloud to the server.
# Returns the server response (OK or ERR) as a string.
def move_by_angle_vc_correct(c1, angle, strength):
    x = (math.cos(angle) * strength * (-1))-c1['Vel']['X']
    y = (math.sin(angle) * strength * (-1))-c1['Vel']['Y']
    return move(x, y)
 
def move_by_angle_vc_correct_cost_xy(c1, angle, strength):
    x = (math.cos(angle) * strength * (-1))-c1['Vel']['X']
    y = (math.sin(angle) * strength * (-1))-c1['Vel']['Y']
    return x,y
 
def move_vc_increase(c1, strength):
    x = (c1['Vel']['X']/c1['Vel']['Y'])*strength*2
    y = (1.0-c1['Vel']['X']/c1['Vel']['Y'])*strength*2
    return move(x, y)


# Kill blasts the controlled cloud and removes it from the game.
# Returns the server response (OK or ERR) as a string.
def kill():
    return cmd_write_read("kill")



# Calculates Delta between two clouds
def delta(c1, c2):
    x2 = math.pow(abs(c1['Pos']['X']-c2['Pos']['X']), 2)
    y2 = math.pow(abs(c1['Pos']['Y']-c2['Pos']['Y']), 2)
    return math.sqrt(x2+y2)-math.sqrt(c1['Vapor'])-math.sqrt(c2['Vapor'])

def wind(c1):
    x2 = math.pow(abs(c1['Vel']['X']), 2)
    y2 = math.pow(abs(c1['Vel']['Y']), 2)
    return math.sqrt(x2+y2)

def delta_vc(c1, c2):
    x2 = math.pow(abs(c1['Vel']['X']-c2['Vel']['X']), 2)
    y2 = math.pow(abs(c1['Vel']['Y']-c2['Vel']['Y']), 2)
    return math.sqrt(x2+y2)

def delta_angel(c1, c2):
    # return  _point_delta_angle(c1['Pos']['X'], c1['Pos']['Y'], c2['Pos']['X'], c2['Pos']['Y'])
    delta_x = c1['Pos']['X'] - c2['Pos']['X']
    delta_y = c1['Pos']['Y'] - c2['Pos']['Y']
    d_a = math.atan2(delta_y, delta_x)
    return d_a

def is_in_sector(c1, c2, w, percentage):
    _d = delta(c1, c2)
    d_max = math.sqrt(w['Width']**2+w['Height']**2)*0.5
    return True if _d < d_max*percentage else False
    
    
def _hunting_update_world_border(world, c1):
    c1_rad = math.sqrt(c1['Vapor'])
    if c1['Pos']['X'] < c1_rad:
        c1['Pos']['X'] = c1_rad
        c1['Vel']['X'] = abs(c1['Vel']['X']) * 0.6
    if c1['Pos']['Y'] < c1_rad:
        c1['Pos']['Y'] = c1_rad
        c1['Vel']['Y'] = abs(c1['Vel']['X']) * 0.6
    if c1['Pos']['X']+c1_rad > world['Width']:
        c1['Pos']['X'] = world['Width']-c1_rad
        c1['Vel']['X'] = abs(c1['Vel']['X']) * -0.6
    if c1['Pos']['Y']+c1_rad > world['Height']:
        c1['Pos']['Y'] = world['Height']-c1_rad
        c1['Vel']['Y'] = abs(c1['Vel']['Y']) * -0.6
    
def hunting_costs(c1, c2, ai_state, world, strength_min, strength_max, stength_percentage, initial_boost, max_hunt_steps):
    _c1 = copy.deepcopy(c1)
    _c2 = copy.deepcopy(c2)
    cost_total = -0.1
    current_hunt_step = 0
    p_wind_1 = 0.0
        
    sim_decay = 0.999
    sim_velo_factor = 0.1
    sim_cost_factor = 0.001
    
    if 'GameSpeed' in world:
        move_update_factor = abs(world['GameSpeed']/ai_state['ai_speed'])
    else:
        move_update_factor = 1.0
    
    while True:
        current_hunt_step+=1
        # Skip hunt, too long
        if current_hunt_step > max_hunt_steps:
            cost_total = -2
            break
        
        # Collison bigger cloud
        for _c3 in world['Clouds']:
            if _c3['Player'] != _c1['Player'] and _c3['Vapor'] > _c1['Vapor'] and delta(_c1, _c3) < (math.sqrt(_c1['Vapor']) + math.sqrt(_c3['Vapor'])):
                cost_total = -1
                break
        if cost_total == -1:
            break
                
        # Hunt ok
        if delta(_c1, _c2) <= 2.0:
            break
        
        # Calulate Movement
        _vstrength_sum = 0
        vapor_percentage = _c1['Vapor']/100.0
    
        if (current_hunt_step == 1):
            _strength = min(max(strength_min,vapor_percentage*stength_percentage*initial_boost), strength_max)
        else:
            _strength = min(max(strength_min,vapor_percentage*stength_percentage), strength_max)
            
        _mx, _my = move_by_angle_vc_correct_cost_xy(_c1, delta_angel(_c1, _c2), _strength)
        # Velocity Strength (= -Vapor)
        _vstrength = math.sqrt(_mx*_mx + _my*_my)
        if _vstrength > 1.0 and  _vstrength < _c1['Vapor']/2.0:
            _c1['Vapor'] -= _vstrength
            _c1['Vel']['X'] += _mx * 5.0/math.sqrt(_c1['Vapor'])
            _c1['Vel']['Y'] += _my * 5.0/math.sqrt(_c1['Vapor'])
        else:
            _vstrength = 0
        _vstrength_sum+=_vstrength
            
        if (current_hunt_step == 1):
            p_wind_1 = math.sqrt(_c1['Vel']['X']*_c1['Vel']['X'] + _c1['Vel']['Y']*_c1['Vel']['Y'])
        
        # Update Round
        _c1['Pos']['X'] += _c1['Vel']['X']*sim_velo_factor
        _c1['Pos']['Y'] += _c1['Vel']['Y']*sim_velo_factor
        _rcost = abs(_c1['Vel']['X'])*sim_cost_factor+abs(_c1['Vel']['Y'])*sim_cost_factor
        _c1['Vel']['X'] *= sim_decay
        _c1['Vel']['Y'] *= sim_decay
       
        _c2['Pos']['X'] += _c2['Vel']['X']*sim_velo_factor
        _c2['Pos']['Y'] += _c2['Vel']['Y']*sim_velo_factor
        _c2['Vel']['X'] *= sim_decay
        _c2['Vel']['Y'] *= sim_decay
       
        # Border Bounce
        _hunting_update_world_border(world, _c1)
        _hunting_update_world_border(world, _c2)
       
        # Add up costs
        cost_total+=(_vstrength_sum/move_update_factor)+_rcost
    
    _c2['Vel']['X'] = 0.0
    _c2['Vel']['Y'] = 0.0
    
    # Approx:
    if cost_total == -2:
        cost_total = c1['Vapor'] * ai_state["cd_factor"] * delta(c1, c2)
        
    return cost_total, _c2, current_hunt_step, p_wind_1


def cloud_score(my_cloud, c1, hunt_calc_steps, hunt_p_wind):
    # Vapor Gain Perventage + Step Pecentage, Weighted
    cur_p_wind = math.sqrt(my_cloud['Vel']['X']*my_cloud['Vel']['X'] + my_cloud['Vel']['Y']*my_cloud['Vel']['Y'])
    return (c1['Vapor']/(my_cloud['Vapor']-c1['Cost']))*0.95 + (1.0-c1['HuntSteps']/hunt_calc_steps)*0.05 + hunt_p_wind/cur_p_wind*0.0
    

def hunt_score(my_cloud, c_old, c_new, hunt_calc_steps, hunt_p_wind):
    return True if cloud_score(my_cloud, c_old, hunt_calc_steps, hunt_p_wind) < cloud_score(my_cloud, c_new, hunt_calc_steps, hunt_p_wind) else False

# ----- Alex TotalCloudWar AI ---------------------------------------------------------------------------------------- #

def turn_baron(world, ai_state):
    # parse some world stats
    world_width = world['Width']  # game board size (DEFAULT: 2048)
    world_height = world['Height']  # game board size (DEFAULT: 1152)
    world_game_speed = world['GameSpeed']  # updates per second (DEFAULT: 60)
    world_iteration = world['Iteration']  # increases with every server update
    world_vapor = world['WorldVapor']  # vapor of all clouds together
    world_alive = world['Alive']  # active clouds
    world_clouds = world['Clouds']  # cloud list
    
    world_vapor = world_vapor if world_vapor > 0 else 1
    
    # cloud list
    my_cloud = [e for e in world_clouds if e['Player'] == CLOUD_NAME][0]
    enemy_list = [e for e in world_clouds if (e != my_cloud and ['Player'] != '' and e["Vapor"] > 1.0)]
    
    # Dead
    if my_cloud['Vapor'] <= 1.0:
        return -1
    
    vapor_percentage = my_cloud['Vapor']/100.0
    
    # Turn-Local Parameters
    adjust_toggle = 0
    
    sector_percentage = 1.0/6
    sector_percentage_increase_step = 1.0/6
    
    target_max_size_percentage = 0.98
    target_min_size_percentage = 0.2
    target_min_size_percentage_decrease_step = 0.0125
    
    stength_percentage = 3.7
    stength_percentage_initial_boost = 1.5
    stength_percentage_avoide_burst = 7.1
    strength_min = 1
    strength_max = 25
    
    hunt_calc_steps_max = 500
    hunt_steps_stall_factor = 1.25
    hunt_calc_time_max = 1000
    hunt_min_efficiency = 1.2
    hunt_min_efficiency_enemy = 0.4
    hunt_interception_wind_max = 25
    hunt_force_enemy = True
    
    player_avoidance_min_dist_percentage = 0.05
    
    sleep_time = 0.005
    
    # Avoid Bigger enemys
    for cloud in enemy_list:
        if (is_in_sector(my_cloud, cloud, world, player_avoidance_min_dist_percentage) and my_cloud['Vapor'] < cloud['Vapor']):
            print("AVOID ENEMY:", cloud['UID'])
            move_by_angle_vc_correct(my_cloud, delta_angel(cloud, my_cloud), min(max(strength_min,vapor_percentage*stength_percentage_avoide_burst), strength_max))
            ai_state["target_uid"] = ""
            return 0
    
    # Hunt targeted Cloud
    hunting_target = [e for e in world_clouds if e['UID'] == ai_state["target_uid"]]
    if len(hunting_target) and hunting_target[0]["Vapor"] >= 1.0:
        target_cloud = hunting_target[0]
        ai_state["target_hunt_cnt"]+=1
        # Cloud still small enought?
        if my_cloud['Vapor']*target_max_size_percentage <= target_cloud['Vapor']:
            print("HUNTING: STOPPED, TARGET TOO BIG")
            ai_state["target_uid"] = ""
            return
        # Cloud lost vapor? (Not Player)
        if (target_cloud['Player'] != '') and ai_state["target_vapor"] > target_cloud['Vapor']:
            if (delta(my_cloud, target_cloud) > 10.0):
                print("HUNTING: STOPPED, CLOUD LOST VAPOR (SUS)")
                ai_state["target_uid"] = ""
                return
        # Prediction <-> Reality Missmatch? (Not Player)
        if (target_cloud['Player'] != '') and ai_state["target_hunt_cnt"] > ai_state["target_hunt_steps"] :
            print("HUNTING: STOPPED, PREDICTION REALITY MISSMATCH")
            ai_state["target_uid"] = ""
            return
            
        if delta(my_cloud, target_cloud) < 10.0 and ai_state["target_start_vapor"] > 0:
            vapor_cost = ai_state["target_start_vapor"]-my_cloud["Vapor"]
            vapor_gain_predicted = ai_state["target_vapor"]
            vapor_gain_real = target_cloud['Vapor']
            vapor_cost_dist_initial_vapor = vapor_cost/ai_state["target_start_dist"]/ai_state["target_start_vapor"]
            print("KILLING TARGET:", ai_state["target_uid"], "DRAIN PRED.:", vapor_gain_predicted, "DRAIN REAL:", vapor_gain_real, "COST REAL:", vapor_cost, "COST_DIST_FACTOR:", vapor_cost_dist_initial_vapor)
            if vapor_cost > 0 and vapor_cost < 0.1:
                ai_state["cd_list"].append(vapor_cost_dist_initial_vapor)
                ai_state["cd_factor"] = statistics.mean(ai_state["cd_list"])
            ai_state["target_start_vapor"]=0
        # Hunt
        if target_cloud:
           if ai_state['target_virtual_pnt']:
                target_cloud = ai_state['target_virtual_pnt'] if delta(my_cloud, ai_state['target_virtual_pnt']) > 1.0 else target_cloud
           move_by_angle_vc_correct(my_cloud, delta_angel(my_cloud, target_cloud), min(max(strength_min,vapor_percentage*stength_percentage), strength_max))
           time.sleep(sleep_time)
           return 0

    ai_state["target_uid"] = ""
    while True:
        # Self-Kill
        if ai_state["self_kill_50p"] and world_vapor>0 and my_cloud["Vapor"]/world_vapor > 0.5:
            alive = False
            kill()
            break
    
        target_cloud = None
        target_cloud_bu = None
        target_delta = 0
        target_eff = 0
        target_vapor = 0
        target_hunt_steps = 1
        
        # Find best Target according to paramters
        _t = datetime.datetime.now()
        for cloud in world_clouds:
            # Limit calc time, use nearest
            if (datetime.datetime.now()-_t).total_seconds()*1000 > hunt_calc_time_max and target_cloud_bu:
                target_cloud = target_cloud_bu if target_cloud is None else target_cloud
                if 'Cost' not in target_cloud:
                    target_cloud['Cost'], target_cloud['_vPos'], target_cloud['HuntSteps'], _ = hunting_costs(my_cloud, target_cloud, ai_state, world, strength_min, strength_max, stength_percentage, stength_percentage_initial_boost, hunt_calc_steps_max)
                    target_cloud['Efficiency'] = target_cloud['Vapor']/target_cloud['Cost']
                target_delta = delta(my_cloud, target_cloud)
                target_vapor = target_cloud['Vapor']
                target_eff = target_cloud['Efficiency']
                break
                
            # Cloud is me
            if cloud == my_cloud:
                continue
            if cloud['Vapor'] < 1.0:
                continue
            # Distance to far?
            if not is_in_sector(my_cloud, cloud, world, sector_percentage):
                continue
            # Not in Size Range?
            if (my_cloud['Vapor']*target_max_size_percentage < cloud['Vapor']) or (my_cloud['Vapor']*target_min_size_percentage > cloud['Vapor']):
                continue       
            # Backup for calculation time
            _t_delta = delta(my_cloud, cloud)
            if _t_delta < target_delta:
                target_cloud_bu = cloud
                target_delta = _t_delta
                
            # Efficiency:
            if 'Cost' not in cloud:
                cloud['Cost'], cloud['_vPos'], cloud['HuntSteps'], cloud['p_wind'] = hunting_costs(my_cloud, cloud, ai_state, world, strength_min, strength_max, stength_percentage, stength_percentage_initial_boost, hunt_calc_steps_max)
                cloud['Efficiency'] = cloud['Vapor']/cloud['Cost']
            if cloud['Efficiency'] < hunt_min_efficiency and ((cloud['Player'] != '') and cloud['Efficiency'] < hunt_min_efficiency_enemy):
                continue
                        
            # Score targets
            if (not hunt_force_enemy) and target_cloud and (not hunt_score(my_cloud, target_cloud, cloud, hunt_calc_steps_max, cloud['p_wind'])):
                continue
            
            target_cloud = cloud
            target_delta = _t_delta
            target_vapor = target_cloud['Vapor']
            target_hunt_steps = target_cloud['HuntSteps']
            target_eff = cloud['Efficiency']
            
            # Force Enemy Hunt
            if hunt_force_enemy:
                break
          
        # Target found: Initiate Hunting
        if target_cloud:
            if '_vPos' in target_cloud:
                new_cost,_,__,___ = hunting_costs(my_cloud, target_cloud['_vPos'], ai_state, world, strength_min, strength_max, stength_percentage, stength_percentage_initial_boost, hunt_calc_steps_max)
            else:
                new_cost = -1
            print("HUNTING:", target_cloud["UID"], "DIST:", target_delta, "VAPOR:", target_vapor, "VAPOR_COST:",  target_cloud["Cost"], "EFFICENCY:", target_cloud['Efficiency'])
            if new_cost > 0 and new_cost < target_cloud["Cost"] and target_cloud['Player'] == '' and wind(target_cloud) < hunt_interception_wind_max:
                move_by_angle_vc_correct(my_cloud, delta_angel(my_cloud, target_cloud['_vPos']), min(max(strength_min,vapor_percentage*stength_percentage), strength_max))
                ai_state['target_virtual_pnt'] = target_cloud['_vPos']
            else:
                move_by_angle_vc_correct(my_cloud, delta_angel(my_cloud, target_cloud), min(max(strength_min,vapor_percentage*stength_percentage*stength_percentage_initial_boost), strength_max))
                ai_state['target_virtual_pnt'] = None
            
            ai_state["target_uid"] = target_cloud["UID"]
            ai_state["target_vapor"] = target_cloud['Vapor']
            ai_state["target_hunt_steps"] = target_cloud['HuntSteps']*hunt_steps_stall_factor
            ai_state["target_hunt_cnt"] = 0
            ai_state["target_start_vapor"] = my_cloud["Vapor"]
            ai_state["target_start_dist"] = target_delta
            break
        
        # Allo more targets
        elif sector_percentage < 1.00 or target_min_size_percentage > 0.0:
            # adjust sector size
            if adjust_toggle == 0:
                sector_percentage = min(1.0, sector_percentage+sector_percentage_increase_step)
            # adjust size range
            else:
               target_min_size_percentage-=target_min_size_percentage_decrease_step
               adjust_toggle = -1
            # Toogliger Toogle
            adjust_toggle+=1
        else:
            break
    
    time.sleep(sleep_time)
    return 0


def turn(ai_state):
    # get world status
    json_str = stat()
    world = json.loads(json_str)
    turn_baron(world, ai_state)

def baron_main():
    # set name, color and start the game
    name(CLOUD_NAME)
    color(CLOUD_COLOR)
    play()
    ai_state = {"self_kill_50p": False,
    "target_uid": '###',
    "target_virtual_pnt": None,
    "target_vapor": 0.0,
    "target_start_dist": 0,
    "target_start_vapor": 0,
    "target_hunt_steps": 0,
    "target_hunt_cnt": 0,
    "cd_list": [],
    "cd_factor": 0.0001,
    "ai_speed": 40}

    _t = datetime.datetime.now()
    _c = 0
    while True:
        if turn(ai_state) == -1:
            exit(0)
        _c += 1
        _d = datetime.datetime.now()-_t
        if (_d.total_seconds()*1000 >= 1000):
            ai_state["ai_speed"] = 0.0+_c/_d.total_seconds()
            _t = datetime.datetime.now()
            _c = 0
        
    
    close()