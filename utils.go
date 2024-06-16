/*
@author: sk
@date: 2024/6/15
*/
package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"unsafe"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

func HandleErr(err error) {
	if err != nil {
		panic(err)
	}
}

func ReadFile(path string) []byte {
	bs, err := os.ReadFile(path)
	HandleErr(err)
	return bs
}

func UnmarshalFile(path string, dst any) {
	bs := ReadFile(path)
	err := json.Unmarshal(bs, dst)
	HandleErr(err)
}

func SliceToPtr[T any](data []T) uintptr {
	return uintptr(unsafe.Pointer(&data[0]))
}

func Assert(val bool, msg string, args ...any) {
	if !val {
		panic(fmt.Sprintf(msg, args...))
	}
}

func PtrToSlice[T any](ptr uintptr, count int) []T {
	return unsafe.Slice((*T)(unsafe.Pointer(ptr)), count)
}

func PtrToStr(ptr uintptr) string {
	bs := PtrToSlice[byte](ptr, 32)
	for i := 0; i < len(bs); i++ {
		if bs[i] == '\x00' {
			return string(bs[:i])
		}
	}
	panic("invalid ptr to str")
}

func ToMotion(data *MotionData1) *Motion {
	curves := make([]*Curve, 0) // 暂时没有管音乐
	for _, item := range data.Curves {
		lastPoint := &Point{
			Time:  item.Segments[0],
			Value: item.Segments[1],
		}
		segments := make([]*Segment, 0)
		i := 2
		for i < len(item.Segments) {
			type0 := item.Segments[i]
			i++
			switch type0 {
			case CurveLinear:
				nextPoint := &Point{Time: item.Segments[i], Value: item.Segments[i+1]}
				segments = append(segments, &Segment{
					Points: []*Point{lastPoint, nextPoint},
					Type:   CurveLinear,
				})
				lastPoint = nextPoint
				i += 2
			case CurveBezier:
				nextPoint := &Point{Time: item.Segments[i+4], Value: item.Segments[i+5]}
				segments = append(segments, &Segment{
					Points: []*Point{
						lastPoint,
						{Time: item.Segments[i], Value: item.Segments[i+1]},
						{Time: item.Segments[i+2], Value: item.Segments[i+3]},
						nextPoint,
					},
					Type: CurveBezier,
				})
				lastPoint = nextPoint
				i += 6
			case CurveStepped:
				nextPoint := &Point{Time: item.Segments[i], Value: item.Segments[i+1]}
				segments = append(segments, &Segment{
					Points: []*Point{lastPoint},
					Type:   CurveStepped,
					Value:  nextPoint.Time,
				})
				lastPoint = nextPoint
				i += 2
			case CurveInverseStepped:
				nextPoint := &Point{Time: item.Segments[i], Value: item.Segments[i+1]}
				segments = append(segments, &Segment{
					Points: []*Point{lastPoint},
					Type:   CurveInverseStepped,
					Value:  lastPoint.Time,
				})
				lastPoint = nextPoint
				i += 2
			default:
				panic(fmt.Sprintf("invalid type0: %v", type0))
			}
		}
		curves = append(curves, &Curve{
			Data:        item,
			FadeInTime:  ElemOrDef(item.FadeInTime, -1),
			FadeOutTime: ElemOrDef(item.FadeOutTime, -1),
			Segments:    segments,
		})
	}
	return &Motion{
		Data:   data,
		Curves: curves,
	}
}

func ElemOrDef[T any](ptr *T, def T) T {
	if ptr == nil {
		return def
	}
	return *ptr
}

func GetFade(motion *Motion, timer float64) (float64, float64) {
	fadeIn := 1.0 // 没有渐入渐出时间就取立即值
	if motion.Data.Data.FadeInTime > 0 {
		fadeIn = GetEasingSine(timer / motion.Data.Data.FadeInTime)
	}
	fadeOut := 1.0
	if motion.Data.Data.FadeOutTime > 0 {
		fadeOut = GetEasingSine(timer / motion.Data.Data.FadeOutTime)
	}
	return fadeIn, fadeOut
}

func GetEasingSine(rate float64) float64 {
	if rate < 0.0 {
		return 0.0
	}
	if rate > 1.0 {
		return 1.0
	}
	return 0.5 - 0.5*math.Cos(rate*math.Pi)
}

func GetRightSegment(segments []*Segment, timer float64) *Segment {
	for _, segment := range segments {
		switch segment.Type {
		case CurveLinear:
			if segment.Points[0].Time <= timer && segment.Points[1].Time >= timer {
				return segment
			}
		case CurveBezier:
			if segment.Points[0].Time <= timer && segment.Points[3].Time >= timer {
				return segment
			}
		case CurveStepped:
			if segment.Points[0].Time <= timer && segment.Value >= timer {
				return segment
			}
		case CurveInverseStepped:
			if segment.Value <= timer && segment.Points[0].Time >= timer {
				return segment
			}
		default:
			panic(fmt.Sprintf("invalid type0: %v", segment.Type))
		}
	}
	return nil
}

func GetSegmentValue(segment *Segment, timer float64) float64 {
	switch segment.Type {
	case CurveLinear:
		rate := (timer - segment.Points[0].Time) / (segment.Points[1].Time - segment.Points[0].Time)
		return segment.Points[0].Value + rate*(segment.Points[1].Value-segment.Points[0].Value)
	case CurveBezier:
		rate := (timer - segment.Points[0].Time) / (segment.Points[3].Time - segment.Points[0].Time)
		// 多次取线性值
		p01 := LerpPoint(segment.Points[0], segment.Points[1], rate)
		p12 := LerpPoint(segment.Points[1], segment.Points[2], rate)
		p23 := LerpPoint(segment.Points[2], segment.Points[3], rate)
		p02 := LerpPoint(p01, p12, rate)
		p13 := LerpPoint(p12, p23, rate)
		return LerpPoint(p02, p13, rate).Value
	case CurveStepped, CurveInverseStepped:
		return segment.Points[0].Value
	default:
		panic(fmt.Sprintf("invalid type0: %v", segment.Type))
	}
}

func LerpPoint(p1 *Point, p2 *Point, rate float64) *Point {
	return &Point{
		Time:  p1.Time + rate*(p2.Time-p1.Time),
		Value: p1.Value + rate*(p2.Value-p1.Value),
	}
}

func HasFlag(flag uint8, mask uint8) bool {
	return flag&mask > 0
}

func OpenImage(path string) *ebiten.Image {
	res, _, err := ebitenutil.NewImageFromFile(path)
	HandleErr(err)
	return res
}

func Max[T float32](v1, v2 T) T {
	if v1 > v2 {
		return v1
	} else {
		return v2
	}
}

func OpenShader(path string) *ebiten.Shader {
	bs := ReadFile(path)
	res, err := ebiten.NewShader(bs)
	HandleErr(err)
	return res
}
