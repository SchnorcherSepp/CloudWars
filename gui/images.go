package gui

import (
	"embed"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	_ "image/png" // needed for NewImageFromFile()
	"log"
)

var (
	bgImage     *ebiten.Image
	blueImage   *ebiten.Image
	grayImage   *ebiten.Image
	logoImage   *ebiten.Image
	orangeImage *ebiten.Image
	purpleImage *ebiten.Image
	redImage    *ebiten.Image
)

//go:embed images
var f embed.FS

func init() {
	bgImage = load("images/bg.png")
	blueImage = load("images/blue.png")
	grayImage = load("images/gray.png")
	logoImage = load("images/logo.png")
	orangeImage = load("images/orange.png")
	purpleImage = load("images/purple.png")
	redImage = load("images/red.png")
}

func load(name string) *ebiten.Image {
	// open reader
	r, err := f.Open(name)
	if err != nil {
		log.Fatalf("load: %v\n", err)
	}
	// get image
	eim, _, err := ebitenutil.NewImageFromReader(r)
	if err != nil {
		log.Fatalf("load: %v\n", err)
	}
	// return
	return eim
}
