package go_param_table

import (
	"fmt"
	"slices"
)

const (
	_HOOK_OFF_CALC    uint32 = 0
	_HOOK_OFF_PLEN    uint32 = 1
	_HOOK_OFF_CLEN    uint32 = 2
	_HOOK_OFF_INSTART uint32 = 3
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

type hookup uint32

func (h hookup) isInit() bool {
	return h != 0
}

func (t *ParamTable) isDerived(idx uint16) bool {
	h := t.hookups[idx]
	if !h.isInit() {
		return false
	}
	return t.hookupData[uint32(h)+_HOOK_OFF_PLEN] != 0
}
func (t *ParamTable) isRoot(idx uint16) bool {
	h := t.hookups[idx]
	if !h.isInit() {
		return false
	}
	return t.hookupData[uint32(h)+_HOOK_OFF_PLEN] == 0
}
func (t *ParamTable) hasChildren(idx uint16) bool {
	h := t.hookups[idx]
	if !h.isInit() {
		return false
	}
	return t.hookupData[uint32(h)+_HOOK_OFF_CLEN] != 0
}

func (t *ParamTable) trigger(idx uint16, prevIdxs []uint16) (updatedPrevIdxs []uint16) {
	h := t.hookups[idx]
	i := uint32(h)
	calcIdx := PIdx_Calc(t.hookupData[i])
	paramsLen := paramsLen(t.hookupData[i+_HOOK_OFF_PLEN])
	inLen := paramsLen.inLen()
	outLen := paramsLen.outLen()
	inout := i + _HOOK_OFF_INSTART
	insEnd := inout + inLen
	ins := t.hookupData[inout:insEnd]
	outs := t.hookupData[insEnd : insEnd+outLen]
	iface := CalcInterface{
		table:    t,
		inputs:   ins,
		outputs:  outs,
		prevIdxs: prevIdxs,
	}
	t.calcs[calcIdx](&iface)
	return iface.prevIdxs
}

func (t *ParamTable) getCalcFromValIdx(idx uint16) (calc ParamCalc) {
	h := t.hookups[idx]
	if !h.isInit() {
		return nil
	}
	calcIdx := t.hookupData[h]
	return t.calcs[calcIdx]
}

func (t *ParamTable) getParents(idx uint16) (parents []uint16) {
	start, end := t.getParentsLimits(idx)
	return t.hookupData[start:end]
}
func (t *ParamTable) getParentsLimits(idx uint16) (start, end uint32) {
	h := t.hookups[idx]
	i := uint32(h)
	if !h.isInit() {
		return 0, 0
	}
	paramsLen := paramsLen(t.hookupData[i+_HOOK_OFF_PLEN])
	inLen := paramsLen.inLen()
	start = i + _HOOK_OFF_INSTART
	end = start + inLen
	return
}

func (t *ParamTable) getSiblings(idx uint16) (siblings []uint16) {
	start, end := t.getSiblingsLimits(idx)
	return t.hookupData[start:end]
}
func (t *ParamTable) getSiblingsLimits(idx uint16) (start, end uint32) {
	h := t.hookups[idx]
	i := uint32(h)
	if !h.isInit() {
		return 0, 0
	}
	paramsLen := paramsLen(t.hookupData[i+_HOOK_OFF_PLEN])
	inLen := paramsLen.inLen()
	outLen := paramsLen.outLen()
	start = i + _HOOK_OFF_INSTART + inLen
	end = start + outLen
	return
}
func (t *ParamTable) getChildren(idx uint16) (children []uint16) {
	start, end := t.getChildrenLimits(idx)
	return t.hookupData[start:end]
}
func (t *ParamTable) getChildrenLimits(idx uint16) (start, end uint32) {
	h := t.hookups[idx]
	if !h.isInit() {
		return 0, 0
	}
	i := uint32(h)
	paramsLen := paramsLen(t.hookupData[i+_HOOK_OFF_PLEN])
	inLen := paramsLen.inLen()
	outLen := paramsLen.outLen()
	childLen := uint32(t.hookupData[i+_HOOK_OFF_CLEN])
	start = i + _HOOK_OFF_INSTART + inLen + outLen
	end = start + childLen
	return
}

func (t *ParamTable) addChild(idx uint16, childIdx uint16) {
	h := t.hookups[idx]
	i := uint32(h)
	if !h.isInit() {
		t.initRootHookupWithChild(idx, childIdx)
		return
	}
	start, end := t.getChildrenLimits(idx)
	for start < end {
		if t.hookupData[start] == childIdx {
			if EnableDebug {
				fmt.Fprintf(DebugWriter, "warn: go_param_table: hookup.addChild(): child idx %d was already inside hookup idx %d (owned by value idx %d)", childIdx, h, idx)
			}
			return
		}
		start += 1
	}
	for i, hh := range t.hookups {
		if uint32(hh) >= end {
			t.hookups[i] += 1
		}
	}
	t.hookupData = slices.Insert(t.hookupData, int(end), childIdx)
	t.hookupData[i+_HOOK_OFF_CLEN] += 1
	t.hookupData[end] = childIdx
}

func (t *ParamTable) removeChild(idx uint16, childIdx uint16) {
	h := t.hookups[idx]
	i := uint32(h)
	if EnableDebug {
		if !h.isInit() {
			fmt.Fprintf(DebugWriter, "warn: go_param_table: hookup.removeChild(): idx %d never had its hookup initialized", idx)
			return
		}
	}
	start, end := t.getChildrenLimits(idx)
	for start < end {
		if t.hookupData[start] == childIdx {
			t.hookupData = slices.Delete(t.hookupData, int(start), int(start+1))
			for i, hh := range t.hookups {
				if uint32(hh) >= end {
					t.hookups[i] -= 1
				}
			}
			t.hookupData[i+_HOOK_OFF_CLEN] -= 1
			return
		}
	}
	if EnableDebug {
		fmt.Fprintf(DebugWriter, "warn: go_param_table: hookup.removeChild(): child idx %d was not inside hookup idx %d (owned by value idx %d)", childIdx, h, idx)
	}
}
