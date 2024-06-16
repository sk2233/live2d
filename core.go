/*
@author: sk
@date: 2024/6/15
*/
package main

import (
	"fmt"
	"path/filepath"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/hajimehoshi/ebiten/v2"
)

type Core struct {
	lib                           uintptr
	csmGetVersion                 func() uint32
	csmHasMocConsistency          func(uintptr, uint64) int64
	csmReviveMocInPlace           func(uintptr, uint64) uintptr
	csmGetSizeofModel             func(uintptr) uint64
	csmInitializeModelInPlace     func(uintptr, uintptr, uint64) uintptr
	csmGetDrawableCount           func(uintptr) uint64
	csmGetDrawableConstantFlags   func(uintptr) uintptr
	csmGetDrawableDynamicFlags    func(uintptr) uintptr
	csmGetDrawableTextureIndices  func(uintptr) uintptr
	csmGetDrawableOpacities       func(uintptr) uintptr
	csmGetDrawableVertexCounts    func(uintptr) uintptr
	csmGetDrawableVertexPositions func(uintptr) uintptr
	csmGetDrawableVertexUvs       func(uintptr) uintptr
	csmGetDrawableIndexCounts     func(uintptr) uintptr
	csmGetDrawableIndices         func(uintptr) uintptr
	csmGetDrawableMaskCounts      func(uintptr) uintptr
	csmGetDrawableMasks           func(uintptr) uintptr
	csmGetDrawableIds             func(uintptr) uintptr
	csmReadCanvasInfo             func(uintptr, uintptr, uintptr, uintptr)
	csmGetDrawableRenderOrders    func(uintptr) uintptr
	csmGetPartOpacities           func(uintptr) uintptr
	csmGetPartCount               func(uintptr) uint64
	csmGetPartIds                 func(uintptr) uintptr
	csmGetParameterCount          func(uintptr) uint64
	csmGetParameterIds            func(uintptr) uintptr
	csmGetParameterValues         func(uintptr) uintptr
	csmResetDrawableDynamicFlags  func(uintptr)
	csmUpdateModel                func(uintptr)
}

func NewCore(path string) *Core {
	// 加载外部 c 库
	lib, err := purego.Dlopen(path, purego.RTLD_NOW|purego.RTLD_GLOBAL)
	HandleErr(err)
	res := &Core{
		lib: lib,
	}
	// 绑定各个函数
	purego.RegisterLibFunc(&res.csmGetVersion, lib, "csmGetVersion")
	purego.RegisterLibFunc(&res.csmHasMocConsistency, lib, "csmHasMocConsistency")
	purego.RegisterLibFunc(&res.csmReviveMocInPlace, lib, "csmReviveMocInPlace")
	purego.RegisterLibFunc(&res.csmGetSizeofModel, lib, "csmGetSizeofModel")
	purego.RegisterLibFunc(&res.csmInitializeModelInPlace, lib, "csmInitializeModelInPlace")
	purego.RegisterLibFunc(&res.csmGetDrawableCount, lib, "csmGetDrawableCount")
	purego.RegisterLibFunc(&res.csmGetDrawableConstantFlags, lib, "csmGetDrawableConstantFlags")
	purego.RegisterLibFunc(&res.csmGetDrawableDynamicFlags, lib, "csmGetDrawableDynamicFlags")
	purego.RegisterLibFunc(&res.csmGetDrawableTextureIndices, lib, "csmGetDrawableTextureIndices")
	purego.RegisterLibFunc(&res.csmGetDrawableOpacities, lib, "csmGetDrawableOpacities")
	purego.RegisterLibFunc(&res.csmGetDrawableVertexCounts, lib, "csmGetDrawableVertexCounts")
	purego.RegisterLibFunc(&res.csmGetDrawableVertexPositions, lib, "csmGetDrawableVertexPositions")
	purego.RegisterLibFunc(&res.csmGetDrawableVertexUvs, lib, "csmGetDrawableVertexUvs")
	purego.RegisterLibFunc(&res.csmGetDrawableIndexCounts, lib, "csmGetDrawableIndexCounts")
	purego.RegisterLibFunc(&res.csmGetDrawableIndices, lib, "csmGetDrawableIndices")
	purego.RegisterLibFunc(&res.csmGetDrawableMaskCounts, lib, "csmGetDrawableMaskCounts")
	purego.RegisterLibFunc(&res.csmGetDrawableMasks, lib, "csmGetDrawableMasks")
	purego.RegisterLibFunc(&res.csmGetDrawableIds, lib, "csmGetDrawableIds")
	purego.RegisterLibFunc(&res.csmReadCanvasInfo, lib, "csmReadCanvasInfo")
	purego.RegisterLibFunc(&res.csmGetDrawableRenderOrders, lib, "csmGetDrawableRenderOrders")
	purego.RegisterLibFunc(&res.csmGetPartOpacities, lib, "csmGetPartOpacities")
	purego.RegisterLibFunc(&res.csmGetPartCount, lib, "csmGetPartCount")
	purego.RegisterLibFunc(&res.csmGetPartIds, lib, "csmGetPartIds")
	purego.RegisterLibFunc(&res.csmGetParameterCount, lib, "csmGetParameterCount")
	purego.RegisterLibFunc(&res.csmGetParameterIds, lib, "csmGetParameterIds")
	purego.RegisterLibFunc(&res.csmGetParameterValues, lib, "csmGetParameterValues")
	purego.RegisterLibFunc(&res.csmResetDrawableDynamicFlags, lib, "csmResetDrawableDynamicFlags")
	purego.RegisterLibFunc(&res.csmUpdateModel, lib, "csmUpdateModel")
	return res
}

func (c *Core) GetVersion() string {
	code := c.csmGetVersion()
	major := code >> 24
	minor := (code >> 16) & 0xFF
	patch := code & 0xFFFF
	return fmt.Sprintf("%d.%d.%d", major, minor, patch)
}

func (c *Core) LoadModel(path string) *Model {
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
	moc := c.LoadMoc(ref.Moc)
	// 加载 drawable资源
	ds := c.GetDrawables(moc.ModelPtr, ref.Textures)
	// 转换 motion信息
	motions := make(map[string][]*Motion)
	for name, datas := range motionDatas {
		for _, data := range datas {
			motions[name] = append(motions[name], ToMotion(data))
		}
	}
	return &Model{
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

func (c *Core) LoadMoc(path string) *Moc {
	moc := &Moc{}
	moc.MocBuff = ReadFile(path)
	// 完整性检查
	res := c.csmHasMocConsistency(SliceToPtr(moc.MocBuff), uint64(len(moc.MocBuff)))
	Assert(res == 1, "moc not consistency")
	// 装载 moc3文件
	moc.MocPtr = c.csmReviveMocInPlace(SliceToPtr(moc.MocBuff), uint64(len(moc.MocBuff)))
	Assert(moc.MocPtr != 0, "moc load fail")
	// 获取模型大小
	size := c.csmGetSizeofModel(moc.MocPtr)
	Assert(size > 0, "moc load fail")
	// 初始化模型
	moc.ModelBuff = make([]byte, size)
	moc.ModelPtr = c.csmInitializeModelInPlace(moc.MocPtr, SliceToPtr(moc.ModelBuff), size)
	Assert(moc.ModelPtr != 0, "init model fail")
	return moc
}

func (c *Core) GetDrawables(modelPtr uintptr, textures []string) []*Drawable {
	// 获取有多少绘制组件
	count := c.csmGetDrawableCount(modelPtr)
	// 获取这些组件信息
	cflags := PtrToSlice[uint8](c.csmGetDrawableConstantFlags(modelPtr), int(count))
	dflags := PtrToSlice[uint8](c.csmGetDrawableDynamicFlags(modelPtr), int(count))
	tIdxs := PtrToSlice[uint32](c.csmGetDrawableTextureIndices(modelPtr), int(count)) // 纹理索引
	opacities := PtrToSlice[float32](c.csmGetDrawableOpacities(modelPtr), int(count))
	orders := PtrToSlice[uint32](c.csmGetDrawableRenderOrders(modelPtr), int(count))
	// 获取每个绘制目标 顶点， uv 与索引，每个绘制对象由多个三角形组成
	vCounts := PtrToSlice[uint32](c.csmGetDrawableVertexCounts(modelPtr), int(count)) // 每个绘制的顶点数
	iCounts := PtrToSlice[uint32](c.csmGetDrawableIndexCounts(modelPtr), int(count))  // 每个绘制对象的索引数
	pos := make([][]Vector2, 0)
	uvs := make([][]Vector2, 0)
	idxs := make([][]uint16, 0)
	posPtr := c.csmGetDrawableVertexPositions(modelPtr)
	uvPtr := c.csmGetDrawableVertexUvs(modelPtr)
	idxPtr := c.csmGetDrawableIndices(modelPtr)
	for i := uint64(0); i < count; i++ {
		// 每个指针占用 8 byte
		pos = append(pos, unsafe.Slice(*(**Vector2)(unsafe.Pointer(posPtr + uintptr(i*8))), int(vCounts[i])))
		uvs = append(uvs, unsafe.Slice(*(**Vector2)(unsafe.Pointer(uvPtr + uintptr(i*8))), int(vCounts[i])))
		idxs = append(idxs, unsafe.Slice(*(**uint16)(unsafe.Pointer(idxPtr + uintptr(i*8))), int(iCounts[i])))
	}
	// 获取 mask信息
	mCounts := PtrToSlice[uint32](c.csmGetDrawableMaskCounts(modelPtr), int(count)) // 每个绘制的 mask 数目
	masks := make([][]uint32, 0)
	maskPtr := c.csmGetDrawableMasks(modelPtr)
	for i := uint64(0); i < count; i++ {
		// 每个指针占用 8 byte
		masks = append(masks, unsafe.Slice(*(**uint32)(unsafe.Pointer(maskPtr + uintptr(i*8))), int(mCounts[i])))
	}
	// 获取 id 信息
	ids := make([]string, 0)
	idPtr := c.csmGetDrawableIds(modelPtr)
	for i := uint64(0); i < count; i++ {
		// 每个指针占用 8 byte
		ptr := *(**byte)(unsafe.Pointer(idPtr + uintptr(i*8))) // 来回转换主要是写入类型信息
		ids = append(ids, PtrToStr(uintptr(unsafe.Pointer(ptr))))
	}
	imgs := make(map[string]*ebiten.Image)
	for _, texture := range textures {
		if _, ok := imgs[texture]; ok {
			continue
		}
		imgs[texture] = OpenImage(texture)
	}
	res := make([]*Drawable, 0)
	for i := uint64(0); i < count; i++ {
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

func (c *Core) GetCanvasInfo(modelPtr uintptr) (*Vector2, *Vector2, float32) {
	size := &Vector2{}
	origin := &Vector2{}
	pixelsPerUnit := float32(0)
	c.csmReadCanvasInfo(modelPtr, uintptr(unsafe.Pointer(size)), uintptr(unsafe.Pointer(origin)), uintptr(unsafe.Pointer(&pixelsPerUnit)))
	return size, origin, pixelsPerUnit
}

func (c *Core) SetPartOpacity(modelPtr uintptr, id string, value float32) {
	ptr := c.csmGetPartOpacities(modelPtr)
	idx := c.GetPartIdIndex(modelPtr, id) // 直接写入公共缓存区
	*(*float32)(unsafe.Pointer(ptr + uintptr(idx*4))) = value
}

func (c *Core) GetPartIdIndex(modelPtr uintptr, id string) uint64 {
	count := c.csmGetPartCount(modelPtr)
	idPtr := c.csmGetPartIds(modelPtr)
	for i := uint64(0); i < count; i++ {
		// 每个指针占用 8 byte
		ptr := *(**byte)(unsafe.Pointer(idPtr + uintptr(i*8))) // 来回转换主要是写入类型信息
		if PtrToStr(uintptr(unsafe.Pointer(ptr))) == id {
			return i
		}
	}
	panic(fmt.Sprintf("id %s not found", id))
}

func (c *Core) GetParameterIdIndex(modelPtr uintptr, id string) uint64 {
	count := c.csmGetParameterCount(modelPtr)
	idPtr := c.csmGetParameterIds(modelPtr)
	for i := uint64(0); i < count; i++ {
		// 每个指针占用 8 byte
		ptr := *(**byte)(unsafe.Pointer(idPtr + uintptr(i*8))) // 来回转换主要是写入类型信息
		if PtrToStr(uintptr(unsafe.Pointer(ptr))) == id {
			return i
		}
	}
	panic(fmt.Sprintf("id %s not found", id))
}

func (c *Core) GetParameterValue(modelPtr uintptr, id string) float32 {
	idx := c.GetParameterIdIndex(modelPtr, id)
	count := c.csmGetParameterCount(modelPtr)
	vals := PtrToSlice[float32](c.csmGetParameterValues(modelPtr), int(count))
	return vals[idx]
}

func (c *Core) SetParameterValue(modelPtr uintptr, id string, value float32) {
	idx := c.GetParameterIdIndex(modelPtr, id)
	ptr := c.csmGetParameterValues(modelPtr)
	*(*float32)(unsafe.Pointer(ptr + uintptr(idx*4))) = value
}

func (c *Core) Update(modelPtr uintptr) {
	c.csmResetDrawableDynamicFlags(modelPtr)
	c.csmUpdateModel(modelPtr)
}

func (c *Core) GetDynamicFlags(modelPtr uintptr) []uint8 {
	count := c.csmGetDrawableCount(modelPtr)
	return PtrToSlice[uint8](c.csmGetDrawableDynamicFlags(modelPtr), int(count))
}

func (c *Core) GetDrawableRenderOrders(modelPtr uintptr) []uint32 {
	count := c.csmGetDrawableCount(modelPtr)
	return PtrToSlice[uint32](c.csmGetDrawableRenderOrders(modelPtr), int(count))
}

func (c *Core) GetDrawableOpacities(modelPtr uintptr) []float32 {
	count := c.csmGetDrawableCount(modelPtr)
	return PtrToSlice[float32](c.csmGetDrawableOpacities(modelPtr), int(count))
}

func (c *Core) GetDrawableVertexPositions(modelPtr uintptr) [][]Vector2 {
	count := c.csmGetDrawableCount(modelPtr)
	vCounts := PtrToSlice[uint32](c.csmGetDrawableVertexCounts(modelPtr), int(count)) // 每个绘制的顶点数
	pos := make([][]Vector2, 0)
	posPtr := c.csmGetDrawableVertexPositions(modelPtr)
	for i := uint64(0); i < count; i++ {
		// 每个指针占用 8 byte
		pos = append(pos, unsafe.Slice(*(**Vector2)(unsafe.Pointer(posPtr + uintptr(i*8))), int(vCounts[i])))
	}
	return pos
}
