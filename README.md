# go_param_table
A parameter heirarchy table, featuring 'root' values and auto-updating 'derived' values based on user registered calculation functions, written in golang

  - [What is it and why?](#what-is-it-and-why)
  - [Installation](#installation)
  - [Examples](#examples)
    - [Example 1: UI Positioning](#example-1-ui-positioning)
  - [Pros/Cons/Caveats](#prosconscaveats)
  - [Changelog/Goals](#changelog)

## What is it and why?

This package provides a means of automatic app state management in a compact form-factor. The user declares constant integer id's for app state parameters, which can either be 'root' values with no 'parent' values, or 'derived' values that are automatically (re)calculated when their parent values are updated, using user-registered calculation functions.

This package uses minimal special features of golang, relying entirely on base syntax features to structure the table to minimize code size, memory footprint, and execution speed.

## Installation
From within your project directory, use the following command:

```
go get github.com/gabe-lee/go_param_table@latest
```

Then simply import it wherever it is needed:
```golang
import "github.com/gabe-lee/go_param_table"
```

[Back to Top](#go_param_table)
## Examples
#### Example 1: UI Positioning

Suppose you had a button that you wanted to ALWAYS position and size relative to it's parent using the following formulas:
  - Parent X Pos:  `Px`
  - Parent Y Pos:  `Py`
  - Parent Width:  `Pw`
  - Parent Height: `Ph`
  - Parent Margins: `Margin`
  - Button X Pos: `Bx = (Px + Margin)`
  - Button Y Pos: `By = (Py + Margin)`
  - Button Width: `Bw = (Pw * 0.50) - (Margin * 2.0)`
  - Button Height: `Bh = (Ph * 0.50) - (Margin * 2.0)`

This can be achieved with this package like so:
```golang
import para "github.com/gabe-lee/go_param_table"

func main() {
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
	var Bw = MyParamTable.InitDerived_F32(false, CALC_HALF_A_MINUS_2B, Pw,Margin)
	var Bh = MyParamTable.InitDerived_F32(false, CALC_HALF_A_MINUS_2B, Ph,Margin)

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

While this seems like a cumbersome way to achieve such a simple calculation, the benefits begin to show themselves when you then need to update the parent dimensions and immediately need the updated button dimensions as a result (for example to draw the next frame)

```golang
func main() {
  // ...previous results
  MyParamTable.SetRoot_F32(Pw, 990.0) // change parent width
	MyParamTable.SetRoot_F32(Py, 333.0) // change parent y position
	MyParamTable.SetRoot_F32(Margin, 48.0) // change parent margins
  printTable(&MyParamTable)
  //   X     Y     W     H
  //P: 100.0 333.0 990.0 600.0
  //B: 148.0 381.0 399.0 204.0
}
```

Thats it! Once the initialization process is complete, you no longer need to worry about triggering all the update functions manually or from within some other user-defined web of structs, methods, etc.

[Back to Top](#go_param_table)

## Pros/Cons/Caveats
#### Pros
  - Automatic updates of derived/calculated values when their parent values (root _or_ derived) change, no matter how deeply nested inside one another
  - Uses no interfaces or type reflection to achieve this
  - Safety checks enabled by default, but can be turned off using a global var in the library (`EnableSafetyChecks`) for more speed
  - Supports types: `bool, uint8, uint16, uint32, uint64, int8, int16, int32, int64, float32, float64, unsafe.Pointer`
  - With enough creativity, you can model nearly anything fully within this library

#### Cons
  - Every parameter incurrs AT LEAST 32 bytes of overhead
    - overhead = 32 + ((parentCount + siblingCount + childCount) * 2)
    - For example, [Example 1](#example-1-ui-positioning) went from 224 bytes uninitialized, to 890 bytes with 2 calcs, 5 root float32 vals, and 4 derived float32 vals
    - I hope to reduce this in the future, sorry :(
  - Limited to 65535 unique parameters
  - Limited to 65535 unique calculation functions
  - Calculation functions can only have a maximium of `len(inputs) + len(outputs) <= 255`
  - Arrays, Slices, and Struct types not directly supported (but can be used by either separating each struct fields into a parameter, or using `unsafe.Pointer` to load/store struct/array/slice pointers/lengths/capacities)
  - Cannot remove children once they are added (yet, this is planned in the future)

#### Caveats
  - Safety checks always cause panics, since most if not all errors
  covered by the safety checking would be due to programmer error when
  performing initialization or type mismatches on `Get_()`/`Set_()` functions. This prevents error handling bloat while providing all the error checking most projects would require, and additional error checking can be user defined inside the function bodies of the calculation functions themselves

[Back to Top](#go_param_table)
## Changelog
#### v0.11.0
  - MAJOR REWORK!
  - All parameter ID's are now dynamically generated!
    - Much better ergonomics
    - Cyclic update loops are now impossible (children can only be initialized from parent ID's that have already been generated, data can only flow 'downward')
    - Enables future (not yet implemented) features such as removing children, removing parents, deleting parameters, etc...
  - Updates are no longer recursive, but are appended to a flat queue instead
  - Sadly, this resulted in a considerable increase in memory footprint per parameter
#### v0.10.0
  - Reduction of memory footprint by around 15-25% (depending on usage)
#### v0.9.0
  - API refactor
#### v0.8.0
  - Initial commited version
#### Future Goals
  - [ ] Optional UI Layout system that uses `ParamTable`
  - [ ] Helper functions for common calculations/patterns
  - [ ] Reduce memory footprint per parameter
  - [ ] Maybe actually support slices?

#### Non-Goals
  - ~~Directly Support arrays/slices as base data types~~
    - ~~This can (and in my opinion _should_) be a concern for optional helper functions, some other external library, or user code. This library has no problem storing a slice pointer/length/capacity using the available data types.~~
