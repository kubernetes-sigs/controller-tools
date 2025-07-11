// Code generated by applyconfiguration. DO NOT EDIT.

package v1

// MinMaxObjectApplyConfiguration represents a declarative configuration of the MinMaxObject type for use
// with apply.
type MinMaxObjectApplyConfiguration struct {
	Foo *string `json:"foo,omitempty"`
	Bar *string `json:"bar,omitempty"`
	Baz *string `json:"baz,omitempty"`
}

// MinMaxObjectApplyConfiguration constructs a declarative configuration of the MinMaxObject type for use with
// apply.
func MinMaxObject() *MinMaxObjectApplyConfiguration {
	return &MinMaxObjectApplyConfiguration{}
}
func (b MinMaxObjectApplyConfiguration) IsApplyConfiguration() {}

// WithFoo sets the Foo field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Foo field is set to the value of the last call.
func (b *MinMaxObjectApplyConfiguration) WithFoo(value string) *MinMaxObjectApplyConfiguration {
	b.Foo = &value
	return b
}

// WithBar sets the Bar field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Bar field is set to the value of the last call.
func (b *MinMaxObjectApplyConfiguration) WithBar(value string) *MinMaxObjectApplyConfiguration {
	b.Bar = &value
	return b
}

// WithBaz sets the Baz field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Baz field is set to the value of the last call.
func (b *MinMaxObjectApplyConfiguration) WithBaz(value string) *MinMaxObjectApplyConfiguration {
	b.Baz = &value
	return b
}
