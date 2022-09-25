package provider

import (
	"bytes"
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func StringValueOrNull(val *string) types.String {
	if val == nil {
		return types.String{Null: true}
	} else {
		return types.String{Value: *val}
	}
}

func ValueStringOrNull(val types.String) *string {
	if val.IsNull() {
		return nil
	} else {
		return &val.Value
	}
}

func MinifyJSONString(val string) string {
	compactContext := &bytes.Buffer{}
	if err := json.Compact(compactContext, []byte(val)); err != nil {
		panic(err)
	}
	return compactContext.String()
}
