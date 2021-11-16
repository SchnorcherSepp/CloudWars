package core

import (
	"math"
)

// Cloud can be a neutral cloud or a player cloud.
// Neutral clouds simply float around passively, while player cloud are controlled by a player.
// Each cloud has the following properties in the simulator:
//   position : 2-component float vector representing the cloud's position
//   velocity : 2-component float vector representing the cloud's velocity
//   vapor : float representing the amount of vapor in the cloud
//   radius : float representing radius of the cloud. This is always the square root of vapor
//   name: string identifying the player who owns the cloud
type Cloud struct {
	world  *World
	Pos    *Position // position
	Vel    *Velocity // speed
	Vapor  float32   // representing the amount of vapor in the cloud
	Player string    // clouds controlled by a player
	Color  string    // blue, red, orange, purple and gray (default: gray)
}

// NewCloud create a new Cloud.
// This function does not automatically add the cloud to the internal cloud list of a world!
func NewCloud(world *World, pos *Position, vel *Velocity, vapor float32, player, color string) *Cloud {
	return &Cloud{
		world:  world,
		Pos:    pos,
		Vel:    vel,
		Vapor:  vapor,
		Player: player,
		Color:  color,
	}
}

//----  GETTER  ------------------------------------------------------------------------------------------------------//

// Radius representing the radius of the cloud. This is always the square root of vapor
func (c *Cloud) Radius() float32 {
	return float32(math.Sqrt(float64(c.Vapor)))
}

// IsDeath true if the cloud is dead. (Vapor < 1)
func (c *Cloud) IsDeath() bool {
	return c.Vapor < 1
}

// isIntersects is true if two clouds overlap.
func (c *Cloud) isIntersects(o *Cloud) bool {
	x := o.Pos.X - c.Pos.X
	y := o.Pos.Y - c.Pos.Y
	return float32(math.Sqrt(float64(x*x+y*y))) < (o.Radius() + c.Radius())
}

// clone creates a new instance of Position and initializes all its fields with exactly the contents.
// Attention: The internal reference to the world is set to nil!
func (c *Cloud) clone() *Cloud {
	return NewCloud(nil, c.Pos.clone(), c.Vel.clone(), c.Vapor, c.Player, c.Color)
}

//----  SETTER  ------------------------------------------------------------------------------------------------------//

// update reduces velocity and adds it to the position.
// This function also calculates the rebound at the edges of the board.
func (c *Cloud) update() {
	// ignore death cloud
	if c.IsDeath() {
		return
	}

	// Movement
	// position += velocity * 0.1;
	simSpeedUp := 1
	if c.world != nil && c.world.SimSpeedUp != 0 {
		simSpeedUp = c.world.SimSpeedUp
	}
	c.Pos.add(c.Vel, 0.1*float32(simSpeedUp))

	// Damping of velocity
	// velocity *= 0.999;
	c.Vel.multi(0.999)

	// Absorbing vapor from others
	for _, o := range c.world.clouds {
		if o == c || o.IsDeath() {
			continue
		}

		var smallest *Cloud
		var biggest *Cloud
		if c.Radius() < o.Radius() {
			smallest = c
			biggest = o
		} else {
			smallest = o
			biggest = c
		}

		// Check for intersection
		for c.isIntersects(o) {
			if c.IsDeath() || o.IsDeath() {
				break
			}
			// Transfer vapor from smallest to biggest
			biggest.Vapor += 1
			smallest.Vapor -= 1
		}
	}

	// Bounce against walls
	if c.Pos.X < c.Radius() {
		c.Pos.X = c.Radius()
		c.Vel.X = float32(math.Abs(float64(c.Vel.X)) * 0.6)
	}
	if c.Pos.Y < c.Radius() {
		c.Pos.Y = c.Radius()
		c.Vel.Y = float32(math.Abs(float64(c.Vel.Y)) * 0.6)
	}
	if c.Pos.X+c.Radius() > float32(c.world.Width()) {
		c.Pos.X = float32(c.world.Width()) - c.Radius()
		c.Vel.X = float32(-math.Abs(float64(c.Vel.X)) * 0.6)
	}
	if c.Pos.Y+c.Radius() > float32(c.world.Height()) {
		c.Pos.Y = float32(c.world.Height()) - c.Radius()
		c.Vel.Y = float32(-math.Abs(float64(c.Vel.Y)) * 0.6)
	}
}

// move implements a move command. Vapor is reduced in order to generate Velocity.
func (c *Cloud) move(wind *Velocity) bool {
	// The Strength of the wind is calculated as sqrt(x*x+y*y)
	strength := wind.Strength()

	// This value is not allowed to be less than 1 or greater than vapor/2.
	// The amount of vapor can't go below 1.
	// If this happens, the WIND command is ignored.
	if strength < 1 || strength > c.Vapor/2 {
		return false
	}

	// The vapor property of the player controlled cloud will be reduced by Strength
	c.Vapor -= strength

	// The vector [(x / Radius) * 5, (y / Radius) * 5] is added to the
	// velocity of the cloud.
	c.Vel.add(wind, 5/c.Radius())

	// exhaust gases
	{
		// A new cloud is spawned with vapor equal to Strength.
		// The distance to spawn the new cloud at is calculated as:
		// (int)((storm_radius + raincloud_radius) * 1.1)
		distance := (c.Radius() + float32(math.Sqrt(float64(strength)))) * 1.1

		// The position of the new cloud is set to
		// [(int)(px - wx * distance), (int)(py - wy * distance)]
		position := NewPosition(c.Pos.X-wind.X/strength*distance, c.Pos.Y-wind.Y/strength*distance)

		// with velocity
		// [-(x / Strength) * 20 + vx, -(y / Strength) * 20 + vy]
		velocity := NewVelocity(-(wind.X/strength)*20+c.Vel.X, -(wind.Y/strength)*20+c.Vel.Y)

		// add to world
		cloud := NewCloud(c.world, position, velocity, strength, "", "")
		c.world.addCloud(cloud)
	}

	// success
	return true
}

// kill is a suicide order. The cloud explodes.
func (c *Cloud) kill() bool {
	if c.IsDeath() {
		c.Vapor = 0  // double kill ;)
		return false // already dead
	}

	// build explosion ring in world
	if c.world != nil {
		for i := float32(0); i < 360 && !c.IsDeath(); i += 10 {
			v := NewVelocityByAngle(i, 2.5)
			c.move(v)
		}
	}

	// kill
	c.Vapor = 0
	return true // successful
}
