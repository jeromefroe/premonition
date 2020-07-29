package main

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	"k8s.io/apimachinery/pkg/util/yaml"
)

const defaultBufferSize = 4096

// Decode decodes objects from a stream until it encounters an EOF. It uses the object
// registry to discover the types of the objects it decodes.
func Decode(r io.Reader) ([]Object, error) {
	return decodeWithRegistry(r, Registry)
}

// decodeWithRegistry contains the actual logic for decoding objects from a stream. It
// accepts an ObjectRegistry as an argument to faciliate testing.
func decodeWithRegistry(r io.Reader, reg ObjectRegistry) ([]Object, error) {
	var (
		objs []Object
		raw  json.RawMessage
		dec  = yaml.NewYAMLOrJSONDecoder(r, defaultBufferSize)
	)
	for {
		raw = raw[:0]
		if err := dec.Decode(&raw); err != nil {
			if err != io.EOF {
				return nil, fmt.Errorf("unable to decode object: %v", err)
			}
			break
		}

		obj, err := findObject(raw, reg)
		if err != nil {
			return nil, err
		}

		// The YAMLOrJSONDecoder will convert objects defined in YAML into JSON so `raw` is
		// guaranteed to hold the JSON representation of the object.
		if err := json.Unmarshal(raw, &obj); err != nil {
			return nil, fmt.Errorf("unable to unmarshal object: %v", err)
		}

		objs = append(objs, obj)
	}

	return objs, nil
}

// findObject attempts to find the `type_name` field in a serialized JSON object
// and uses that information to look up the runtime type of the object in an
// ObjectRegistry.
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


