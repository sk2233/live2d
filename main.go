/*
@author: sk
@date: 2024/6/15
*/
package main

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	fmt.Println(GetVersion())
	model := LoadModel("res/haru/haru.model3.json")
	size, _, _ := GetCanvasInfo(model.Moc.Model)
	scale := float32(1)
	WinW, WinH = size.X*scale, size.Y*scale
	motionManager := NewMotionManager(model, WinW, WinH)
	motionManager.PlayMotion("Idle", true)
	ebiten.SetWindowSize(int(WinW), int(WinH))
	ebiten.SetWindowDecorated(false)
	ebiten.SetWindowFloating(true)
	//ebiten.SetWindowMousePassthrough(true)
	err := ebiten.RunGameWithOptions(NewApp(motionManager, scale),
		&ebiten.RunGameOptions{ScreenTransparent: true})
	HandleErr(err)
}
