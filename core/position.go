package core

// Position is a 2-component float32 vector representing the cloud's position
type Position struct {
	X float32
	Y float32
}

// NewPosition create a new float32 vector (see Position)
func NewPosition(x, y float32) *Position {
	return &Position{
		X: x,
		Y: y,
	}
}

//----  GETTER  ------------------------------------------------------------------------------------------------------//

// clone creates a new instance of Position and initializes all its fields with exactly the contents.
func (p *Position) clone() *Position {
	return NewPosition(p.X, p.Y)
}

//----  SETTER  ------------------------------------------------------------------------------------------------------//

// add the velocity with a multiplier to the current position.
func (p *Position) add(v *Velocity, multi float32) {
	if v == nil {
		return
	}
	p.X += v.X * multi
	p.Y += v.Y * multi
}
