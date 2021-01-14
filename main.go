package main

import (
	"flag"
	"fmt"
	"image/color"
	"math/rand"
	"os"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// rgb is the colour palette to use for the Fire.
// It's ordered from "coldest" to "hottest"
var rgb = []color.RGBA{
	{0x07, 0x07, 0x07, 0xFF},
	{0x1F, 0x07, 0x07, 0xFF},
	{0x2F, 0x0F, 0x07, 0xFF},
	{0x47, 0x0F, 0x07, 0xFF},
	{0x57, 0x17, 0x07, 0xFF},
	{0x67, 0x1F, 0x07, 0xFF},
	{0x77, 0x1f, 0x07, 0xFF},
	{0x8f, 0x27, 0x07, 0xFF},
	{0x9F, 0x2F, 0x07, 0xFF},
	{0xAF, 0x3F, 0x07, 0xFF},
	{0xBF, 0x47, 0x07, 0xFF},
	{0xC7, 0x47, 0x07, 0xFF},
	{0xDF, 0x4F, 0x07, 0xFF},
	{0xDF, 0x57, 0x07, 0xFF},
	{0xDF, 0x57, 0x07, 0xFF},
	{0xDF, 0x57, 0x07, 0xFF},
	{0xD7, 0x67, 0x0F, 0xFF},
	{0xCF, 0x6F, 0x0F, 0xFF},
	{0xCF, 0x77, 0x0F, 0xFF},
	{0xCF, 0x7F, 0x0F, 0xFF},
	{0xCF, 0x7F, 0x0F, 0xFF},
	{0xCF, 0x87, 0x17, 0xFF},
	{0xC7, 0x87, 0x17, 0xFF},
	{0xC7, 0x8F, 0x17, 0xFF},
	{0xC7, 0x97, 0x1F, 0xFF},
	{0xBF, 0x9F, 0x1F, 0xFF},
	{0xBF, 0x9F, 0x1F, 0xFF},
	{0xBF, 0xA7, 0x27, 0xFF},
	{0xBF, 0xA7, 0x27, 0xFF},
	{0xBF, 0xAF, 0x2F, 0xFF},
	{0xB7, 0xAF, 0x2F, 0xFF},
	{0xB7, 0xB7, 0x2F, 0xFF},
	{0xB7, 0xB7, 0x37, 0xFF},
	{0xCF, 0xCF, 0x6F, 0xFF},
	{0xDF, 0xDF, 0x9F, 0xFF},
	{0xEF, 0xEF, 0xC7, 0xFF},
	{0xFF, 0xFF, 0xFF, 0xFF},
}

var (
	widthFlag  = flag.Int("width", 600, "width (in pixels)")
	heightFlag = flag.Int("height", 400, "height (in pixels)")
)

// Doom implements the ebiten.Game interface.
type Doom struct {
	width, height int
	firePixels    []int
	screenBuffer  []byte
	waitGroup     sync.WaitGroup
	printDebug    bool
}

// NewDoom creates a new instance of Doom
func NewDoom(width, height int) *Doom {
	d := &Doom{width: width, height: height}
	d.firePixels = make([]int, d.width*d.height)
	d.screenBuffer = make([]byte, d.width*d.height*4)
	// Set whole screen to 0 (color: 0x07,0x07,0x07)
	for i := 0; i < d.width*d.height; i++ {
		d.firePixels[i] = 0
	}

	// Set bottom line to 37 (color white: 0xFF,0xFF,0xFF)
	for i := 0; i < d.width; i++ {
		d.firePixels[(d.height-1)*d.width+i] = len(rgb) - 1
	}
	return d
}

func (d *Doom) spreadFire(src int) {
	pixel := d.firePixels[src]
	if pixel == 0 {
		d.firePixels[src-d.width] = 0
		return
	}
	randIdx := rand.Intn(3)
	dst := src - randIdx + 1
	i := dst - d.width
	if i > 0 {
		d.firePixels[i] = pixel - (randIdx & 1)
	}
}

// Update applies the fire spread on each frame.
func (d *Doom) Update() error {
	for x := 0; x < d.width; x++ {
		d.waitGroup.Add(1)
		go func(x int) {
			defer d.waitGroup.Done()
			for y := 1; y < d.height; y++ {
				d.spreadFire(y*d.width + x)
			}
		}(x)
	}
	d.waitGroup.Wait()

	if inpututil.IsKeyJustPressed(ebiten.KeyD) {
		d.printDebug = !d.printDebug
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		os.Exit(0)
	}

	return nil
}

// Draw plots the current fire framebuffer.
func (d *Doom) Draw(scr *ebiten.Image) {
	l := d.width * d.height
	for i := 0; i < l; i++ {
		c := d.firePixels[i]
		r, g, b, a := rgb[c].RGBA()
		d.screenBuffer[4*i] = byte(r)
		d.screenBuffer[4*i+1] = byte(g)
		d.screenBuffer[4*i+2] = byte(b)
		d.screenBuffer[4*i+3] = byte(a)
	}
	scr.ReplacePixels(d.screenBuffer)

	if d.printDebug {
		ebitenutil.DebugPrint(scr, fmt.Sprintf("TPS: %0.2f", ebiten.CurrentTPS()))
	}
}

// Layout takes the outside size (e.g., the window size) and returns the (logical) screen size.
func (d *Doom) Layout(outsideWidth, outsideHeight int) (int, int) {
	return d.width, d.height
}

func main() {
	flag.Parse()
	ebiten.SetWindowTitle("DOOM")
	ebiten.SetMaxTPS(ebiten.UncappedTPS)
	ebiten.SetWindowSize(*widthFlag*2, *heightFlag*2)
	d := NewDoom(*widthFlag, *heightFlag)

	if err := ebiten.RunGame(d); err != nil {
		panic(err)
	}
}
