# CloudWars
CloudWars is a GO implementation of the Gathering [Harcore programming competition 2011](https://archive.gathering.org/tg11/en/creative/competitions/hardcore-programming/cloudwars/). This implementation offers more configuration options, a larger game board and the server-client protocol has been changed.

![img](https://archive.gathering.org/tg11/files/content/images/creativia/Screenshot.png)

## Summary
**The assignment is to create an Artificial Intelligence (AI) that plays a game called "CloudWars", described here. In the game, every player controls a thunderstorm, and the goal is to becomme the biggest thunderstorm by absorbing vapor from clouds and other players.**

The game contains two types of clouds; thunderstorms controlled by players, and rainclouds that just float around with no player control.

To move the thunderstorms around, players can create wind forces by expelling vapor. When a wind is created, small rainclouds containing the expelled vapor are created and propelled in the opposite direction of the force. This way, the total amount of vapor available in the game world is always constant.

When two clouds collide, the bigger clouds absorbs vapor from the smaller. The size of each cloud is proportional to the amount of vapor it contains. Absorption of vapor occurs until the clouds no longer overlap.


## How to compete
Download the GO source from [Github](https://github.com/SchnorcherSepp/CloudWars/) or the fully compiled binaries from the [release page](https://github.com/SchnorcherSepp/CloudWars/releases). The simulator will act as a TCP/IP game server and simulate and visualize the game according to the formal game rules, found below.

Each player AI is a separate application written in the [language of your choice](https://github.com/SchnorcherSepp/CloudWars/tree/master/examples) that connects to the simulator via TCP/IP. The clients (player AIs) and server communicate via a simple ASCII protocol over this connection. This protocol is described in the fomal game rules.

The simulator supports several game modes (AI vs AI, AI vs Human). Feel free to try or train your AI against human players or AI's made by others entering the competition ahead of the compo tournament.

The source code for the simulator is also provided. Feel free to modify it to accommodate any type of testing process you prefer. You are also free to create your own simulator from scratch, if you wish to do so.


## Formal game rules

### Game initialization
- The game is played in a 2D coordinate system. Default is a size of 2048 x 1152 units.
- 32-bit float values are used to represent positions and dimenisons within this coordinate system.
- The game contains two types of clouds called rainclouds and thunderstorms. Rainclouds simply float around passively, while Thunderstorms are controlled by a player.
- In the game logic, each cloud is treated as a perfect circle, even if the visual representation in the simulator is not always a perfect circle.
- The game is initialized random, containting 1-4 thunderstorms and 0-100 rainclouds with arbitrary amounts of vapor.
- Each cloud has the following properties in the simulator:
  - position : 2-component float vector representing the cloud's position
  - velocity : 2-component float vector representing the cloud's velocity
  - vapor : float representing the amount of vapor in the cloud
  - radius : float representing radius of the cloud. This is always the square root of vapor.
  - In addition, thunderstorms have string names, identifying the player who owns the thunderstorm.
- The simulator acts as a TCP/IP server, while the players act as TCP/IP clients that connect to the server. Clients for this protocol are implemented in the [example code](https://github.com/SchnorcherSepp/CloudWars/tree/master/examples).
- The game is executed in iterations. When the game starts, the iteration index is zero.
- After the initialization procedure described in the protocol, the simulator then enters a loop executing iterations of the game logic until the game is finsihed.
- The game runs in "real time" with seemingly smooth movements when visualized. However, internally, the simulator runs discrete iterations of the game logic, typically locked to 60 iterations per second. What happens for each iteration is described below.

### Game iteration loop

In the beginning of each iteration, a list called clouds is constructed by taking all thunderstorms in the order of their indicies first, and then appending all rainclouds in the order of their indices.

For each cloud A in the order of the list clouds:

1) Process input
   If the thunderstorm is controlled by an AI, the queued commands at the TCP/IP socket are processed according to the rules defined in the protocol. If the thunderstorm is controlled by a human player, the queue of events accumulated from the input device (e.g. mouse) is processed.

2) Movement
   The velocity vector is added to the position of each thunderstorm, scaled with a factor of 0.1.
   `position += velcoity * 0.1`

3) Damping of velocity
   The velocity is damped to make it more natural.
   `velocity *= 0.999`

4) Absorbing vapor from others
   For each cloud B in the order of the list clouds different from A. If the two cloud A and B's circles intersect, the following happens:
   One unit (1.0) of vapour is removed from the smallest cloud and added to the biggest cloud in a loop until their circles are no longer intersecting or the smallest cloud dies. If the smallest cloud's amount of vapor goes below 1.0, the cloud dies. If the clouds have exactly the same amount of vapor, it is random (but not undefined) which one is considered the largest with 50% probability for both A and B.

5) Bounce against the boundaries of the game world
   If the circle of A exceeds the boundaries of the game world, the position of A is affected as follows:
   - `if (position.x < radius) { position.x = radius; velocity.x = abs(velocity.x) * 0.6; }`
   - `if (position.y < radius) { position.y = radius; velocity.y = abs(velocity.y) * 0.6; }`
   - `if (position.x+radius > width) { position.x = width-radius; velocity.x = -abs(velocity.x) * 0.6; }`
   - `if (position.y+radius > height) { position.y = height-radius; velocity.y = -abs(velocity.y) * 0.6; }`

### Removal of dead clouds
At this point, all dead clouds (`vapor < 1.0`) are removed from their respective lists.

### Winning conditions
The game ends when a cloud unites more than 50% of the world's mass and can no longer be swallowed.

## Network protocol specification

### General conventions
1) The client sends a command to the server as a single line of text.
2) A command always consists of 4 characters followed by the payload without separators such as spaces.
3) The server responds by sending a single line of text.
4) A line of text must always be a string of ASCII characters terminated by a single, unix-style new line character: '\n'
5) All integers represented as ASCII text.
6) All floating point numbers are represented as ASCII text on the form 13.37
7) The client is always the active party while the server is always the reactive party. The server never sends anything without first receiving a command from the client.

### Initialization
Prior to game start, the simulator does the following:
- First the player should set a name with the _name_ command with the syntax:
  `name{YourNameHere}\n`
- Optionally, a color can be selected with the _type_ command with the syntax:
  `type{myColor}\n`. Valid values are `blue`, `gray`, `orange`, `purple` or `red`.
- The Server waits for all player send a _play_ command with the syntax:
  `play\n`

The status of the world should be queried continuously in order to recognize via the world iteration that the game has started.

### In-game commands
The following list contains the commands that the client can send to the server, and for each command a list of the possible responses from the server and their meanings.

#### Command: `list\n`
Polls the server for the current game state.

The client can poll this up to a maximum of 10 times per second. This means that the client can not keep 100% up to date and needs to choose how often it wants to poll the game state. Polling too often can lead to disqualification.

The server responds with a JSON on a single line:
```
{
   "Width":2048,     // game board size (DEFAULT: 2048)
   "Height":1152,    // game board size (DEFAULT: 1152)
   "GameSpeed":60,   // updates per second (DEFAULT: 60)
   "Iteration":0,    // increases with every server update
   "WorldVapor":0,   // vapor of all clouds together
   "Alive":0,        // active clouds
   "Clouds":[        // cloud list
      {
         "Pos":{
            "X":1262.5076, // x position
            "Y":364.86002  // y position
         },
         "Vel":{
            "X":0,  // x velocity (speed)
            "Y":0   // y velocity (speed)
         },
         " Vapor":600,     // cloud vapor (mass)
         "Player":"Hansi", // only player controlled clouds have names
         "Color":"blue"    // cloud color
      },
      {
        ...
      },
   ],
}
```

#### Command: `move{x};{y}\n`
Expels vapor from the player's thunderstorm and converts it into velocity for the thunderstorm.

This has the following effects:
- The strength of the wind is calculated as `sqrt(x*x+y*y)`.
- This value is not allowed to be less than `1.0` or greater than `vapor/2`. If this happens, the _move_ command is ignored.
- The vapor property of the thunderstorm will be reduced by strength.
- If the thunderstorm's amount of vapor goes below `1.0`, the player dies.
- The vector `[(x / radius)*5, (y / radius)*5]` is added to the velocity of the thunderstorm.
- The vector `[wx, wy]` is calculated as `[x / strength, y / strength]`.
- The vector `[vx, vy]` represent the velocity of the thunderstorm.
- A new raincloud is spawned with vapor equal to strength
- The distance to spawn the new raincloud at is calculated as `(storm_radius + raincloud_radius) * 1.1`
- The position of the new raincloud is set to `[px - wx * distance, py - wy * distance]` with velocity `[-(x / strength)*20 + vx, -(y / strength)*20 + vy]`

The server replies as follows:
- `ok\n` or
- `err: invalid move\n`


#### Command: `kill\n`
Kill blasts the controlled cloud and removes it from the game.


#### Command: `quit\n`
Quit disconnects from the server. The controlled cloud remains unchanged.

