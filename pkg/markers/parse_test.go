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

package markers_test

import (
	"bytes"
	"reflect"
	sc "text/scanner"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "sigs.k8s.io/controller-tools/pkg/markers"
)

func mustDefine(reg *Registry, name string, target TargetType, obj interface{}) {
	Expect(reg.Define(name, target, obj)).To(Succeed())
}

type wrappedMarkerVal string
type multiFieldStruct struct {
	Str          string
	Int          int
	Bool         bool
	Any          interface{}
	PtrOpt       *string
	NormalOpt    string `marker:",optional"`
	DiffNamed    string `marker:"other"`
	BothTags     string `marker:"both,optional"`
	Slice        []int
	SliceOfSlice [][]int
}

type allOptionalStruct struct {
	OptStr string `marker:",optional"`
	OptInt *int
}

var _ = Describe("Parsing", func() {
	var reg *Registry

	Context("of full markers", func() {
		BeforeEach(func() {
			reg = &Registry{}

			mustDefine(reg, "testing:empty", DescribesPackage, struct{}{})
			mustDefine(reg, "testing:anonymous:literal", DescribesPackage, "")
			mustDefine(reg, "testing:anonymous:named", DescribesPackage, wrappedMarkerVal(""))
			mustDefine(reg, "testing:raw", DescribesPackage, RawArguments(""))
			mustDefine(reg, "testing:multiField", DescribesPackage, multiFieldStruct{})
			mustDefine(reg, "testing:allOptional", DescribesPackage, allOptionalStruct{})
			mustDefine(reg, "testing:anonymousOptional", DescribesPackage, (*int)(nil))
			mustDefine(reg, "testing:multi:segment", DescribesPackage, 0)
			mustDefine(reg, "testing:parent", DescribesPackage, allOptionalStruct{})
			mustDefine(reg, "testing:parent:nested", DescribesPackage, "")
			mustDefine(reg, "testing:tripleDefined", DescribesPackage, 0)
			mustDefine(reg, "testing:tripleDefined", DescribesField, "")
			mustDefine(reg, "testing:tripleDefined", DescribesType, false)
		})

		It("should parse name-only markers", parseTestCase{reg: &reg, raw: "+testing:empty", output: struct{}{}}.Run)

		Context("when parsing anonymous markers", func() {
			It("should parse into literal-typed values", parseTestCase{reg: &reg, raw: "+testing:anonymous:literal=foo", output: "foo"}.Run)
			It("should parse into named-typed values", parseTestCase{reg: &reg, raw: "+testing:anonymous:named=foo", output: wrappedMarkerVal("foo")}.Run)
			It("shouldn't require any argument to an optional-valued marker", parseTestCase{reg: &reg, raw: "+testing:anonymousOptional", output: (*int)(nil)}.Run)
		})

		It("should parse raw-argument markers", parseTestCase{reg: &reg, raw: "+testing:raw=this;totally,doesn't;get,parsed", output: RawArguments("this;totally,doesn't;get,parsed")}.Run)

		Context("when parsing multi-field markers", func() {
			definitelyAString := "optional string"
			It("should support parsing multiple fields with the appropriate names", parseTestCase{
				reg: &reg,
				raw: `+testing:multiField:str=some str,int=42,bool=true,any=21,ptrOpt="optional string",normalOpt="other string",other="yet another",both="and one more",slice=99;104;101;101;115;101,sliceOfSlice={{1,1},{2,3},{5,8}}`,
				output: multiFieldStruct{
					Str:          "some str",
					Bool:         true,
					Any:          21,
					PtrOpt:       &definitelyAString,
					NormalOpt:    "other string",
					DiffNamed:    "yet another",
					BothTags:     "and one more",
					Slice:        []int{99, 104, 101, 101, 115, 101},
					SliceOfSlice: [][]int{{1, 1}, {2, 3}, {5, 8}},
					Int:          42,
				},
			}.Run)

			It("shouldn't matter what order the fields are specified in", parseTestCase{
				reg: &reg,
				// just some random order with everything out of order
				raw: `+testing:multiField:int=42,str=some str,any=21,bool=true,any=21,sliceOfSlice={{1,1},{2,3},{5,8}},slice=99;104;101;101;115;101,other="yet another"`,
				output: multiFieldStruct{
					Str:          "some str",
					Bool:         true,
					Any:          21,
					DiffNamed:    "yet another",
					Slice:        []int{99, 104, 101, 101, 115, 101},
					SliceOfSlice: [][]int{{1, 1}, {2, 3}, {5, 8}},
					Int:          42,
				},
			}.Run)

			It("should support leaving out optional fields", parseTestCase{
				reg: &reg,
				raw: `+testing:multiField:str=some str,bool=true,any=21,other="yet another",slice=99;104;101;101;115;101,sliceOfSlice={{1,1},{2,3},{5,8}},int=42`,
				output: multiFieldStruct{
					Str:          "some str",
					Bool:         true,
					Any:          21,
					DiffNamed:    "yet another",
					Slice:        []int{99, 104, 101, 101, 115, 101},
					SliceOfSlice: [][]int{{1, 1}, {2, 3}, {5, 8}},
					Int:          42,
				},
			}.Run)
			It("should error out for missing values", func() {
				By("looking up the marker definition")
				raw := "+testing:multiField:str=`hi`"
				defn := reg.Lookup(raw, DescribesPackage)
				Expect(defn).NotTo(BeNil())

				By("trying to parse the marker")
				_, err := defn.Parse(raw)
				Expect(err).To(HaveOccurred())
			})
			It("shouldn't require any arguments to an optional-valued marker", parseTestCase{reg: &reg, raw: "+testing:allOptional", output: allOptionalStruct{}}.Run)
		})

		It("should support markers with multiple segments in the name", parseTestCase{reg: &reg, raw: "+testing:multi:segment=42", output: 42}.Run)

		Context("when dealing with disambiguating anonymous markers", func() {
			It("should favor the shorter-named one", parseTestCase{reg: &reg, raw: "+testing:parent", output: allOptionalStruct{}}.Run)
			It("should still allow fetching the longer-named one", parseTestCase{reg: &reg, raw: "+testing:parent:nested=some string", output: "some string"}.Run)
			It("should consider anonymously-named ones before considering fields", parseTestCase{reg: &reg, raw: "+testing:parent:optStr=other string", output: allOptionalStruct{OptStr: "other string"}}.Run)
		})

		Context("when dealing with markers describing multiple things", func() {
			It("should properly parse the package-level one", parseTestCase{reg: &reg, raw: "+testing:tripleDefined=42", output: 42, target: DescribesPackage}.Run)
			It("should properly parse the field-level one", parseTestCase{reg: &reg, raw: "+testing:tripleDefined=foo", output: "foo", target: DescribesField}.Run)
			It("should properly parse the type-level one", parseTestCase{reg: &reg, raw: "+testing:tripleDefined=true", output: true, target: DescribesType}.Run)
		})
	})

	Context("of individual arguments", func() {
		It("should support bare strings", argParseTestCase{arg: Argument{Type: StringType}, raw: `some string here!`, output: "some string here!"}.Run)
		It("should support double-quoted strings", argParseTestCase{arg: Argument{Type: StringType}, raw: `"some; string, \nhere"`, output: "some; string, \nhere"}.Run)
		It("should support raw strings", argParseTestCase{arg: Argument{Type: StringType}, raw: "`some; string, \\nhere`", output: `some; string, \nhere`}.Run)
		It("should support integers", argParseTestCase{arg: Argument{Type: IntType}, raw: "42", output: 42}.Run)
		XIt("should support negative integers", argParseTestCase{arg: Argument{Type: IntType}, raw: "-42", output: -42}.Run)
		It("should support false booleans", argParseTestCase{arg: Argument{Type: BoolType}, raw: "false", output: false}.Run)
		It("should support true booleans", argParseTestCase{arg: Argument{Type: BoolType}, raw: "true", output: true}.Run)

		sliceOSlice := Argument{Type: SliceType, ItemType: &Argument{Type: SliceType, ItemType: &Argument{Type: IntType}}}
		sliceOSliceOut := [][]int{{1, 1}, {2, 3}, {5, 8}}

		It("should support bare slices", argParseTestCase{arg: Argument{Type: SliceType, ItemType: &Argument{Type: StringType}}, raw: "hi;hello;hey y'all", output: []string{"hi", "hello", "hey y'all"}}.Run)
		It("should support delimitted slices", argParseTestCase{arg: Argument{Type: SliceType, ItemType: &Argument{Type: StringType}}, raw: "{hi,hello,hey y'all}", output: []string{"hi", "hello", "hey y'all"}}.Run)
		It("should support delimitted slices of bare slices", argParseTestCase{arg: sliceOSlice, raw: "{1;1,2;3,5;8}", output: sliceOSliceOut}.Run)
		It("should support delimitted slices of delimitted slices", argParseTestCase{arg: sliceOSlice, raw: "{{1,1},{2,3},{5,8}}", output: sliceOSliceOut}.Run)

		Context("with any value", func() {
			anyArg := Argument{Type: AnyType}
			It("should support bare strings", argParseTestCase{arg: anyArg, raw: `some string here!`, output: "some string here!"}.Run)
			It("should support double-quoted strings", argParseTestCase{arg: anyArg, raw: `"some; string, \nhere"`, output: "some; string, \nhere"}.Run)
			It("should support raw strings", argParseTestCase{arg: anyArg, raw: "`some; string, \\nhere`", output: `some; string, \nhere`}.Run)
			It("should support integers", argParseTestCase{arg: anyArg, raw: "42", output: 42}.Run)
			XIt("should support negative integers", argParseTestCase{arg: anyArg, raw: "-42", output: -42}.Run)
			It("should support false booleans", argParseTestCase{arg: anyArg, raw: "false", output: false}.Run)
			It("should support true booleans", argParseTestCase{arg: anyArg, raw: "true", output: true}.Run)

			sliceOSliceOut := [][]int{{1, 1}, {2, 3}, {5, 8}}

			It("should support bare slices", argParseTestCase{arg: anyArg, raw: "hi;hello;hey y'all", output: []string{"hi", "hello", "hey y'all"}}.Run)
			It("should support delimitted slices", argParseTestCase{arg: anyArg, raw: "{hi,hello,hey y'all}", output: []string{"hi", "hello", "hey y'all"}}.Run)
			It("should support delimitted slices of bare slices", argParseTestCase{arg: anyArg, raw: "{1;1,2;3,5;8}", output: sliceOSliceOut}.Run)
			It("should support delimitted slices of delimitted slices", argParseTestCase{arg: anyArg, raw: "{{1,1},{2,3},{5,8}}", output: sliceOSliceOut}.Run)
		})
	})
})

type parseTestCase struct {
	reg    **Registry
	raw    string
	output interface{}
	target TargetType // NB(directxman12): iota is DescribesPackage
}

func (tc parseTestCase) Run() {
	reg := *tc.reg
	By("looking up the marker definition")
	defn := reg.Lookup(tc.raw, tc.target)
	Expect(defn).NotTo(BeNil())

	By("parsing the marker")
	outVal, err := defn.Parse(tc.raw)
	Expect(err).NotTo(HaveOccurred())

	By("checking for the expected output")
	Expect(outVal).To(Equal(tc.output))
}

type argParseTestCase struct {
	arg    Argument
	raw    string
	output interface{}
}

func (tc argParseTestCase) Run() {
	scanner := sc.Scanner{}
	scanner.Init(bytes.NewBufferString(tc.raw))
	scanner.Mode = sc.ScanIdents | sc.ScanInts | sc.ScanStrings | sc.ScanRawStrings | sc.SkipComments
	scanner.Error = func(_ *sc.Scanner, msg string) {
		Fail(msg)
	}

	var actualOut reflect.Value
	if tc.arg.Type == AnyType {
		actualOut = reflect.Indirect(reflect.New(reflect.TypeOf((*interface{})(nil)).Elem()))
	} else {
		actualOut = reflect.Indirect(reflect.New(reflect.TypeOf(tc.output)))
	}

	By("parsing the raw argument")
	tc.arg.Parse(&scanner, tc.raw, actualOut)

	By("checking that it equals the expected output")
	Expect(actualOut.Interface()).To(Equal(tc.output))
}
