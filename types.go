package go_param_table

// type bufidx = uint32
type validx = uint16

// type hookup_data_block = uint16

type (
	ParamCalc func(c *CalcInterface)
	ParamID   uint16
	CalcID    uint16
	// PIdx_U64  uint16
	// PIdx_I64  uint16
	// PIdx_F64  uint16
	// PIdx_Ptr  uint16
	// PIdx_U32  uint16
	// PIdx_I32  uint16
	// PIdx_F32  uint16
	// PIdx_U16  uint16
	// PIdx_I16  uint16
	// PIdx_U8   uint16
	// PIdx_I8   uint16
	// PIdx_Bool uint16
	// PIdx_Calc uint16
)

// const (
// 	IDX_NULL uint16 = 65535
// )

type ParamType uint8

const (
	TypeNone ParamType = iota
	TypeU64
	TypeI64
	TypeF64
	TypePtr
	TypeU32
	TypeI32
	TypeF32
	TypeU16
	TypeI16
	TypeU8
	TypeI8
	TypeBool
	typeCount
)

var typeNames = [typeCount]string{
	TypeNone: "<None>",
	TypeU64:  "uint64",
	TypeI64:  "int64",
	TypeF64:  "float64",
	TypePtr:  "unsafe.Pointer",
	TypeU32:  "uint32",
	TypeI32:  "int32",
	TypeF32:  "float32",
	TypeU16:  "uint16",
	TypeI16:  "int16",
	TypeU8:   "uint8",
	TypeI8:   "int8",
	TypeBool: "bool",
}

// const (
// 	size64   uint32 = 8
// 	sizePtr  uint32 = uint32(unsafe.Sizeof(unsafe.Pointer(nil)))
// 	sizeUint uint32 = uint32(unsafe.Sizeof(uint(0)))
// 	size32   uint32 = 4
// 	size16   uint32 = 2
// 	size8    uint32 = 1
// )

// var sizeTable = [typeCount]uint32{
// 	TypeU64:  size64,
// 	TypeI64:  size64,
// 	TypeF64:  size64,
// 	TypePtr:  sizePtr,
// 	TypeU32:  size32,
// 	TypeI32:  size32,
// 	TypeF32:  size32,
// 	TypeU16:  size16,
// 	TypeI16:  size16,
// 	TypeU8:   size8,
// 	TypeI8:   size8,
// 	TypeBool: size8,
// }
