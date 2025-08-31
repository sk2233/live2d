/*
@author: sk
@date: 2024/6/15
*/
package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type App struct {
	MotionManager *MotionManager
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
	if ebiten.IsKeyPressed(ebiten.KeyW) {
		Origin.Y -= 20
	} else if ebiten.IsKeyPressed(ebiten.KeyS) {
		Origin.Y += 20
	} else if ebiten.IsKeyPressed(ebiten.KeyA) {
		Origin.X -= 20
	} else if ebiten.IsKeyPressed(ebiten.KeyD) {
		Origin.X += 20
	} else if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		fmt.Println(Origin.X, Origin.Y)
	}
	return nil
}

func (a *App) Draw(screen *ebiten.Image) {
	a.MotionManager.Draw(screen)
}

func (a *App) Layout(w, h int) (int, int) {
	return w, h
}

func NewApp(motionManager *MotionManager) *App {
	return &App{MotionManager: motionManager, AnimIndex: 0, AnimNames: motionManager.GetAllMotions()}
}
