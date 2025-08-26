package go_param_table

import (
	"fmt"
	"unsafe"

	ll "github.com/gabe-lee/go_list_like"
)

type paramsLen uint16

func newParamsLen(inLen uint32, outLen uint32) paramsLen {
	return (paramsLen(inLen) << 8) | paramsLen(outLen)
}

func (l paramsLen) inLen() uint32 {
	return uint32(l >> 8)
}
func (l paramsLen) outLen() uint32 {
	return uint32(l & 0b11111111)
}

type metadata struct {
	hookupsRaw    ll.SliceAdapter[uint16]
	calcId        CalcID
	valIdx        validx
	pType         ParamType
	flags         pflags
	childrenStart uint8
	siblingsStart uint8
}

type hookups struct {
	calc     ParamCalc
	parents  []ParamID
	siblings []ParamID
	children []ParamID
}

type pflags uint8

const (
	_FLAG_OFF_IS_USED pflags = iota
	_FLAG_OFF_ALWAYS_UPDATE
	_FLAG_OFF_HAS_CHILDREN
	_FLAG_OFF_HAS_PARENT
	_FLAG_OFF_HAS_SIBLINGS
	_FLAG_OFF_HAS_CALCULATION
)

const (
	_FLAG_IS_USED         pflags = 1 << _FLAG_OFF_IS_USED
	_FLAG_ALWAYS_UPDATE   pflags = 1 << _FLAG_OFF_ALWAYS_UPDATE
	_FLAG_HAS_CHILDREN    pflags = 1 << _FLAG_OFF_HAS_CHILDREN
	_FLAG_HAS_PARENT      pflags = 1 << _FLAG_OFF_HAS_PARENT
	_FLAG_HAS_SIBLINGS    pflags = 1 << _FLAG_OFF_HAS_SIBLINGS
	_FLAG_HAS_CALCULATION pflags = 1 << _FLAG_OFF_HAS_CALCULATION
)

func (m metadata) is_free() bool {
	return m.flags&_FLAG_IS_USED == 0
}
func (m metadata) is_used() bool {
	return m.flags&_FLAG_IS_USED == _FLAG_IS_USED
}
func (m *metadata) set_free() {
	m.flags &= ^_FLAG_IS_USED
}
func (m *metadata) set_used() {
	m.flags |= _FLAG_IS_USED
}

func (m metadata) should_always_update() bool {
	return m.flags&_FLAG_ALWAYS_UPDATE == _FLAG_ALWAYS_UPDATE
}
func (m metadata) only_update_on_change() bool {
	return m.flags&_FLAG_ALWAYS_UPDATE == 0
}
func (m *metadata) set_always_update(state bool) {
	if state {
		m.flags |= _FLAG_ALWAYS_UPDATE
	} else {
		m.flags &= ^_FLAG_ALWAYS_UPDATE
	}
}

func (m metadata) has_children() bool {
	return m.flags&_FLAG_HAS_CHILDREN == _FLAG_HAS_CHILDREN
}
func (m metadata) no_children() bool {
	return m.flags&_FLAG_HAS_CHILDREN == 0
}
func (m *metadata) set_has_children(state bool) {
	if state {
		m.flags |= _FLAG_HAS_CHILDREN
	} else {
		m.flags &= ^_FLAG_HAS_CHILDREN
	}
}

func (m metadata) has_siblings() bool {
	return m.flags&_FLAG_HAS_SIBLINGS == _FLAG_HAS_SIBLINGS
}
func (m metadata) no_siblings() bool {
	return m.flags&_FLAG_HAS_SIBLINGS == 0
}
func (m *metadata) set_has_siblings(state bool) {
	if state {
		m.flags |= _FLAG_HAS_SIBLINGS
	} else {
		m.flags &= ^_FLAG_HAS_SIBLINGS
	}
}

func (m metadata) has_parent() bool {
	return m.flags&_FLAG_HAS_PARENT == _FLAG_HAS_PARENT
}
func (m metadata) no_parent() bool {
	return m.flags&_FLAG_HAS_PARENT == 0
}
func (m *metadata) set_has_parent(state bool) {
	if state {
		m.flags |= _FLAG_HAS_PARENT
	} else {
		m.flags &= ^_FLAG_HAS_PARENT
	}
}

func (m metadata) has_calc() bool {
	return m.flags&_FLAG_HAS_CALCULATION == _FLAG_HAS_CALCULATION
}
func (m metadata) no_calc() bool {
	return m.flags&_FLAG_HAS_CALCULATION == 0
}
func (m *metadata) set_has_calc(state bool) {
	if state {
		m.flags |= _FLAG_HAS_CALCULATION
	} else {
		m.flags &= ^_FLAG_HAS_CALCULATION
	}
}

func (m metadata) is_type(t ParamType) bool {
	return t == m.pType
}

func (meta metadata) getHookups(table *ParamTable) hookups {
	p := meta.hookupsRaw.GoSlice()[:meta.siblingsStart]
	s := meta.hookupsRaw.GoSlice()[meta.siblingsStart:meta.childrenStart]
	c := meta.hookupsRaw.GoSlice()[meta.childrenStart:]
	h := hookups{
		parents:  *(*[]ParamID)(unsafe.Pointer(&p)),
		siblings: *(*[]ParamID)(unsafe.Pointer(&s)),
		children: *(*[]ParamID)(unsafe.Pointer(&c)),
	}
	if meta.has_calc() {
		h.calc = table.calcs.Get(int(meta.calcId))
	}
	return h
}

func (meta metadata) getChildren() []ParamID {
	c := meta.hookupsRaw.GoSlice()[meta.childrenStart:]
	return *(*[]ParamID)(unsafe.Pointer(&c))
}

func (meta *metadata) appendChild(child ParamID) (childIdx int) {
	realIdx := ll.PushGetIdx(&meta.hookupsRaw, uint16(child))
	return realIdx - int(meta.childrenStart)
}

func (meta *metadata) appendChildren(children []ParamID) (firstChildIdx int) {
	cuint16 := *(*[]uint16)(unsafe.Pointer(&children))
	realIdx := ll.AppendGetStartIdxV(&meta.hookupsRaw, cuint16...)
	return realIdx - int(meta.childrenStart)
}

func (meta *metadata) insertChild(insertIdx int, child ParamID) {
	realIdx := int(meta.childrenStart) + insertIdx
	ll.InsertV(&meta.hookupsRaw, realIdx, uint16(child))
}
func (meta *metadata) insertChildren(insertIdx int, children []ParamID) {
	realIdx := int(meta.childrenStart) + insertIdx
	cuint16 := *(*[]uint16)(unsafe.Pointer(&children))
	ll.InsertV(&meta.hookupsRaw, realIdx, cuint16...)
}

func (meta *metadata) deleteChild(childIdx int) {
	realIdx := int(meta.childrenStart) + childIdx
	ll.Delete(&meta.hookupsRaw, realIdx, 1)
}
func (meta *metadata) deleteChildren(firstChildIdx int, childCount int) {
	realIdx := int(meta.childrenStart) + firstChildIdx
	ll.Delete(&meta.hookupsRaw, realIdx, childCount)
}

func (meta *metadata) checkAddParentsOrSiblings(add int) {
	if EnableSafetyChecks {
		result := int(meta.childrenStart) + add
		if result > 255 {
			fmt.Fprintf(DebugWriter, "fatal: go_param_table: adding %d siblings or parents would increase the total count of parents+siblings beyound the maximum (255), (ParamType %s, ValIdx %d)", add, typeNames[meta.pType], meta.valIdx)
			panic("FATAL")
		}
	}
}

func (meta *metadata) checkDeleteSiblings(start, count int) {
	if EnableSafetyChecks {
		sibStart := int(meta.siblingsStart) + start
		result := sibStart + count
		if result > int(meta.childrenStart) {
			fmt.Fprintf(DebugWriter, "fatal: go_param_table: cannot delete %d siblings starting at index %d, only %d siblings exist after that index, (ParamType %s, ValIdx %d)", count, start, int(meta.childrenStart)-sibStart, typeNames[meta.pType], meta.valIdx)
			panic("FATAL")
		}
	}
}
func (meta *metadata) checkDeleteParents(start, count int) {
	if EnableSafetyChecks {
		parStart := start
		result := parStart + count
		if result > int(meta.siblingsStart) {
			fmt.Fprintf(DebugWriter, "fatal: go_param_table: cannot delete %d parents starting at index %d, only %d parents exist after that index, (ParamType %s, ValIdx %d)", count, start, int(meta.siblingsStart)-parStart, typeNames[meta.pType], meta.valIdx)
			panic("FATAL")
		}
	}
}

func (meta *metadata) appendSibling(sibling ParamID) (siblingIdx int) {
	realIdx := int(meta.childrenStart)
	meta.checkAddParentsOrSiblings(1)
	meta.childrenStart += 1
	ll.InsertV(&meta.hookupsRaw, realIdx, uint16(sibling))
	return realIdx - int(meta.siblingsStart)
}

func (meta *metadata) appendSiblings(siblings []ParamID) (firstSiblingIdx int) {
	usiblings := *(*[]uint16)(unsafe.Pointer(&siblings))
	realIdx := int(meta.childrenStart)
	meta.checkAddParentsOrSiblings(len(siblings))
	meta.childrenStart += uint8(len(siblings))
	ll.InsertV(&meta.hookupsRaw, realIdx, usiblings...)
	return realIdx - int(meta.siblingsStart)
}

func (meta *metadata) insertSibling(insertIdx int, sibling ParamID) {
	realIdx := int(meta.siblingsStart) + insertIdx
	meta.checkAddParentsOrSiblings(1)
	meta.childrenStart += 1
	ll.InsertV(&meta.hookupsRaw, realIdx, uint16(sibling))
}
func (meta *metadata) insertSiblings(insertIdx int, siblings []ParamID) {
	realIdx := int(meta.siblingsStart) + insertIdx
	usiblings := *(*[]uint16)(unsafe.Pointer(&siblings))
	meta.checkAddParentsOrSiblings(len(siblings))
	meta.childrenStart += uint8(len(siblings))
	ll.InsertV(&meta.hookupsRaw, realIdx, usiblings...)
}

func (meta *metadata) deleteSibling(siblingIdx int) {
	realIdx := int(meta.siblingsStart) + siblingIdx
	meta.checkDeleteSiblings(siblingIdx, 1)
	ll.Delete(&meta.hookupsRaw, realIdx, 1)
	meta.childrenStart -= 1
}
func (meta *metadata) deleteSiblings(firstSiblingIdx int, siblingCount int) {
	realIdx := int(meta.siblingsStart) + firstSiblingIdx
	meta.checkDeleteSiblings(firstSiblingIdx, siblingCount)
	ll.Delete(&meta.hookupsRaw, realIdx, siblingCount)
	meta.childrenStart -= uint8(siblingCount)
}

func (meta *metadata) appendParent(parent ParamID) (parentIdx int) {
	realIdx := int(meta.siblingsStart)
	meta.checkAddParentsOrSiblings(1)
	meta.childrenStart += 1
	meta.siblingsStart += 1
	ll.InsertV(&meta.hookupsRaw, realIdx, uint16(parent))
	return realIdx
}

func (meta *metadata) appendParents(parents []ParamID) (firstSiblingIdx int) {
	uparents := *(*[]uint16)(unsafe.Pointer(&parents))
	realIdx := int(meta.siblingsStart)
	meta.checkAddParentsOrSiblings(len(parents))
	meta.childrenStart += uint8(len(parents))
	meta.siblingsStart += uint8(len(parents))
	ll.InsertV(&meta.hookupsRaw, realIdx, uparents...)
	return realIdx
}

func (meta *metadata) insertParent(insertIdx int, parent ParamID) {
	meta.checkAddParentsOrSiblings(1)
	meta.childrenStart += 1
	meta.siblingsStart += 1
	ll.InsertV(&meta.hookupsRaw, insertIdx, uint16(parent))
}

func (meta *metadata) insertParents(insertIdx int, parents []ParamID) {
	uparents := *(*[]uint16)(unsafe.Pointer(&parents))
	meta.checkAddParentsOrSiblings(len(parents))
	meta.childrenStart += uint8(len(parents))
	meta.siblingsStart += uint8(len(parents))
	ll.InsertV(&meta.hookupsRaw, insertIdx, uparents...)
}

func (meta *metadata) deleteParent(parentIdx int) {
	meta.checkDeleteParents(parentIdx, 1)
	ll.Delete(&meta.hookupsRaw, parentIdx, 1)
	meta.childrenStart -= 1
	meta.siblingsStart -= 1
}
func (meta *metadata) deleteParents(firstParentIdx int, parentCount int) {
	meta.checkDeleteParents(firstParentIdx, parentCount)
	ll.Delete(&meta.hookupsRaw, firstParentIdx, parentCount)
	meta.childrenStart -= uint8(parentCount)
	meta.siblingsStart -= uint8(parentCount)
}
