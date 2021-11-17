package gui

import (
	"CloudWars/core"
	"CloudWars/remote"
	"errors"
	"fmt"
	"github.com/golang/freetype/truetype"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text"
	"golang.org/x/image/font/gofont/gomonobold"
	"image/color"
	"time"
)

// interface check: ebiten.Game
var _ ebiten.Game = (*Game)(nil)

type Game struct {
	screenWidth       int
	screenHeight      int
	world             *core.World
	localPlayer       *core.Cloud // manual control for this player cloud
	maxUpdateTime     time.Duration
	externWorldUpdate bool
	remoteMove        *remote.TcpClient
}

// RunGame creates a GUI. The game can be watched in the window or a player cloud can be controlled with the mouse.
// title is for the window title. screenWidth and screenHeight define the dimensions of the window.
// The window can also update the game logic of world. gameSpeed defines how often an update is called per second.
// Is externWorldUpdate true, no game logic is updated. If a local cloud is set with localPlayer, it can be controlled
// with the mouse. Is remoteMove set, the move command is send to a remote server with remote.TcpClient.
func RunGame(title string, screenWidth, screenHeight, gameSpeed int, world *core.World, localPlayer *core.Cloud, externWorldUpdate bool, remoteMove *remote.TcpClient) error {
	// world check
	if world == nil {
		return errors.New("world is nul")
	}

	// config game
	game := &Game{
		screenWidth:       screenWidth,
		screenHeight:      screenHeight,
		world:             world,
		localPlayer:       localPlayer,
		maxUpdateTime:     0, // 16ms is fast enough for 60 updates per second
		externWorldUpdate: externWorldUpdate,
		remoteMove:        remoteMove,
	}

	// config window
	ebiten.SetWindowTitle(title)
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowResizable(true)
	ebiten.SetMaxTPS(gameSpeed) // default: 60

	// BLOCKING run
	return ebiten.RunGame(game)
}

//--------------------------------------------------------------------------------------------------------------------//

// Layout accepts a native outside size in device-independent pixels and returns the game's logical screen
// size.
//
// On desktops, the outside is a window or a monitor (fullscreen mode). On browsers, the outside is a body
// element. On mobiles, the outside is the view's size.
//
// Even though the outside size and the screen size differ, the rendering scale is automatically adjusted to
// fit with the outside.
//
// Layout is called almost every frame.
//
// It is ensured that Layout is invoked before Update is called in the first frame.
//
// If Layout returns non-positive numbers, the caller can panic.
//
// You can return a fixed screen size if you don't care, or you can also return a calculated screen size
// adjusted with the given outside size.
func (g *Game) Layout(_, _ int) (int, int) {
	return g.screenWidth, g.screenHeight
}

// Update updates a game by one tick. The given argument represents a screen image.
//
// Update updates only the game logic and Draw draws the screen.
//
// In the first frame, it is ensured that Update is called at least once before Draw. You can use Update
// to initialize the game state.
//
// After the first frame, Update might not be called or might be called once
// or more for one frame. The frequency is determined by the current TPS (tick-per-second).
func (g *Game) Update() error {
	start := time.Now()

	// player control
	if g.localPlayer != nil && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		if g.remoteMove == nil {
			// local command
			c := g.localPlayer
			v := core.NewVelocity((float32(x)-c.Pos.X)/100, (float32(y)-c.Pos.Y)/100)
			g.world.Move(c, v)
		} else {
			// remote command
			c := g.world.Me(g.localPlayer.Player)
			v := core.NewVelocity((float32(x)-c.Pos.X)/100, (float32(y)-c.Pos.Y)/100)
			g.remoteMove.Move(v)
		}
	}

	// local world update
	if !g.externWorldUpdate {
		g.world.Update()
	}

	// watch maxUpdateTime
	duration := time.Since(start)
	if g.maxUpdateTime.Microseconds() < duration.Microseconds() {
		g.maxUpdateTime = duration
		if g.maxUpdateTime > 16*time.Millisecond {
			// 16ms is fast enough for 60 updates per second
			fmt.Println("WARNING:", "maxUpdateTime", duration)
		}
	}
	return nil
}

// Draw draws the game screen by one frame.
//
// The give argument represents a screen image. The updated content is adopted as the game screen.
func (g *Game) Draw(screen *ebiten.Image) {
	// background image
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(g.screenWidth)/1273.0, float64(g.screenHeight)/720.0) // bgImage is 1273px * 720px
	op.Filter = ebiten.FilterLinear                                             // Specify linear filter.
	screen.DrawImage(bgImage, op)

	// cloud images
	for _, c := range g.world.Clouds() {
		// calc for image placing
		radius := c.Radius()
		size := radius * 2 / 400 // the image is 512px, but the ball is 400px
		off := radius * 1.28     // 512px / 400px = 1.28

		// prepare image
		op = &ebiten.DrawImageOptions{}
		op.GeoM.Scale(float64(size), float64(size))
		op.GeoM.Translate(float64(c.Pos.X-off), float64(c.Pos.Y-off))
		op.Filter = ebiten.FilterLinear // Specify linear filter.

		// select color
		switch c.Color {
		case "blue":
			screen.DrawImage(blueImage, op)
		case "orange":
			screen.DrawImage(orangeImage, op)
		case "purple":
			screen.DrawImage(purpleImage, op)
		case "red":
			screen.DrawImage(redImage, op)
		default:
			screen.DrawImage(grayImage, op)
		}

		// print player name
		// char height is 10px
		// char width is 6px
		name := c.Player
		if name != "" && !c.IsDeath() {
			ebitenutil.DebugPrintAt(screen, name, int(c.Pos.X-float32(6/2*len(name))), int(c.Pos.Y+5+radius))
		}
	}

	// DEBUG text
	iteration, worldVapor, alive, winCondition, leader := g.world.Stats()
	msg := fmt.Sprintf("\n  round=%d/%d, alive=%d, worldVapor=%.0f, maxUpdateTime=%v\n", iteration, g.world.MaxIterations(), alive, worldVapor, g.maxUpdateTime)
	for _, c := range g.world.Clouds() {
		if c.Player != "" {
			if !c.IsDeath() {
				msg += fmt.Sprintf("  > %s: %.0f (%.0f%%)\n", c.Player, c.Vapor, c.Vapor/worldVapor*100)
			} else {
				msg += fmt.Sprintf("  > %s: dead\n", c.Player)
			}
		}
	}
	ebitenutil.DebugPrint(screen, msg)

	// PRINT WINNER
	// font size is 44
	// char height is 32px
	// char width is 24px
	if winCondition {
		fnt, _ := truetype.Parse(gomonobold.TTF)
		face := truetype.NewFace(fnt, &truetype.Options{Size: 44})

		winnerMsg := fmt.Sprintf("Victory: %s", leader)
		x := g.world.Width()/2 - 24/2*len(winnerMsg) - 20
		y := g.world.Height()/2 - 32/2
		clr := color.White

		text.Draw(screen, winnerMsg, face, x, y, clr)
	}
}
