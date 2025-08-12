# goparatable
A parameter heirarchy table, featuring 'root' values and auto-updating 'derived' values based on user registered calculation functions, written in golang

  - [What is it and why?](#what-is-it-and-why)
  - [Installation](#installation)
  - [Examples](#examples)
    - [Example 1: UI Positioning](#example-1-ui-positioning)
  - [Pros/Cons/Caveats](#prosconscaveats)
  - [Quickstart/Template](#quickstarttemplate)
  - [Future Plans/TODO](#future-planstodo)

## What is it and why?

This package provides a means of automatic app state management in a compact form-factor. The user declares constant integer id's for app state parameters, which can either be 'root' values with no 'parent' values, or 'derived' values that are automatically (re)calculated when their parent values are updated, using user-registered calculation functions.

This package uses minimal special features of golang, relying entirely on base syntax features to structure the table to minimize code size, memory footprint, and execution speed.

## Installation
From within your project directory, use the following command:

```go get github.com/gabe-lee/goparatable@latest```

Then simply import it wherever it is needed:
```golang
import "github.com/gabe-lee/goparatable"
```

[Back to Top](#goparatable)
## Examples
#### Example 1: UI Positioning

Suppose you had a button that you wanted to ALWAYS position and size relative to it's parent using the following formulas:
  - Parent X Pos:  `Px`
  - Parent Y Pos:  `Py`
  - Parent Width:  `Pw`
  - Parent Height: `Ph`
  - Button X Pos: `Bx = (Px + 32.0)`
  - Button Y Pos: `By = (Py + 32.0)`
  - Button Width: `Bw = (Pw * 0.50) - 64.0`
  - Button Height: `Bh = (Ph * 0.50) - 64.0`

This can be achieved with this package like so:

```golang
import para "github.com/gabe-lee/goparatable"

type (
    PIdx_F32 = para.PIdx_F32
    ParamTable = para.ParamTable
    CalcInterface = para.CalcInterface
)

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
    para.EnableDebug = true
    // each parameter denotes the END index for the data type, see the full template for details
    table := para.NewParamTable(
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

func main() {
    printTable(&MyParamTable)
    //    X     Y     W     H
    // P: 100.0 200.0 800.0 600.0
    // B: 132.0 232.0 336.0 236.0
}

func printTable(table *ParamTable) {
    fmt.Printf(`
   X     Y     W     H
P: %.1f %.1f %.1f %.1f 
B: %.1f %.1f %.1f %.1f`,
        table.Get_F32(Px), table.Get_F32(Py), table.Get_F32(Pw), table.Get_F32(Ph),
        table.Get_F32(Bx), table.Get_F32(By), table.Get_F32(Bw), table.Get_F32(Bh),
    )
}
```

While this seems like an incredibly convoluted and cumbersome way to achieve such a simple calculation, the benefits begin to show themselves when you then need to update the parent dimensions and immediately need the updated button dimensions as a result (for example to draw the next frame)

```golang
func main() {
    // ...previous results
    MyParamTable.SetRoot_F32(Pw, 990.0) // change parent width
	MyParamTable.SetRoot_F32(Py, 333.0) // change parent y position
    printTable(&MyParamTable)
    //    X     Y     W     H
    // P: 100.0 333.0 990.0 600.0
    // B: 132.0 365.0 431.0 236.0
}
```

Thats it! Once the cumbersome initialization process is complete, you no longer need to worry about triggering all the update functions manually or from within some other user-defined web of structs, methods, etc.

[Back to Top](#goparatable)

## Pros/Cons/Caveats
#### Pros
  - Automatic updates of derived/calculated values when their parent values (root _or_ derived) change, no matter how deeply nested
  - Relatively small memory footprint for the functionality provided
  - Parameter ID's that are adjactent to each other are _also_ cache-local to one another
  - Uses no interfaces or type reflection
  - Safety checks enabled by default, but can be turned off using a global var in the library (`EnableDebug`) for more speed
  - Supports types: `bool, uint8, uint16, uint32, uint64, uintptr, int8, int16, int32, int64, float32, float64`
  - Zero external dependancies, bare minimum of standard library imports
  - With enough creativity, you can model nearly anything fully within this library

#### Cons
  - Limited to 65535 unique parameters
  - Limited to 65535 unique calculation functions
  - Calculation functions can have a maximum of 255 inputs and 255 outputs
  - Cumbersome (but straight-forward) to initially set-up
  - Arrays, Slices, and Struct types not directly supported (but can be used by either separating each struct fields into a parameter, or using `uintptr` and `unsafe` to load/store struct/array/slice pointers/lengths/capacities)
  - Currently updates are performed recursively (this is planned to change in the future)

#### Caveats
  - Safety checks always cause panics, since most if not all errors
  covered by the safety checking would be due to programmer error when
  performing initialization or type mismatches on `Get_()`/`Set_()` functions. This prevents error handling bloat while providing all the error checking most projects would require, and additional error checking can be user defined inside the function bodies of the calculation functions themselves

[Back to Top](#goparatable)
## Quickstart/Template
Below is provided a template for jump-starting a new ParamTable
```golang
import "github.com/gabe-lee/goparatable"

type (
	ParamCalc     = goparatable.ParamCalc
	PIdx_U64      = goparatable.PIdx_U64
	PIdx_I64      = goparatable.PIdx_I64
	PIdx_F64      = goparatable.PIdx_F64
	PIdx_Addr     = goparatable.PIdx_Addr
	PIdx_U32      = goparatable.PIdx_U32
	PIdx_I32      = goparatable.PIdx_I32
	PIdx_F32      = goparatable.PIdx_F32
	PIdx_U16      = goparatable.PIdx_U16
	PIdx_I16      = goparatable.PIdx_I16
	PIdx_U8       = goparatable.PIdx_U8
	PIdx_I8       = goparatable.PIdx_I8
	PIdx_Bool     = goparatable.PIdx_Bool
	PIdx_Calc     = goparatable.PIdx_Calc
	ParamTable    = goparatable.ParamTable
	CalcInterface = goparatable.CalcInterface
)

const (
	FIRST_U64_PARAM PIdx_U64 = PIdx_U64(iota)
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
	FIRST_ADDR_PARAM PIdx_Addr = PIdx_Addr(iota + _F64_PARAMS_END)
	// ... more uintptr param indexes
	_ADDR_PARAMS_END
)

const (
	FIRST_U32_PARAM PIdx_U32 = PIdx_U32(iota + _ADDR_PARAMS_END)
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
	_FIRST_CALC = PIdx_Calc(iota)
	// ... more calculation indexes
	_CALC_COUNT
)

// ************
// FIRST CALC *
// ************

// user-defined helper consts for _FIRST_CALC input indexes
// not strictly necessary if you are sure to put the parameters in the correct order
// at all call sites
const (
	_IN_FIRST_CALC_A uint16 = iota
	_IN_FIRST_CALC_B
)

// user-defined helper consts for _FIRST_CALC output indexes
// not strictly necessary if you are sure to put the parameters in the correct order
// at all call sites
const (
	_OUT_FIRST_CALC uint16 = iota
)

// User-created helper func for initializing derived values that use _FIRST_CALC
func initFirstCalcDerived(table *ParamTable, alwaysUpdate bool, a PIdx_U64, b PIdx_I64, out PIdx_I32) {
	ins := []uint16{
		_IN_FIRST_CALC_A: uint16(a),
		_IN_FIRST_CALC_B: uint16(b),
	}
	outs := []uint16{
		_OUT_FIRST_CALC: uint16(out),
	}
	table.InitDerived_I32(out, alwaysUpdate, _FIRST_CALC, ins, outs)
}

func InitMyParamTable() ParamTable {
	// This is already the default, but you can set to `false` after you have tested your
	// table and want more speed
	goparatable.EnableDebug = true
	// Initialize table with type index ends (each parameter except the last (calcsCount) must be >= the previous parameter)
	table := goparatable.NewParamTable(_U64_PARAMS_END, _I64_PARAMS_END, _F64_PARAMS_END, _ADDR_PARAMS_END, _U32_PARAMS_END, _I32_PARAMS_END, _F32_PARAMS_END, _U16_PARAMS_END, _I16_PARAMS_END, _U8_PARAMS_END, _I8_PARAMS_END, _BOOL_PARAMS_END, _CALC_COUNT)
	// Register all calculations first
	table.RegisterCalc(_FIRST_CALC, func(t *CalcInterface) {
		vala := t.GetInput_U64(_IN_FIRST_CALC_A) // first input
		valb := t.GetInput_I64(_IN_FIRST_CALC_B) // second input
		valout := int32(vala) + int32(valb)
		t.SetOutput_I32(_OUT_FIRST_CALC, valout)
	})
	// Init root values second
	table.InitRoot_U64(FIRST_U64_PARAM, 1, false)
	table.InitRoot_I64(FIRST_I64_PARAM, -2, true)
	// Init derived values last
    // Here the user-defined helper function is used,
    // but you can call `table.InitDerived_XXX()` manually,
    // or with any other helper code pattern desired
	initFirstCalcDerived(&table, false, FIRST_U64_PARAM, FIRST_I64_PARAM, FIRST_I32_PARAM)
	return table
}

var MyParamTable = InitMyParamTable()
```
[Back to Top](#goparatable)
## Future Plans/TODO
  - [x] Core functionality `ParamTable`
  - [ ] Optional UI Layout system that uses `ParamTable`
  - [ ] Additional reduction of memory footprint?
  - [ ] Non-recursive update algorithm
  - [ ] Direct support for arrays/pointers?