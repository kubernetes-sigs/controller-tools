// Code generated by controller-gen. DO NOT EDIT.

package cronjob

// AssociativeTypeApplyConfiguration represents an declarative configuration of the AssociativeType type for use
// with apply.
type AssociativeTypeApplyConfiguration struct {
	Name      *string `json:"name,omitempty"`
	Secondary *int    `json:"secondary,omitempty"`
	Foo       *string `json:"foo,omitempty"`
}

// AssociativeTypeApplyConfiguration constructs an declarative configuration of the AssociativeType type for use with
// apply.
func AssociativeType() *AssociativeTypeApplyConfiguration {
	return &AssociativeTypeApplyConfiguration{}
}

// WithName sets the Name field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Name field is set to the value of the last call.
func (b *AssociativeTypeApplyConfiguration) WithName(value string) *AssociativeTypeApplyConfiguration {
	b.Name = &value
	return b
}

// WithSecondary sets the Secondary field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Secondary field is set to the value of the last call.
func (b *AssociativeTypeApplyConfiguration) WithSecondary(value int) *AssociativeTypeApplyConfiguration {
	b.Secondary = &value
	return b
}

// WithFoo sets the Foo field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Foo field is set to the value of the last call.
func (b *AssociativeTypeApplyConfiguration) WithFoo(value string) *AssociativeTypeApplyConfiguration {
	b.Foo = &value
	return b
}
