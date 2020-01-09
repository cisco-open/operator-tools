package utils

import (
	"fmt"
	"hash/fnv"
	"sort"

	"emperror.dev/errors"
	"github.com/iancoleman/orderedmap"
)

// MergeLabels merge into map[string]string map
func MergeLabels(labelGroups ...map[string]string) map[string]string {
	mergedLabels := make(map[string]string)
	for _, labels := range labelGroups {
		for k, v := range labels {
			mergedLabels[k] = v
		}
	}
	return mergedLabels
}

// IntPointer converts int32 to *int32
func IntPointer(i int32) *int32 {
	return &i
}

// IntPointer converts int64 to *int64
func IntPointer64(i int64) *int64 {
	return &i
}

// BoolPointer converts bool to *bool
func BoolPointer(b bool) *bool {
	return &b
}

// StringPointer converts string to *string
func StringPointer(s string) *string {
	return &s
}

// OrderedStringMap
func OrderedStringMap(original map[string]string) *orderedmap.OrderedMap {
	o := orderedmap.New()
	for k, v := range original {
		o.Set(k, v)
	}
	o.SortKeys(sort.Strings)
	return o
}

// Contains check if a string item exists in []string
func Contains(s []string, e string) bool {
	for _, i := range s {
		if i == e {
			return true
		}
	}
	return false
}

// Hash32 calculate for string
func Hash32(in string) (string, error) {
	hasher := fnv.New32()
	_, err := hasher.Write([]byte(in))
	if err != nil {
		return "", errors.WrapIf(err, "failed to calculate hash for the configmap data")
	}
	return fmt.Sprintf("%x", hasher.Sum32()), nil
}
