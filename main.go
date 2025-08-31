/*
@author: sk
@date: 2024/6/15
*/
package main

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

var (
	Size       *Vector2
	Origin     *Vector2
	Scale      float32
	IsAzurLane bool // 大部分顶点都是 -0.5～0.5 的小数，不过 碧蓝航线 需要特殊处理
)

// 空格切换动画 鼠标拖动位置

func main() {
	fmt.Println(GetVersion())
	model := LoadModel("res/kewei/kewei_4.model3.json")
	Size, Origin, _ = GetCanvasInfo(model.Moc.Model)
	Scale = min(1440/Size.X, 810/Size.Y) // 宽高限制在 0~1280 0~720
	Size.X, Size.Y, Origin.X, Origin.Y = Size.X*Scale, Size.Y*Scale, Origin.X*Scale, Origin.Y*Scale
	// chaijun
	//IsAzurLane = true
	//Scale = 1.0 / 28.0
	//Origin.X, Origin.Y = 14703, 11465
	// kewei
	IsAzurLane = true
	Scale = 1.0 / 25.0
	Origin.X, Origin.Y = 8123, 9365
	motionManager := NewMotionManager(model)
	motionManager.PlayMotion("Idle", true)
	ebiten.SetWindowSize(int(Size.X), int(Size.Y))
	ebiten.SetWindowDecorated(false)
	ebiten.SetWindowFloating(true)
	//ebiten.SetWindowMousePassthrough(true)
	err := ebiten.RunGameWithOptions(NewApp(motionManager),
		&ebiten.RunGameOptions{ScreenTransparent: true})
	HandleErr(err)
}
