/*
@author: sk
@date: 2024/6/15
*/
package main

const ( // 几种 Curve 的类型
	CurveLinear         = 0
	CurveBezier         = 1
	CurveStepped        = 2
	CurveInverseStepped = 3
)

const (
	TargetPartOpacity = "PartOpacity"
	TargetParameter   = "Parameter"
)

const (
	DFlagVisible = 1 << iota
	DFlagVisibilityChange
	DFlagOpacityChange
	DFlagDrawOrderChange
	DFlagRenderOrderChange
	DFlagVertexPositionChange
	DFlagBlendColorChange
)

const (
	WinW = 280
	WinH = 800
)
