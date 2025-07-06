package utils

import (
	"github.com/getkin/kin-openapi/openapi3"
)

// 递归设置所有嵌套Schema的additionalProperties为false并使所有字段required
func MakeAllFieldsRequired(schema *openapi3.Schema) {
	if schema == nil {
		return
	}

	// 设置additionalProperties为false
	schema.AdditionalProperties = openapi3.AdditionalProperties{
		Has: &[]bool{false}[0],
	}

	// 如果有属性，将所有属性设为required
	if schema.Properties != nil {
		schema.Required = []string{}
		for propName, propSchema := range schema.Properties {
			schema.Required = append(schema.Required, propName)
			// 递归处理嵌套Schema
			MakeAllFieldsRequired(propSchema.Value)
		}
	}

	// 处理数组类型
	if schema.Items != nil {
		MakeAllFieldsRequired(schema.Items.Value)
	}
}
