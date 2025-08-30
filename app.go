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
	AnimIndex     int
	AnimNames     []string
}

var (
	lastX = 0
	lastY = 0
)

func (a *App) Update() error {
	a.MotionManager.Update(1.0 / float64(ebiten.TPS()))
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		a.AnimIndex = (a.AnimIndex + 1) % len(a.AnimNames)
		a.MotionManager.PlayMotion(a.AnimNames[a.AnimIndex], true)
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
	return &App{MotionManager: motionManager, Scale: float64(scale), AnimIndex: 0, AnimNames: motionManager.GetAllMotions()}
}
