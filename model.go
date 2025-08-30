/*
@author: sk
@date: 2024/6/15
*/
package main

import "github.com/hajimehoshi/ebiten/v2"

type Model struct {
	RootDir         string
	ModelData       *ModelData
	ExpressionDatas []*ExpressionData1
	MotionDatas     map[string][]*MotionData1
	Moc             *Moc
	Drawables       []*Drawable
	Motions         map[string][]*Motion
	// 暂时没有用到的数据
	DisplayData *DisplayData
	PhysicData  *PhysicData
	PoseData    *PoseData
	UserData    *UserData0
}

type Moc struct {
	// 这些 byte空间由 c 占用，不能写入或提前释放
	Moc       Moc0
	MocBuff   []byte
	Model     Model0
	ModelBuff []byte
}

type Drawable struct {
	// 静态属性
	Id      string
	Texture string
	Image   *ebiten.Image
	Uvs     []Vector2
	Idxs    []uint16
	CFlag   uint8
	Masks   []uint32
	// 动态属性，每帧需要更新的属性
	DFlag   uint8
	Order   int32
	Opacity float32
	Pos     []Vector2
}

type Vector2 struct {
	X float32 `json:"X"`
	Y float32 `json:"Y"`
}

type Motion struct {
	Data   *MotionData1
	Curves []*Curve
}

type Curve struct {
	Data        *CurveData
	FadeInTime  float64
	FadeOutTime float64
	Segments    []*Segment
}

type Point struct {
	Time  float64
	Value float64
}

type Segment struct {
	Points []*Point
	Type   int
	Value  float64
}
