package go_param_table

import (
	"testing"
)

func TestREADME_Ex1(t *testing.T) {
	EnableSafetyChecks = true

	var MyParamTable = NewParamTable(8)

	var CALC_A_PLUS_B = MyParamTable.RegisterCalc(func(c *CalcInterface) {
		a := c.GetInput_F32(0) // get the first function parameter as a float32
		b := c.GetInput_F32(1) // get the second function parameter as a float32
		out := a + b
		c.SetOutput_F32(0, out) // set the first function output
	})

	var CALC_HALF_A_MINUS_2B = MyParamTable.RegisterCalc(func(c *CalcInterface) {
		a := c.GetInput_F32(0) // get the first function parameter as a float32
		b := c.GetInput_F32(1) // get the second function parameter as a float32
		out := (a * 0.5) - (b * 2.0)
		c.SetOutput_F32(0, out) // set the first function output
	})

	var Px = MyParamTable.InitRoot_F32(100.0, false)
	var Py = MyParamTable.InitRoot_F32(200.0, false)
	var Pw = MyParamTable.InitRoot_F32(800.0, false)
	var Ph = MyParamTable.InitRoot_F32(600.0, false)

	var Margin = MyParamTable.InitRoot_F32(32.0, false)

	// Derived variables must be declared AFTER the calculations they use
	// and the root/derived values they are children of
	var Bx = MyParamTable.InitDerived_F32(false, CALC_A_PLUS_B, Px, Margin)
	var By = MyParamTable.InitDerived_F32(false, CALC_A_PLUS_B, Py, Margin)
	var Bw = MyParamTable.InitDerived_F32(false, CALC_HALF_A_MINUS_2B, Pw, Margin)
	var Bh = MyParamTable.InitDerived_F32(false, CALC_HALF_A_MINUS_2B, Ph, Margin)

	// var print = func(table *ParamTable) {
	// 	fmt.Printf(`
	//    X     Y     W     H
	// P: %.1f %.1f %.1f %.1f
	// B: %.1f %.1f %.1f %.1f`,
	// 		table.Get_F32(Px), table.Get_F32(Py), table.Get_F32(Pw), table.Get_F32(Ph),
	// 		table.Get_F32(Bx), table.Get_F32(By), table.Get_F32(Bw), table.Get_F32(Bh),
	// 	)
	// }

	// print(&MyParamTable)

	if v := MyParamTable.Get_F32(Bx); v != 132.0 {
		t.Fatalf("example failed: Bx: %f != 132.0", v)
	}
	if v := MyParamTable.Get_F32(By); v != 232.0 {
		t.Fatalf("example failed: By: %f != 232.0", v)
	}
	if v := MyParamTable.Get_F32(Bw); v != 336.0 {
		t.Fatalf("example failed: Bw: %f != 336.0", v)
	}
	if v := MyParamTable.Get_F32(Bh); v != 236.0 {
		t.Fatalf("example failed: Bh: %f != 236.0", v)
	}
	MyParamTable.SetRoot_F32(Pw, 990.0)
	MyParamTable.SetRoot_F32(Py, 333.0)
	MyParamTable.SetRoot_F32(Margin, 48.0)
	if MyParamTable.Get_F32(Bx) != 148.0 {
		t.Fatalf("example failed: Bx: %f != 148.0", MyParamTable.Get_F32(Bx))
	}
	if MyParamTable.Get_F32(By) != 381.0 {
		t.Fatalf("example failed: By: %f != 381.0", MyParamTable.Get_F32(By))
	}
	if MyParamTable.Get_F32(Bw) != 399.0 {
		t.Fatalf("example failed: Bw: %f != 399.0", MyParamTable.Get_F32(Bw))
	}
	if MyParamTable.Get_F32(Bh) != 204.0 {
		t.Fatalf("example failed: Bh: %f != 204.0", MyParamTable.Get_F32(Bh))
	}
	// print(&MyParamTable)
	t.Logf("TestTable MEM: %d", MyParamTable.TotalMemoryFootprint())
}
