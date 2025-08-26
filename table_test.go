package go_param_table

import (
	"testing"
)

func TestParamTable(t *testing.T) {
	EnableSafetyChecks = true
	const (
		RECT_WIDTH_1          = iota // example root val
		RECT_HEIGHT_1                // example root val
		RECT_AREA_1                  // example derived val
		RECT_AREA_DUPLICATE_1        // example derived val
		RECT_VOLUME_1                // example derived val
		RECT_WIDTH_2                 // example root val
		RECT_HEIGHT_2                // example root val
		RECT_AREA_2                  // example derived val
		RECT_VOLUME_2                // example derived val
		RECT_SUM_VOLUME              // example derived val
		RECT_AREA_CYCLIC             // example derived val
		RECT_DEPTH_1                 // example root val
		RECT_DEPTH_2                 // example root val
		// ... more param indexes
		_PARAM_COUNT
	)

	var ID [_PARAM_COUNT]ParamID

	const (
		CALC_AREA_OF_RECTANGLE      = iota // example calc
		CALC_VOLUME_OF_RECTANGLE           // example calc
		CALC_SUM_OF_VOLUME_OF_RECTS        // example calc
		// ... more calculation indexes
		_CALC_COUNT
	)

	var CALC [_CALC_COUNT]CalcID

	const (
		_IN_AREA_OF_RECTANGLE_WIDTH uint16 = iota
		_IN_AREA_OF_RECTANGLE_HEIGHT
	)
	const (
		_OUT_AREA_OF_RECTANGLE_AREA uint16 = iota
	)

	const (
		_IN_VOLUME_OF_RECTANGLE_AREA uint16 = iota
		_IN_VOLUME_OF_RECTANGLE_DEPTH
	)

	const (
		_OUT_VOLUME_OF_RECTANGLE_VOLUME uint16 = iota
	)

	const (
		_IN_SUM_OF_RECT_VOLUME_1 uint16 = iota
		_IN_SUM_OF_RECT_VOLUME_2
	)

	const (
		_OUT_SUM_OF_RECT_VOLUME_SUM uint16 = iota
	)

	var tooLongHookup [277]ParamID
	var tooLongHookupLinked [277]TypeInit

	var MyParamTable = NewParamTable(_PARAM_COUNT)
	var InitMyParamTable func() = func() {
		// Init root values first (or second)
		ID[RECT_WIDTH_1] = MyParamTable.InitRoot_U64(600, false)  // width
		ID[RECT_HEIGHT_1] = MyParamTable.InitRoot_U64(777, false) // height
		ID[RECT_DEPTH_1] = MyParamTable.InitRoot_U32(0, false)    // depth
		ID[RECT_WIDTH_2] = MyParamTable.InitRoot_U64(42, false)   // width
		ID[RECT_HEIGHT_2] = MyParamTable.InitRoot_U64(99, false)  // height_IN_VOLUME_OF_RECTANGLE_DEPTH
		ID[RECT_DEPTH_2] = MyParamTable.InitRoot_U32(0, false)    // depth
		// Register all calculations second (or first)
		CALC[CALC_AREA_OF_RECTANGLE] = MyParamTable.RegisterCalc(func(t *CalcInterface) {
			width := t.GetInput_U64(_IN_AREA_OF_RECTANGLE_WIDTH)   // first input
			height := t.GetInput_U64(_IN_AREA_OF_RECTANGLE_HEIGHT) // second input
			area := width * height
			t.SetOutput_U64(_OUT_AREA_OF_RECTANGLE_AREA, area)
		})
		CALC[CALC_VOLUME_OF_RECTANGLE] = MyParamTable.RegisterCalc(func(t *CalcInterface) {
			area := t.GetInput_U64(_IN_VOLUME_OF_RECTANGLE_AREA)   // first input
			depth := t.GetInput_U32(_IN_VOLUME_OF_RECTANGLE_DEPTH) // second input
			volume := area * uint64(depth)
			t.SetOutput_U64(_OUT_VOLUME_OF_RECTANGLE_VOLUME, volume)
		})
		CALC[CALC_SUM_OF_VOLUME_OF_RECTS] = MyParamTable.RegisterCalc(func(t *CalcInterface) {
			vol1 := t.GetInput_U64(_IN_SUM_OF_RECT_VOLUME_1) // first input
			vol2 := t.GetInput_U64(_IN_SUM_OF_RECT_VOLUME_2) // second input
			sumVol := vol1 + vol2
			t.SetOutput_U64(_OUT_SUM_OF_RECT_VOLUME_SUM, sumVol)
		})
		// Init derived values last, derived values that require other derived values must be initialized after the ones they require
		ID[RECT_AREA_1] = MyParamTable.InitDerived_U64(false, CALC[CALC_AREA_OF_RECTANGLE], []ParamID{_IN_AREA_OF_RECTANGLE_HEIGHT: ID[RECT_HEIGHT_1], _IN_AREA_OF_RECTANGLE_WIDTH: ID[RECT_WIDTH_1]}...)
		ID[RECT_VOLUME_1] = MyParamTable.InitDerived_U64(false, CALC[CALC_VOLUME_OF_RECTANGLE], []ParamID{_IN_VOLUME_OF_RECTANGLE_AREA: ID[RECT_AREA_1], _IN_VOLUME_OF_RECTANGLE_DEPTH: ID[RECT_DEPTH_1]}...)
		ID[RECT_AREA_2] = MyParamTable.InitDerived_U64(false, CALC[CALC_AREA_OF_RECTANGLE], []ParamID{_IN_AREA_OF_RECTANGLE_HEIGHT: ID[RECT_HEIGHT_2], _IN_AREA_OF_RECTANGLE_WIDTH: ID[RECT_WIDTH_2]}...)
		ID[RECT_VOLUME_2] = MyParamTable.InitDerived_U64(false, CALC[CALC_VOLUME_OF_RECTANGLE], []ParamID{_IN_VOLUME_OF_RECTANGLE_AREA: ID[RECT_AREA_2], _IN_VOLUME_OF_RECTANGLE_DEPTH: ID[RECT_DEPTH_2]}...)
		ID[RECT_SUM_VOLUME] = MyParamTable.InitDerived_U64(false, CALC[CALC_SUM_OF_VOLUME_OF_RECTS], []ParamID{_IN_SUM_OF_RECT_VOLUME_1: ID[RECT_VOLUME_1], _IN_SUM_OF_RECT_VOLUME_2: ID[RECT_VOLUME_2]}...)
		ID[RECT_AREA_DUPLICATE_1] = MyParamTable.InitDerived_U64(false, CALC[CALC_AREA_OF_RECTANGLE], []ParamID{_IN_AREA_OF_RECTANGLE_HEIGHT: ID[RECT_HEIGHT_1], _IN_AREA_OF_RECTANGLE_WIDTH: ID[RECT_WIDTH_1]}...)
	}
	InitMyParamTable()
	var expectRoot_U64 = func(idx ParamID, val uint64) {
		gotVal := MyParamTable.Get_U64(ID[idx])
		if gotVal != val {
			t.Errorf("root value error:\n\tEXP: %d\n\tGOT: %d", val, gotVal)
		}
	}
	var expectRoot_U32 = func(idx ParamID, val uint32) {
		gotVal := MyParamTable.Get_U32(ID[idx])
		if gotVal != val {
			t.Errorf("root value error:\n\tEXP: %d\n\tGOT: %d", val, gotVal)
		}
	}
	var expectDerived_U64 = func(idx ParamID, val uint64) {
		gotVal := MyParamTable.Get_U64(ID[idx])
		if gotVal != val {
			t.Errorf("derived value error:\n\tEXP: %d\n\tGOT: %d", val, gotVal)
		}
	}
	expectRoot_U64(RECT_WIDTH_1, 600)
	expectRoot_U64(RECT_HEIGHT_1, 777)
	expectDerived_U64(RECT_AREA_1, 466200)
	expectDerived_U64(RECT_AREA_DUPLICATE_1, 466200)
	expectRoot_U32(RECT_DEPTH_1, 0)
	expectDerived_U64(RECT_VOLUME_1, 0)
	expectRoot_U64(RECT_WIDTH_2, 42)
	expectRoot_U64(RECT_HEIGHT_2, 99)
	expectDerived_U64(RECT_AREA_2, 4158)
	expectRoot_U32(RECT_DEPTH_2, 0)
	expectDerived_U64(RECT_VOLUME_2, 0)
	expectDerived_U64(RECT_SUM_VOLUME, 0)
	MyParamTable.SetRoot_U64(ID[RECT_WIDTH_1], 555)
	expectRoot_U64(RECT_WIDTH_1, 555)
	expectRoot_U64(RECT_HEIGHT_1, 777)
	expectDerived_U64(RECT_AREA_1, 431235)
	expectDerived_U64(RECT_AREA_DUPLICATE_1, 431235)
	expectRoot_U32(RECT_DEPTH_1, 0)
	expectDerived_U64(RECT_VOLUME_1, 0)
	expectRoot_U64(RECT_WIDTH_2, 42)
	expectRoot_U64(RECT_HEIGHT_2, 99)
	expectDerived_U64(RECT_AREA_2, 4158)
	expectRoot_U32(RECT_DEPTH_2, 0)
	expectDerived_U64(RECT_VOLUME_2, 0)
	expectDerived_U64(RECT_SUM_VOLUME, 0)
	MyParamTable.SetRoot_U32(ID[RECT_DEPTH_1], 411)
	expectRoot_U64(RECT_WIDTH_1, 555)
	expectRoot_U64(RECT_HEIGHT_1, 777)
	expectDerived_U64(RECT_AREA_1, 431235)
	expectDerived_U64(RECT_AREA_DUPLICATE_1, 431235)
	expectRoot_U32(RECT_DEPTH_1, 411)
	expectDerived_U64(RECT_VOLUME_1, 177237585)
	expectRoot_U64(RECT_WIDTH_2, 42)
	expectRoot_U64(RECT_HEIGHT_2, 99)
	expectDerived_U64(RECT_AREA_2, 4158)
	expectRoot_U32(RECT_DEPTH_2, 0)
	expectDerived_U64(RECT_VOLUME_2, 0)
	expectDerived_U64(RECT_SUM_VOLUME, 177237585)
	MyParamTable.SetRoot_U32(ID[RECT_DEPTH_2], 35)
	expectRoot_U64(RECT_WIDTH_1, 555)
	expectRoot_U64(RECT_HEIGHT_1, 777)
	expectDerived_U64(RECT_AREA_1, 431235)
	expectDerived_U64(RECT_AREA_DUPLICATE_1, 431235)
	expectRoot_U32(RECT_DEPTH_1, 411)
	expectDerived_U64(RECT_VOLUME_1, 177237585)
	expectRoot_U64(RECT_WIDTH_2, 42)
	expectRoot_U64(RECT_HEIGHT_2, 99)
	expectDerived_U64(RECT_AREA_2, 4158)
	expectRoot_U32(RECT_DEPTH_2, 35)
	expectDerived_U64(RECT_VOLUME_2, 145530)
	expectDerived_U64(RECT_SUM_VOLUME, 177383115)
	MyParamTable.SetRoot_U64(ID[RECT_HEIGHT_1], 51)
	expectRoot_U64(RECT_WIDTH_1, 555)
	expectRoot_U64(RECT_HEIGHT_1, 51)
	expectDerived_U64(RECT_AREA_1, 28305)
	expectDerived_U64(RECT_AREA_DUPLICATE_1, 28305)
	expectRoot_U32(RECT_DEPTH_1, 411)
	expectDerived_U64(RECT_VOLUME_1, 11633355)
	expectRoot_U64(RECT_WIDTH_2, 42)
	expectRoot_U64(RECT_HEIGHT_2, 99)
	expectDerived_U64(RECT_AREA_2, 4158)
	expectRoot_U32(RECT_DEPTH_2, 35)
	expectDerived_U64(RECT_VOLUME_2, 145530)
	expectDerived_U64(RECT_SUM_VOLUME, 11778885)
	// bad conditions
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("setting root value with wrong type did not cause panic with EnableDebug == true")
			}
		}()
		MyParamTable.SetRoot_Bool(ID[RECT_HEIGHT_1], true)
	}()
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("setting root value with idx out of range did not cause panic with EnableDebug == true")
			}
		}()
		MyParamTable.SetRoot_Bool(1000, true)
	}()
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("directly setting derived value did not cause panic with EnableDebug == true")
			}
		}()
		MyParamTable.SetRoot_U64(RECT_AREA_1, 1)
	}()
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("getting unregistered calc did not cause panic with EnableDebug == true")
			}
		}()
		var _ = MyParamTable.calcs.Get(int(MyParamTable.meta_table.Get(int(ID[RECT_HEIGHT_1])).calcId))
	}()
	// func() {
	// 	defer func() {
	// 		if r := recover(); r == nil {
	// 			t.Errorf("causing derived cyclic loop did not cause panic with EnableDebug == true")
	// 		}
	// 	}()
	// 	areaCyclic := MyParamTable.InitDerived_U64(false, CALC_AREA_OF_RECTANGLE, []ParamID{ID[RECT_WIDTH_1], ID[RECT_HEIGHT_1]})
	// 	MyParamTable.InitDerived_U64(false, CALC_VOLUME_OF_RECTANGLE, []ParamID{areaCyclic, ID[RECT_DEPTH_1]})
	// 	metaCyclic := MyParamTable.meta_table.Get(int(areaCyclic))
	// 	_, _, c := MyParamTable.getMetaSlices(metaCyclic)
	// 	c[0] = uint16(ID[RECT_WIDTH_1])
	// 	MyParamTable.SetRoot_U64(ID[RECT_HEIGHT_1], 1)
	// }()
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("overlong parent length did not cause panic with EnableDebug == true")
			}
		}()
		MyParamTable.InitDerived_F64(false, CALC_AREA_OF_RECTANGLE, tooLongHookup[:]...)
	}()
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("overlong outputs length did not cause panic with EnableDebug == true")
			}
		}()
		MyParamTable.InitDerived_Linked(CALC_AREA_OF_RECTANGLE, []ParamID{ID[RECT_WIDTH_1], ID[RECT_HEIGHT_1]}, tooLongHookupLinked[:])
	}()
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("getting uninit val did not cause panic with EnableDebug == true")
			}
		}()
		MyParamTable.Get_I16(7)
	}()
	t.Logf("TestTable MEM: %d", MyParamTable.TotalMemoryFootprint())
}
