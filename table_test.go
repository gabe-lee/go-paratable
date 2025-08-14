package go_param_table

import (
	"testing"
)

func TestParamTable(t *testing.T) {
	EnableDebug = true
	const (
		FIRST_U64_PARAM       PIdx_U64 = PIdx_U64(iota)
		RECT_WIDTH_1                   // example root val
		RECT_HEIGHT_1                  // example root val
		RECT_AREA_1                    // example derived val
		RECT_AREA_DUPLICATE_1          // example derived val
		RECT_VOLUME_1                  // example derived val
		RECT_WIDTH_2                   // example root val
		RECT_HEIGHT_2                  // example root val
		RECT_AREA_2                    // example derived val
		RECT_VOLUME_2                  // example derived val
		RECT_SUM_VOLUME                // example derived val
		RECT_AREA_CYCLIC               // example derived val
		// ... more uint64 param indexes
		_U64_PARAMS_END
	)

	const (
		FIRST_I64_PARAM PIdx_I64 = PIdx_I64(iota + _U64_PARAMS_END)
		// ... more uint64 param indexes
		_I64_PARAMS_END
	)

	const (
		FIRST_F64_PARAM PIdx_F64 = PIdx_F64(iota + _I64_PARAMS_END)
		// ... more float64 param indexes
		_F64_PARAMS_END
	)

	const (
		FIRST_PTR_PARAM PIdx_Ptr = PIdx_Ptr(iota + _I64_PARAMS_END)
		// ... more float64 param indexes
		_PTR_PARAMS_END
	)

	const (
		FIRST_U32_PARAM PIdx_U32 = PIdx_U32(iota + _PTR_PARAMS_END)
		RECT_DEPTH_1             // example root val
		RECT_DEPTH_2             // example root val
		// ... more uint32 param indexes
		_U32_PARAMS_END
	)

	const (
		FIRST_I32_PARAM PIdx_I32 = PIdx_I32(iota + _U32_PARAMS_END)
		// ... more int32 param indexes
		_I32_PARAMS_END
	)

	const (
		FIRST_F32_PARAM PIdx_F32 = PIdx_F32(iota + _I32_PARAMS_END)
		// ... more float32 param indexes
		_F32_PARAMS_END
	)

	const (
		FIRST_U16_PARAM PIdx_U16 = PIdx_U16(iota + _F32_PARAMS_END)
		// ... more uint16 param indexes
		_U16_PARAMS_END
	)

	const (
		FIRST_I16_PARAM PIdx_I16 = PIdx_I16(iota + _U16_PARAMS_END)
		// ... more uint16 param indexes
		_I16_PARAMS_END
	)

	const (
		FIRST_U8_PARAM PIdx_U8 = PIdx_U8(iota + _I16_PARAMS_END)
		// ... more uint8 param indexes
		_U8_PARAMS_END
	)

	const (
		FIRST_I8_PARAM PIdx_I8 = PIdx_I8(iota + _U8_PARAMS_END)
		// ... more uint8 param indexes
		_I8_PARAMS_END
	)

	const (
		FIRST_BOOL_PARAM PIdx_Bool = PIdx_Bool(iota + _I8_PARAMS_END)
		// ... more bool param indexes
		_BOOL_PARAMS_END
	)

	const (
		_CALC_INVALID                PIdx_Calc = PIdx_Calc(iota) //example uninit calc for testing
		_CALC_AREA_OF_RECTANGLE                                  // example calc
		_CALC_VOLUME_OF_RECTANGLE                                // example calc
		_CALC_SUM_OF_VOLUME_OF_RECTS                             // example calc
		// ... more calculation indexes
		_CALC_COUNT
	)

	const (
		_IN_AREA_OF_RECTANGLE_WIDTH uint16 = iota
		_IN_AREA_OF_RECTANGLE_HEIGHT
	)
	const (
		_OUT_AREA_OF_RECTANGLE_AREA uint16 = iota
	)

	var _INS_AREA_OF_RECTANGLE_1 = [...]uint16{
		_IN_AREA_OF_RECTANGLE_WIDTH:  uint16(RECT_WIDTH_1),
		_IN_AREA_OF_RECTANGLE_HEIGHT: uint16(RECT_HEIGHT_1),
	}
	var _OUTS_AREA_OF_RECTANGLE_1 = [...]uint16{
		_OUT_AREA_OF_RECTANGLE_AREA: uint16(RECT_AREA_1),
	}
	var _OUTS_AREA_OF_RECTANGLE_1_DUPE = [...]uint16{
		_OUT_AREA_OF_RECTANGLE_AREA: uint16(RECT_AREA_DUPLICATE_1),
	}

	var _INS_AREA_OF_RECTANGLE_2 = [...]uint16{
		_IN_AREA_OF_RECTANGLE_WIDTH:  uint16(RECT_WIDTH_2),
		_IN_AREA_OF_RECTANGLE_HEIGHT: uint16(RECT_HEIGHT_2),
	}
	var _OUTS_AREA_OF_RECTANGLE_2 = [...]uint16{
		_OUT_AREA_OF_RECTANGLE_AREA: uint16(RECT_AREA_2),
	}

	var _INS_AREA_OF_RECTANGLE_CYCLIC = [...]uint16{
		_IN_AREA_OF_RECTANGLE_WIDTH:  uint16(RECT_AREA_CYCLIC),
		_IN_AREA_OF_RECTANGLE_HEIGHT: uint16(RECT_HEIGHT_2),
	}
	var _OUTS_AREA_OF_RECTANGLE_CYCLIC = [...]uint16{
		_OUT_AREA_OF_RECTANGLE_AREA: uint16(RECT_AREA_CYCLIC),
	}

	const (
		_IN_VOLUME_OF_RECTANGLE_AREA uint16 = iota
		_IN_VOLUME_OF_RECTANGLE_DEPTH
	)

	const (
		_OUT_VOLUME_OF_RECTANGLE_VOLUME uint16 = iota
	)

	var _INS_VOLUME_OF_RECTANGLE_1 = [...]uint16{
		_IN_VOLUME_OF_RECTANGLE_AREA:  uint16(RECT_AREA_1),
		_IN_VOLUME_OF_RECTANGLE_DEPTH: uint16(RECT_DEPTH_1),
	}
	var _OUTS_VOLUME_OF_RECTANGLE_1 = [...]uint16{
		_OUT_VOLUME_OF_RECTANGLE_VOLUME: uint16(RECT_VOLUME_1),
	}

	var _INS_VOLUME_OF_RECTANGLE_2 = [...]uint16{
		_IN_VOLUME_OF_RECTANGLE_AREA:  uint16(RECT_AREA_2),
		_IN_VOLUME_OF_RECTANGLE_DEPTH: uint16(RECT_DEPTH_2),
	}
	var _OUTS_VOLUME_OF_RECTANGLE_2 = [...]uint16{
		_OUT_VOLUME_OF_RECTANGLE_VOLUME: uint16(RECT_VOLUME_2),
	}

	const (
		_IN_SUM_OF_RECT_VOLUME_1 uint16 = iota
		_IN_SUM_OF_RECT_VOLUME_2
	)

	const (
		_OUT_SUM_OF_RECT_VOLUME_SUM uint16 = iota
	)

	var _INS_SUM_OF_RECT_VOLUMES_1_2 = [...]uint16{
		_IN_SUM_OF_RECT_VOLUME_1: uint16(RECT_VOLUME_1),
		_IN_SUM_OF_RECT_VOLUME_2: uint16(RECT_VOLUME_2),
	}
	var _OUTS_SUM_OF_RECT_VOLUMES_1_2 = [...]uint16{
		_OUT_SUM_OF_RECT_VOLUME_SUM: uint16(RECT_SUM_VOLUME),
	}

	var tooLongHookup [277]uint16

	var MyParamTable = NewParamTable(_U64_PARAMS_END, _I64_PARAMS_END, _F64_PARAMS_END, _PTR_PARAMS_END, _U32_PARAMS_END, _I32_PARAMS_END, _F32_PARAMS_END, _U16_PARAMS_END, _I16_PARAMS_END, _U8_PARAMS_END, _I8_PARAMS_END, _BOOL_PARAMS_END, _CALC_COUNT)
	var InitMyParamTable func() = func() {
		// Register all calculations first
		MyParamTable.RegisterCalc(_CALC_AREA_OF_RECTANGLE, func(t *CalcInterface) {
			width := t.GetInput_U64(_IN_AREA_OF_RECTANGLE_WIDTH)   // first input
			height := t.GetInput_U64(_IN_AREA_OF_RECTANGLE_HEIGHT) // second input
			area := width * height
			t.SetOutput_U64(_OUT_AREA_OF_RECTANGLE_AREA, area)
		})
		MyParamTable.RegisterCalc(_CALC_VOLUME_OF_RECTANGLE, func(t *CalcInterface) {
			area := t.GetInput_U64(_IN_VOLUME_OF_RECTANGLE_AREA)   // first input
			depth := t.GetInput_U32(_IN_VOLUME_OF_RECTANGLE_DEPTH) // second input
			volume := area * uint64(depth)
			t.SetOutput_U64(_OUT_VOLUME_OF_RECTANGLE_VOLUME, volume)
		})
		MyParamTable.RegisterCalc(_CALC_SUM_OF_VOLUME_OF_RECTS, func(t *CalcInterface) {
			vol1 := t.GetInput_U64(_IN_SUM_OF_RECT_VOLUME_1) // first input
			vol2 := t.GetInput_U64(_IN_SUM_OF_RECT_VOLUME_2) // second input
			sumVol := vol1 + vol2
			t.SetOutput_U64(_OUT_SUM_OF_RECT_VOLUME_SUM, sumVol)
		})
		// Init root values
		MyParamTable.InitRoot_U64(RECT_WIDTH_1, 600, false)  // width
		MyParamTable.InitRoot_U64(RECT_HEIGHT_1, 777, false) // height
		MyParamTable.InitRoot_U32(RECT_DEPTH_1, 0, false)    // depth
		MyParamTable.InitRoot_U64(RECT_WIDTH_2, 42, false)   // width
		MyParamTable.InitRoot_U64(RECT_HEIGHT_2, 99, false)  // height
		MyParamTable.InitRoot_U32(RECT_DEPTH_2, 0, false)    // depth
		// Init derived values
		MyParamTable.InitDerived_U64(RECT_AREA_1, false, _CALC_AREA_OF_RECTANGLE, _INS_AREA_OF_RECTANGLE_1[:], _OUTS_AREA_OF_RECTANGLE_1[:])
		MyParamTable.InitDerived_U64(RECT_VOLUME_1, false, _CALC_VOLUME_OF_RECTANGLE, _INS_VOLUME_OF_RECTANGLE_1[:], _OUTS_VOLUME_OF_RECTANGLE_1[:])
		MyParamTable.InitDerived_U64(RECT_AREA_2, false, _CALC_AREA_OF_RECTANGLE, _INS_AREA_OF_RECTANGLE_2[:], _OUTS_AREA_OF_RECTANGLE_2[:])
		MyParamTable.InitDerived_U64(RECT_VOLUME_2, false, _CALC_VOLUME_OF_RECTANGLE, _INS_VOLUME_OF_RECTANGLE_2[:], _OUTS_VOLUME_OF_RECTANGLE_2[:])
		MyParamTable.InitDerived_U64(RECT_SUM_VOLUME, false, _CALC_SUM_OF_VOLUME_OF_RECTS, _INS_SUM_OF_RECT_VOLUMES_1_2[:], _OUTS_SUM_OF_RECT_VOLUMES_1_2[:])
		MyParamTable.InitDerived_U64(RECT_AREA_DUPLICATE_1, false, _CALC_AREA_OF_RECTANGLE, _INS_AREA_OF_RECTANGLE_1[:], _OUTS_AREA_OF_RECTANGLE_1_DUPE[:])
	}
	InitMyParamTable()
	var expectRoot_U64 = func(idx PIdx_U64, val uint64) {
		gotVal := MyParamTable.Get_U64(idx)
		if gotVal != val {
			t.Errorf("root value error:\n\tEXP: %d\n\tGOT: %d", val, gotVal)
		}
	}
	var expectRoot_U32 = func(idx PIdx_U32, val uint32) {
		gotVal := MyParamTable.Get_U32(idx)
		if gotVal != val {
			t.Errorf("root value error:\n\tEXP: %d\n\tGOT: %d", val, gotVal)
		}
	}
	var expectDerived_U64 = func(idx PIdx_U64, val uint64) {
		gotVal := MyParamTable.Get_U64(idx)
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
	MyParamTable.SetRoot_U64(RECT_WIDTH_1, 555)
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
	MyParamTable.SetRoot_U32(RECT_DEPTH_1, 411)
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
	MyParamTable.SetRoot_U32(RECT_DEPTH_2, 35)
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
	MyParamTable.SetRoot_U64(RECT_HEIGHT_1, 51)
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
		MyParamTable.SetRoot_Bool(PIdx_Bool(RECT_HEIGHT_1), true)
	}()
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("setting root value with idx out of range did not cause panic with EnableDebug == true")
			}
		}()
		MyParamTable.SetRoot_Bool(PIdx_Bool(1000), true)
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
		var _ = MyParamTable.getCalc(_CALC_INVALID)
	}()
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("re-registering already registered calc did not cause panic with EnableDebug == true")
			}
		}()
		MyParamTable.RegisterCalc(_CALC_AREA_OF_RECTANGLE, func(t *CalcInterface) {})
	}()
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("causing derived cyclic loop did not cause panic with EnableDebug == true")
			}
		}()
		MyParamTable.InitDerived_U64(RECT_AREA_CYCLIC, false, _CALC_AREA_OF_RECTANGLE, _INS_AREA_OF_RECTANGLE_CYCLIC[:], _OUTS_AREA_OF_RECTANGLE_CYCLIC[:])
	}()
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("out-of-order NewParamTable did not cause panic with EnableDebug == true")
			}
		}()
		var _ = NewParamTable(PIdx_U64(_U32_PARAMS_END), _I64_PARAMS_END, _F64_PARAMS_END, _PTR_PARAMS_END, PIdx_U32(_U64_PARAMS_END), _I32_PARAMS_END, _F32_PARAMS_END, _U16_PARAMS_END, _I16_PARAMS_END, _U8_PARAMS_END, _I8_PARAMS_END, _BOOL_PARAMS_END, _CALC_COUNT)
	}()
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("overlong parent length did not cause panic with EnableDebug == true")
			}
		}()
		MyParamTable.InitDerived_F64(FIRST_F64_PARAM, false, _CALC_AREA_OF_RECTANGLE, tooLongHookup[:], _OUTS_AREA_OF_RECTANGLE_CYCLIC[:])
	}()
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("overlong outputs length did not cause panic with EnableDebug == true")
			}
		}()
		MyParamTable.InitDerived_I8(FIRST_I8_PARAM, false, _CALC_AREA_OF_RECTANGLE, _INS_AREA_OF_RECTANGLE_CYCLIC[:], tooLongHookup[:])
	}()
	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("getting uninit val did not cause panic with EnableDebug == true")
			}
		}()
		MyParamTable.Get_I16(FIRST_I16_PARAM)
	}()
	t.Logf("TestTable MEM: %d", MyParamTable.TotalMemoryFootprint())
}
