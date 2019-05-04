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

package markers

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	sc "text/scanner"
	"unicode"

	"sigs.k8s.io/controller-tools/pkg/loader"
)

// expect checks that the next token of the scanner is the given token, adding an error
// to the scanner if not.  It returns whether the token was as expected.
func expect(scanner *sc.Scanner, expected rune, errDesc string) bool {
	tok := scanner.Scan()
	if tok != expected {
		scanner.Error(scanner, fmt.Sprintf("expected %s, got %q", errDesc, scanner.TokenText()))
		return false
	}
	return true
}

var (
	// interfaceType is a pre-computed reflect.Type representing the empty interface.
	interfaceType = reflect.TypeOf((*interface{})(nil)).Elem()
	rawArgsType   = reflect.TypeOf((*RawArguments)(nil)).Elem()
)

// lowerCamelCase converts PascalCase string to
// a camelCase string (by lowering the first rune).
func lowerCamelCase(in string) string {
	isFirst := true
	return strings.Map(func(inRune rune) rune {
		if isFirst {
			isFirst = false
			return unicode.ToLower(inRune)
		}
		return inRune
	}, in)
}

// RawArguments is a special type that can be used for a marker
// to receive *all* raw, underparsed argument data for a marker.
// You probably want to use `interface{}` to match any type instead.
// Use *only* for legacy markers that don't follow Definition's normal
// parsing logic.  It should *not* be used as a field in a marker struct.
type RawArguments []byte

// ArgumentType is the kind of a marker argument type.
// It's roughly analogous to a subset of reflect.Kind, with
// an extra "AnyType" to represent the empty interface.
type ArgumentType int

const (
	// Invalid represents a type that can't be parsed, and should never be used.
	InvalidType ArgumentType = iota
	// IntType is an int
	IntType
	// StringType is a string
	StringType
	// BoolType is a bool
	BoolType
	// AnyType is the empty interface, and matches the rest of the content
	AnyType
	// SliceType is any slice constructed of the ArgumentTypes
	SliceType
	// RawType represents content that gets passed directly to the marker
	// without any parsing. It should *only* be used with anonymous markers.
	RawType
)

// Argument is the type of a marker argument.
type Argument struct {
	Type     ArgumentType
	Optional bool
	Pointer  bool

	ItemType *Argument
}

// typeString contains the internals of TypeString.
func (a Argument) typeString(out *strings.Builder) {
	if a.Pointer {
		out.WriteRune('*')
	}

	switch a.Type {
	case InvalidType:
		out.WriteString("<invalid>")
	case IntType:
		out.WriteString("int")
	case StringType:
		out.WriteString("string")
	case BoolType:
		out.WriteString("bool")
	case AnyType:
		out.WriteString("<any>")
	case SliceType:
		out.WriteString("[]")
		// arguments can't be non-pointer optional, so just call into typeString again.
		a.ItemType.typeString(out)
	case RawType:
		out.WriteString("<raw>")
	}
}

// TypeString returns a string roughly equivalent
// (but not identical) to the underlying Go type that
// this argument would parse to.  It's mainly useful
// for user-friendly formatting of this argument (e.g.
// help strings).
func (a Argument) TypeString() string {
	out := &strings.Builder{}
	a.typeString(out)
	return out.String()
}

func (a Argument) String() string {
	if a.Optional {
		return fmt.Sprintf("<optional arg %s>", a.TypeString())
	}
	return fmt.Sprintf("<arg %s>", a.TypeString())
}

// castAndSet casts val to out's type if needed,
// then sets out to val.
func castAndSet(out, val reflect.Value) {
	outType := out.Type()
	if outType != val.Type() {
		val = val.Convert(outType)
	}
	out.Set(val)
}

// makeSliceType makes a reflect.Type for a slice of the given type.
// Useful for constructing the out value for when AnyType's guess returns a slice.
func makeSliceType(itemType Argument) (reflect.Type, error) {
	var itemReflectedType reflect.Type
	switch itemType.Type {
	case IntType:
		itemReflectedType = reflect.TypeOf(int(0))
	case StringType:
		itemReflectedType = reflect.TypeOf("")
	case BoolType:
		itemReflectedType = reflect.TypeOf(false)
	case SliceType:
		subItemType, err := makeSliceType(*itemType.ItemType)
		if err != nil {
			return nil, err
		}
		itemReflectedType = subItemType
	default:
		return nil, fmt.Errorf("invalid type when constructing guessed slice out: %v", itemType.Type)
	}

	if itemType.Pointer {
		itemReflectedType = reflect.PtrTo(itemReflectedType)
	}

	return reflect.SliceOf(itemReflectedType), nil
}

// guessType takes an educated guess about the type of the next field.  If allowSlice
// is false, it will not guess slices.  It's less efficient than parsing with actual
// type information, since we need to allocate to peek ahead full tokens, and the scanner
// only allows peeking ahead one character.
func guessType(scanner *sc.Scanner, raw string, allowSlice bool) *Argument {
	if allowSlice {
		maybeItem := guessType(scanner, raw, false)

		subRaw := raw[scanner.Pos().Offset:]
		subScanner := parserScanner(subRaw, scanner.Error)

		var tok rune
		for tok = subScanner.Scan(); tok != ',' && tok != sc.EOF && tok != ';'; tok = subScanner.Scan() {
			// wait till we get something interesting
		}

		// semicolon means it's a legacy slice
		if tok == ';' {
			return &Argument{
				Type:     SliceType,
				ItemType: maybeItem,
			}
		}

		return maybeItem
	}

	// first, try the easy case -- quoted strings strings
	hint := scanner.Peek()
	switch hint {
	case '"', '\'', '`':
		return &Argument{Type: StringType}
	}

	// everything else needs a duplicate scanner to scan properly
	// (so we don't consume our scanner tokens until we actually
	// go to use this -- Go doesn't like scanners that can be rewound).
	subRaw := raw[scanner.Pos().Offset:]
	subScanner := parserScanner(subRaw, scanner.Error)

	// next, check for slices
	if hint == '{' {
		subScanner.Scan()
		return &Argument{
			Type:     SliceType,
			ItemType: guessType(subScanner, subRaw, false),
		}
	}

	// then, bools...
	probablyString := false
	if hint == 't' || hint == 'f' {
		// maybe a bool
		if nextTok := subScanner.Scan(); nextTok == sc.Ident {
			switch subScanner.TokenText() {
			case "true", "false":
				// definitely a bool
				return &Argument{Type: BoolType}
			}
			// probably a string
			probablyString = true
		} else {
			// we shouldn't ever get here
			scanner.Error(scanner, fmt.Sprintf("got a token (%q) that looked like an ident, but was not", scanner.TokenText()))
			return &Argument{Type: InvalidType}
		}
	}

	if !probablyString {
		if nextTok := subScanner.Scan(); nextTok == sc.Int {
			return &Argument{Type: IntType}
		}
	}

	// otherwise assume bare strings
	return &Argument{Type: StringType}
}

// parseString parses either of the two accepted string forms (quoted, or bare tokens).
func (a *Argument) parseString(scanner *sc.Scanner, raw string, out reflect.Value) {
	// strings are a bit weird -- the "easy" case is quoted strings (tokenized as strings),
	// the "hard" case (present for backwards compat) is a bare sequence of tokens that aren't
	// a comma.
	tok := scanner.Scan()
	if tok == sc.String || tok == sc.RawString {
		// the easy case
		val, err := strconv.Unquote(scanner.TokenText())
		if err != nil {
			scanner.Error(scanner, fmt.Sprintf("unable to parse string: %v", err))
			return
		}
		castAndSet(out, reflect.ValueOf(val))
		return
	}

	// the "hard" case -- bare tokens not including ',' (the argument
	// separator), ';' (the slice separator), or '}' (delimitted slice
	// ender)
	startPos := scanner.Position.Offset
	for hint := scanner.Peek(); hint != ',' && hint != ';' && hint != '}' && hint != sc.EOF; hint = scanner.Peek() {
		// skip this token
		scanner.Scan()
	}
	endPos := scanner.Position.Offset + len(scanner.TokenText())
	castAndSet(out, reflect.ValueOf(raw[startPos:endPos]))
}

// parseSlice parses either of the two slice forms (curly-brace-delimitted and semicolon-separated).
func (a *Argument) parseSlice(scanner *sc.Scanner, raw string, out reflect.Value) {
	// slices have two supported formats, like string:
	// - `{val, val, val}` (preferred)
	// - `val;val;val` (legacy)
	resSlice := reflect.Zero(out.Type())
	elem := reflect.Indirect(reflect.New(out.Type().Elem()))

	// preferred case
	if scanner.Peek() == '{' {
		// NB(directxman12): supporting delimitted slices in bare slices
		// would require an extra look-ahead here :-/

		scanner.Scan() // skip '{'
		for hint := scanner.Peek(); hint != '}' && hint != sc.EOF; hint = scanner.Peek() {
			a.ItemType.parse(scanner, raw, elem, true)
			resSlice = reflect.Append(resSlice, elem)
			tok := scanner.Peek()
			if tok == '}' {
				break
			}
			if !expect(scanner, ',', "comma") {
				return
			}
		}
		if !expect(scanner, '}', "close curly brace") {
			return
		}
		castAndSet(out, resSlice)
		return
	}

	// legacy case
	for hint := scanner.Peek(); hint != ',' && hint != '}' && hint != sc.EOF; hint = scanner.Peek() {
		a.ItemType.parse(scanner, raw, elem, true)
		resSlice = reflect.Append(resSlice, elem)
		tok := scanner.Peek()
		if tok == ',' || tok == '}' || tok == sc.EOF {
			break
		}
		scanner.Scan()
		if tok != ';' {
			scanner.Error(scanner, fmt.Sprintf("expected comma, got %q", scanner.TokenText()))
			return
		}
	}
	castAndSet(out, resSlice)
}

// parse functions like Parse, except that it allows passing down whether or not we're
// already in a slice, to avoid duplicate legacy slice detection for AnyType
func (a *Argument) parse(scanner *sc.Scanner, raw string, out reflect.Value, inSlice bool) {
	// nolint:gocyclo
	if a.Type == InvalidType {
		scanner.Error(scanner, fmt.Sprintf("cannot parse invalid type"))
		return
	}
	if a.Pointer {
		out.Set(reflect.New(out.Type().Elem()))
		out = reflect.Indirect(out)
	}
	switch a.Type {
	case RawType:
		// raw consumes everything else
		castAndSet(out, reflect.ValueOf(raw[scanner.Pos().Offset:]))
		// consume everything else
		for tok := scanner.Scan(); tok != sc.EOF; tok = scanner.Scan() {
		}
	case IntType:
		if !expect(scanner, sc.Int, "integer") {
			return
		}
		// TODO(directxman12): respect the size when parsing
		val, err := strconv.Atoi(scanner.TokenText())
		if err != nil {
			scanner.Error(scanner, fmt.Sprintf("unable to parse integer: %v", err))
			return
		}
		castAndSet(out, reflect.ValueOf(val))
	case StringType:
		// strings are a bit weird -- the "easy" case is quoted strings (tokenized as strings),
		// the "hard" case (present for backwards compat) is a bare sequence of tokens that aren't
		// a comma.
		a.parseString(scanner, raw, out)
	case BoolType:
		if !expect(scanner, sc.Ident, "true or false") {
			return
		}
		switch scanner.TokenText() {
		case "true":
			castAndSet(out, reflect.ValueOf(true))
		case "false":
			castAndSet(out, reflect.ValueOf(false))
		default:
			scanner.Error(scanner, fmt.Sprintf("expected true or false, got %q", scanner.TokenText()))
			return
		}
	case AnyType:
		guessedType := guessType(scanner, raw, !inSlice)
		newOut := out
		if guessedType.Type == SliceType {
			// we need to be able to construct the right element types, below
			// in parse, so construct a concretely-typed value to use as "out"
			newType, err := makeSliceType(*guessedType.ItemType)
			if err != nil {
				scanner.Error(scanner, err.Error())
				return
			}
			newOut = reflect.Indirect(reflect.New(newType))
		}
		if !newOut.CanSet() {
			panic("at the disco")
		}
		guessedType.Parse(scanner, raw, newOut)
		castAndSet(out, newOut)
	case SliceType:
		// slices have two supported formats, like string:
		// - `{val, val, val}` (preferred)
		// - `val;val;val` (legacy)
		a.parseSlice(scanner, raw, out)
	}
}

// Parse attempts to consume the argument from the given scanner (based on the given
// raw input as well for collecting ranges of content), and places the output value
// in the given reflect.Value.  Errors are reported via the given scanner.
func (a *Argument) Parse(scanner *sc.Scanner, raw string, out reflect.Value) {
	a.parse(scanner, raw, out, false)
}

// ArgumentFromType constructs an Argument by examining the given
// raw reflect.Type.  It can construct arguments from the Go types
// corresponding to any of the types listed in ArgumentType.
func ArgumentFromType(rawType reflect.Type) (Argument, error) {
	if rawType == rawArgsType {
		return Argument{
			Type: RawType,
		}, nil
	}

	if rawType == interfaceType {
		return Argument{
			Type: AnyType,
		}, nil
	}

	arg := Argument{}
	if rawType.Kind() == reflect.Ptr {
		rawType = rawType.Elem()
		arg.Pointer = true
		arg.Optional = true
	}

	switch rawType.Kind() {
	case reflect.String:
		arg.Type = StringType
	case reflect.Int, reflect.Int32: // NB(directxman12): all ints in kubernetes are int32, so explicitly support that
		arg.Type = IntType
	case reflect.Bool:
		arg.Type = BoolType
	case reflect.Slice:
		arg.Type = SliceType
		itemType, err := ArgumentFromType(rawType.Elem())
		if err != nil {
			return Argument{}, fmt.Errorf("bad slice item type: %v", err)
		}
		arg.ItemType = &itemType
	default:
		return Argument{}, fmt.Errorf("type has unsupported kind %s", rawType.Kind())
	}

	return arg, nil
}

// TargetType describes which kind of node a given marker is associated with.
type TargetType int

const (
	// DescribesPackage indicates that a marker is associated with a package.
	DescribesPackage TargetType = iota
	// DescribesType indicates that a marker is associated with a type declaration.
	DescribesType
	// DescribesField indicates that a marker is associated with a struct field.
	DescribesField
)

// Definition is a parsed definition of a marker.
type Definition struct {
	// Output is the deserialized Go type of the marker.
	Output reflect.Type
	// Name is the marker's name.
	Name string
	// Target indicates which kind of node this marker can be associated with.
	Target TargetType
	// Fields lists out the types of each field that this marker has, by
	// argument name as used in the marker (if the output type isn't a struct,
	// it'll have a single, blank field name).  This only lists exported fields,
	// (as per reflection rules).
	Fields map[string]Argument
	// FieldNames maps argument names (as used in the marker) to struct field name
	// in the output type.
	FieldNames map[string]string
	// Strict indicates that this definition should error out when parsing if
	// not all non-optional fields were seen.
	Strict bool
}

// AnonymousField indicates that the definition has one field,
// (actually the original object), and thus the field
// doesn't get named as part of the name.
func (d *Definition) AnonymousField() bool {
	if len(d.Fields) != 1 {
		return false
	}
	_, hasAnonField := d.Fields[""]
	return hasAnonField
}

// Empty indicates that this definition has no fields.
func (d *Definition) Empty() bool {
	return len(d.Fields) == 0
}

// loadFields uses reflection to populate argument information from the Output type.
func (d *Definition) loadFields() error {
	if d.Fields == nil {
		d.Fields = make(map[string]Argument)
		d.FieldNames = make(map[string]string)
	}
	if d.Output.Kind() != reflect.Struct {
		// anonymous field type
		argType, err := ArgumentFromType(d.Output)
		if err != nil {
			return err
		}
		d.Fields[""] = argType
		d.FieldNames[""] = ""
		return nil
	}

	for i := 0; i < d.Output.NumField(); i++ {
		field := d.Output.Field(i)
		if field.PkgPath != "" {
			// as per the reflect package docs, pkgpath is empty for exported fields,
			// so non-empty package path means a private field, which we should skip
			continue
		}
		argName := lowerCamelCase(field.Name)
		markerTag, tagSpecified := field.Tag.Lookup("marker")
		markerTagParts := strings.Split(markerTag, ",")

		if tagSpecified && markerTagParts[0] != "" {
			// allow overriding to support legacy cases where we don't follow camelCase conventions
			argName = markerTagParts[0]
		}

		argType, err := ArgumentFromType(field.Type)
		if err != nil {
			return fmt.Errorf("unable to extract type information for field %q: %v", field.Name, err)
		}

		if argType.Type == RawType {
			return fmt.Errorf("RawArguments must be the direct type of a marker, and not a field")
		}

		for _, tagOption := range markerTagParts[1:] {
			switch tagOption {
			case "optional":
				argType.Optional = true
			}
		}

		d.Fields[argName] = argType
		d.FieldNames[argName] = field.Name
	}

	return nil
}

// parserScanner makes a new scanner appropriate for use in parsing definitions and arguments.
func parserScanner(raw string, err func(*sc.Scanner, string)) *sc.Scanner {
	scanner := &sc.Scanner{}
	scanner.Init(bytes.NewBufferString(raw))
	scanner.Mode = sc.ScanIdents | sc.ScanInts | sc.ScanStrings | sc.ScanRawStrings | sc.SkipComments
	scanner.Error = err

	return scanner
}

// Parse uses the type information in this Definition to parse the given
// raw marker in the form `+a:b:c=arg,d=arg` into an output object of the
// type specified in the definition.
func (d *Definition) Parse(rawMarker string) (interface{}, error) {
	name, anonName, fields := splitMarker(rawMarker)

	out := reflect.Indirect(reflect.New(d.Output))

	// if we're a not a struct or have no arguments, treat the full `a:b:c` as the name,
	// otherwise, treat `c` as a field name, and `a:b` as the marker name.
	if !d.AnonymousField() && !d.Empty() && len(anonName) >= len(name)+1 {
		fields = anonName[len(name)+1:] + "=" + fields
	}

	var errs []error
	scanner := parserScanner(fields, func(_ *sc.Scanner, msg string) {
		errs = append(errs, errors.New(msg))
	})

	// TODO(directxman12): strict parsing where we error out if certain fields aren't optional
	seen := make(map[string]struct{}, len(d.Fields))
	if d.AnonymousField() && scanner.Peek() != sc.EOF {
		// no need for trying to parse field names if we're not a struct
		field := d.Fields[""]
		field.Parse(scanner, fields, out)
		seen[""] = struct{}{} // mark as seen for strict definitions
	} else if !d.Empty() && scanner.Peek() != sc.EOF {
		// if we expect *and* actually have arguments passed
		for {
			// parse the argument name
			if !expect(scanner, sc.Ident, "argument name") {
				break
			}
			argName := scanner.TokenText()
			if !expect(scanner, '=', "equals") {
				break
			}

			// make sure we know the field
			fieldName, known := d.FieldNames[argName]
			if !known {
				scanner.Error(scanner, fmt.Sprintf("unknown argument %q", argName))
				break
			}
			fieldType, known := d.Fields[argName]
			if !known {
				scanner.Error(scanner, fmt.Sprintf("unknown argument %q", argName))
				break
			}
			seen[argName] = struct{}{} // mark as seen for strict definitions

			// parse the field value
			fieldVal := out.FieldByName(fieldName)
			if !fieldVal.CanSet() {
				scanner.Error(scanner, fmt.Sprintf("cannot set field %q (might not exist)", fieldName))
				break
			}
			fieldType.Parse(scanner, fields, fieldVal)

			if len(errs) > 0 {
				break
			}

			if scanner.Peek() == sc.EOF {
				break
			}
			if !expect(scanner, ',', "comma") {
				break
			}
		}
	}

	if tok := scanner.Scan(); tok != sc.EOF {
		scanner.Error(scanner, fmt.Sprintf("extra arguments provided: %q", fields[scanner.Position.Offset:]))
	}

	if d.Strict {
		for argName, arg := range d.Fields {
			if _, wasSeen := seen[argName]; !wasSeen && !arg.Optional {
				scanner.Error(scanner, fmt.Sprintf("missing argument %q", argName))
			}
		}
	}

	return out.Interface(), loader.MaybeErrList(errs)
}

// MakeDefinition constructs a definition from a name, type, and the output type.
// All such definitions are strict by default.  If a struct is passed as the output
// type, its public fields will automatically be populated into Fields (and similar
// fields in Definition).  Other values will have a single, empty-string-named Fields
// entry.
func MakeDefinition(name string, target TargetType, output interface{}) (*Definition, error) {
	def := &Definition{
		Name:   name,
		Target: target,
		Output: reflect.TypeOf(output),
		Strict: true,
	}

	if err := def.loadFields(); err != nil {
		return nil, err
	}

	return def, nil
}

// splitMarker takes a marker in the form of `+a:b:c=arg,d=arg` and splits it
// into the name (`a:b`), the name if it's not a struct (`a:b:c`), and the parts
// that are definitely fields (`arg,d=arg`).
func splitMarker(raw string) (name string, anonymousName string, restFields string) {
	raw = raw[1:] // get rid of the leading '+'
	nameFieldParts := strings.SplitN(raw, "=", 2)
	if len(nameFieldParts) == 1 {
		return nameFieldParts[0], nameFieldParts[0], ""
	}
	anonymousName = nameFieldParts[0]
	name = anonymousName
	restFields = nameFieldParts[1]

	nameParts := strings.Split(name, ":")
	if len(nameParts) > 1 {
		name = strings.Join(nameParts[:len(nameParts)-1], ":")
	}
	return name, anonymousName, restFields
}

// Registry keeps track of registered definitions, and allows for easy lookup.
// It's thread-safe, and the zero-value can be safely used.
type Registry struct {
	forPkg   map[string]*Definition
	forType  map[string]*Definition
	forField map[string]*Definition

	mu       sync.RWMutex
	initOnce sync.Once
}

func (r *Registry) init() {
	r.initOnce.Do(func() {
		if r.forPkg == nil {
			r.forPkg = make(map[string]*Definition)
		}
		if r.forType == nil {
			r.forType = make(map[string]*Definition)
		}
		if r.forField == nil {
			r.forField = make(map[string]*Definition)
		}
	})
}

// Define defines a new marker with the given name, target, and output type.
// It's a shortcut around
//  r.Register(MakeDefinition(name, target, obj))
func (r *Registry) Define(name string, target TargetType, obj interface{}) error {
	def, err := MakeDefinition(name, target, obj)
	if err != nil {
		return err
	}
	return r.Register(def)
}

// Register registers the given marker definition with this registry for later lookup.
func (r *Registry) Register(def *Definition) error {
	r.init()

	r.mu.Lock()
	defer r.mu.Unlock()

	switch def.Target {
	case DescribesPackage:
		r.forPkg[def.Name] = def
	case DescribesType:
		r.forType[def.Name] = def
	case DescribesField:
		r.forField[def.Name] = def
	default:
		return fmt.Errorf("unknown target type %v", def.Target)
	}
	return nil
}

// Lookup fetches the definition corresponding to the given name and target type.
func (r *Registry) Lookup(name string, target TargetType) *Definition {
	r.init()

	r.mu.RLock()
	defer r.mu.RUnlock()

	switch target {
	case DescribesPackage:
		return tryAnonLookup(name, r.forPkg)
	case DescribesType:
		return tryAnonLookup(name, r.forType)
	case DescribesField:
		return tryAnonLookup(name, r.forField)
	default:
		return nil
	}
}

// AllDefinitions returns all marker definitions known to this registry.
func (r *Registry) AllDefinitions() []*Definition {
	res := make([]*Definition, 0, len(r.forPkg)+len(r.forType)+len(r.forField))
	for _, def := range r.forPkg {
		res = append(res, def)
	}
	for _, def := range r.forType {
		res = append(res, def)
	}
	for _, def := range r.forField {
		res = append(res, def)
	}
	return res
}

// tryAnonLookup tries looking up the given marker as both an struct-based
// marker and an anonymous marker, returning whichever format matches first,
// preferring the longer (anonymous) name in case of conflicts.
func tryAnonLookup(name string, defs map[string]*Definition) *Definition {
	// NB(directxman12): we look up anonymous names first to work with
	// legacy style marker definitions that have a namespaced approach
	// (e.g. deepcopy-gen, which uses `+k8s:deepcopy-gen=foo,bar` *and*
	// `+k8s.io:deepcopy-gen:interfaces=foo`).
	name, anonName, _ := splitMarker(name)
	if def, exists := defs[anonName]; exists {
		return def
	}

	return defs[name]
}

// Must panics on errors creating definitions.
func Must(def *Definition, err error) *Definition {
	if err != nil {
		panic(err)
	}
	return def
}
