package go_param_table

import (
	"testing"
)

func TestREADME_Ex1(t *testing.T) {

	const (
		// 'root' values
		Px PIdx_F32 = PIdx_F32(iota)
		Py
		Pw
		Ph
		// 'derived' values
		Bx
		By
		Bw
		Bh
		// important, used for table initialization, can have any name
		_F32_PARAMS_END
	)

	const (
		CalcVal_Plus_32_0 = PIdx_Calc(iota)
		CalcVal_Mult_0_50_Sub_64_0
		// important, used for table initialization, can have any name
		_CALC_COUNT
	)

	const _end uint16 = uint16(_F32_PARAMS_END)

	var MyParamTable = func() ParamTable {
		// This is the default library setting, used for identifying table issues during development
		EnableDebug = true
		// each parameter denotes the END index for the data type, see the full template for details
		table := NewParamTable(
			PIdx_U64(0),
			PIdx_I64(0),
			PIdx_F64(0),
			PIdx_Addr(0),
			PIdx_U32(0),
			PIdx_I32(0),
			_F32_PARAMS_END,
			PIdx_U16(_end),
			PIdx_I16(_end),
			PIdx_U8(_end),
			PIdx_I8(_end),
			PIdx_Bool(_end),
			_CALC_COUNT)
		// First, register your calculation/update functions
		table.RegisterCalc(CalcVal_Plus_32_0, func(c *CalcInterface) {
			inVal := c.GetInput_F32(0) // get the first function parameter as a float32
			outVal := inVal + 32.0
			c.SetOutput_F32(0, outVal) // set the first function output to outVal
		})
		table.RegisterCalc(CalcVal_Mult_0_50_Sub_64_0, func(c *CalcInterface) {
			inVal := c.GetInput_F32(0) // get the first function parameter as a float32
			outVal := (inVal * 0.5) - 64.0
			c.SetOutput_F32(0, outVal) // set the first function output to outVal
		})
		// Root values are initialized next
		table.InitRoot_F32(Px, 100.0, false)
		table.InitRoot_F32(Py, 200.0, false)
		table.InitRoot_F32(Pw, 800.0, false)
		table.InitRoot_F32(Ph, 600.0, false)
		// Derived values are initialized last
		//   - (These function calls can be cumbersome, it is recomended to use the pattern(s)
		//      shown in the template sample instead)
		table.InitDerived_F32(Bx, false, CalcVal_Plus_32_0, []uint16{uint16(Px)}, []uint16{uint16(Bx)})
		table.InitDerived_F32(By, false, CalcVal_Plus_32_0, []uint16{uint16(Py)}, []uint16{uint16(By)})
		table.InitDerived_F32(Bw, false, CalcVal_Mult_0_50_Sub_64_0, []uint16{uint16(Pw)}, []uint16{uint16(Bw)})
		table.InitDerived_F32(Bh, false, CalcVal_Mult_0_50_Sub_64_0, []uint16{uint16(Ph)}, []uint16{uint16(Bh)})
		return table
	}()

	// 	var print = func(table *ParamTable) {
	// 		fmt.Printf(`
	//    X     Y     W     H
	// P: %.1f %.1f %.1f %.1f
	// B: %.1f %.1f %.1f %.1f`,
	// 			table.Get_F32(Px), table.Get_F32(Py), table.Get_F32(Pw), table.Get_F32(Ph),
	// 			table.Get_F32(Bx), table.Get_F32(By), table.Get_F32(Bw), table.Get_F32(Bh),
	// 		)
	// 	}

	// print(&MyParamTable)

	if MyParamTable.Get_F32(Bx) != 132.0 {
		t.Errorf("example failed: Bx: %f != 132.0", MyParamTable.Get_F32(Bx))
	}
	if MyParamTable.Get_F32(By) != 232.0 {
		t.Errorf("example failed: By: %f != 232.0", MyParamTable.Get_F32(By))
	}
	if MyParamTable.Get_F32(Bw) != 336.0 {
		t.Errorf("example failed: Bw: %f != 336.0", MyParamTable.Get_F32(Bw))
	}
	if MyParamTable.Get_F32(Bh) != 236.0 {
		t.Errorf("example failed: Bw: %f != 336.0", MyParamTable.Get_F32(Bh))
	}
	MyParamTable.SetRoot_F32(Pw, 990.0)
	MyParamTable.SetRoot_F32(Py, 333.0)
	if MyParamTable.Get_F32(Bx) != 132.0 {
		t.Errorf("example failed: Bx: %f != 132.0", MyParamTable.Get_F32(Bx))
	}
	if MyParamTable.Get_F32(By) != 365.0 {
		t.Errorf("example failed: By: %f != 365.0", MyParamTable.Get_F32(By))
	}
	if MyParamTable.Get_F32(Bw) != 431.0 {
		t.Errorf("example failed: Bw: %f != 431.0", MyParamTable.Get_F32(Bw))
	}
	if MyParamTable.Get_F32(Bh) != 236.0 {
		t.Errorf("example failed: Bw: %f != 336.0", MyParamTable.Get_F32(Bh))
	}

	// print(&MyParamTable)
}
