package core

import (
	"math"
)

// Velocity is a 2-component float32 vector representing the cloud's velocity
type Velocity struct {
	X float32
	Y float32
}

// NewVelocity create a new float32 vector (see Velocity)
func NewVelocity(x, y float32) *Velocity {
	return &Velocity{
		X: x,
		Y: y,
	}
}

// NewVelocityByAngle create a new Velocity from angle and strength.
func NewVelocityByAngle(angle, strength float32) *Velocity {
	x := math.Cos(float64(math.Pi/180*angle)) * float64(strength) * (-1)
	y := math.Sin(float64(math.Pi/180*angle)) * float64(strength) * (-1)
	return NewVelocity(float32(x), float32(y))
}

//----  GETTER  ------------------------------------------------------------------------------------------------------//

// Strength is the hypotenuse of x and y and stands for the speed in one direction.
func (v *Velocity) Strength() float32 {
	return float32(math.Sqrt(float64(v.X*v.X + v.Y*v.Y)))
}

// clone creates a new instance of Velocity and initializes all its fields with exactly the contents.
func (v *Velocity) clone() *Velocity {
	return NewVelocity(v.X, v.Y)
}

//----  SETTER  ------------------------------------------------------------------------------------------------------//

// add the velocity with a multiplier to the current velocity.
func (v *Velocity) add(o *Velocity, multi float32) {
	if o == nil {
		return
	}
	v.X += o.X * multi
	v.Y += o.Y * multi
}

// multi the values x and y by the factor m
func (v *Velocity) multi(m float32) {
	v.X *= m
	v.Y *= m
}
