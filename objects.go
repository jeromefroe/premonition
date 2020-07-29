package main

func init() {
	MustRegisterObject(AppleTypeMeta, &Apple{})
	MustRegisterObject(BananaTypeMeta, &Banana{})
}

// AppleTypeName is the type name of an Apple object.
const AppleTypeName = "Apple"

// AppleTypeMeta is the type information for an Apple object.
var AppleTypeMeta = TypeMeta{TypeName: AppleTypeName}

// Apple is an example object.
type Apple struct {
	TypeMeta `json:",inline"`

	Color string `json:"color"`
}

// BananaTypeName is the type name of a Banana object.
const BananaTypeName = "Banana"

// BananaTypeMeta is the type information for a Banana object.
var BananaTypeMeta = TypeMeta{TypeName: BananaTypeName}

// Banana is an example object.
type Banana struct {
	TypeMeta `json:",inline"`

	Ripe bool `json:"ripe"`
}
