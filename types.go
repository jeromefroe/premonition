package main

// TypeMeta contains the metadata required to identify the type of an object.
type TypeMeta struct {
	// TypeName is the name of an object's type.
	TypeName string `json:"type_name,omitempty"`
}

// Object is the interface that all types must fulfill.
type Object interface {
	Type() TypeMeta
}

func (obj *TypeMeta) Type() TypeMeta { return *obj }
