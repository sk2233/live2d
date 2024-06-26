/*
@author: sk
@date: 2024/6/15
*/
package main

import (
	"fmt"
	"image/color"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
)

type MotionManager struct {
	Core   *Core
	Model  *Model
	Motion *Motion
	Timer  float64
	Loop   bool
	W, H   float32
	Shader *ebiten.Shader
	// shader中使用的图片必须等大小，这里必须要先把图片绘制到另一个图片上
	Mask *ebiten.Image
	Src  *ebiten.Image
}

func (m *MotionManager) PlayMotion(name string, idx int, loop bool) {
	m.Motion = m.Model.Motions[name][idx]
	m.Timer = 0
	m.Loop = loop
}

func (m *MotionManager) StopMotion() {
	m.Motion = nil
}

func (m *MotionManager) Update(delta float64) {
	m.UpdateMotion(delta)
	m.Core.Update(m.Model.Moc.ModelPtr)
	m.UpdateModel()
}

func (m *MotionManager) UpdateMotion(delta float64) {
	if m.Motion == nil {
		return
	}
	m.Timer += delta
	if m.Timer > m.Motion.Data.Meta.Duration {
		if !m.Loop {
			m.Motion = nil
			return
		}
		m.Timer = 0
	}
	// 整体的渐入渐出设置
	fadeIn, fadeOut := GetFade(m.Motion, m.Timer)
	for _, curve := range m.Motion.Curves {
		// 每个曲线控制一个部分，一个曲线分为多段，循环获取当前时间对应的段
		segment := GetRightSegment(curve.Segments, m.Timer)
		value := GetSegmentValue(segment, m.Timer) // 获取所在段当前时间的值
		switch curve.Data.Target {
		case TargetPartOpacity:
			m.Core.SetPartOpacity(m.Model.Moc.ModelPtr, curve.Data.Id, float32(value))
		case TargetParameter:
			oldValue := m.Core.GetParameterValue(m.Model.Moc.ModelPtr, curve.Data.Id)
			fin, fout := fadeIn, fadeOut // 默认都取全局默认值，我们认为 FadeInTime<0 FadeOutTime<0 是默认值
			if curve.FadeInTime > 0 {
				fin = GetEasingSine(m.Timer / curve.FadeInTime)
			} else if curve.FadeInTime == 0 { // 不需要时间直接就是最终状态
				fin = 1
			}
			if curve.FadeOutTime > 0 {
				fout = GetEasingSine((m.Motion.Data.Meta.Duration - m.Timer) / curve.FadeOutTime)
			} else if curve.FadeOutTime == 0 {
				fout = 1
			}
			newValue := oldValue + float32(fin*fout)*(float32(value)-oldValue)
			m.Core.SetParameterValue(m.Model.Moc.ModelPtr, curve.Data.Id, newValue)
		default:
			panic(fmt.Sprintf("invalid target: %v", curve.Data.Target))
		}
	}
}

func (m *MotionManager) UpdateModel() {
	dflags := m.Core.GetDynamicFlags(m.Model.Moc.ModelPtr)
	// 通过动态 flag判断任何一个有改变就就进行一次同步数据
	drawOrderChange := false
	renderOrderChange := false
	opacityChange := false
	vertexPositionsChange := false
	for _, dflag := range dflags {
		if HasFlag(dflag, DFlagDrawOrderChange) {
			drawOrderChange = true
		}
		if HasFlag(dflag, DFlagRenderOrderChange) {
			renderOrderChange = true
		}
		if HasFlag(dflag, DFlagOpacityChange) {
			opacityChange = true
		}
		if HasFlag(dflag, DFlagVertexPositionChange) {
			vertexPositionsChange = true
		}
	} // 绘图顺序改变
	if drawOrderChange || renderOrderChange { // 渲染顺序才是我们需要的
		orders := m.Core.GetDrawableRenderOrders(m.Model.Moc.ModelPtr)
		for i, order := range orders {
			m.Model.Drawables[i].Order = order
		}
	} // 透明度改变
	if opacityChange {
		opacities := m.Core.GetDrawableOpacities(m.Model.Moc.ModelPtr)
		for i, opacity := range opacities {
			m.Model.Drawables[i].Opacity = opacity
		}
	} // 顶点变化
	if vertexPositionsChange {
		pos := m.Core.GetDrawableVertexPositions(m.Model.Moc.ModelPtr)
		for i, item := range pos {
			m.Model.Drawables[i].Pos = item
		}
	}
}

func (m *MotionManager) Draw(screen *ebiten.Image, scale float64) {
	// 临时排序进行渲染
	orderDs := make([]*Drawable, 0)
	for _, drawable := range m.Model.Drawables {
		orderDs = append(orderDs, drawable)
	}
	sort.Slice(orderDs, func(i, j int) bool {
		return orderDs[i].Order < orderDs[j].Order
	})
	vts := make([][]ebiten.Vertex, 0)
	for _, drawable := range orderDs {
		vts = append(vts, m.ToVertexes(drawable))
	}
	for i, drawable := range orderDs { // order用法太奇怪了，建议挪出Drawable
		if !HasFlag(drawable.DFlag, DFlagVisible) {
			continue
		}
		if len(drawable.Masks) > 0 {
			option := &ebiten.DrawTrianglesOptions{}
			// 清理 mask 并绘制遮罩
			m.Mask.Fill(color.RGBA{}) // 使用透明色覆盖
			for _, mask := range drawable.Masks {
				temp := m.Model.Drawables[mask]
				index := GetOrderDsIndex(orderDs, temp.Id)
				m.Mask.DrawTriangles(vts[index], temp.Idxs, temp.Image, option)
			}
			// 清理目标纹理 重新绘制目标纹理
			m.Src.Fill(color.RGBA{})
			m.Src.DrawTriangles(vts[i], drawable.Idxs, drawable.Image, option)
			// 进行合并
			//shaderOption := &ebiten.DrawTrianglesShaderOptions{}
			//shaderOption.Images[0] = m.Src
			//shaderOption.Images[1] = m.Mask
			//screen.DrawTrianglesShader(vts[i], drawable.Idxs, m.Shader, shaderOption)
			options := &ebiten.DrawRectShaderOptions{}
			options.Images[0] = m.Src
			options.Images[1] = m.Mask
			screen.DrawRectShader(int(m.W), int(m.H), m.Shader, options)
		} else {
			option := &ebiten.DrawTrianglesOptions{}
			screen.DrawTriangles(vts[i], drawable.Idxs, drawable.Image, option)
		}
	}
}

func GetOrderDsIndex(orderDs []*Drawable, id string) int {
	for i, drawable := range orderDs {
		if drawable.Id == id {
			return i
		}
	}
	panic(fmt.Sprintf("invalid id: %v", id))
}

func (m *MotionManager) ToVertexes(drawable *Drawable) []ebiten.Vertex {
	bound := drawable.Image.Bounds()
	w, h := float32(bound.Dx()), float32(bound.Dy())
	res := make([]ebiten.Vertex, 0)
	size := Max(m.W, m.H)
	for i := 0; i < len(drawable.Pos); i++ {
		// 主要注意绘图坐标系 y 轴反转
		// 注意最终图片是绘制出正方形的，而视口是长方形，要进行一定调整
		res = append(res, ebiten.Vertex{
			DstX:   (drawable.Pos[i].X+1)*size/2 - (size-WinW)/2,
			DstY:   (1-drawable.Pos[i].Y)*size/2 - (size-WinH)/2,
			SrcX:   drawable.Uvs[i].X * w,
			SrcY:   (1 - drawable.Uvs[i].Y) * h,
			ColorR: 1,
			ColorG: 1,
			ColorB: 1,
			ColorA: 1,
		})
	}
	return res
}

func NewMotionManager(core *Core, model *Model, w float32, h float32) *MotionManager {
	return &MotionManager{Core: core, Model: model, W: w, H: h,
		Mask: ebiten.NewImage(int(w), int(h)), Src: ebiten.NewImage(int(w), int(h)),
		Shader: OpenShader("mask.kage")}
}
