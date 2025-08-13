package go_param_table

import (
	"fmt"
	"slices"
	"unsafe"
)

// Whether or not any ParamTable's insert additional sanity checks on input/output operations
//   - true (default): recommended for develpoment, or any time your program may be stuck in an infinite loop or parameters are not behaving as expected
//   - false: if you want slightly faster speed in production, have already fully tested your param table with EnableDebug set to true, _AND_ your end users do not have direct access to your ParamTable
var EnableDebug bool = true

type (
	ParamCalc func(c *CalcInterface)
	PIdx_U64  uint16
	PIdx_I64  uint16
	PIdx_F64  uint16
	PIdx_Ptr  uint16
	PIdx_U32  uint16
	PIdx_I32  uint16
	PIdx_F32  uint16
	PIdx_U16  uint16
	PIdx_I16  uint16
	PIdx_U8   uint16
	PIdx_I8   uint16
	PIdx_Bool uint16
	PIdx_Calc uint16
)

const (
	typeU64 = iota
	typeI64
	typeF64
	typePtr
	typeU32
	typeI32
	typeF32
	typeU16
	typeI16
	typeU8
	typeI8
	typeBool
	typeCount
)

const (
	size64  uint32 = 8
	sizePtr uint32 = uint32(unsafe.Sizeof(unsafe.Pointer(nil)))
	size32  uint32 = 4
	size16  uint32 = 2
	size8   uint32 = 1
)

const (
	PIDX_NULL uint16 = 65535
)

var sizeTable = [typeCount]uint32{
	typeU64:  size64,
	typeI64:  size64,
	typeF64:  size64,
	typePtr:  sizePtr,
	typeU32:  size32,
	typeI32:  size32,
	typeF32:  size32,
	typeU16:  size16,
	typeI16:  size16,
	typeU8:   size8,
	typeI8:   size8,
	typeBool: size8,
}

type paramFlags uint64

const (
	_PFLAG_INIT paramFlags = 1 << iota
	_PFLAG_ALWAYS_UPDATE

	_PFLAG_BITS                    = iota
	_PFLAG_MASK                    = (1 << _PFLAG_BITS) - 1
	_PFLAG_CHUNK_BITS              = 64
	_PFLAG_SUB_PER_CHUNK           = _PFLAG_CHUNK_BITS / _PFLAG_BITS
	_PFLAG_SUB_PER_CHUNK_SHIFT     = 5
	_PFLAG_SUB_PER_CHUNK_MINUS_ONE = _PFLAG_SUB_PER_CHUNK - 1
)

func (f paramFlags) IsInit() bool {
	return f&_PFLAG_INIT == _PFLAG_INIT
}
func (f paramFlags) AlwaysUpdate() bool {
	return f&_PFLAG_ALWAYS_UPDATE == _PFLAG_ALWAYS_UPDATE
}

func getFlag(elemIdx uint16, blocks []paramFlags) paramFlags {
	bIdx := elemIdx >> _PFLAG_SUB_PER_CHUNK_SHIFT
	sIdx := elemIdx % _PFLAG_SUB_PER_CHUNK
	block := blocks[bIdx]
	return (block >> (sIdx * _PFLAG_BITS)) & _PFLAG_MASK
}

func setFlag(elemIdx uint16, blocks []paramFlags, val paramFlags) {
	bIdx := elemIdx >> _PFLAG_SUB_PER_CHUNK_SHIFT
	sIdx := elemIdx % _PFLAG_SUB_PER_CHUNK
	block := val << paramFlags(sIdx*_PFLAG_BITS)
	blocks[bIdx] |= block
}

func initFlagLen(elemCount uint16) int {
	return int(elemCount+_PFLAG_SUB_PER_CHUNK_MINUS_ONE) >> _PFLAG_SUB_PER_CHUNK_SHIFT
}

type paramHookups struct {
	parentsStart     uint32
	calcOutputsStart uint32
	childrenStart    uint32
	parentsLen       uint8
	calcOutputsLen   uint8
	calculation      PIdx_Calc
}

type ParamTable struct {
	values      []byte
	hookups     []paramHookups
	flags       []paramFlags
	children    []uint16
	parents     []uint16
	outputs     []uint16
	calcs       []ParamCalc
	byteOffsets [typeCount]uint32
	idxOffsets  [typeCount]uint16
}

func NewParamTable(typeU64End PIdx_U64, typeI64End PIdx_I64, typeF64End PIdx_F64, typePtrEnd PIdx_Ptr, typeU32End PIdx_U32, typeI32End PIdx_I32, typeF32End PIdx_F32, typeU16End PIdx_U16, typeI16End PIdx_I16, typeU8End PIdx_U8, typeI8End PIdx_I8, typeBoolEnd PIdx_Bool, calcsCount PIdx_Calc) ParamTable {
	if typeU64End > PIdx_U64(typeI64End) || typeI64End > PIdx_I64(typeF64End) || typeF64End > PIdx_F64(typePtrEnd) ||
		typePtrEnd > PIdx_Ptr(typeU32End) ||
		typeU32End > PIdx_U32(typeI32End) || typeI32End > PIdx_I32(typeF32End) || typeF32End > PIdx_F32(typeU16End) ||
		typeU16End > PIdx_U16(typeI16End) || typeI16End > PIdx_I16(typeU8End) ||
		typeU8End > PIdx_U8(typeI8End) || typeI8End > PIdx_I8(typeBoolEnd) {
		panic(`error: parameter table: indexes not in order: all parameter index ends MUST be in this EXACT order from smallest to largest:
	typeU64End <= typeI64End <= typeF64End <=
	typePtrEnd <=
	typeU32End <= typeI32End <= typeF32End <=
	typeU16End <= typeI16End <=
	typeU8End <= typeI8End <= typeBoolEnd
For an example template that fulfills this requirement, see the function body of 'paratable.TestParamTable(t *testing.T)' or the doc-comment of 'paratable.PARAM_TABLE_TEMPLATE_DOC_COMMENT'`)
	}
	var idxOffsets = [typeCount]uint16{
		typeU64:  0,
		typeI64:  uint16(typeU64End),
		typeF64:  uint16(typeI64End),
		typePtr:  uint16(typePtrEnd),
		typeU32:  uint16(typeF64End),
		typeI32:  uint16(typeU32End),
		typeF32:  uint16(typeI32End),
		typeU16:  uint16(typeF32End),
		typeI16:  uint16(typeU16End),
		typeU8:   uint16(typeI16End),
		typeI8:   uint16(typeU8End),
		typeBool: uint16(typeI8End),
	}
	var valuesIdxLen = typeBoolEnd
	var byteOffsets = [typeCount]uint32{
		typeU64:  0,
		typeI64:  uint32(typeU64End) * 8,
		typeF64:  uint32(typeI64End) * 8,
		typePtr:  uint32(typeF64End) * 8,
		typeU32:  (uint32(typeF64End) * 8) + (uint32(PIdx_F64(typePtrEnd)-typeF64End) * sizePtr),
		typeI32:  (uint32(typeF64End) * 8) + (uint32(PIdx_F64(typePtrEnd)-typeF64End) * sizePtr) + (uint32(PIdx_Ptr(typeU32End)-typePtrEnd) * 4),
		typeF32:  (uint32(typeF64End) * 8) + (uint32(PIdx_F64(typePtrEnd)-typeF64End) * sizePtr) + (uint32(PIdx_Ptr(typeI32End)-typePtrEnd) * 4),
		typeU16:  (uint32(typeF64End) * 8) + (uint32(PIdx_F64(typePtrEnd)-typeF64End) * sizePtr) + (uint32(PIdx_Ptr(typeF32End)-typePtrEnd) * 4),
		typeI16:  (uint32(typeF64End) * 8) + (uint32(PIdx_F64(typePtrEnd)-typeF64End) * sizePtr) + (uint32(PIdx_Ptr(typeF32End)-typePtrEnd) * 4) + (uint32(PIdx_F32(typeU16End)-typeF32End) * 2),
		typeU8:   (uint32(typeF64End) * 8) + (uint32(PIdx_F64(typePtrEnd)-typeF64End) * sizePtr) + (uint32(PIdx_Ptr(typeF32End)-typePtrEnd) * 4) + (uint32(PIdx_F32(typeI16End)-typeF32End) * 2),
		typeI8:   (uint32(typeF64End) * 8) + (uint32(PIdx_F64(typePtrEnd)-typeF64End) * sizePtr) + (uint32(PIdx_Ptr(typeF32End)-typePtrEnd) * 4) + (uint32(PIdx_F32(typeI16End)-typeF32End) * 2) + uint32(PIdx_I16(typeU8End)-typeI16End),
		typeBool: (uint32(typeF64End) * 8) + (uint32(PIdx_F64(typePtrEnd)-typeF64End) * sizePtr) + (uint32(PIdx_Ptr(typeF32End)-typePtrEnd) * 4) + (uint32(PIdx_F32(typeI16End)-typeF32End) * 2) + uint32(PIdx_I16(typeI8End)-typeI16End),
	}
	var valuesByteLen = byteOffsets[typeBool] + uint32(PIdx_I16(typeBoolEnd)-typeI16End)
	valuesSlice := make([]byte, valuesByteLen)
	hookupsSlice := make([]paramHookups, valuesIdxLen)
	calcsSlice := make([]ParamCalc, calcsCount)
	childrenSlice := make([]uint16, 1)
	parentsSlice := make([]uint16, 1)
	calcOutputs := make([]uint16, 1)
	flagsLen := initFlagLen(uint16(valuesIdxLen))
	flags := make([]paramFlags, flagsLen)
	return ParamTable{
		values:      valuesSlice,
		hookups:     hookupsSlice,
		flags:       flags,
		children:    childrenSlice,
		parents:     parentsSlice,
		outputs:     calcOutputs,
		calcs:       calcsSlice,
		byteOffsets: byteOffsets,
		idxOffsets:  idxOffsets,
	}
}

func (t *ParamTable) TotalMemoryFootprint() uintptr {
	size := unsafe.Sizeof(*t)
	size += uintptr(cap(t.values))
	size += uintptr(cap(t.hookups)) * unsafe.Sizeof(paramHookups{})
	size += uintptr(cap(t.flags)) * unsafe.Sizeof(paramFlags(0))
	size += uintptr(cap(t.children)) * 2
	size += uintptr(cap(t.outputs)) * 2
	size += uintptr(cap(t.calcs)) * unsafe.Sizeof((ParamCalc)(nil))
	return size
}

func (t *ParamTable) checkInit(idx uint16) {
	if EnableDebug {
		f := getFlag(idx, t.flags)
		if !f.IsInit() {
			panic(fmt.Sprintf("error: parameter table: parameter index %d was never initialized", idx))
		}
	}
}

func (t *ParamTable) checkIdxType(idx uint16, name string, validType int, final bool, canBeDerived bool) {
	if EnableDebug {
		if idx >= uint16(len(t.hookups)) {
			panic(fmt.Sprintf("error: parameter table: index %d is outside bounds of parameter list (len %d)", idx, len(t.hookups)))
		}
		if final {
			if idx < t.idxOffsets[validType] {
				panic(fmt.Sprintf("error: parameter table: index %d is not a %s value: %s values are in range [%d, %d)", idx, name, name, t.idxOffsets[validType], len(t.hookups)))
			}
		} else {
			if idx < t.idxOffsets[validType] && idx >= t.idxOffsets[validType+1] {
				panic(fmt.Sprintf("error: parameter table: index %d is not a %s value: %s values are in range [%d, %d)", idx, name, name, t.idxOffsets[validType], t.idxOffsets[validType+1]))
			}
		}
		if !canBeDerived {
			if t.hookups[idx].parentsStart != 0 || t.hookups[idx].parentsLen != 0 || t.hookups[idx].calcOutputsStart != 0 || t.hookups[idx].calcOutputsLen != 0 {
				panic(fmt.Sprintf("error: parameter table: index %d is a derived value (has parents and calculation func), cannot update directly", idx))
			}
		}
	}
}

func (t *ParamTable) getBytePtr(idx uint16, typeIdx int) (ptr *byte, subIdx uint16) {
	subIdx = idx - t.idxOffsets[typeIdx]
	memOffset := t.byteOffsets[typeIdx] + (uint32(subIdx) * sizeTable[typeIdx])
	return &t.values[memOffset], subIdx
}

func (t *ParamTable) Get_U8(idx PIdx_U8) uint8 {
	_idx := uint16(idx)
	t.checkIdxType(_idx, "Uint8", typeU8, false, true)
	t.checkInit(_idx)
	memPtr, _ := t.getBytePtr(_idx, typeU8)
	return *memPtr
}

func (t *ParamTable) Get_I8(idx PIdx_I8) int8 {
	_idx := uint16(idx)
	t.checkIdxType(_idx, "Int8", typeI8, false, true)
	t.checkInit(_idx)
	memPtr, _ := t.getBytePtr(_idx, typeI8)
	return *(*int8)(unsafe.Pointer(memPtr))
}

func (t *ParamTable) Get_Bool(idx PIdx_Bool) bool {
	_idx := uint16(idx)
	t.checkIdxType(_idx, "Bool", typeBool, true, true)
	t.checkInit(_idx)
	memPtr, _ := t.getBytePtr(_idx, typeBool)
	return *(*bool)(unsafe.Pointer(memPtr))
}

func (t *ParamTable) Get_U16(idx PIdx_U16) uint16 {
	_idx := uint16(idx)
	t.checkIdxType(_idx, "Uint16", typeU16, false, true)
	t.checkInit(_idx)
	memPtr, _ := t.getBytePtr(_idx, typeU16)
	return *(*uint16)(unsafe.Pointer(memPtr))
}

func (t *ParamTable) Get_I16(idx PIdx_I16) int16 {
	_idx := uint16(idx)
	t.checkIdxType(_idx, "Int16", typeI16, false, true)
	t.checkInit(_idx)
	memPtr, _ := t.getBytePtr(_idx, typeI16)
	return *(*int16)(unsafe.Pointer(memPtr))
}

func (t *ParamTable) Get_U32(idx PIdx_U32) uint32 {
	_idx := uint16(idx)
	t.checkIdxType(_idx, "Uint32", typeU32, false, true)
	t.checkInit(_idx)
	memPtr, _ := t.getBytePtr(_idx, typeU32)
	return *(*uint32)(unsafe.Pointer(memPtr))
}

func (t *ParamTable) Get_I32(idx PIdx_I32) int32 {
	_idx := uint16(idx)
	t.checkIdxType(_idx, "Int32", typeI32, false, true)
	t.checkInit(_idx)
	memPtr, _ := t.getBytePtr(_idx, typeI32)
	return *(*int32)(unsafe.Pointer(memPtr))
}

func (t *ParamTable) Get_F32(idx PIdx_F32) float32 {
	_idx := uint16(idx)
	t.checkIdxType(_idx, "Float32", typeF32, false, true)
	t.checkInit(_idx)
	memPtr, _ := t.getBytePtr(_idx, typeF32)
	return *(*float32)(unsafe.Pointer(memPtr))
}

func (t *ParamTable) Get_U64(idx PIdx_U64) uint64 {
	_idx := uint16(idx)
	t.checkIdxType(_idx, "Uint64", typeU64, false, true)
	t.checkInit(_idx)
	memPtr, _ := t.getBytePtr(_idx, typeU64)
	return *(*uint64)(unsafe.Pointer(memPtr))
}

func (t *ParamTable) Get_I64(idx PIdx_I64) int64 {
	_idx := uint16(idx)
	t.checkIdxType(_idx, "Int64", typeI64, false, true)
	t.checkInit(_idx)
	memPtr, _ := t.getBytePtr(_idx, typeI64)
	return *(*int64)(unsafe.Pointer(memPtr))
}

func (t *ParamTable) Get_F64(idx PIdx_F64) float64 {
	_idx := uint16(idx)
	t.checkIdxType(_idx, "Float64", typeF64, false, true)
	t.checkInit(_idx)
	memPtr, _ := t.getBytePtr(_idx, typeF64)
	return *(*float64)(unsafe.Pointer(memPtr))
}

func (t *ParamTable) Get_Ptr(idx PIdx_Ptr) unsafe.Pointer {
	_idx := uint16(idx)
	t.checkIdxType(_idx, "unsafe.Pointer", typePtr, false, true)
	t.checkInit(_idx)
	memPtr, _ := t.getBytePtr(_idx, typePtr)
	return unsafe.Pointer(memPtr)
}

func (t *ParamTable) set_U8(idx uint16, val uint8, canBeDerived bool, prevIdxs []uint16) (newPrevIdxs []uint16) {
	t.checkIdxType(idx, "Uint8", typeU8, false, canBeDerived)
	f := getFlag(idx, t.flags)
	memPtr, _ := t.getBytePtr(idx, typeU8)
	oldVal := *memPtr
	*memPtr = val
	newPrevIdxs = prevIdxs
	if f.AlwaysUpdate() || oldVal != val {
		newPrevIdxs = t.updateChildren(idx, newPrevIdxs)
	}
	return
}

func (t *ParamTable) set_I8(idx uint16, val int8, canBeDerived bool, prevIdxs []uint16) (newPrevIdxs []uint16) {
	t.checkIdxType(idx, "Int8", typeI8, false, canBeDerived)
	f := getFlag(idx, t.flags)
	memPtr, _ := t.getBytePtr(idx, typeI8)
	valPtr := (*int8)(unsafe.Pointer(memPtr))
	oldVal := *valPtr
	*valPtr = val
	newPrevIdxs = prevIdxs
	if f.AlwaysUpdate() || oldVal != val {
		newPrevIdxs = t.updateChildren(idx, newPrevIdxs)
	}
	return
}

func (t *ParamTable) set_Bool(idx uint16, val bool, canBeDerived bool, prevIdxs []uint16) (newPrevIdxs []uint16) {
	t.checkIdxType(idx, "Bool", typeBool, true, canBeDerived)
	f := getFlag(idx, t.flags)
	memPtr, _ := t.getBytePtr(idx, typeBool)
	valPtr := (*bool)(unsafe.Pointer(memPtr))
	oldVal := *valPtr
	*valPtr = val
	newPrevIdxs = prevIdxs
	if f.AlwaysUpdate() || oldVal != val {
		newPrevIdxs = t.updateChildren(idx, newPrevIdxs)
	}
	return
}

func (t *ParamTable) set_U16(idx uint16, val uint16, canBeDerived bool, prevIdxs []uint16) (newPrevIdxs []uint16) {
	t.checkIdxType(idx, "Uint16", typeU16, false, canBeDerived)
	f := getFlag(idx, t.flags)
	memPtr, _ := t.getBytePtr(idx, typeU16)
	valPtr := (*uint16)(unsafe.Pointer(memPtr))
	oldVal := *valPtr
	*valPtr = val
	newPrevIdxs = prevIdxs
	if f.AlwaysUpdate() || oldVal != val {
		newPrevIdxs = t.updateChildren(idx, newPrevIdxs)
	}
	return
}

func (t *ParamTable) set_I16(idx uint16, val int16, canBeDerived bool, prevIdxs []uint16) (newPrevIdxs []uint16) {
	t.checkIdxType(idx, "Int16", typeI16, false, canBeDerived)
	f := getFlag(idx, t.flags)
	memPtr, _ := t.getBytePtr(idx, typeI16)
	valPtr := (*int16)(unsafe.Pointer(memPtr))
	oldVal := *valPtr
	*valPtr = val
	newPrevIdxs = prevIdxs
	if f.AlwaysUpdate() || oldVal != val {
		newPrevIdxs = t.updateChildren(idx, newPrevIdxs)
	}
	return
}

func (t *ParamTable) set_U32(idx uint16, val uint32, canBeDerived bool, prevIdxs []uint16) (newPrevIdxs []uint16) {
	t.checkIdxType(idx, "Uint32", typeU32, false, canBeDerived)
	f := getFlag(idx, t.flags)
	memPtr, _ := t.getBytePtr(idx, typeU32)
	valPtr := (*uint32)(unsafe.Pointer(memPtr))
	oldVal := *valPtr
	*valPtr = val
	newPrevIdxs = prevIdxs
	if f.AlwaysUpdate() || oldVal != val {
		newPrevIdxs = t.updateChildren(idx, newPrevIdxs)
	}
	return
}

func (t *ParamTable) set_I32(idx uint16, val int32, canBeDerived bool, prevIdxs []uint16) (newPrevIdxs []uint16) {
	t.checkIdxType(idx, "Int32", typeI32, false, canBeDerived)
	f := getFlag(idx, t.flags)
	memPtr, _ := t.getBytePtr(idx, typeI32)
	valPtr := (*int32)(unsafe.Pointer(memPtr))
	oldVal := *valPtr
	*valPtr = val
	newPrevIdxs = prevIdxs
	if f.AlwaysUpdate() || oldVal != val {
		newPrevIdxs = t.updateChildren(idx, newPrevIdxs)
	}
	return
}

func (t *ParamTable) set_F32(idx uint16, val float32, canBeDerived bool, prevIdxs []uint16) (newPrevIdxs []uint16) {
	t.checkIdxType(idx, "Float32", typeF32, false, canBeDerived)
	f := getFlag(idx, t.flags)
	memPtr, _ := t.getBytePtr(idx, typeF32)
	valPtr := (*float32)(unsafe.Pointer(memPtr))
	oldVal := *valPtr
	*valPtr = val
	newPrevIdxs = prevIdxs
	if f.AlwaysUpdate() || oldVal != val {
		newPrevIdxs = t.updateChildren(idx, newPrevIdxs)
	}
	return
}

func (t *ParamTable) set_U64(idx uint16, val uint64, canBeDerived bool, prevIdxs []uint16) (newPrevIdxs []uint16) {
	t.checkIdxType(idx, "Uint64", typeU64, false, canBeDerived)
	f := getFlag(idx, t.flags)
	memPtr, _ := t.getBytePtr(idx, typeU64)
	valPtr := (*uint64)(unsafe.Pointer(memPtr))
	oldVal := *valPtr
	*valPtr = val
	newPrevIdxs = prevIdxs
	if f.AlwaysUpdate() || oldVal != val {
		newPrevIdxs = t.updateChildren(idx, newPrevIdxs)
	}
	return
}

func (t *ParamTable) set_I64(idx uint16, val int64, canBeDerived bool, prevIdxs []uint16) (newPrevIdxs []uint16) {
	t.checkIdxType(idx, "Int64", typeI64, false, canBeDerived)
	f := getFlag(idx, t.flags)
	memPtr, _ := t.getBytePtr(idx, typeI64)
	valPtr := (*int64)(unsafe.Pointer(memPtr))
	oldVal := *valPtr
	*valPtr = val
	newPrevIdxs = prevIdxs
	if f.AlwaysUpdate() || oldVal != val {
		newPrevIdxs = t.updateChildren(idx, newPrevIdxs)
	}
	return
}

func (t *ParamTable) set_F64(idx uint16, val float64, canBeDerived bool, prevIdxs []uint16) (newPrevIdxs []uint16) {
	t.checkIdxType(idx, "Float64", typeF64, false, canBeDerived)
	f := getFlag(idx, t.flags)
	memPtr, _ := t.getBytePtr(idx, typeF64)
	valPtr := (*float64)(unsafe.Pointer(memPtr))
	oldVal := *valPtr
	*valPtr = val
	newPrevIdxs = prevIdxs
	if f.AlwaysUpdate() || oldVal != val {
		newPrevIdxs = t.updateChildren(idx, newPrevIdxs)
	}
	return
}

func (t *ParamTable) set_Ptr(idx uint16, val unsafe.Pointer, canBeDerived bool, prevIdxs []uint16) (newPrevIdxs []uint16) {
	t.checkIdxType(idx, "unsafe.Pointer", typePtr, false, canBeDerived)
	f := getFlag(idx, t.flags)
	memPtr, _ := t.getBytePtr(idx, typePtr)
	valPtr := (*unsafe.Pointer)(unsafe.Pointer(memPtr))
	oldVal := *valPtr
	*valPtr = val
	newPrevIdxs = prevIdxs
	if f.AlwaysUpdate() || oldVal != val {
		newPrevIdxs = t.updateChildren(idx, newPrevIdxs)
	}
	return
}

func (t *ParamTable) SetRoot_U8(idx PIdx_U8, val uint8) {
	_idx := uint16(idx)
	prev := []uint16{_idx}
	t.checkInit(_idx)
	t.set_U8(_idx, val, false, prev)
}

func (t *ParamTable) SetRoot_I8(idx PIdx_I8, val int8) {
	_idx := uint16(idx)
	prev := []uint16{_idx}
	t.checkInit(_idx)
	t.set_I8(_idx, val, false, prev)
}

func (t *ParamTable) SetRoot_Bool(idx PIdx_Bool, val bool) {
	_idx := uint16(idx)
	prev := []uint16{_idx}
	t.checkInit(_idx)
	t.set_Bool(_idx, val, false, prev)
}

func (t *ParamTable) SetRoot_U16(idx PIdx_U16, val uint16) {
	_idx := uint16(idx)
	prev := []uint16{_idx}
	t.checkInit(_idx)
	t.set_U16(_idx, val, false, prev)
}

func (t *ParamTable) SetRoot_I16(idx PIdx_I16, val int16) {
	_idx := uint16(idx)
	prev := []uint16{_idx}
	t.checkInit(_idx)
	t.set_I16(_idx, val, false, prev)
}

func (t *ParamTable) SetRoot_U32(idx PIdx_U32, val uint32) {
	_idx := uint16(idx)
	prev := []uint16{_idx}
	t.set_U32(_idx, val, false, prev)
}

func (t *ParamTable) SetRoot_I32(idx PIdx_I32, val int32) {
	_idx := uint16(idx)
	prev := []uint16{_idx}
	t.set_I32(_idx, val, false, prev)
}

func (t *ParamTable) SetRoot_F32(idx PIdx_F32, val float32) {
	_idx := uint16(idx)
	prev := []uint16{_idx}
	t.set_F32(_idx, val, false, prev)
}

func (t *ParamTable) SetRoot_U64(idx PIdx_U64, val uint64) {
	_idx := uint16(idx)
	prev := []uint16{_idx}
	t.set_U64(_idx, val, false, prev)
}

func (t *ParamTable) SetRoot_I64(idx PIdx_I64, val int64) {
	_idx := uint16(idx)
	prev := []uint16{_idx}
	t.set_I64(_idx, val, false, prev)
}

func (t *ParamTable) SetRoot_F64(idx PIdx_F64, val float64) {
	_idx := uint16(idx)
	prev := []uint16{_idx}
	t.set_F64(_idx, val, false, prev)
}

func (t *ParamTable) SetRoot_Ptr(idx PIdx_Ptr, val unsafe.Pointer) {
	_idx := uint16(idx)
	prev := []uint16{_idx}
	t.set_Ptr(_idx, val, false, prev)
}

func (t *ParamTable) InitRoot_U8(idx PIdx_U8, val uint8, alwaysUpdate bool) {
	_idx := uint16(idx)
	prev := []uint16{_idx}
	f := _PFLAG_INIT
	if alwaysUpdate {
		f |= _PFLAG_ALWAYS_UPDATE
	}
	setFlag(_idx, t.flags, f)
	t.set_U8(_idx, val, false, prev)
}

func (t *ParamTable) InitRoot_I8(idx PIdx_I8, val int8, alwaysUpdate bool) {
	_idx := uint16(idx)
	prev := []uint16{_idx}
	f := _PFLAG_INIT
	if alwaysUpdate {
		f |= _PFLAG_ALWAYS_UPDATE
	}
	setFlag(_idx, t.flags, f)
	t.set_I8(_idx, val, false, prev)
}

func (t *ParamTable) InitRoot_Bool(idx PIdx_Bool, val bool, alwaysUpdate bool) {
	_idx := uint16(idx)
	prev := []uint16{_idx}
	f := _PFLAG_INIT
	if alwaysUpdate {
		f |= _PFLAG_ALWAYS_UPDATE
	}
	setFlag(_idx, t.flags, f)
	t.set_Bool(_idx, val, false, prev)
}

func (t *ParamTable) InitRoot_U16(idx PIdx_U16, val uint16, alwaysUpdate bool) {
	_idx := uint16(idx)
	prev := []uint16{_idx}
	f := _PFLAG_INIT
	if alwaysUpdate {
		f |= _PFLAG_ALWAYS_UPDATE
	}
	setFlag(_idx, t.flags, f)
	t.set_U16(_idx, val, false, prev)
}

func (t *ParamTable) InitRoot_I16(idx PIdx_I16, val int16, alwaysUpdate bool) {
	_idx := uint16(idx)
	prev := []uint16{_idx}
	f := _PFLAG_INIT
	if alwaysUpdate {
		f |= _PFLAG_ALWAYS_UPDATE
	}
	setFlag(_idx, t.flags, f)
	t.set_I16(_idx, val, false, prev)
}

func (t *ParamTable) InitRoot_U32(idx PIdx_U32, val uint32, alwaysUpdate bool) {
	_idx := uint16(idx)
	prev := []uint16{_idx}
	f := _PFLAG_INIT
	if alwaysUpdate {
		f |= _PFLAG_ALWAYS_UPDATE
	}
	setFlag(_idx, t.flags, f)
	t.set_U32(_idx, val, false, prev)
}

func (t *ParamTable) InitRoot_I32(idx PIdx_I32, val int32, alwaysUpdate bool) {
	_idx := uint16(idx)
	prev := []uint16{_idx}
	f := _PFLAG_INIT
	if alwaysUpdate {
		f |= _PFLAG_ALWAYS_UPDATE
	}
	setFlag(_idx, t.flags, f)
	t.set_I32(_idx, val, false, prev)
}

func (t *ParamTable) InitRoot_F32(idx PIdx_F32, val float32, alwaysUpdate bool) {
	_idx := uint16(idx)
	prev := []uint16{_idx}
	f := _PFLAG_INIT
	if alwaysUpdate {
		f |= _PFLAG_ALWAYS_UPDATE
	}
	setFlag(_idx, t.flags, f)
	t.set_F32(_idx, val, false, prev)
}

func (t *ParamTable) InitRoot_U64(idx PIdx_U64, val uint64, alwaysUpdate bool) {
	_idx := uint16(idx)
	prev := []uint16{_idx}
	f := _PFLAG_INIT
	if alwaysUpdate {
		f |= _PFLAG_ALWAYS_UPDATE
	}
	setFlag(_idx, t.flags, f)
	t.set_U64(_idx, val, false, prev)
}

func (t *ParamTable) InitRoot_I64(idx PIdx_I64, val int64, alwaysUpdate bool) {
	_idx := uint16(idx)
	prev := []uint16{_idx}
	f := _PFLAG_INIT
	if alwaysUpdate {
		f |= _PFLAG_ALWAYS_UPDATE
	}
	setFlag(_idx, t.flags, f)
	t.set_I64(_idx, val, false, prev)
}

func (t *ParamTable) InitRoot_F64(idx PIdx_F64, val float64, alwaysUpdate bool) {
	_idx := uint16(idx)
	prev := []uint16{_idx}
	f := _PFLAG_INIT
	if alwaysUpdate {
		f |= _PFLAG_ALWAYS_UPDATE
	}
	setFlag(_idx, t.flags, f)
	t.set_F64(_idx, val, false, prev)
}

func (t *ParamTable) InitRoot_Ptr(idx PIdx_F64, val unsafe.Pointer, alwaysUpdate bool) {
	_idx := uint16(idx)
	prev := []uint16{_idx}
	f := _PFLAG_INIT
	if alwaysUpdate {
		f |= _PFLAG_ALWAYS_UPDATE
	}
	setFlag(_idx, t.flags, f)
	t.set_Ptr(_idx, val, false, prev)
}

func (t *ParamTable) initDerivedHookups(idx uint16, alwaysUpdate bool, calcIdx PIdx_Calc, parents []uint16, outputs []uint16) {
	t.hookups[idx].calculation = calcIdx
	t.hookups[idx].parentsStart = uint32(len(t.parents))
	t.hookups[idx].calcOutputsStart = uint32(len(t.outputs))
	f := _PFLAG_INIT
	if alwaysUpdate {
		f |= _PFLAG_ALWAYS_UPDATE
	}
	setFlag(idx, t.flags, f)
	if EnableDebug {
		if len(parents) > 255 {
			panic(fmt.Sprintf("derived values can only have a maximum of 255 parents (calculation inputs), got parent len %d", len(parents)))
		}
		if len(outputs) > 255 {
			panic(fmt.Sprintf("derived values can only have a maximum of 255 calculation outputs, got output len %d", len(outputs)))
		}
	}
	t.hookups[idx].parentsLen = uint8(len(parents))
	t.hookups[idx].calcOutputsLen = uint8(len(outputs))
	t.parents = append(t.parents, parents...)
	t.outputs = append(t.outputs, outputs...)
nextParent:
	for _, parent := range parents {
		if t.hookups[parent].childrenStart == 0 {
			newStart := uint32(len(t.children))
			t.hookups[parent].childrenStart = newStart
			t.children = append(t.children, idx, 0)
		} else {
			parentChildrenEnd := t.hookups[parent].childrenStart
			for {
				if t.children[parentChildrenEnd] == 0 {
					break
				}
				if t.children[parentChildrenEnd] == idx {
					continue nextParent
				}
				parentChildrenEnd += 1
			}
			t.children = slices.Insert(t.children, int(parentChildrenEnd), idx)
			for i := range t.hookups {
				if t.hookups[i].childrenStart >= parentChildrenEnd {
					t.hookups[i].childrenStart += 1
				}
			}
		}
	}
	calc := t.getCalc(calcIdx)
	recalc := CalcInterface{table: t, inputs: parents, outputs: outputs}
	calc(&recalc)
	t.updateChildren(idx, []uint16{idx})
}

func (t *ParamTable) getCalc(calcIdx PIdx_Calc) ParamCalc {
	if EnableDebug {
		if t.calcs[calcIdx] == nil {
			panic(fmt.Sprintf("error: parameter table: calc index %d has not been registered", calcIdx))
		}
	}
	return t.calcs[calcIdx]
}

func (t *ParamTable) RegisterCalc(calcIdx PIdx_Calc, calc ParamCalc) {
	if EnableDebug {
		if calcIdx > PIdx_Calc(len(t.calcs)) {
			panic(fmt.Sprintf("error: parameter table: calc index %d is outside bounds of calc list (len %d)", calcIdx, uint16(len(t.calcs))))
		}
		if t.calcs[calcIdx] != nil {
			panic(fmt.Sprintf("error: parameter table: calc index %d is already registered", calcIdx))
		}
	}
	t.calcs[calcIdx] = calc
}

func (t *ParamTable) InitDerived_U8(idx PIdx_U8, alwaysUpdate bool, calcIdx PIdx_Calc, inputs []uint16, outputs []uint16) {
	t.checkIdxType(uint16(idx), "Uint8", typeU8, false, true)
	t.initDerivedHookups(uint16(idx), alwaysUpdate, calcIdx, inputs, outputs)
}

func (t *ParamTable) InitDerived_I8(idx PIdx_I8, alwaysUpdate bool, calcIdx PIdx_Calc, inputs []uint16, outputs []uint16) {
	t.checkIdxType(uint16(idx), "Int8", typeI8, false, true)
	t.initDerivedHookups(uint16(idx), alwaysUpdate, calcIdx, inputs, outputs)
}

func (t *ParamTable) InitDerived_Bool(idx PIdx_Bool, alwaysUpdate bool, calcIdx PIdx_Calc, inputs []uint16, outputs []uint16) {
	t.checkIdxType(uint16(idx), "Bool", typeBool, true, true)
	t.initDerivedHookups(uint16(idx), alwaysUpdate, calcIdx, inputs, outputs)
}

func (t *ParamTable) InitDerived_U16(idx PIdx_U16, alwaysUpdate bool, calcIdx PIdx_Calc, inputs []uint16, outputs []uint16) {
	t.checkIdxType(uint16(idx), "Uint16", typeU16, false, true)
	t.initDerivedHookups(uint16(idx), alwaysUpdate, calcIdx, inputs, outputs)
}

func (t *ParamTable) InitDerived_I16(idx PIdx_I16, alwaysUpdate bool, calcIdx PIdx_Calc, inputs []uint16, outputs []uint16) {
	t.checkIdxType(uint16(idx), "Int16", typeI16, false, true)
	t.initDerivedHookups(uint16(idx), alwaysUpdate, calcIdx, inputs, outputs)
}

func (t *ParamTable) InitDerived_U32(idx PIdx_U32, alwaysUpdate bool, calcIdx PIdx_Calc, inputs []uint16, outputs []uint16) {
	t.checkIdxType(uint16(idx), "Uint32", typeU32, false, true)
	t.initDerivedHookups(uint16(idx), alwaysUpdate, calcIdx, inputs, outputs)
}

func (t *ParamTable) InitDerived_I32(idx PIdx_I32, alwaysUpdate bool, calcIdx PIdx_Calc, inputs []uint16, outputs []uint16) {
	t.checkIdxType(uint16(idx), "Int32", typeI32, false, true)
	t.initDerivedHookups(uint16(idx), alwaysUpdate, calcIdx, inputs, outputs)
}

func (t *ParamTable) InitDerived_F32(idx PIdx_F32, alwaysUpdate bool, calcIdx PIdx_Calc, inputs []uint16, outputs []uint16) {
	t.checkIdxType(uint16(idx), "Float32", typeF32, false, true)
	t.initDerivedHookups(uint16(idx), alwaysUpdate, calcIdx, inputs, outputs)
}

func (t *ParamTable) InitDerived_U64(idx PIdx_U64, alwaysUpdate bool, calcIdx PIdx_Calc, inputs []uint16, outputs []uint16) {
	t.checkIdxType(uint16(idx), "Uint64", typeU64, false, true)
	t.initDerivedHookups(uint16(idx), alwaysUpdate, calcIdx, inputs, outputs)
}

func (t *ParamTable) InitDerived_I64(idx PIdx_I64, alwaysUpdate bool, calcIdx PIdx_Calc, inputs []uint16, outputs []uint16) {
	t.checkIdxType(uint16(idx), "Int64", typeI64, false, true)
	t.initDerivedHookups(uint16(idx), alwaysUpdate, calcIdx, inputs, outputs)
}

func (t *ParamTable) InitDerived_F64(idx PIdx_F64, alwaysUpdate bool, calcIdx PIdx_Calc, inputs []uint16, outputs []uint16) {
	t.checkIdxType(uint16(idx), "Float64", typeF64, false, true)
	t.initDerivedHookups(uint16(idx), alwaysUpdate, calcIdx, inputs, outputs)
}

func (t *ParamTable) InitDerived_Addr(idx PIdx_Ptr, alwaysUpdate bool, calcIdx PIdx_Calc, inputs []uint16, outputs []uint16) {
	t.checkIdxType(uint16(idx), "Uintptr", typePtr, false, true)
	t.initDerivedHookups(uint16(idx), alwaysUpdate, calcIdx, inputs, outputs)
}

func (t *ParamTable) updateChildren(idx uint16, prevIdxs []uint16) (newPrevIdxs []uint16) {
	newPrevIdxs = prevIdxs
	childIdxIdx := t.hookups[idx].childrenStart
	if childIdxIdx == 0 {
		return
	}
	if childIdxIdx >= uint32(len(t.children)) {
		return
	}
	childIdx := t.children[childIdxIdx]
	for childIdx != 0 {
		if EnableDebug {
			for _, prevIdx := range prevIdxs {
				if childIdx == prevIdx {
					panic(fmt.Sprintf("error: parameter table: cyclic update loop: during update, idx %d was updated higher (previous) in the heirarchy, but idx %d had previous idx %d as a child, creating an infinite loop", prevIdx, idx, prevIdx))
				}
			}
			newPrevIdxs = append(newPrevIdxs, childIdx)
		}
		hookup := t.hookups[childIdx]
		calc := t.calcs[hookup.calculation]
		recalc := CalcInterface{
			table:    t,
			inputs:   t.parents[hookup.parentsStart : hookup.parentsStart+uint32(hookup.parentsLen)],
			outputs:  t.outputs[hookup.calcOutputsStart : hookup.calcOutputsStart+uint32(hookup.calcOutputsLen)],
			prevIdxs: newPrevIdxs,
		}
		calc(&recalc)
		newPrevIdxs = recalc.prevIdxs
		childIdxIdx += 1
		if childIdxIdx >= uint32(len(t.children)) {
			return
		}
		childIdx = t.children[childIdxIdx]
	}
	return
}

type CalcInterface struct {
	table    *ParamTable
	inputs   []uint16
	outputs  []uint16
	prevIdxs []uint16
}

func (t CalcInterface) GetInput_U8(inputIdx uint16) uint8 {
	idx := t.inputs[inputIdx]
	return t.table.Get_U8(PIdx_U8(idx))
}
func (t CalcInterface) GetInput_I8(inputIdx uint16) int8 {
	idx := t.inputs[inputIdx]
	return t.table.Get_I8(PIdx_I8(idx))
}
func (t CalcInterface) GetInput_Bool(inputIdx uint16) bool {
	idx := t.inputs[inputIdx]
	return t.table.Get_Bool(PIdx_Bool(idx))
}
func (t CalcInterface) GetInput_U16(inputIdx uint16) uint16 {
	idx := t.inputs[inputIdx]
	return t.table.Get_U16(PIdx_U16(idx))
}
func (t CalcInterface) GetInput_I16(inputIdx uint16) int16 {
	idx := t.inputs[inputIdx]
	return t.table.Get_I16(PIdx_I16(idx))
}
func (t CalcInterface) GetInput_U32(inputIdx uint16) uint32 {
	idx := t.inputs[inputIdx]
	return t.table.Get_U32(PIdx_U32(idx))
}
func (t CalcInterface) GetInput_I32(inputIdx uint16) int32 {
	idx := t.inputs[inputIdx]
	return t.table.Get_I32(PIdx_I32(idx))
}
func (t CalcInterface) GetInput_F32(inputIdx uint16) float32 {
	idx := t.inputs[inputIdx]
	return t.table.Get_F32(PIdx_F32(idx))
}
func (t CalcInterface) GetInput_U64(inputIdx uint16) uint64 {
	idx := t.inputs[inputIdx]
	return t.table.Get_U64(PIdx_U64(idx))
}
func (t CalcInterface) GetInput_I64(inputIdx uint16) int64 {
	idx := t.inputs[inputIdx]
	return t.table.Get_I64(PIdx_I64(idx))
}
func (t CalcInterface) GetInput_F64(inputIdx uint16) float64 {
	idx := t.inputs[inputIdx]
	return t.table.Get_F64(PIdx_F64(idx))
}
func (t CalcInterface) GetInput_Ptr(inputIdx uint16) unsafe.Pointer {
	idx := t.inputs[inputIdx]
	return t.table.Get_Ptr(PIdx_Ptr(idx))
}
func (t CalcInterface) GetAllInputs() []uint16 {
	return t.inputs
}
func (t CalcInterface) GetInputRangeStart(start uint16) []uint16 {
	return t.inputs[start:]
}
func (t CalcInterface) GetInputRangeEnd(end uint16) []uint16 {
	return t.inputs[:end]
}
func (t CalcInterface) GetInputRangeStartEnd(start, end uint16) []uint16 {
	return t.inputs[start:end]
}

func (t *CalcInterface) SetOutput_U8(outputIdx uint16, val uint8) {
	idx := t.outputs[outputIdx]
	t.prevIdxs = t.table.set_U8(idx, val, true, t.prevIdxs)
}
func (t *CalcInterface) SetOutput_I8(outputIdx uint16, val int8) {
	idx := t.outputs[outputIdx]
	t.prevIdxs = t.table.set_I8(idx, val, true, t.prevIdxs)
}
func (t *CalcInterface) SetOutput_Bool(outputIdx uint16, val bool) {
	idx := t.outputs[outputIdx]
	t.prevIdxs = t.table.set_Bool(idx, val, true, t.prevIdxs)
}
func (t *CalcInterface) SetOutput_U16(outputIdx uint16, val uint16) {
	idx := t.outputs[outputIdx]
	t.prevIdxs = t.table.set_U16(idx, val, true, t.prevIdxs)
}
func (t *CalcInterface) SetOutput_I16(outputIdx uint16, val int16) {
	idx := t.outputs[outputIdx]
	t.prevIdxs = t.table.set_I16(idx, val, true, t.prevIdxs)
}
func (t *CalcInterface) SetOutput_U32(outputIdx uint16, val uint32) {
	idx := t.outputs[outputIdx]
	t.prevIdxs = t.table.set_U32(idx, val, true, t.prevIdxs)
}
func (t *CalcInterface) SetOutput_I32(outputIdx uint16, val int32) {
	idx := t.outputs[outputIdx]
	t.prevIdxs = t.table.set_I32(idx, val, true, t.prevIdxs)
}
func (t *CalcInterface) SetOutput_F32(outputIdx uint16, val float32) {
	idx := t.outputs[outputIdx]
	t.prevIdxs = t.table.set_F32(idx, val, true, t.prevIdxs)
}
func (t *CalcInterface) SetOutput_U64(outputIdx uint16, val uint64) {
	idx := t.outputs[outputIdx]
	t.prevIdxs = t.table.set_U64(idx, val, true, t.prevIdxs)
}
func (t *CalcInterface) SetOutput_I64(outputIdx uint16, val int64) {
	idx := t.outputs[outputIdx]
	t.prevIdxs = t.table.set_I64(idx, val, true, t.prevIdxs)
}
func (t *CalcInterface) SetOutput_F64(outputIdx uint16, val float64) {
	idx := t.outputs[outputIdx]
	t.prevIdxs = t.table.set_F64(idx, val, true, t.prevIdxs)
}
func (t *CalcInterface) SetOutput_Ptr(outputIdx uint16, val unsafe.Pointer) {
	idx := t.outputs[outputIdx]
	t.prevIdxs = t.table.set_Ptr(idx, val, true, t.prevIdxs)
}
