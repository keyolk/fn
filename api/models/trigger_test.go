package models

import (
	"encoding/json"
	"testing"
)

var openEmptyJson = `{"id":"","name":"","app_id":"","fn_id":"","created_at":"0001-01-01T00:00:00.000Z","updated_at":"0001-01-01T00:00:00.000Z","type":"Unknown","source":""`

var triggerJsonCases = []struct {
	val       *Trigger
	valString string
}{
	{val: &Trigger{}, valString: openEmptyJson + "}"},
	{val: &Trigger{Extensions: map[string]interface{}{"foo": "bar"}}, valString: openEmptyJson + `,"extensions":{"foo":"bar"}}`},
	{val: &Trigger{Extensions: map[string]interface{}{"baz": 1}}, valString: openEmptyJson + `,"extensions":{"baz":1}}`},
}

func TestTriggerJsonMarshalling(t *testing.T) {
	for _, tc := range triggerJsonCases {
		v, err := json.Marshal(tc.val)
		if err != nil {
			t.Fatalf("Failed to marshal json into %s: %v", tc.valString, err)
		}
		if string(v) != tc.valString {
			t.Errorf("Invalid trigger value, expected %s, got %s", tc.valString, string(v))
		}
	}
}

var httpTrigger = &Trigger{Name: "name", AppID: "foo", FnID: "bar", Type: HTTP, Source: "baz"}

var httpTriggerWithExtension = &Trigger{Name: "name", AppID: "foo", FnID: "bar", Type: HTTP, Source: "baz"}

func (t *Trigger) SetExtensions(extension map[string]interface{}) {
	t.Extensions = extension
}

func init() {
	httpTriggerWithExtension.SetExtensions(map[string]interface{}{"foo": "bar"})
}

var triggerValidateCases = []struct {
	val   *Trigger
	valid bool
}{
	{val: &Trigger{}, valid: false},
	{val: httpTrigger, valid: true},
	{val: httpTriggerWithExtension, valid: false},
}

func TestTriggerValidate(t *testing.T) {
	for _, tc := range triggerValidateCases {
		v := tc.val.Validate()
		if v != nil && tc.valid {
			t.Errorf("Expected Trigger to be valid, but err (%s) returned. Trigger: %#v", v, tc.val)
		}
		if v == nil && !tc.valid {
			t.Errorf("Expected Trigger to be invalid, but no err returned. Trigger: %#v", tc.val)
		}
	}
}
