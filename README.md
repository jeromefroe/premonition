# Premonition

An example repo of how to unmarshal structs in Go without knowing their types ahead of time.
To run the example:

```bash
go build *.go
./decode
```

## Problem

When unmarshaling a struct in Go, one often knows the type of the struct that is being unmarshaled
and writes code like the following, which comes from [the standard library's `encoding/json`
package's example of it's `Unmarshal` method]:

```go
var animals []Animal
err := json.Unmarshal(jsonBlob, &animals)
```

However, what if you need to unmarshal a struct, but don't know the type of the struct ahead of
time? This situation occurs, for instance, in the Kubernetes ecosystem, where resources can be
defined in configuration files, and tools need to be able to read those configuration files to
act upon the resources they contain. [The solution that Kubernetes adopted], and that this repo
recreates, is [to require a metadata type to be included in all supported structs] that can be used
to identify, at runtime, which struct is being unmarshaled via a registry.

## Solution

### Define Type Metadata

The first step is to define a type metadata struct, `TypeMeta`, that all structs that need to be
unmarshaled must contain and will provide the necessary information to determine, at runtime, the
type of the struct that we are unmarshaling. We also define an `Object` interface, that `TypeMeta`
implements that, that returns the type information. From [types.go]:

```go
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
```

### Define Registry

Our next step is to define a registry that will provide a mapping from the type metadata of a
struct, defined using our `TypeMeta` struct, to the struct's type representation defined via
[`reflect.Type`]. We then instantiate an instance of this registry which will serve as the
registry that all supported structs should be added to, and create a method for doing so. From
[registry.go]:

```go
// ObjectRegistry is a map from the name of an object's type to it's actual Go type.
type ObjectRegistry = map[TypeMeta]reflect.Type

// Registry contains the type information for all registered objects.
var Registry = make(ObjectRegistry)

// MustRegisterObject registers an object. `meta` must define the object's type
// and there cannot be a different object with the same name in Registry already.
// If any of these conditions aren't met the function will panic. MustRegisterObject
// is intended to be called in init functions to register all valid types at startup.
func MustRegisterObject(meta TypeMeta, obj Object) {
  if err := registerObject(meta, obj, Registry); err != nil {
    panic(fmt.Sprintf("Unable to register Object: %v.", err))
  }
}
```

### Define Decode function

With our type metadata and registry defined, we can implement a function for decoding any registered
struct. The function will perform three main:

1. Unmarshal input as a `TypeMeta` to get the type's metadata.
2. Use the type metadata to look up the corresponding type representation of the struct in our
   registry.
3. Use the type representation to create an instance of the struct and unmarshal the input into
   that struct.

From [decode.go]:

```go
func decodeWithRegistry(r io.Reader, reg ObjectRegistry) ([]Object, error) {
  ...

  obj, err := findObject(raw, reg)
  if err != nil {
    return nil, err
  }

  if err := json.Unmarshal(raw, &obj); err != nil {
    return nil, fmt.Errorf("unable to unmarshal object: %v", err)
  }

  ...
}

func findObject(data []byte, reg ObjectRegistry) (Object, error) {
  var meta TypeMeta
  if err := json.Unmarshal(data, &meta); err != nil {
    return nil, fmt.Errorf("could not find \"type_name\", json parse error: %v", err)
  }

  t, ok := reg[meta]
  if !ok {
    return nil, fmt.Errorf("no registered type found for object with type name: %v", meta.TypeName)
  }

  return reflect.New(t).Interface().(Object), nil
}
```

### Define and Register Types

We can now define and register structs that we want to unmarshal. In the following block, for
example, we define a type `Apple` and register it in an init function. From [objects.go]:

```go
func init() {
  MustRegisterObject(AppleTypeMeta, &Apple{})
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
```

### Put it All Together

Finally, we are in a position where we can put everything together and demonstrate how to
unmarshal structs dynamically. From [main.go]:

```go
const input = `
type_name: Apple
color: Red
---
type_name: Banana
ripe: true
`

func main() {
  objs, err := Decode(strings.NewReader(input))
  if err != nil {
    log.Fatalf("unable to decode objects: %v", err)
  }

  for _, obj := range objs {
    fmt.Printf("%+v (%T)\n", obj, obj)
  }
}
```

And running our main function produces the following output where we can see each object was
unmarshaled appropriately:

```text
&{TypeMeta:{TypeName:Apple} Color:Red} (*main.Apple)
&{TypeMeta:{TypeName:Banana} Ripe:true} (*main.Banana)
```

[the standard library's `encoding/json` package's example of it's `unmarshal` method]: https://golang.org/pkg/encoding/json/#example_Unmarshal
[the solution that kubernetes adopted]: https://github.com/kubernetes/apimachinery/blob/master/pkg/runtime/doc.go
[to require a metadata type to be included in all supported structs]: https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#metadata
[types.go]: ./types.go
[`reflect.type`]: https://golang.org/pkg/reflect/#Type
[registry.go]: ./registry.go
[decode.go]: ./decode.go
[objects.go]: ./objects.go
[main.go]: ./main.go
