package uri

import (
	"fmt"
	"net/url"
	"reflect"
	"strings"
)

const (
	pathTag  = "path"
	queryTag = "query"
)

type URIBuilder struct {
}

func New() *URIBuilder {
	return &URIBuilder{}
}

func (b *URIBuilder) EncodeParams(path string, params any) string {
	epath := encodePath(path, params)
	equeries := encodeQuery(params)
	if len(equeries) != 0 {
		epath += "?" + equeries
	}
	return epath
}

func encodePath(path string, params interface{}) string {
	tagVals := iterateParams(params, pathTag)
	for k, v := range tagVals {
		path = strings.ReplaceAll(path, fmt.Sprintf("{%s}", k), url.PathEscape(v))
	}
	return path
}

func encodeQuery(params interface{}) string {
	tagVals := iterateParams(params, queryTag)

	queries := url.Values{}
	for k, v := range tagVals {
		queries.Add(k, v)
	}
	return queries.Encode()
}

func iterateParams(params interface{}, tagType string) map[string]string {
	tagVals := make(map[string]string)

	pv := reflect.ValueOf(params)
	if pv.Kind() == reflect.Ptr {
		pv = pv.Elem()
	}
	pt := pv.Type()
	for i := 0; i < pv.NumField(); i++ {
		field := pv.Field(i)
		tagName := pt.Field(i).Tag.Get(tagType)
		fv := formatFieldValue(field)
		if tagName != "" && fv != "" {
			tagVals[tagName] = fv
		}
	}
	return tagVals
}

func formatFieldValue(field reflect.Value) string {
	if field.Kind() == reflect.Ptr {
		if field.IsNil() {
			return ""
		}
		field = field.Elem()
	}

	//switch typedValue := field.Interface().(type) {
	//case pmodels.Date:
	//	return typedValue.PathFormat()
	//case emodels.NumericCIK:
	//	return typedValue.Pad()
	//}

	switch field.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16:
		return fmt.Sprintf("%d", field.Int())
	case reflect.Float32, reflect.Float64:
		return fmt.Sprintf("%g", field.Float())
	default:
		return fmt.Sprintf("%v", field.Interface())
	}
}
