package pingone

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func HashByMapKey(key string) func(v interface{}) int {
	return func(v interface{}) int {
		m := v.(map[string]interface{})
		return schema.HashString(m[key])
	}
}

func marshalInterfaceToString(v []interface{}) []string {
	w := make([]string, 0)
	for _, y := range v {
		w = append(w, y.(string))
	}

	return w
}
