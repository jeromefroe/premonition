package main

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	errTypeNameMissing = errors.New("cannot register an Object that doesn't have a TypeName")
	errInvalidType     = errors.New("can only register types that are pointers to structs")
)

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

// registerObject contains the actual logic for registering an Object in an ObjectRegistry.
func registerObject(meta TypeMeta, obj Object, r ObjectRegistry) error {
	if meta.TypeName == "" {
		return errTypeNameMissing
	}

	t := reflect.TypeOf(obj)
	if t.Kind() != reflect.Ptr {
		return errInvalidType
	}
	t = t.Elem()
	if t.Kind() != reflect.Struct {
		return errInvalidType
	}

	if oldT, found := r[meta]; found && oldT != t {
		return fmt.Errorf(
			"Double registration of different types for %v: old=%v.%v, new=%v.%v",
			meta, oldT.PkgPath(), oldT.Name(), t.PkgPath(), t.Name(),
		)
	}
	r[meta] = t

	return nil
}
