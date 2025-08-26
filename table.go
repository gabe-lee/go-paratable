package go_param_table

import (
	"fmt"
	"io"
	"math"
	"os"
	"unsafe"

	// _ "github.com/gabe-lee/go_effect_sort"
	ll "github.com/gabe-lee/go_list_like"

	// mem "github.com/gabe-lee/go_manual_memory"
	rq "github.com/gabe-lee/go_ring_queue"
)

// Whether or not any ParamTable's insert additional sanity/safety checks on input/output operations
//   - true (default): recommended for develpoment, or any time your program may be stuck in an infinite loop or parameters are not behaving as expected
//   - false: if you want slightly faster speed in production, have already fully tested your param table with EnableSafetyChecks set to true, _AND_ your end users do not have direct access to your ParamTable
var EnableSafetyChecks bool = true

// The output writer for debug messages. If `EnableDebug == false`, this will not be used at all
//
// Defaults to `os.Stderr`
var DebugWriter io.Writer = os.Stderr

// // The Allocator to use for the parameter table
// //
// // Defaults to GoAllocator, which is a simple wrapper around the standard
// // golang allocation strategy
// var DataAllocator mem.Allocator = mem.NewGoAllocator()

// type paramFlags uint64

// const (
// 	_PFLAG_INIT paramFlags = 1 << iota
// 	_PFLAG_ALWAYS_UPDATE

// 	_PFLAG_BITS                    = iota
// 	_PFLAG_MASK                    = (1 << _PFLAG_BITS) - 1
// 	_PFLAG_CHUNK_BITS              = 64
// 	_PFLAG_SUB_PER_CHUNK           = _PFLAG_CHUNK_BITS / _PFLAG_BITS
// 	_PFLAG_SUB_PER_CHUNK_SHIFT     = 5
// 	_PFLAG_SUB_PER_CHUNK_MINUS_ONE = _PFLAG_SUB_PER_CHUNK - 1
// )

// func (f paramFlags) IsInit() bool {
// 	return f&_PFLAG_INIT == _PFLAG_INIT
// }
// func (f paramFlags) AlwaysUpdate() bool {
// 	return f&_PFLAG_ALWAYS_UPDATE == _PFLAG_ALWAYS_UPDATE
// }

// func getFlag(elemIdx uint16, blocks []paramFlags) paramFlags {
// 	bIdx := elemIdx >> _PFLAG_SUB_PER_CHUNK_SHIFT
// 	sIdx := elemIdx % _PFLAG_SUB_PER_CHUNK
// 	block := blocks[bIdx]
// 	return (block >> (sIdx * _PFLAG_BITS)) & _PFLAG_MASK
// }

// func setFlag(elemIdx uint16, blocks []paramFlags, val paramFlags) {
// 	bIdx := elemIdx >> _PFLAG_SUB_PER_CHUNK_SHIFT
// 	sIdx := elemIdx % _PFLAG_SUB_PER_CHUNK
// 	block := val << paramFlags(sIdx*_PFLAG_BITS)
// 	blocks[bIdx] |= block
// }

// func initFlagLen(elemCount uint16) int {
// 	return int(elemCount+_PFLAG_SUB_PER_CHUNK_MINUS_ONE) >> _PFLAG_SUB_PER_CHUNK_SHIFT
// }

type cInit bool

const (
	_mustBeInit  cInit = true
	_canBeUninit cInit = false
)

type cDerived bool

const (
	_cannotBeDerived cDerived = true
	_canBeDerived    cDerived = false
)

type cRoot bool

const (
	_cannotBeRoot cRoot = true
	_canBeRoot    cRoot = false
)

type ParamTable struct {
	// alloc        mem.Allocator
	meta_table   ll.SliceAdapter[metadata]
	vals_8       ll.SliceAdapter[uint8]
	vals_16      ll.SliceAdapter[uint16]
	vals_32      ll.SliceAdapter[uint32]
	vals_64      ll.SliceAdapter[uint64]
	vals_ptr     ll.SliceAdapter[unsafe.Pointer]
	calcs        ll.SliceAdapter[ParamCalc]
	update_queue rq.RingQueue[ParamID]
	prev_idxs    ll.SliceAdapter[ParamID]
}

func NewParamTable(initCap int) ParamTable {
	return ParamTable{
		// alloc:        alloc,
		meta_table:   ll.EmptySliceAdapter[metadata](initCap),
		vals_8:       ll.EmptySliceAdapter[uint8](0),
		vals_16:      ll.EmptySliceAdapter[uint16](0),
		vals_32:      ll.EmptySliceAdapter[uint32](0),
		vals_64:      ll.EmptySliceAdapter[uint64](0),
		vals_ptr:     ll.EmptySliceAdapter[unsafe.Pointer](0),
		calcs:        ll.EmptySliceAdapter[ParamCalc](0),
		update_queue: rq.New[ParamID](1),
		prev_idxs:    ll.EmptySliceAdapter[ParamID](0),
	}
}

func (t *ParamTable) TotalMemoryFootprint() uintptr {
	size := unsafe.Sizeof(ParamTable{})

	size += uintptr(t.meta_table.Cap()) * unsafe.Sizeof(metadata{})
	size += uintptr(t.vals_8.Cap())
	size += uintptr(t.vals_16.Cap()) * 2
	size += uintptr(t.vals_32.Cap()) * 4
	size += uintptr(t.vals_64.Cap()) * 8
	size += uintptr(t.vals_ptr.Cap()) * unsafe.Sizeof(unsafe.Pointer(uintptr(0)))
	size += uintptr(t.calcs.Cap()) * unsafe.Sizeof(func() {})
	size += uintptr(t.update_queue.Cap()) * 2
	size += uintptr(t.prev_idxs.Cap()) * 2
	ll.DoActionOnAllItems(&t.meta_table, func(slice ll.SliceLike[metadata], idx int, item metadata) {
		size += uintptr(item.hookupsRaw.Cap()) * 2
	})
	return size
}

func (t *ParamTable) beginUpdate(rootIdx ParamID) {
	if EnableSafetyChecks {
		ll.Clear(&t.prev_idxs)
		ll.AppendV(&t.prev_idxs, rootIdx)
	}
	t.update_queue.Clear()
}

func (t *ParamTable) finishUpdate() {
	nextIdx, hasNext := t.update_queue.Dequeue()
	for hasNext {
		meta := t.meta_table.Get(int(nextIdx))
		t.doUpdate(meta)
		nextIdx, hasNext = t.update_queue.Dequeue()
	}
}

func (t *ParamTable) doUpdate(meta metadata) {
	hooks := meta.getHookups(t)
	iface := CalcInterface{
		table:   t,
		inputs:  hooks.parents,
		outputs: hooks.siblings,
	}
	hooks.calc(&iface)
}

func (t *ParamTable) newValU8(val uint8) (valIdx uint16) {
	valIdx = uint16(ll.PushGetIdx(&t.vals_8, val))
	return
}
func (t *ParamTable) newValI8(val int8) (valIdx uint16) {
	valIdx = uint16(ll.PushGetIdx(&t.vals_8, 0))
	ll.SetUnsafeCast(&t.vals_8, int(valIdx), val)
	return
}
func (t *ParamTable) newValBool(val bool) (valIdx uint16) {
	valIdx = uint16(ll.PushGetIdx(&t.vals_8, 0))
	ll.SetUnsafeCast(&t.vals_8, int(valIdx), val)
	return
}
func (t *ParamTable) newValU16(val uint16) (valIdx uint16) {
	valIdx = uint16(ll.PushGetIdx(&t.vals_16, val))
	return
}
func (t *ParamTable) newValI16(val int16) (valIdx uint16) {
	valIdx = uint16(ll.PushGetIdx(&t.vals_16, 0))
	ll.SetUnsafeCast(&t.vals_16, int(valIdx), val)
	return
}
func (t *ParamTable) newValU32(val uint32) (valIdx uint16) {
	valIdx = uint16(ll.PushGetIdx(&t.vals_32, val))
	return
}
func (t *ParamTable) newValI32(val int32) (valIdx uint16) {
	valIdx = uint16(ll.PushGetIdx(&t.vals_32, 0))
	ll.SetUnsafeCast(&t.vals_32, int(valIdx), val)
	return
}
func (t *ParamTable) newValF32(val float32) (valIdx uint16) {
	valIdx = uint16(ll.PushGetIdx(&t.vals_32, 0))
	ll.SetUnsafeCast(&t.vals_32, int(valIdx), val)
	return
}
func (t *ParamTable) newValU64(val uint64) (valIdx uint16) {
	valIdx = uint16(ll.PushGetIdx(&t.vals_64, val))
	return
}
func (t *ParamTable) newValI64(val int64) (valIdx uint16) {
	valIdx = uint16(ll.PushGetIdx(&t.vals_64, 0))
	ll.SetUnsafeCast(&t.vals_64, int(valIdx), val)
	return
}
func (t *ParamTable) newValF64(val float64) (valIdx uint16) {
	valIdx = uint16(ll.PushGetIdx(&t.vals_64, 0))
	ll.SetUnsafeCast(&t.vals_64, int(valIdx), val)
	return
}
func (t *ParamTable) newValPtr(val unsafe.Pointer) (valIdx uint16) {
	valIdx = uint16(ll.PushGetIdx(&t.vals_ptr, val))
	return
}

func (t *ParamTable) newMetadataRoot(valIdx uint16, setType ParamType, alwaysUpdate bool) (id ParamID) {
	meta := metadata{
		hookupsRaw:    ll.EmptySliceAdapter[uint16](0),
		calcId:        math.MaxUint16,
		childrenStart: 0,
		siblingsStart: 0,
		valIdx:        valIdx,
		pType:         setType,
		flags:         _FLAG_IS_USED,
	}
	meta.set_always_update(alwaysUpdate)
	id = ParamID(ll.PushGetIdx(&t.meta_table, meta))
	return
}

func (t *ParamTable) newMetadataDerivedSingle(valIdx uint16, setType ParamType, alwaysUpdate bool, calcId CalcID, parents []ParamID) (meta metadata, id ParamID) {
	meta = metadata{
		hookupsRaw:    ll.EmptySliceAdapter[uint16](0),
		calcId:        calcId,
		childrenStart: 0,
		siblingsStart: 0,
		valIdx:        valIdx,
		pType:         setType,
		flags:         _FLAG_IS_USED,
	}
	meta.set_has_calc(true)
	meta.set_has_parent(true)
	meta.set_has_siblings(true)
	meta.set_always_update(alwaysUpdate)
	meta.appendParents(parents)
	id = ParamID(ll.PushGetIdx(&t.meta_table, meta))
	for _, p := range parents {
		pmeta := t.meta_table.Get(int(p))
		pmeta.appendChild(id)
		t.meta_table.Set(int(p), pmeta)
	}
	meta.appendSibling(id)
	ll.Set(&t.meta_table, int(id), meta)
	return
}

func (t *ParamTable) newMetadataDerivedLinkedUninitSiblings(valIdx uint16, setType ParamType, alwaysUpdate bool, calcId CalcID, parents []ParamID) (id ParamID) {
	meta := metadata{
		hookupsRaw:    ll.EmptySliceAdapter[uint16](0),
		calcId:        calcId,
		childrenStart: 0,
		siblingsStart: 0,
		valIdx:        valIdx,
		pType:         setType,
		flags:         _FLAG_IS_USED,
	}
	meta.set_has_calc(true)
	meta.set_has_parent(true)
	meta.set_has_siblings(true)
	meta.set_always_update(alwaysUpdate)
	meta.appendParents(parents)
	id = ParamID(ll.PushGetIdx(&t.meta_table, meta))
	return
}

func (t *ParamTable) getMetaWithCheck(idx ParamID, validType ParamType, mustBeInit cInit, cannotBeDerived cDerived, cannotBeRoot cRoot) (meta metadata) {
	if EnableSafetyChecks {
		if idx >= ParamID(t.meta_table.Len()) {
			fmt.Fprintf(DebugWriter, "fatal: go_param_table: index %d is outside bounds of metadata list (len %d)", idx, t.meta_table.Len())
			panic("FATAL")
		}
		meta = ll.Get(&t.meta_table, int(idx))
		if mustBeInit == _mustBeInit && meta.is_free() {
			fmt.Fprintf(DebugWriter, "fatal: go_param_table: index %d is a free metadata index (previously used, but deleted)", idx)
			panic("FATAL")
		}
		if meta.pType != validType {
			fmt.Fprintf(DebugWriter, "fatal: go_param_table: index %d is not a %s value, (found %s value)", idx, typeNames[validType], typeNames[meta.pType])
			panic("FATAL")
		}
		if cannotBeDerived == _cannotBeDerived && (meta.has_parent() || meta.has_calc()) {
			fmt.Fprintf(DebugWriter, "fatal: go_param_table: index %d is a derived value (has parents and/or calculation func), expected root value", idx)
			panic("FATAL")
		}
		if cannotBeRoot == _cannotBeRoot && !meta.has_parent() && !meta.has_calc() {
			fmt.Fprintf(DebugWriter, "fatal: go_param_table: index %d is a root value (has no parents or calculation func), expected derived value", idx)
			panic("FATAL")
		}
	} else {
		meta = ll.Get(&t.meta_table, int(idx))
	}
	return
}

func (t *ParamTable) Get_U8(idx ParamID) uint8 {
	meta := t.getMetaWithCheck(idx, TypeU8, _mustBeInit, _canBeDerived, _canBeRoot)
	return ll.Get(&t.vals_8, int(meta.valIdx))
}

func (t *ParamTable) Get_I8(idx ParamID) int8 {
	meta := t.getMetaWithCheck(idx, TypeI8, _mustBeInit, _canBeDerived, _canBeRoot)
	return ll.GetUnsafeCast[uint8, int8](&t.vals_8, int(meta.valIdx))
}

func (t *ParamTable) Get_Bool(idx ParamID) bool {
	meta := t.getMetaWithCheck(idx, TypeBool, _mustBeInit, _canBeDerived, _canBeRoot)
	return ll.GetUnsafeCast[uint8, bool](&t.vals_8, int(meta.valIdx))
}

func (t *ParamTable) Get_U16(idx ParamID) uint16 {
	meta := t.getMetaWithCheck(idx, TypeU8, _mustBeInit, _canBeDerived, _canBeRoot)
	return ll.Get(&t.vals_16, int(meta.valIdx))
}

func (t *ParamTable) Get_I16(idx ParamID) int16 {
	meta := t.getMetaWithCheck(idx, TypeI16, _mustBeInit, _canBeDerived, _canBeRoot)
	return ll.GetUnsafeCast[uint16, int16](&t.vals_16, int(meta.valIdx))
}

func (t *ParamTable) Get_U32(idx ParamID) uint32 {
	meta := t.getMetaWithCheck(idx, TypeU32, _mustBeInit, _canBeDerived, _canBeRoot)
	return ll.Get(&t.vals_32, int(meta.valIdx))
}

func (t *ParamTable) Get_I32(idx ParamID) int32 {
	meta := t.getMetaWithCheck(idx, TypeI32, _mustBeInit, _canBeDerived, _canBeRoot)
	return ll.GetUnsafeCast[uint32, int32](&t.vals_32, int(meta.valIdx))
}

func (t *ParamTable) Get_F32(idx ParamID) float32 {
	meta := t.getMetaWithCheck(idx, TypeF32, _mustBeInit, _canBeDerived, _canBeRoot)
	return ll.GetUnsafeCast[uint32, float32](&t.vals_32, int(meta.valIdx))
}

func (t *ParamTable) Get_U64(idx ParamID) uint64 {
	meta := t.getMetaWithCheck(idx, TypeU64, _mustBeInit, _canBeDerived, _canBeRoot)
	return ll.Get(&t.vals_64, int(meta.valIdx))
}

func (t *ParamTable) Get_I64(idx ParamID) int64 {
	meta := t.getMetaWithCheck(idx, TypeI64, _mustBeInit, _canBeDerived, _canBeRoot)
	return ll.GetUnsafeCast[uint64, int64](&t.vals_64, int(meta.valIdx))
}

func (t *ParamTable) Get_F64(idx ParamID) float64 {
	meta := t.getMetaWithCheck(idx, TypeF64, _mustBeInit, _canBeDerived, _canBeRoot)
	return ll.GetUnsafeCast[uint64, float64](&t.vals_64, int(meta.valIdx))
}

func (t *ParamTable) Get_Ptr(idx ParamID) unsafe.Pointer {
	meta := t.getMetaWithCheck(idx, TypePtr, _mustBeInit, _canBeDerived, _canBeRoot)
	return ll.Get(&t.vals_ptr, int(meta.valIdx))
}

func (t *ParamTable) queueChildren(meta metadata) {
	children := meta.getChildren()
	//VERIFY due to the new architechture of ParamTable, cyclic references should be impossible to initialize?
	// if EnableSafetyChecks {
	// 	var i = 0
	// 	var lim = t.prev_idxs.Len()
	// 	for i < lim {
	// 		prev := ll.Get(&t.prev_idxs, i)
	// 		for _, c := range children {
	// 			if prev == c {
	// 				fmt.Fprintf(DebugWriter, "fatal: go_param_table: cyclic update loop: during update, idx %d was updated higher (previous) in the heirarchy, but idx %d had previous idx %d as a child, creating an infinite loop", prev, c, prev)
	// 				panic("FATAL")
	// 			}
	// 		}
	// 	}
	// 	ll.AppendV(&t.prev_idxs, children...)
	// }
	t.update_queue.QueueMany(children...)
}

func (t *ParamTable) set_U8(idx ParamID, val uint8, mustBeInit cInit, cannotBeDerived cDerived, cannotBeRoot cRoot) {
	meta := t.getMetaWithCheck(idx, TypeU8, mustBeInit, cannotBeDerived, cannotBeRoot)
	changed := ll.SetChanged(&t.vals_8, int(meta.valIdx), val)
	if changed || meta.should_always_update() {
		t.queueChildren(meta)
	}
}

func (t *ParamTable) set_I8(idx ParamID, val int8, mustBeInit cInit, cannotBeDerived cDerived, cannotBeRoot cRoot) {
	meta := t.getMetaWithCheck(idx, TypeI8, mustBeInit, cannotBeDerived, cannotBeRoot)
	changed := ll.SetUnsafeCastChanged(&t.vals_8, int(meta.valIdx), val)
	if changed || meta.should_always_update() {
		t.queueChildren(meta)
	}
}

func (t *ParamTable) set_Bool(idx ParamID, val bool, mustBeInit cInit, cannotBeDerived cDerived, cannotBeRoot cRoot) {
	meta := t.getMetaWithCheck(idx, TypeBool, mustBeInit, cannotBeDerived, cannotBeRoot)
	changed := ll.SetUnsafeCastChanged(&t.vals_8, int(meta.valIdx), val)
	if changed || meta.should_always_update() {
		t.queueChildren(meta)
	}
}

func (t *ParamTable) set_U16(idx ParamID, val uint16, mustBeInit cInit, cannotBeDerived cDerived, cannotBeRoot cRoot) {
	meta := t.getMetaWithCheck(idx, TypeU16, mustBeInit, cannotBeDerived, cannotBeRoot)
	changed := ll.SetChanged(&t.vals_16, int(meta.valIdx), val)
	if changed || meta.should_always_update() {
		t.queueChildren(meta)
	}
}

func (t *ParamTable) set_I16(idx ParamID, val int16, mustBeInit cInit, cannotBeDerived cDerived, cannotBeRoot cRoot) {
	meta := t.getMetaWithCheck(idx, TypeI16, mustBeInit, cannotBeDerived, cannotBeRoot)
	changed := ll.SetUnsafeCastChanged(&t.vals_16, int(meta.valIdx), val)
	if changed || meta.should_always_update() {
		t.queueChildren(meta)
	}
}

func (t *ParamTable) set_U32(idx ParamID, val uint32, mustBeInit cInit, cannotBeDerived cDerived, cannotBeRoot cRoot) {
	meta := t.getMetaWithCheck(idx, TypeU32, mustBeInit, cannotBeDerived, cannotBeRoot)
	changed := ll.SetChanged(&t.vals_32, int(meta.valIdx), val)
	if changed || meta.should_always_update() {
		t.queueChildren(meta)
	}
}

func (t *ParamTable) set_I32(idx ParamID, val int32, mustBeInit cInit, cannotBeDerived cDerived, cannotBeRoot cRoot) {
	meta := t.getMetaWithCheck(idx, TypeI32, mustBeInit, cannotBeDerived, cannotBeRoot)
	changed := ll.SetUnsafeCastChanged(&t.vals_32, int(meta.valIdx), val)
	if changed || meta.should_always_update() {
		t.queueChildren(meta)
	}
}

func (t *ParamTable) set_F32(idx ParamID, val float32, mustBeInit cInit, cannotBeDerived cDerived, cannotBeRoot cRoot) {
	meta := t.getMetaWithCheck(idx, TypeF32, mustBeInit, cannotBeDerived, cannotBeRoot)
	changed := ll.SetUnsafeCastChanged(&t.vals_32, int(meta.valIdx), val)
	if changed || meta.should_always_update() {
		t.queueChildren(meta)
	}
}

func (t *ParamTable) set_U64(idx ParamID, val uint64, mustBeInit cInit, cannotBeDerived cDerived, cannotBeRoot cRoot) {
	meta := t.getMetaWithCheck(idx, TypeU64, mustBeInit, cannotBeDerived, cannotBeRoot)
	changed := ll.SetChanged(&t.vals_64, int(meta.valIdx), val)

	if changed || meta.should_always_update() {

		t.queueChildren(meta)
	}
}

func (t *ParamTable) set_I64(idx ParamID, val int64, mustBeInit cInit, cannotBeDerived cDerived, cannotBeRoot cRoot) {
	meta := t.getMetaWithCheck(idx, TypeI64, mustBeInit, cannotBeDerived, cannotBeRoot)
	changed := ll.SetUnsafeCastChanged(&t.vals_64, int(meta.valIdx), val)
	if changed || meta.should_always_update() {
		t.queueChildren(meta)
	}
}

func (t *ParamTable) set_F64(idx ParamID, val float64, mustBeInit cInit, cannotBeDerived cDerived, cannotBeRoot cRoot) {
	meta := t.getMetaWithCheck(idx, TypeF64, mustBeInit, cannotBeDerived, cannotBeRoot)
	changed := ll.SetUnsafeCastChanged(&t.vals_64, int(meta.valIdx), val)
	if changed || meta.should_always_update() {
		t.queueChildren(meta)
	}
}

func (t *ParamTable) set_Ptr(idx ParamID, val unsafe.Pointer, mustBeInit cInit, cannotBeDerived cDerived, cannotBeRoot cRoot) {
	meta := t.getMetaWithCheck(idx, TypePtr, mustBeInit, cannotBeDerived, cannotBeRoot)
	changed := ll.SetChanged(&t.vals_ptr, int(meta.valIdx), val)
	if changed || meta.should_always_update() {
		t.queueChildren(meta)
	}
}

func (t *ParamTable) SetRoot_U8(idx ParamID, val uint8) {
	t.beginUpdate(idx)
	t.set_U8(idx, val, _mustBeInit, _cannotBeDerived, _canBeRoot)
	t.finishUpdate()
}

func (t *ParamTable) SetRoot_I8(idx ParamID, val int8) {
	t.beginUpdate(idx)
	t.set_I8(idx, val, _mustBeInit, _cannotBeDerived, _canBeRoot)
	t.finishUpdate()
}

func (t *ParamTable) SetRoot_Bool(idx ParamID, val bool) {
	t.beginUpdate(idx)
	t.set_Bool(idx, val, _mustBeInit, _cannotBeDerived, _canBeRoot)
	t.finishUpdate()
}

func (t *ParamTable) SetRoot_U16(idx ParamID, val uint16) {
	t.beginUpdate(idx)
	t.set_U16(idx, val, _mustBeInit, _cannotBeDerived, _canBeRoot)
	t.finishUpdate()
}

func (t *ParamTable) SetRoot_I16(idx ParamID, val int16) {
	t.beginUpdate(idx)
	t.set_I16(idx, val, _mustBeInit, _cannotBeDerived, _canBeRoot)
	t.finishUpdate()
}

func (t *ParamTable) SetRoot_U32(idx ParamID, val uint32) {
	t.beginUpdate(idx)
	t.set_U32(idx, val, _mustBeInit, _cannotBeDerived, _canBeRoot)
	t.finishUpdate()
}

func (t *ParamTable) SetRoot_I32(idx ParamID, val int32) {
	t.beginUpdate(idx)
	t.set_I32(idx, val, _mustBeInit, _cannotBeDerived, _canBeRoot)
	t.finishUpdate()
}

func (t *ParamTable) SetRoot_F32(idx ParamID, val float32) {
	t.beginUpdate(idx)
	t.set_F32(idx, val, _mustBeInit, _cannotBeDerived, _canBeRoot)
	t.finishUpdate()
}

func (t *ParamTable) SetRoot_U64(idx ParamID, val uint64) {
	t.beginUpdate(idx)
	t.set_U64(idx, val, _mustBeInit, _cannotBeDerived, _canBeRoot)
	t.finishUpdate()
}

func (t *ParamTable) SetRoot_I64(idx ParamID, val int64) {
	t.beginUpdate(idx)
	t.set_I64(idx, val, _mustBeInit, _cannotBeDerived, _canBeRoot)
	t.finishUpdate()
}

func (t *ParamTable) SetRoot_F64(idx ParamID, val float64) {
	t.beginUpdate(idx)
	t.set_F64(idx, val, _mustBeInit, _cannotBeDerived, _canBeRoot)
	t.finishUpdate()
}

func (t *ParamTable) SetRoot_Ptr(idx ParamID, val unsafe.Pointer) {
	t.beginUpdate(idx)
	t.set_Ptr(idx, val, _mustBeInit, _cannotBeDerived, _canBeRoot)
	t.finishUpdate()
}

func (t *ParamTable) InitRoot_U8(val uint8, alwaysUpdate bool) (id ParamID) {
	valIdx := t.newValU8(val)
	id = t.newMetadataRoot(valIdx, TypeU8, alwaysUpdate)
	return
}

func (t *ParamTable) InitRoot_I8(val int8, alwaysUpdate bool) (id ParamID) {
	valIdx := t.newValI8(val)
	id = t.newMetadataRoot(valIdx, TypeI8, alwaysUpdate)
	return
}

func (t *ParamTable) InitRoot_Bool(val bool, alwaysUpdate bool) (id ParamID) {
	valIdx := t.newValBool(val)
	id = t.newMetadataRoot(valIdx, TypeBool, alwaysUpdate)
	return
}

func (t *ParamTable) InitRoot_U16(val uint16, alwaysUpdate bool) (id ParamID) {
	valIdx := t.newValU16(val)
	id = t.newMetadataRoot(valIdx, TypeU16, alwaysUpdate)
	return
}

func (t *ParamTable) InitRoot_I16(val int16, alwaysUpdate bool) (id ParamID) {
	valIdx := t.newValI16(val)
	id = t.newMetadataRoot(valIdx, TypeI16, alwaysUpdate)
	return
}

func (t *ParamTable) InitRoot_U32(val uint32, alwaysUpdate bool) (id ParamID) {
	valIdx := t.newValU32(val)
	id = t.newMetadataRoot(valIdx, TypeU32, alwaysUpdate)
	return
}

func (t *ParamTable) InitRoot_I32(val int32, alwaysUpdate bool) (id ParamID) {
	valIdx := t.newValI32(val)
	id = t.newMetadataRoot(valIdx, TypeI32, alwaysUpdate)
	return
}

func (t *ParamTable) InitRoot_F32(val float32, alwaysUpdate bool) (id ParamID) {
	valIdx := t.newValF32(val)

	id = t.newMetadataRoot(valIdx, TypeF32, alwaysUpdate)
	return
}

func (t *ParamTable) InitRoot_U64(val uint64, alwaysUpdate bool) (id ParamID) {
	valIdx := t.newValU64(val)
	id = t.newMetadataRoot(valIdx, TypeU64, alwaysUpdate)
	return
}

func (t *ParamTable) InitRoot_I64(val int64, alwaysUpdate bool) (id ParamID) {
	valIdx := t.newValI64(val)
	id = t.newMetadataRoot(valIdx, TypeI64, alwaysUpdate)
	return
}

func (t *ParamTable) InitRoot_F64(val float64, alwaysUpdate bool) (id ParamID) {
	valIdx := t.newValF64(val)
	id = t.newMetadataRoot(valIdx, TypeF64, alwaysUpdate)
	return
}

func (t *ParamTable) InitRoot_Ptr(val unsafe.Pointer, alwaysUpdate bool) (id ParamID) {
	valIdx := t.newValPtr(val)
	id = t.newMetadataRoot(valIdx, TypePtr, alwaysUpdate)
	return
}

func (t *ParamTable) RegisterCalc(calc ParamCalc) (calcIdx CalcID) {
	calcIdx = CalcID(ll.PushGetIdx(&t.calcs, calc))
	return
}

type TypeInit struct {
	Ptype        ParamType
	AlwaysUpdate bool
}

func NewDerivedInit(ptype ParamType, alwaysUpate bool) TypeInit {
	return TypeInit{
		Ptype:        ptype,
		AlwaysUpdate: alwaysUpate,
	}
}

func (t *ParamTable) InitDerived_Linked(calcIdx CalcID, inputs []ParamID, outputTypes []TypeInit) (outputIDs ll.SliceAdapter[ParamID]) {
	outputIDs = ll.EmptySliceAdapter[ParamID](len(outputTypes))
	for i, out := range outputTypes {
		switch out.Ptype {
		case TypeU8, TypeI8, TypeBool:
			valIdx := t.newValU8(0)
			id := t.newMetadataDerivedLinkedUninitSiblings(valIdx, out.Ptype, out.AlwaysUpdate, calcIdx, inputs)
			ll.Push(&outputIDs, id)
		case TypeU16, TypeI16:
			valIdx := t.newValU16(0)
			id := t.newMetadataDerivedLinkedUninitSiblings(valIdx, out.Ptype, out.AlwaysUpdate, calcIdx, inputs)
			ll.Push(&outputIDs, id)
		case TypeU32, TypeI32, TypeF32:
			valIdx := t.newValU32(0)
			id := t.newMetadataDerivedLinkedUninitSiblings(valIdx, out.Ptype, out.AlwaysUpdate, calcIdx, inputs)
			ll.Push(&outputIDs, id)
		case TypeU64, TypeI64, TypeF64:
			valIdx := t.newValU64(0)
			id := t.newMetadataDerivedLinkedUninitSiblings(valIdx, out.Ptype, out.AlwaysUpdate, calcIdx, inputs)
			ll.Push(&outputIDs, id)
		case TypePtr:
			valIdx := t.newValPtr(nil)
			id := t.newMetadataDerivedLinkedUninitSiblings(valIdx, out.Ptype, out.AlwaysUpdate, calcIdx, inputs)
			ll.Push(&outputIDs, id)
		default:
			if EnableSafetyChecks {
				fmt.Fprintf(DebugWriter, "fatal: go_param_table: InitDerived_Linked(): output value at idx %d had invalid ParamType %d (largest valid ParamType is %d)", i, out.Ptype, typeCount-1)
				panic("FATAL")
			}
		}
	}
	for _, p := range inputs {
		pmeta := t.meta_table.Get(int(p))
		pmeta.appendChildren(outputIDs.GoSlice())
		t.meta_table.Set(int(p), pmeta)
	}
	for i := range outputIDs.Len() {
		id := int(outputIDs.Get(i))
		meta := t.meta_table.Get(id)
		meta.appendSiblings(outputIDs.GoSlice())
		t.meta_table.Set(id, meta)
	}
	iface := CalcInterface{
		table:   t,
		inputs:  inputs,
		outputs: outputIDs.GoSlice(),
	}
	calc := t.calcs.Get(int(calcIdx))
	calc(&iface)
	return
}

func (t *ParamTable) InitDerived_U8(alwaysUpdate bool, calcIdx CalcID, inputs ...ParamID) (id ParamID) {
	valIdx := t.newValU8(0)
	meta, id := t.newMetadataDerivedSingle(valIdx, TypeU8, alwaysUpdate, calcIdx, inputs)
	t.doUpdate(meta)
	return id
}

func (t *ParamTable) InitDerived_I8(alwaysUpdate bool, calcIdx CalcID, inputs ...ParamID) (id ParamID) {
	valIdx := t.newValI8(0)
	meta, id := t.newMetadataDerivedSingle(valIdx, TypeI8, alwaysUpdate, calcIdx, inputs)
	t.doUpdate(meta)
	return id
}

func (t *ParamTable) InitDerived_Bool(alwaysUpdate bool, calcIdx CalcID, inputs ...ParamID) (id ParamID) {
	valIdx := t.newValBool(false)
	meta, id := t.newMetadataDerivedSingle(valIdx, TypeBool, alwaysUpdate, calcIdx, inputs)
	t.doUpdate(meta)
	return id
}

func (t *ParamTable) InitDerived_U16(alwaysUpdate bool, calcIdx CalcID, inputs ...ParamID) (id ParamID) {
	valIdx := t.newValU16(0)
	meta, id := t.newMetadataDerivedSingle(valIdx, TypeU16, alwaysUpdate, calcIdx, inputs)
	t.doUpdate(meta)
	return id
}

func (t *ParamTable) InitDerived_I16(alwaysUpdate bool, calcIdx CalcID, inputs ...ParamID) (id ParamID) {
	valIdx := t.newValI16(0)
	meta, id := t.newMetadataDerivedSingle(valIdx, TypeI16, alwaysUpdate, calcIdx, inputs)
	t.doUpdate(meta)
	return id
}

func (t *ParamTable) InitDerived_U32(alwaysUpdate bool, calcIdx CalcID, inputs ...ParamID) (id ParamID) {
	valIdx := t.newValU32(0)
	meta, id := t.newMetadataDerivedSingle(valIdx, TypeU32, alwaysUpdate, calcIdx, inputs)
	t.doUpdate(meta)
	return id
}

func (t *ParamTable) InitDerived_I32(alwaysUpdate bool, calcIdx CalcID, inputs ...ParamID) (id ParamID) {
	valIdx := t.newValU32(0)
	meta, id := t.newMetadataDerivedSingle(valIdx, TypeI32, alwaysUpdate, calcIdx, inputs)
	t.doUpdate(meta)
	return id
}

func (t *ParamTable) InitDerived_F32(alwaysUpdate bool, calcIdx CalcID, inputs ...ParamID) (id ParamID) {
	valIdx := t.newValU32(0)
	meta, id := t.newMetadataDerivedSingle(valIdx, TypeF32, alwaysUpdate, calcIdx, inputs)
	t.doUpdate(meta)
	return id
}

func (t *ParamTable) InitDerived_U64(alwaysUpdate bool, calcIdx CalcID, inputs ...ParamID) (id ParamID) {
	valIdx := t.newValU64(0)
	meta, id := t.newMetadataDerivedSingle(valIdx, TypeU64, alwaysUpdate, calcIdx, inputs)
	t.doUpdate(meta)
	return id
}

func (t *ParamTable) InitDerived_I64(alwaysUpdate bool, calcIdx CalcID, inputs ...ParamID) (id ParamID) {
	valIdx := t.newValU64(0)
	meta, id := t.newMetadataDerivedSingle(valIdx, TypeI64, alwaysUpdate, calcIdx, inputs)
	t.doUpdate(meta)
	return id
}

func (t *ParamTable) InitDerived_F64(alwaysUpdate bool, calcIdx CalcID, inputs ...ParamID) (id ParamID) {
	valIdx := t.newValU64(0)
	meta, id := t.newMetadataDerivedSingle(valIdx, TypeF64, alwaysUpdate, calcIdx, inputs)
	t.doUpdate(meta)
	return id
}

func (t *ParamTable) InitDerived_Ptr(alwaysUpdate bool, calcIdx CalcID, inputs ...ParamID) (id ParamID) {
	valIdx := t.newValPtr(nil)
	meta, id := t.newMetadataDerivedSingle(valIdx, TypePtr, alwaysUpdate, calcIdx, inputs)
	t.doUpdate(meta)
	return id
}

type CalcInterface struct {
	table   *ParamTable
	inputs  []ParamID
	outputs []ParamID
}

func (t CalcInterface) GetInput_U8(inputIdx uint16) uint8 {
	idx := t.inputs[inputIdx]
	return t.table.Get_U8(idx)
}
func (t CalcInterface) GetInput_I8(inputIdx uint16) int8 {
	idx := t.inputs[inputIdx]
	return t.table.Get_I8(idx)
}
func (t CalcInterface) GetInput_Bool(inputIdx uint16) bool {
	idx := t.inputs[inputIdx]
	return t.table.Get_Bool(idx)
}
func (t CalcInterface) GetInput_U16(inputIdx uint16) uint16 {
	idx := t.inputs[inputIdx]
	return t.table.Get_U16(idx)
}
func (t CalcInterface) GetInput_I16(inputIdx uint16) int16 {
	idx := t.inputs[inputIdx]
	return t.table.Get_I16(idx)
}
func (t CalcInterface) GetInput_U32(inputIdx uint16) uint32 {
	idx := t.inputs[inputIdx]
	return t.table.Get_U32(idx)
}
func (t CalcInterface) GetInput_I32(inputIdx uint16) int32 {
	idx := t.inputs[inputIdx]
	return t.table.Get_I32(idx)
}
func (t CalcInterface) GetInput_F32(inputIdx uint16) float32 {
	idx := t.inputs[inputIdx]
	return t.table.Get_F32(idx)
}
func (t CalcInterface) GetInput_U64(inputIdx uint16) uint64 {
	idx := t.inputs[inputIdx]
	return t.table.Get_U64(idx)
}
func (t CalcInterface) GetInput_I64(inputIdx uint16) int64 {
	idx := t.inputs[inputIdx]
	return t.table.Get_I64(idx)
}
func (t CalcInterface) GetInput_F64(inputIdx uint16) float64 {
	idx := t.inputs[inputIdx]
	return t.table.Get_F64(idx)
}
func (t CalcInterface) GetInput_Ptr(inputIdx uint16) unsafe.Pointer {
	idx := t.inputs[inputIdx]
	return t.table.Get_Ptr(idx)
}
func (t CalcInterface) GetAllInputs() []ParamID {
	return t.inputs
}
func (t CalcInterface) GetInputRangeStart(start uint16) []ParamID {
	return t.inputs[start:]
}
func (t CalcInterface) GetInputRangeEnd(end uint16) []ParamID {
	return t.inputs[:end]
}
func (t CalcInterface) GetInputRangeStartEnd(start, end uint16) []ParamID {
	return t.inputs[start:end]
}

func (t *CalcInterface) SetOutput_U8(outputIdx uint16, val uint8) {
	idx := t.outputs[outputIdx]
	t.table.set_U8(idx, val, _mustBeInit, _canBeDerived, _cannotBeRoot)
}
func (t *CalcInterface) SetOutput_I8(outputIdx uint16, val int8) {
	idx := t.outputs[outputIdx]
	t.table.set_I8(idx, val, _mustBeInit, _canBeDerived, _cannotBeRoot)
}
func (t *CalcInterface) SetOutput_Bool(outputIdx uint16, val bool) {
	idx := t.outputs[outputIdx]
	t.table.set_Bool(idx, val, _mustBeInit, _canBeDerived, _cannotBeRoot)
}
func (t *CalcInterface) SetOutput_U16(outputIdx uint16, val uint16) {
	idx := t.outputs[outputIdx]
	t.table.set_U16(idx, val, _mustBeInit, _canBeDerived, _cannotBeRoot)
}
func (t *CalcInterface) SetOutput_I16(outputIdx uint16, val int16) {
	idx := t.outputs[outputIdx]
	t.table.set_I16(idx, val, _mustBeInit, _canBeDerived, _cannotBeRoot)
}
func (t *CalcInterface) SetOutput_U32(outputIdx uint16, val uint32) {
	idx := t.outputs[outputIdx]
	t.table.set_U32(idx, val, _mustBeInit, _canBeDerived, _cannotBeRoot)
}
func (t *CalcInterface) SetOutput_I32(outputIdx uint16, val int32) {
	idx := t.outputs[outputIdx]
	t.table.set_I32(idx, val, _mustBeInit, _canBeDerived, _cannotBeRoot)
}
func (t *CalcInterface) SetOutput_F32(outputIdx uint16, val float32) {
	idx := t.outputs[outputIdx]
	t.table.set_F32(idx, val, _mustBeInit, _canBeDerived, _cannotBeRoot)
}
func (t *CalcInterface) SetOutput_U64(outputIdx uint16, val uint64) {
	idx := t.outputs[outputIdx]
	t.table.set_U64(idx, val, _mustBeInit, _canBeDerived, _cannotBeRoot)
}
func (t *CalcInterface) SetOutput_I64(outputIdx uint16, val int64) {
	idx := t.outputs[outputIdx]
	t.table.set_I64(idx, val, _mustBeInit, _canBeDerived, _cannotBeRoot)
}
func (t *CalcInterface) SetOutput_F64(outputIdx uint16, val float64) {
	idx := t.outputs[outputIdx]
	t.table.set_F64(idx, val, _mustBeInit, _canBeDerived, _cannotBeRoot)
}
func (t *CalcInterface) SetOutput_Ptr(outputIdx uint16, val unsafe.Pointer) {
	idx := t.outputs[outputIdx]
	t.table.set_Ptr(idx, val, _mustBeInit, _canBeDerived, _cannotBeRoot)
}
