/*
@author: sk
@date: 2024/6/15
*/
package main

/*
#cgo CFLAGS: -I./cubism_sdk/include
#cgo LDFLAGS: -L./cubism_sdk/lib -lLive2DCubismCore

#include "Live2DCubismCore.h"
*/
import "C" // 采用静态链接，可以打包为一个文件，性能更好
import (
	"fmt"
	"path/filepath"
	"unsafe"

	"github.com/hajimehoshi/ebiten/v2"
)

// 数据对齐
const (
	AlignofMoc   = C.csmAlignofMoc
	AlignofModel = C.csmAlignofModel
)

// moc3 Version
const (
	MocVersionUnknown = C.csmMocVersion_Unknown
	MocVersion30      = C.csmMocVersion_30
	MocVersion33      = C.csmMocVersion_33
	MocVersion40      = C.csmMocVersion_40
	MocVersion42      = C.csmMocVersion_42
	MocVersion50      = C.csmMocVersion_50
)

type (
	Moc0   *C.csmMoc
	Model0 *C.csmModel
)

func GetVersion() string {
	code := uint32(C.csmGetVersion())
	major := code >> 24
	minor := (code >> 16) & 0xFF
	patch := code & 0xFFFF
	return fmt.Sprintf("%d.%d.%d", major, minor, patch)
}

// 获取支持的最新版本
func GetLatestMocVersion() uint32 {
	res := C.csmGetLatestMocVersion()
	return uint32(res)
}

// 获取当前数据版本
func GetMocVersion(data []byte) uint32 {
	addr := SliceToPtr(data)
	size := C.uint(len(data))
	res := C.csmGetMocVersion(addr, size)
	return uint32(res)
}

func LoadModel(path string) *Model {
	dir := filepath.Dir(path)
	// 加载入口资源
	modelData := &ModelData{}
	UnmarshalFile(path, modelData)
	ref := modelData.FileReferences
	// 加载其他关联资源
	physicData := &PhysicData{}
	if len(ref.Physics) > 0 { // 物理效果
		ref.Physics = filepath.Join(dir, ref.Physics)
		UnmarshalFile(ref.Physics, physicData)
	}
	poseData := &PoseData{}
	if len(ref.Pose) > 0 { // pose数据
		ref.Pose = filepath.Join(dir, ref.Pose)
		UnmarshalFile(ref.Pose, poseData)
	}
	displayData := &DisplayData{}
	if len(ref.DisplayInfo) > 0 { // 展示信息
		ref.DisplayInfo = filepath.Join(dir, ref.DisplayInfo)
		UnmarshalFile(ref.DisplayInfo, displayData)
	}
	expressionDatas := make([]*ExpressionData1, 0)
	for _, item := range ref.Expressions { // 表情信息
		item.File = filepath.Join(dir, item.File)
		expressionData := &ExpressionData1{}
		UnmarshalFile(item.File, expressionData)
		expressionData.Name = item.Name
		expressionDatas = append(expressionDatas, expressionData)
	}
	motionDatas := make(map[string][]*MotionData1)
	for name, motions := range ref.Motions { // 动作信息
		for _, motion := range motions {
			motion.File = filepath.Join(dir, motion.File)
			motionData := &MotionData1{}
			UnmarshalFile(motion.File, motionData)
			motionData.Data = motion
			motionDatas[name] = append(motionDatas[name], motionData)
		}
	}
	userData := &UserData0{}
	if len(ref.UserData) > 0 { // 用户自定义数据，一般没啥用
		ref.UserData = filepath.Join(dir, ref.UserData)
		UnmarshalFile(ref.UserData, userData)
	}
	// 转换路径，方便后面使用
	for i, texture := range ref.Textures {
		ref.Textures[i] = filepath.Join(dir, texture)
	}
	ref.Moc = filepath.Join(dir, ref.Moc)
	// 加载 moc文件
	moc := LoadMoc(ref.Moc)
	// 加载 drawable资源
	ds := GetDrawables(moc.Model, ref.Textures)
	// 转换 motion信息
	motions := make(map[string][]*Motion)
	for name, datas := range motionDatas {
		for _, data := range datas {
			motions[name] = append(motions[name], ToMotion(data))
		}
	}
	return &Model{
		RootDir:         dir,
		ModelData:       modelData,
		PhysicData:      physicData,
		PoseData:        poseData,
		DisplayData:     displayData,
		ExpressionDatas: expressionDatas,
		MotionDatas:     motionDatas,
		UserData:        userData,
		Moc:             moc,
		Drawables:       ds,
		Motions:         motions,
	}
}

func LoadMoc(path string) *Moc {
	moc := &Moc{}
	moc.MocBuff = AlignByte(ReadFile(path), AlignofMoc)
	// 完整性检查
	res := C.csmHasMocConsistency(SliceToPtr(moc.MocBuff), C.uint(len(moc.MocBuff)))
	Assert(res == 1, "moc not consistency")
	maxVersion := GetLatestMocVersion()
	currVersion := GetMocVersion(moc.MocBuff)
	Assert(currVersion != 0 && currVersion <= maxVersion, "core %v not support version %v", maxVersion, currVersion)
	// 装载 moc3文件
	moc.Moc = C.csmReviveMocInPlace(SliceToPtr(moc.MocBuff), C.uint(len(moc.MocBuff)))
	// 获取模型大小
	size := C.csmGetSizeofModel(moc.Moc)
	Assert(size > 0, "moc load fail")
	// 初始化模型
	moc.ModelBuff = AlignByte(make([]byte, size), AlignofModel)
	moc.Model = C.csmInitializeModelInPlace(moc.Moc, SliceToPtr(moc.ModelBuff), size)
	return moc
}

func GetDrawables(model Model0, textures []string) []*Drawable {
	// 获取有多少绘制组件
	count := int32(C.csmGetDrawableCount(model))
	// 获取这些组件信息
	cflags := PtrToSlice[uint8](unsafe.Pointer(C.csmGetDrawableConstantFlags(model)), count)
	dflags := PtrToSlice[uint8](unsafe.Pointer(C.csmGetDrawableDynamicFlags(model)), count)
	tIdxs := PtrToSlice[int32](unsafe.Pointer(C.csmGetDrawableTextureIndices(model)), count) // 纹理索引
	opacities := PtrToSlice[float32](unsafe.Pointer(C.csmGetDrawableOpacities(model)), count)
	orders := PtrToSlice[int32](unsafe.Pointer(C.csmGetDrawableRenderOrders(model)), count)
	// 获取每个绘制目标 顶点， uv 与索引，每个绘制对象由多个三角形组成
	vCounts := PtrToSlice[int32](unsafe.Pointer(C.csmGetDrawableVertexCounts(model)), count) // 每个绘制的顶点数
	iCounts := PtrToSlice[int32](unsafe.Pointer(C.csmGetDrawableIndexCounts(model)), count)  // 每个绘制对象的索引数
	pos := PtrToSlice2[Vector2](unsafe.Pointer(C.csmGetDrawableVertexPositions(model)), vCounts)
	uvs := PtrToSlice2[Vector2](unsafe.Pointer(C.csmGetDrawableVertexUvs(model)), vCounts)
	idxs := PtrToSlice2[uint16](unsafe.Pointer(C.csmGetDrawableIndices(model)), iCounts)
	// 获取 mask信息
	mCounts := PtrToSlice[int32](unsafe.Pointer(C.csmGetDrawableMaskCounts(model)), count) // 每个绘制的 mask 数目
	masks := PtrToSlice2[uint32](unsafe.Pointer(C.csmGetDrawableMasks(model)), mCounts)    // 使用那些 绘制对象当做遮罩
	// 获取 id 信息
	ids := make([]string, 0)
	idPtr := unsafe.Pointer(C.csmGetDrawableIds(model))
	for i := int32(0); i < count; i++ {
		// 每个指针占用 8 byte
		ptr := *(**byte)(unsafe.Pointer(uintptr(idPtr) + uintptr(i*8))) // 来回转换主要是写入类型信息
		ids = append(ids, PtrToStr(unsafe.Pointer(ptr)))
	}
	imgs := make(map[string]*ebiten.Image)
	for _, texture := range textures {
		if _, ok := imgs[texture]; ok {
			continue
		}
		imgs[texture] = OpenImage(texture)
	}
	res := make([]*Drawable, 0)
	for i := int32(0); i < count; i++ {
		res = append(res, &Drawable{
			Id:      ids[i],
			Texture: textures[tIdxs[i]],
			Image:   imgs[textures[tIdxs[i]]],
			Pos:     pos[i],
			Uvs:     uvs[i],
			Idxs:    idxs[i],
			CFlag:   cflags[i],
			DFlag:   dflags[i],
			Opacity: opacities[i],
			Masks:   masks[i],
			Order:   orders[i],
		})
	}
	return res
}

func GetCanvasInfo(model Model0) (*Vector2, *Vector2, float32) {
	var cSize C.csmVector2
	var cOrigin C.csmVector2
	var cPixelsPerUnit C.float
	C.csmReadCanvasInfo(model, &cSize, &cOrigin, &cPixelsPerUnit)
	return &Vector2{
			X: float32(cSize.X),
			Y: float32(cSize.Y),
		}, &Vector2{
			X: float32(cOrigin.X),
			Y: float32(cOrigin.Y),
		}, float32(cPixelsPerUnit)
}

func SetPartOpacity(model Model0, id string, value float32) {
	ptr := unsafe.Pointer(C.csmGetPartOpacities(model))
	idx := GetPartIdIndex(model, id) // 直接写入公共缓存区
	*(*float32)(unsafe.Pointer(uintptr(ptr) + uintptr(idx*4))) = value
}

func GetPartIdIndex(model Model0, id string) int32 {
	count := int32(C.csmGetPartCount(model))
	idPtr := unsafe.Pointer(C.csmGetPartIds(model))
	for i := int32(0); i < count; i++ {
		// 每个指针占用 8 byte
		ptr := *(**byte)(unsafe.Pointer(uintptr(idPtr) + uintptr(i*8))) // 来回转换主要是写入类型信息
		if PtrToStr(unsafe.Pointer(ptr)) == id {
			return i
		}
	}
	panic(fmt.Sprintf("id %s not found", id))
}

func GetParameterIdIndex(model Model0, id string) int32 {
	count := int32(C.csmGetParameterCount(model))
	idPtr := unsafe.Pointer(C.csmGetParameterIds(model))
	for i := int32(0); i < count; i++ {
		// 每个指针占用 8 byte
		ptr := *(**byte)(unsafe.Pointer(uintptr(idPtr) + uintptr(i*8))) // 来回转换主要是写入类型信息
		if PtrToStr(unsafe.Pointer(ptr)) == id {
			return i
		}
	}
	panic(fmt.Sprintf("id %s not found", id))
}

// 这里的参数值是控制多个关联对象的
func GetParameterValue(model Model0, id string) float32 {
	idx := GetParameterIdIndex(model, id)
	count := int32(C.csmGetParameterCount(model))
	vals := PtrToSlice[float32](unsafe.Pointer(C.csmGetParameterValues(model)), count)
	return vals[idx]
}

// 分别获取参数可取的最大/最小/默认值

func GetParameterMaximumValues(model Model0, id string) float32 {
	idx := GetParameterIdIndex(model, id)
	count := int32(C.csmGetParameterCount(model))
	vals := PtrToSlice[float32](unsafe.Pointer(C.csmGetParameterMaximumValues(model)), count)
	return vals[idx]
}

func GetParameterMinimumValues(model Model0, id string) float32 {
	idx := GetParameterIdIndex(model, id)
	count := int32(C.csmGetParameterCount(model))
	vals := PtrToSlice[float32](unsafe.Pointer(C.csmGetParameterMinimumValues(model)), count)
	return vals[idx]
}

func GetParameterDefaultValues(model Model0, id string) float32 {
	idx := GetParameterIdIndex(model, id)
	count := int32(C.csmGetParameterCount(model))
	vals := PtrToSlice[float32](unsafe.Pointer(C.csmGetParameterDefaultValues(model)), count)
	return vals[idx]
}

func SetParameterValue(model Model0, id string, value float32) {
	idx := GetParameterIdIndex(model, id)
	ptr := unsafe.Pointer(C.csmGetParameterValues(model))
	*(*float32)(unsafe.Pointer(uintptr(ptr) + uintptr(idx*4))) = value
}

func Update(model Model0) {
	C.csmResetDrawableDynamicFlags(model)
	C.csmUpdateModel(model)
}

func GetDynamicFlags(model Model0) []uint8 {
	count := int32(C.csmGetDrawableCount(model))
	return PtrToSlice[uint8](unsafe.Pointer(C.csmGetDrawableDynamicFlags(model)), count)
}

// 还有一个 csmGetDrawableDrawOrders 是界面上用户填写值无序使用
func GetDrawableRenderOrders(model Model0) []int32 {
	count := int32(C.csmGetDrawableCount(model))
	return PtrToSlice[int32](unsafe.Pointer(C.csmGetDrawableRenderOrders(model)), count)
}

func GetDrawableOpacities(model Model0) []float32 {
	count := int32(C.csmGetDrawableCount(model))
	return PtrToSlice[float32](unsafe.Pointer(C.csmGetDrawableOpacities(model)), count)
}

func GetDrawableVertexPositions(model Model0) [][]Vector2 {
	count := int32(C.csmGetDrawableCount(model))
	vCounts := PtrToSlice[int32](unsafe.Pointer(C.csmGetDrawableVertexCounts(model)), count) // 每个绘制的顶点数
	return PtrToSlice2[Vector2](unsafe.Pointer(C.csmGetDrawableVertexPositions(model)), vCounts)
}
