package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"io"
	"log"
	"time"
)

// A palette with black and white colors.
var palette = [][]byte{
	{255, 255, 255},
	{165, 165, 165},
	{82, 82, 82},
	{0, 0, 0},
}

const (
	audioSampleRate = 48000
	bytesPerPixel   = 4
	numPixels       = DisplayWidth * DisplayHeight
	windowScale     = 3
	// A bit of a tradeoff: a large buffer size provides more stable audio, but increases the delay between an audio
	// change and when the new audio is actually played.
	audioBufferSize = 100 * time.Millisecond
)

type Game struct {
	pixels       [numPixels * bytesPerPixel]byte
	keysListener KeysListener
	audioStream  io.Reader
	audioContext *audio.Context
	audioPlayer  *audio.Player
}

func (g *Game) Update() error {
	if g.audioPlayer == nil && g.audioStream != nil {
		var err error
		g.audioPlayer, err = g.audioContext.NewPlayerF32(g.audioStream)
		if err != nil {
			return err
		}
		g.audioPlayer.SetBufferSize(audioBufferSize)
		g.audioPlayer.Play()
	}

	if g.keysListener != nil {
		keys := PressedKeys{
			ebiten.IsKeyPressed(ebiten.KeyUp),
			ebiten.IsKeyPressed(ebiten.KeyDown),
			ebiten.IsKeyPressed(ebiten.KeyLeft),
			ebiten.IsKeyPressed(ebiten.KeyRight),
			ebiten.IsKeyPressed(ebiten.KeyA),
			ebiten.IsKeyPressed(ebiten.KeyS),
			ebiten.IsKeyPressed(ebiten.KeyEnter),
			ebiten.IsKeyPressed(ebiten.KeyShiftRight),
		}
		g.keysListener.SetPressedKeys(keys)
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.WritePixels(g.pixels[:])
}

func (g *Game) Layout(int, int) (screenWidth int, screenHeight int) {
	return DisplayWidth, DisplayHeight
}

func (g *Game) Run() {
	if err := ebiten.RunGame(g); err != nil {
		log.Fatalf("Game failed to start: %v", err)
	}
}

func (g *Game) SetPixel(r int, c int, color byte) {
	pixelIndex := (r*DisplayWidth + c) * bytesPerPixel
	copy(g.pixels[pixelIndex:pixelIndex+3], palette[color])
}

func (g *Game) SetKeysListener(listener KeysListener) {
	g.keysListener = listener
}

func (g *Game) SetAudioStream(audioStream io.Reader) {
	g.audioStream = audioStream
}

func MakeGame() *Game {
	game := Game{}
	game.audioContext = audio.NewContext(audioSampleRate)
	ebiten.SetWindowSize(DisplayWidth*windowScale, DisplayHeight*windowScale)
	ebiten.SetWindowTitle("Good Boy")
	return &game
}
