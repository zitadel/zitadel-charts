package assert

import (
	"fmt"
	"reflect"

	"github.com/onsi/gomega/types"
)

// Partial wraps an Assertable as a gomega matcher for use with
// ContainElement and other gomega combinators. It performs the same
// partial matching as AssertPartial — only fields with a non-nil Val
// or Matcher are checked; zero-valued fields are skipped.
//
// Example:
//
//	gomega.ContainElement(assert.Partial(assert.VolumeAssertion{
//	    Name: assert.Some("server-ssl-crt"),
//	}))
func Partial(assertion Assertable) types.GomegaMatcher {
	return &partialMatcher{assertion: assertion}
}

type partialMatcher struct {
	assertion Assertable
}

func (m *partialMatcher) Match(actual interface{}) (bool, error) {
	return tryMatch(reflect.ValueOf(actual), reflect.ValueOf(m.assertion))
}

func (m *partialMatcher) FailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected\n\t%+v\nto partially match\n\t%+v", actual, m.assertion)
}

func (m *partialMatcher) NegatedFailureMessage(actual interface{}) string {
	return fmt.Sprintf("Expected\n\t%+v\nnot to partially match\n\t%+v", actual, m.assertion)
}

func tryMatch(actualVal, assertionVal reflect.Value) (bool, error) {
	for actualVal.Kind() == reflect.Ptr {
		if actualVal.IsNil() {
			return false, nil
		}
		actualVal = actualVal.Elem()
	}
	for assertionVal.Kind() == reflect.Ptr {
		if assertionVal.IsNil() {
			return false, nil
		}
		assertionVal = assertionVal.Elem()
	}

	if actualVal.Kind() != reflect.Struct {
		return false, fmt.Errorf("actual must be a struct, got %v (%v)", actualVal.Kind(), actualVal.Type())
	}
	if assertionVal.Kind() != reflect.Struct {
		return false, fmt.Errorf("assertion must be a struct, got %v", assertionVal.Kind())
	}

	assertionType := assertionVal.Type()
	for i := 0; i < assertionType.NumField(); i++ {
		fieldInfo := assertionType.Field(i)
		fieldVal := assertionVal.Field(i)

		actualField := actualVal.FieldByName(fieldInfo.Name)
		if !actualField.IsValid() {
			return false, fmt.Errorf("actual struct %v does not have field %q", actualVal.Type(), fieldInfo.Name)
		}

		// Nested Assertable struct — recurse if non-zero
		if fieldVal.Type().Implements(assertableType) {
			if fieldVal.IsZero() {
				continue
			}
			ok, err := tryMatch(actualField, fieldVal)
			if !ok || err != nil {
				return false, err
			}
			continue
		}

		// Opt[T] field — check Matcher, then Val
		matcherField := fieldVal.FieldByName("Matcher")
		if matcherField.IsValid() && !matcherField.IsNil() {
			matcher := matcherField.Interface().(types.GomegaMatcher)
			ok, err := matcher.Match(actualField.Interface())
			if !ok || err != nil {
				return false, err
			}
			continue
		}

		valField := fieldVal.FieldByName("Val")
		if !valField.IsValid() {
			return false, fmt.Errorf("field %q of type %v is neither Assertable nor Opt[T]", fieldInfo.Name, fieldVal.Type())
		}
		if valField.IsNil() {
			continue
		}

		expectedVal := valField.Elem()

		// Slice of Assertable: compare element-by-element with partial matching
		if expectedVal.Kind() == reflect.Slice && expectedVal.Type().Elem().Implements(assertableType) {
			if expectedVal.Len() != actualField.Len() {
				return false, nil
			}
			for j := 0; j < expectedVal.Len(); j++ {
				ok, err := tryMatch(actualField.Index(j), expectedVal.Index(j))
				if !ok || err != nil {
					return false, err
				}
			}
			continue
		}

		// Map subset matching: all expected keys must exist in actual with matching values
		if expectedVal.Kind() == reflect.Map {
			if expectedVal.Len() == 0 {
				if actualField.Kind() == reflect.Map && actualField.Len() > 0 {
					return false, nil
				}
				continue
			}
			for _, k := range expectedVal.MapKeys() {
				actualEntry := actualField.MapIndex(k)
				if !actualEntry.IsValid() {
					return false, nil
				}
				if !reflect.DeepEqual(expectedVal.MapIndex(k).Interface(), actualEntry.Interface()) {
					return false, nil
				}
			}
			continue
		}

		if !reflect.DeepEqual(expectedVal.Interface(), actualField.Interface()) {
			return false, nil
		}
	}
	return true, nil
}
