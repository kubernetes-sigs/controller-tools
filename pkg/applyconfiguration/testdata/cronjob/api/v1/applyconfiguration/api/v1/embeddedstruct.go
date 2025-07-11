// Code generated by applyconfiguration. DO NOT EDIT.

package v1

// EmbeddedStructApplyConfiguration represents a declarative configuration of the EmbeddedStruct type for use
// with apply.
type EmbeddedStructApplyConfiguration struct {
	FromEmbedded *string `json:"fromEmbedded,omitempty"`
}

// EmbeddedStructApplyConfiguration constructs a declarative configuration of the EmbeddedStruct type for use with
// apply.
func EmbeddedStruct() *EmbeddedStructApplyConfiguration {
	return &EmbeddedStructApplyConfiguration{}
}
func (b EmbeddedStructApplyConfiguration) IsApplyConfiguration() {}

// WithFromEmbedded sets the FromEmbedded field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the FromEmbedded field is set to the value of the last call.
func (b *EmbeddedStructApplyConfiguration) WithFromEmbedded(value string) *EmbeddedStructApplyConfiguration {
	b.FromEmbedded = &value
	return b
}
