/*
@author: sk
@date: 2024/6/15
*/
package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type App struct {
	MotionManager *MotionManager
	Scale         float64
	IdleIndex     int
	OnceIndex     int
}

var (
	onceAnims = []string{"Flick", "FlickDown", "Tap", "Tap@Body", "Flick@Body"}
)

var (
	lastX = 0
	lastY = 0
)

func (a *App) Update() error {
	a.MotionManager.Update(1.0 / float64(ebiten.TPS()))
	if inpututil.IsKeyJustPressed(ebiten.KeyI) {
		a.IdleIndex = (a.IdleIndex + 1) % 3
		a.MotionManager.PlayMotion("Idle", a.IdleIndex, true)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		a.OnceIndex = (a.OnceIndex + 1) % len(onceAnims)
		a.MotionManager.PlayMotion(onceAnims[a.OnceIndex], 0, false)
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		lastX, lastY = ebiten.CursorPosition()
	} else if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		currX, currY := ebiten.CursorPosition()
		x, y := ebiten.WindowPosition()
		ebiten.SetWindowPosition(x+currX-lastX, y+currY-lastY)
	}
	return nil
}

func (a *App) Draw(screen *ebiten.Image) {
	a.MotionManager.Draw(screen, a.Scale)
}

func (a *App) Layout(w, h int) (int, int) {
	return w, h
}

func NewApp(motionManager *MotionManager, scale float32) *App {
	return &App{MotionManager: motionManager, Scale: float64(scale)}
}
