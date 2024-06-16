/*
@author: sk
@date: 2024/6/15
*/
package main

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// https://www.live2d.com/en/sdk/download/native/ sdk
// https://docs.live2d.com/en/cubism-sdk-manual/cubism-core-api-reference/ doc
// https://www.live2d.com/en/learn/sample/ official resource

func main() {
	core := NewCore("res/libLive2DCubismCore.dylib")
	model := core.LoadModel("res/hiyori/hiyori_free_t08.model3.json")
	size, _, _ := core.GetCanvasInfo(model.Moc.ModelPtr)
	scale := float32(0.28)
	w, h := size.X*scale, size.Y*scale
	motionManager := NewMotionManager(core, model, w, h)
	motionManager.PlayMotion("Idle", 2, true)
	ebiten.SetWindowSize(WinW, WinH)
	ebiten.SetWindowDecorated(false)
	ebiten.SetWindowFloating(true)
	//ebiten.SetWindowMousePassthrough(true)
	err := ebiten.RunGameWithOptions(NewApp(motionManager, scale),
		&ebiten.RunGameOptions{ScreenTransparent: true})
	HandleErr(err)
}
