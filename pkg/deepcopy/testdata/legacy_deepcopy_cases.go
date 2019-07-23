/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cronjob

// NB(directxman12): this file contains tests copied over from the legacy deepcopy-gen
// They should all continue to pass -- non-covered cases have already been pruned.

// +kubebuilder:object:root=true
type Foo struct {
	X int
}

type Builtin int
type Slice []int
type Pointer *int
type PointerAlias *Builtin
type Struct Foo
type Map map[string]int

type FooAlias Foo
type FooSlice []Foo
type FooPointer *Foo
type FooMap map[string]Foo

type AliasBuiltin Builtin
type AliasSlice Slice
type AliasPointer Pointer
type AliasStruct Struct
type AliasMap Map

// Aliases
type Ttest struct {
	Builtin      Builtin
	Slice        Slice
	Pointer      Pointer
	PointerAlias PointerAlias
	Struct       Struct
	Map          Map
	SliceSlice   []Slice
	MapSlice     map[string]Slice

	FooAlias   FooAlias
	FooSlice   FooSlice
	FooPointer FooPointer
	FooMap     FooMap

	AliasBuiltin AliasBuiltin
	AliasSlice   AliasSlice
	AliasPointer AliasPointer
	AliasStruct  AliasStruct
	AliasMap     AliasMap
}

type TestBuiltins struct {
	Byte byte
	Int8    int8
	Int16   int16
	Int32   int32
	Int64   int64
	Uint8   uint8
	Uint16  uint16
	Uint32  uint32
	Uint64  uint64
	Float32 float32
	Float64 float64
	String  string
}

type TestMaps struct {
	Byte map[string]byte
	Int8    map[string]int8
	Int16        map[string]int16
	Int32        map[string]int32
	Int64        map[string]int64
	Uint8        map[string]uint8
	Uint16       map[string]uint16
	Uint32       map[string]uint32
	Uint64       map[string]uint64
	Float32      map[string]float32
	Float64      map[string]float64
	String       map[string]string
	StringPtr    map[string]*string
	StringPtrPtr map[string]**string
	Map          map[string]map[string]string
	MapPtr       map[string]*map[string]string
	Slice        map[string][]string
	SlicePtr     map[string]*[]string
	Struct       map[string]Ttest
	StructPtr    map[string]*Ttest
}


type TestPointers struct {
	Builtin   *string
	Ptr       **string
	Map       *map[string]string
	Slice     *[]string
	MapPtr    **map[string]string
	SlicePtr  **[]string
	Struct    *Ttest
	StructPtr **Ttest
}

type TestSlices struct {
	Byte []byte
	Int8    []int8 //TODO: int8 becomes byte in SnippetWriter
	Int16        []int16
	Int32        []int32
	Int64        []int64
	Uint8        []uint8
	Uint16       []uint16
	Uint32       []uint32
	Uint64       []uint64
	Float32      []float32
	Float64      []float64
	String       []string
	StringPtr    []*string
	StringPtrPtr []**string
	Map          []map[string]string
	MapPtr       []*map[string]string
	Slice        [][]string
	SlicePtr     []*[]string
	Struct       []Ttest
	StructPtr    []*Ttest
}

type Inner struct {
	Byte byte
	Int8    int8 //TODO: int8 becomes byte in SnippetWriter
	Int16   int16
	Int32   int32
	Int64   int64
	Uint8   uint8
	Uint16  uint16
	Uint32  uint32
	Uint64  uint64
	Float32 float32
	Float64 float64
	String  string
}

type TestInner struct {
	Inner1 Inner
	Inner2 Inner
}

// Trivial
type Struct_Empty struct{}

// Only primitives
type Struct_Primitives struct {
	BoolField   bool
	IntField    int
	StringField string
	FloatField  float64
}
type Struct_Primitives_Alias Struct_Primitives
type Struct_Embed_Struct_Primitives struct {
	Struct_Primitives
}
type Struct_Embed_Int struct {
	int
}
type Struct_Struct_Primitives struct {
	StructField Struct_Primitives
}

// Manual DeepCopy method
type ManualStruct struct {
	StringField string
}

func (m ManualStruct) DeepCopy() ManualStruct {
	return m
}

type ManualStruct_Alias ManualStruct

type Struct_Embed_ManualStruct struct {
	ManualStruct
}

// Only pointers to primitives
type Struct_PrimitivePointers struct {
	BoolPtrField   *bool
	IntPtrField    *int
	StringPtrField *string
	FloatPtrField  *float64
}
type Struct_PrimitivePointers_Alias Struct_PrimitivePointers
type Struct_Embed_Struct_PrimitivePointers struct {
	Struct_PrimitivePointers
}
type Struct_Embed_Pointer struct {
	*int
}
type Struct_Struct_PrimitivePointers struct {
	StructField Struct_PrimitivePointers
}

// Manual DeepCopy method
type ManualSlice []string

func (m ManualSlice) DeepCopy() ManualSlice {
	r := make(ManualSlice, len(m))
	copy(r, m)
	return r
}

// Slices
type Struct_Slices struct {
	SliceBoolField                         []bool
	SliceByteField                         []byte
	SliceIntField                          []int
	SliceStringField                       []string
	SliceFloatField                        []float64
	SliceStructPrimitivesField             []Struct_Primitives
	SliceStructPrimitivesAliasField        []Struct_Primitives_Alias
	SliceStructPrimitivePointersField      []Struct_PrimitivePointers
	SliceStructPrimitivePointersAliasField []Struct_PrimitivePointers_Alias
	SliceSliceIntField                     [][]int
	SliceManualStructField                 []ManualStruct
	ManualSliceField                       ManualSlice
}
type Struct_Slices_Alias Struct_Slices
type Struct_Embed_Struct_Slices struct {
	Struct_Slices
}
type Struct_Struct_Slices struct {
	StructField Struct_Slices
}

// Everything
type Struct_Everything struct {
	BoolField                 bool
	IntField                  int
	StringField               string
	FloatField                float64
	StructField               Struct_Primitives
	EmptyStructField          Struct_Empty
	ManualStructField         ManualStruct
	ManualStructAliasField    ManualStruct_Alias
	BoolPtrField              *bool
	IntPtrField               *int
	StringPtrField            *string
	FloatPtrField             *float64
	PrimitivePointersField    Struct_PrimitivePointers
	ManualStructPtrField      *ManualStruct
	ManualStructAliasPtrField *ManualStruct_Alias
	SliceBoolField            []bool
	SliceByteField            []byte
	SliceIntField             []int
	SliceStringField          []string
	SliceFloatField           []float64
	SlicesField               Struct_Slices
	SliceManualStructField    []ManualStruct
	ManualSliceField          ManualSlice
}

// An Object
// +kubebuilder:object:root=true
type Struct_ExplicitObject struct {
	x int
}

// +k8s:deepcopy-gen=false
type Struct_TypeMeta struct {
}

// +kubebuilder:object:root=true
type Struct_ExplicitSelectorExplicitObject struct {
	Struct_TypeMeta
}

// Another type in another file.
type Struct_B struct{}
