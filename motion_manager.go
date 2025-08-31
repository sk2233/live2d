/*
@author: sk
@date: 2024/6/15
*/
package main

import (
	"fmt"
	"image/color"
	"math/rand"
	"sort"

	"github.com/hajimehoshi/ebiten/v2"
)

type MotionManager struct {
	Model       *Model
	Motion      *Motion
	Timer       float64
	Loop        bool
	Shader      *ebiten.Shader
	AudioPlayer *AudioPlayer
	// shader中使用的图片必须等大小，这里必须要先把图片绘制到另一个图片上
	Mask *ebiten.Image
	Src  *ebiten.Image
}

func (m *MotionManager) PlayMotion(name string, loop bool) {
	motions := m.Model.Motions[name]
	idx := rand.Intn(len(motions))
	m.Motion = motions[idx] // 有多个动作进行随机
	m.Timer = 0
	m.Loop = loop
	if sound := m.Motion.Data.Data.Sound; len(sound) > 0 {
		m.AudioPlayer.Play(sound) // 只播放一次
	}
	fmt.Printf("name %s idx %d file %s\n", name, idx, m.Motion.Data.Data.File)
}

func (m *MotionManager) GetAllMotions() []string {
	names := make([]string, 0)
	for name := range m.Model.Motions {
		names = append(names, name)
	}
	return names
}

func (m *MotionManager) StopMotion() {
	m.Motion = nil
}

func (m *MotionManager) Update(delta float64) {
	m.UpdateMotion(delta)
	Update(m.Model.Moc.Model)
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
		segment := GetRightSegments(curve.Segments, m.Timer)
		if segment == nil { // 可能没有需要修改的参数
			continue
		}
		value := GetSegmentValue(segment, m.Timer)
		switch curve.Data.Target {
		case TargetPartOpacity:
			SetPartOpacity(m.Model.Moc.Model, curve.Data.Id, float32(value))
		case TargetParameter:
			oldValue := GetParameterValue(m.Model.Moc.Model, curve.Data.Id)
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
			SetParameterValue(m.Model.Moc.Model, curve.Data.Id, newValue)
		case TargetModel:
			// TODO
		default:
			panic(fmt.Sprintf("invalid target: %v", curve.Data.Target))
		}
	}
}

func (m *MotionManager) UpdateModel() {
	dflags := GetDynamicFlags(m.Model.Moc.Model)
	// 通过动态 flag判断任何一个有改变就就进行一次同步数据
	drawOrderChange := false
	renderOrderChange := false
	opacityChange := false
	vertexPositionsChange := false
	for i, dflag := range dflags {
		m.Model.Drawables[i].DFlag = dflag // 下面有使用，要更新上
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
		orders := GetDrawableRenderOrders(m.Model.Moc.Model)
		for i, order := range orders {
			m.Model.Drawables[i].Order = order
		}
	} // 透明度改变
	if opacityChange {
		opacities := GetDrawableOpacities(m.Model.Moc.Model)
		for i, opacity := range opacities {
			m.Model.Drawables[i].Opacity = opacity
		}
	} // 顶点变化
	if vertexPositionsChange {
		pos := GetDrawableVertexPositions(m.Model.Moc.Model)
		for i, item := range pos {
			m.Model.Drawables[i].Pos = item
		}
	}
}

func (m *MotionManager) Draw(screen *ebiten.Image) {
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
			// 最终绘制
			options := &ebiten.DrawRectShaderOptions{}
			options.Images[0] = m.Src
			options.Images[1] = m.Mask
			screen.DrawRectShader(int(Size.X), int(Size.Y), m.Shader, options)
		} else {
			option := &ebiten.DrawTrianglesOptions{}
			option.ColorM.Scale(1, 1, 1, float64(drawable.Opacity))
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
	size := min(Size.X, Size.Y)
	for i := 0; i < len(drawable.Pos); i++ {
		// 主要注意绘图坐标系 y 轴反转
		// 注意最终图片是绘制出正方形的，而视口是长方形，要进行一定调整
		// Uvs.XY  0~1
		if IsAzurLane { // IsAzurLane=true
			res = append(res, ebiten.Vertex{
				DstX:   (drawable.Pos[i].X*size + Origin.X) * Scale,
				DstY:   (-drawable.Pos[i].Y*size + Origin.Y) * Scale,
				SrcX:   drawable.Uvs[i].X * w,
				SrcY:   (1 - drawable.Uvs[i].Y) * h,
				ColorR: 1,
				ColorG: 1,
				ColorB: 1,
				ColorA: 1,
			})
		} else { // IsAzurLane=false Pos.XY  -0.5~0.5 一般情况
			res = append(res, ebiten.Vertex{
				DstX:   drawable.Pos[i].X*size + Origin.X,
				DstY:   -drawable.Pos[i].Y*size + Origin.Y,
				SrcX:   drawable.Uvs[i].X * w,
				SrcY:   (1 - drawable.Uvs[i].Y) * h,
				ColorR: 1,
				ColorG: 1,
				ColorB: 1,
				ColorA: 1,
			})
		}
	}
	return res
}

func NewMotionManager(model *Model) *MotionManager {
	return &MotionManager{Model: model,
		Mask: ebiten.NewImage(int(Size.X), int(Size.Y)), Src: ebiten.NewImage(int(Size.X), int(Size.Y)),
		Shader: OpenShader("mask.kage"), AudioPlayer: NewAudioPlayer(model.RootDir)}
}
