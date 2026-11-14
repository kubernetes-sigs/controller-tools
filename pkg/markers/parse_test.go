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
	"fmt"
	"math"
	"reflect"
	sc "text/scanner"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "sigs.k8s.io/controller-tools/pkg/markers"
)

func mustDefine(reg *Registry, name string, target TargetType, obj any) {
	Expect(reg.Define(name, target, obj)).To(Succeed())
}

type wrappedMarkerVal string
type multiFieldStruct struct {
	Str          string
	Int          int
	Bool         bool
	Any          any
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

type CustomType struct {
	Value any
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
			mustDefine(reg, "testing:map", DescribesPackage, map[string]string{})

			defn, err := MakeAnyTypeDefinition("testing:custom", DescribesPackage, CustomType{})
			Expect(err).NotTo(HaveOccurred())

			Expect(reg.Register(defn)).To(Succeed())
		})

		It("should work with fiddled field names", parseTestCase{reg: &reg, raw: "+testing:custom={hi}", output: CustomType{Value: []string{"hi"}}}.Run)

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

		Context("when parsing map markers", func() {
			It("should error out for a map missing a comma", func() {
				By("looking up the marker definition")
				raw := "+testing:map={cpu: 500m memory: 1Gi}"
				defn := reg.Lookup(raw, DescribesPackage)
				Expect(defn).NotTo(BeNil())

				By("trying to parse the marker")
				_, err := defn.Parse(raw)
				Expect(err).To(HaveOccurred())
			})
			It("should error out for a map key with a missing value before a comma", func() {
				By("looking up the marker definition")
				raw := "+testing:map={replicas: , enabled: true}"
				defn := reg.Lookup(raw, DescribesPackage)
				Expect(defn).NotTo(BeNil())

				By("trying to parse the marker")
				_, err := defn.Parse(raw)
				Expect(err).To(HaveOccurred())
			})
			It("should error out for a map key with a missing value before a comma and no spaces", func() {
				By("looking up the marker definition")
				raw := "+testing:map={replicas:,enabled:true}"
				defn := reg.Lookup(raw, DescribesPackage)
				Expect(defn).NotTo(BeNil())

				By("trying to parse the marker")
				_, err := defn.Parse(raw)
				Expect(err).To(HaveOccurred())
			})
			It("should error out for an unquoted map value with a colon and space", parseErrorTestCase{reg: &reg, raw: "+testing:map={msg: hello world: again}"}.Run)
			It("should keep a quoted map value with a colon and space", parseTestCase{reg: &reg, raw: `+testing:map={msg: "hello world: again"}`, output: map[string]string{"msg": "hello world: again"}}.Run)
		})

		Context("when a value outside a map has a colon and space", func() {
			one := 1
			It("should keep it in a string", parseTestCase{reg: &reg, raw: "+testing:anonymous:literal=Use v2: it is better", output: "Use v2: it is better"}.Run)
			It("should keep it in a string field", parseTestCase{reg: &reg, raw: "+testing:parent:optStr=bad value: fix it", output: allOptionalStruct{OptStr: "bad value: fix it"}}.Run)
			It("should keep it in a string field before another field", parseTestCase{reg: &reg, raw: "+testing:allOptional:optStr=bad value: fix it,optInt=1", output: allOptionalStruct{OptStr: "bad value: fix it", OptInt: &one}}.Run)
			It("should keep it in a value of any type", parseTestCase{reg: &reg, raw: "+testing:custom=key: value", output: CustomType{Value: "key: value"}}.Run)
		})

		Context("when a value cannot be read", func() {
			It("should error out for an integer above the largest int64", parseErrorTestCase{reg: &reg, raw: "+testing:custom=9223372036854775808"}.Run)
			It("should error out for an integer below the smallest int64", parseErrorTestCase{reg: &reg, raw: "+testing:custom=-9223372036854775809"}.Run)
			It("should error out for a float too large for a float64", parseErrorTestCase{reg: &reg, raw: "+testing:custom=1e1000"}.Run)
			It("should error out for a word given to an integer marker", parseErrorTestCase{reg: &reg, raw: "+testing:multi:segment=abc"}.Run)
			It("should error out for digits followed by letters given to an integer marker", parseErrorTestCase{reg: &reg, raw: "+testing:multi:segment=12abc"}.Run)
			It("should error out for a map that is never closed", parseErrorTestCase{reg: &reg, raw: "+testing:custom={replicas: 3"}.Run)
			It("should error out for a map key with a missing value", parseErrorTestCase{reg: &reg, raw: "+testing:custom={replicas: }"}.Run)
		})

		Context("when dealing with markers describing multiple things", func() {
			It("should properly parse the package-level one", parseTestCase{reg: &reg, raw: "+testing:tripleDefined=42", output: 42, target: DescribesPackage}.Run)
			It("should properly parse the field-level one", parseTestCase{reg: &reg, raw: "+testing:tripleDefined=foo", output: "foo", target: DescribesField}.Run)
			It("should properly parse the type-level one", parseTestCase{reg: &reg, raw: "+testing:tripleDefined=true", output: true, target: DescribesType}.Run)
		})
	})

	Context("of individual arguments", func() {
		It("should support bare strings", argParseTestCase{arg: Argument{Type: StringType}, raw: `some string here!`, output: "some string here!"}.Run)
		It("should support bare strings containing number-ish values 1", argParseTestCase{arg: Argument{Type: StringType}, raw: `aa 0Baaa aaa`, output: "aa 0Baaa aaa"}.Run)
		It("should support bare strings containing number-ish values 2", argParseTestCase{arg: Argument{Type: StringType}, raw: `/tmp/tmp-CHECKCRD-0B7LDeoZta`, output: "/tmp/tmp-CHECKCRD-0B7LDeoZta"}.Run)
		It("should support bare strings containing number-ish values 3", argParseTestCase{arg: Argument{Type: StringType}, raw: `.0B7LDeoZt`, output: ".0B7LDeoZt"}.Run)
		It("should support double-quoted strings", argParseTestCase{arg: Argument{Type: StringType}, raw: `"some; string, \nhere"`, output: "some; string, \nhere"}.Run)
		It("should support raw strings", argParseTestCase{arg: Argument{Type: StringType}, raw: "`some; string, \\nhere`", output: `some; string, \nhere`}.Run)
		It("should support integers", argParseTestCase{arg: Argument{Type: IntType}, raw: "42", output: 42}.Run)
		It("should support negative integers", argParseTestCase{arg: Argument{Type: IntType}, raw: "-42", output: -42}.Run)
		It("should support numbers with grouped digits", argParseTestCase{arg: Argument{Type: NumberType}, raw: "1_000", output: 1000.0}.Run)
		It("should support negative numbers with a space after the sign", argParseTestCase{arg: Argument{Type: NumberType}, raw: "- 7", output: -7.0}.Run)
		It("should support false booleans", argParseTestCase{arg: Argument{Type: BoolType}, raw: "false", output: false}.Run)
		It("should support true booleans", argParseTestCase{arg: Argument{Type: BoolType}, raw: "true", output: true}.Run)

		sliceOSlice := Argument{Type: SliceType, ItemType: &Argument{Type: SliceType, ItemType: &Argument{Type: IntType}}}
		sliceOSliceOut := [][]int{{1, 1}, {2, 3}, {5, 8}}

		It("should support bare slices", argParseTestCase{arg: Argument{Type: SliceType, ItemType: &Argument{Type: StringType}}, raw: "hi;hello;hey y'all", output: []string{"hi", "hello", "hey y'all"}}.Run)
		It("should support delimitted slices", argParseTestCase{arg: Argument{Type: SliceType, ItemType: &Argument{Type: StringType}}, raw: "{hi,hello,hey y'all}", output: []string{"hi", "hello", "hey y'all"}}.Run)
		It("should support delimitted slices of bare slices", argParseTestCase{arg: sliceOSlice, raw: "{1;1,2;3,5;8}", output: sliceOSliceOut}.Run)
		It("should support delimitted slices of delimitted slices", argParseTestCase{arg: sliceOSlice, raw: "{{1,1},{2,3},{5,8}}", output: sliceOSliceOut}.Run)

		It("should support maps", argParseTestCase{arg: Argument{Type: MapType, ItemType: &Argument{Type: StringType}}, raw: "{formal: hello, `informal`: `hi!`}", output: map[string]string{"formal": "hello", "informal": "hi!"}}.Run)
		It("should work with empty maps (which are equal to empty lists in the output)", argParseTestCase{arg: Argument{Type: MapType, ItemType: &Argument{Type: StringType}}, raw: "{}", output: map[string]string{}}.Run)

		Context("with any value", func() {
			anyArg := Argument{Type: AnyType}
			It("should support bare strings", argParseTestCase{arg: anyArg, raw: `some string here!`, output: "some string here!"}.Run)
			It("should keep a bare string whole when a space comes before the next item", argParseTestCase{arg: anyArg, raw: `{small , large}`, output: []string{"small", "large"}}.Run)
			It("should keep a number-ish string whole when a space comes before the next item", argParseTestCase{arg: anyArg, raw: `{100Mi , 200Mi}`, output: []string{"100Mi", "200Mi"}}.Run)
			It("should support double-quoted strings", argParseTestCase{arg: anyArg, raw: `"some; string, \nhere"`, output: "some; string, \nhere"}.Run)
			It("should support raw strings", argParseTestCase{arg: anyArg, raw: "`some; string, \\nhere`", output: `some; string, \nhere`}.Run)
			It("should support integers", argParseTestCase{arg: anyArg, raw: "42", output: 42}.Run)
			It("should support negative integers", argParseTestCase{arg: anyArg, raw: "-42", output: -42}.Run)
			It("should support floats", argParseTestCase{arg: anyArg, raw: "1.5", output: 1.5}.Run)
			It("should support negative floats", argParseTestCase{arg: anyArg, raw: "-1.5", output: -1.5}.Run)
			It("should support false booleans", argParseTestCase{arg: anyArg, raw: "false", output: false}.Run)
			It("should support true booleans", argParseTestCase{arg: anyArg, raw: "true", output: true}.Run)

			sliceOSliceOut := [][]int{{1, 1}, {2, 3}, {5, 8}}

			It("should support bare slices", argParseTestCase{arg: anyArg, raw: "hi;hello;hey y'all", output: []string{"hi", "hello", "hey y'all"}}.Run)
			It("should support delimitted slices", argParseTestCase{arg: anyArg, raw: "{hi,hello,hey y'all}", output: []string{"hi", "hello", "hey y'all"}}.Run)
			It("should support delimitted slices of bare slices", argParseTestCase{arg: anyArg, raw: "{1;1,2;3,5;8}", output: sliceOSliceOut}.Run)
			It("should support delimitted slices of delimitted slices", argParseTestCase{arg: anyArg, raw: "{{1,1},{2,3},{5,8}}", output: sliceOSliceOut}.Run)

			complexMap := map[string]any{
				"text":     "abc",
				"len":      3,
				"as bytes": []int{97, 98, 99},
				"props": map[string]any{
					"encoding": "ascii",
					"nullsafe": true,
					"tags":     []string{"triple", "in a row"},
				},
			}
			It("should support non-uniform, nested maps (and always guess as such)", argParseTestCase{arg: anyArg, raw: `{text: "abc", len: 3, "as bytes": 97;98;99, props: {encoding: ascii, nullsafe: true, tags: {"triple", "in a row"}}}`, output: complexMap}.Run)

			It("should parse number-ish strings with invalid octal prefix as string", argParseTestCase{arg: anyArg, raw: `087bdd`, output: "087bdd"}.Run)
			It("should parse path-like strings as string", argParseTestCase{arg: anyArg, raw: `/tmp/tmp.4paHkdEcpK`, output: "/tmp/tmp.4paHkdEcpK"}.Run)
			It("should parse numeric path as string", argParseTestCase{arg: anyArg, raw: `123/config`, output: "123/config"}.Run)
			It("should parse invalid octal number as string", argParseTestCase{arg: anyArg, raw: `0987`, output: "0987"}.Run)
			It("should parse resource quantity as string", argParseTestCase{arg: anyArg, raw: `100Mi`, output: "100Mi"}.Run)
			It("should parse date-like string as string", argParseTestCase{arg: anyArg, raw: `2024-01-15`, output: "2024-01-15"}.Run)
			It("should parse SHA prefix as string", argParseTestCase{arg: anyArg, raw: `sha256:08a7b`, output: "sha256:08a7b"}.Run)
			It("should parse UUID-like string as string", argParseTestCase{arg: anyArg, raw: `00000000-1234-5678`, output: "00000000-1234-5678"}.Run)
			It("should parse alphanumeric starting with number as string", argParseTestCase{arg: anyArg, raw: `123abc`, output: "123abc"}.Run)
			It("should parse bool-prefixed hyphenated string as string", argParseTestCase{arg: anyArg, raw: `true-positive`, output: "true-positive"}.Run)
			It("should parse bool-prefixed dotted string as string", argParseTestCase{arg: anyArg, raw: `false.example.com`, output: "false.example.com"}.Run)
			It("should parse bool-prefixed path as string", argParseTestCase{arg: anyArg, raw: `true/false`, output: "true/false"}.Run)
			It("should parse slices of bool-prefixed strings", argParseTestCase{arg: anyArg, raw: `true-positive;false-positive`, output: []string{"true-positive", "false-positive"}}.Run)
			It("should still parse booleans in slices", argParseTestCase{arg: anyArg, raw: `{true, false}`, output: []bool{true, false}}.Run)
			It("should parse dotted digits with a slash as string", argParseTestCase{arg: anyArg, raw: `10.96.0.0/12`, output: "10.96.0.0/12"}.Run)
			It("should parse a hex prefix with no digits as string", argParseTestCase{arg: anyArg, raw: `0x`, output: "0x"}.Run)
			It("should parse an exponent with no digits as string", argParseTestCase{arg: anyArg, raw: `1e`, output: "1e"}.Run)
			It("should parse a trailing digit separator as string", argParseTestCase{arg: anyArg, raw: `1_`, output: "1_"}.Run)
			It("should parse a negative float followed by a letter as string", argParseTestCase{arg: anyArg, raw: `-1e5x`, output: "-1e5x"}.Run)
			It("should parse a negative integer followed by letters as string", argParseTestCase{arg: anyArg, raw: `-42nd`, output: "-42nd"}.Run)
			It("should parse a bare slice of number-ish strings", argParseTestCase{arg: anyArg, raw: `2024-01-15;2024-06-01`, output: []string{"2024-01-15", "2024-06-01"}}.Run)
			It("should parse a delimitted slice of number-ish strings", argParseTestCase{arg: anyArg, raw: `{100Mi,200Mi,500Mi}`, output: []string{"100Mi", "200Mi", "500Mi"}}.Run)
			It("should parse a map of number-ish strings", argParseTestCase{arg: anyArg, raw: `{memory: 100Mi, cpu: 500m}`, output: map[string]any{"memory": "100Mi", "cpu": "500m"}}.Run)
			It("should parse a nested map of number-ish strings", argParseTestCase{arg: anyArg, raw: `{host: 10.0.0.1, labels: {commit: 087bdd}}`, output: map[string]any{"host": "10.0.0.1", "labels": map[string]any{"commit": "087bdd"}}}.Run)
			It("should keep a slice item with a colon and space", argParseTestCase{arg: anyArg, raw: `a: b;c`, output: []string{"a: b", "c"}}.Run)
			It("should keep a colon inside a map value", argParseTestCase{arg: anyArg, raw: `{digest: sha256:08a7b, image: registry.k8s.io/pause:3.9, time: 12:30}`, output: map[string]any{"digest": "sha256:08a7b", "image": "registry.k8s.io/pause:3.9", "time": "12:30"}}.Run)

			It("should keep the largest int64 exact", argParseTestCase{arg: anyArg, raw: `9223372036854775807`, output: math.MaxInt64}.Run)
			It("should keep the smallest int64 exact", argParseTestCase{arg: anyArg, raw: `-9223372036854775808`, output: math.MinInt64}.Run)
			It("should keep a slice of integers above 2^53 exact", argParseTestCase{arg: anyArg, raw: `{-9223372036854775808,9007199254740993}`, output: []int{math.MinInt64, 9007199254740993}}.Run)
			It("should support floats with grouped digits", argParseTestCase{arg: anyArg, raw: `1_000.5`, output: 1000.5}.Run)
			It("should support floats with no leading digit", argParseTestCase{arg: anyArg, raw: `.5`, output: 0.5}.Run)
			It("should support floats with an exponent", argParseTestCase{arg: anyArg, raw: `1e3`, output: 1000.0}.Run)
			It("should support floats with a fraction and an exponent", argParseTestCase{arg: anyArg, raw: `1.5e3`, output: 1500.0}.Run)
			It("should support hexadecimal floats", argParseTestCase{arg: anyArg, raw: `0x1p2`, output: 4.0}.Run)
			It("should support negative integers with a space after the sign", argParseTestCase{arg: anyArg, raw: `- 7`, output: -7}.Run)
			It("should support negative floats with a space after the sign", argParseTestCase{arg: anyArg, raw: `- 1.5`, output: -1.5}.Run)
			It("should keep integers when a space comes before the next item", argParseTestCase{arg: anyArg, raw: `{42 , 43}`, output: []int{42, 43}}.Run)
			It("should keep floats when a space comes before the next item", argParseTestCase{arg: anyArg, raw: `{1.5 , 2.5}`, output: []float64{1.5, 2.5}}.Run)
			It("should keep booleans when a space comes before the next item", argParseTestCase{arg: anyArg, raw: `{true , false}`, output: []bool{true, false}}.Run)
			It("should keep map value types when a space comes before the next item", argParseTestCase{arg: anyArg, raw: `{replicas: 42 , enabled: true , ratio: 1.5 }`, output: map[string]any{"replicas": 42, "enabled": true, "ratio": 1.5}}.Run)
			It("should parse a value starting with a plus sign as string", argParseTestCase{arg: anyArg, raw: `+0530`, output: "+0530"}.Run)
			It("should parse a slice of values starting with a plus sign as strings", argParseTestCase{arg: anyArg, raw: `+1;+44;+81`, output: []string{"+1", "+44", "+81"}}.Run)
			It("should parse a short flag as string", argParseTestCase{arg: anyArg, raw: `-v`, output: "-v"}.Run)
			It("should parse a long flag as string", argParseTestCase{arg: anyArg, raw: `--leader-elect`, output: "--leader-elect"}.Run)
			It("should parse a boolean word after a dash as string", argParseTestCase{arg: anyArg, raw: `-true`, output: "-true"}.Run)
		})
	})
})

type parseTestCase struct {
	reg    **Registry
	raw    string
	output any
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

type parseErrorTestCase struct {
	reg    **Registry
	raw    string
	target TargetType
}

func (tc parseErrorTestCase) Run() {
	reg := *tc.reg
	By("looking up the marker definition")
	defn := reg.Lookup(tc.raw, tc.target)
	Expect(defn).NotTo(BeNil())

	By("trying to parse the marker")
	_, err := defn.Parse(tc.raw)
	Expect(err).To(HaveOccurred())
}

type argParseTestCase struct {
	arg    Argument
	raw    string
	output any
}

func (tc argParseTestCase) Run() {
	scanner := sc.Scanner{}
	scanner.Init(bytes.NewBufferString(tc.raw))
	scanner.Mode = sc.ScanIdents | sc.ScanInts | sc.ScanFloats | sc.ScanStrings | sc.ScanRawStrings | sc.SkipComments
	scanner.Error = func(scanner *sc.Scanner, msg string) {
		Fail(fmt.Sprintf("%s (at %s)", msg, scanner.Position))
	}

	var actualOut reflect.Value
	if tc.arg.Type == AnyType {
		actualOut = reflect.Indirect(reflect.New(reflect.TypeFor[*any]().Elem()))
	} else {
		actualOut = reflect.Indirect(reflect.New(reflect.TypeOf(tc.output)))
	}

	By("parsing the raw argument")
	tc.arg.Parse(&scanner, tc.raw, actualOut)

	By("checking that it equals the expected output")
	Expect(actualOut.Interface()).To(Equal(tc.output))
}
