package utils

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetFullCrdFQDN(obj *metav1.TypeMeta) string {
	return strings.ToLower(obj.GetObjectKind().GroupVersionKind().GroupKind().String())
}

func Map[T, V any](fn func(T) V, ts []T) []V {
	result := make([]V, len(ts))
	for i, t := range ts {
		result[i] = fn(t)
	}
	return result
}

func MapQuote(ts []string) []string {
	return Map(strconv.Quote, ts)
}

type ConfigKvMapping map[string]any

func (s *ConfigKvMapping) Apply() map[string]any {
	return map[string]any{}
}

func HclExport(object interface{}) (map[string]any, error) {
	_m := map[string]any{}
	val := reflect.Indirect(reflect.ValueOf(object))
	for i := range val.NumField() {
		field := reflect.TypeOf(object).Field(i)
		if hcl, ok := field.Tag.Lookup("hcl"); ok {
			if v, err := HclEscape(reflect.ValueOf(object).Field(i)); err != nil {
				return nil, err
			} else if v != nil {
				_m[hcl] = v
			}
		}
	}
	return _m, nil
}

func HclEscape(v reflect.Value) (any, error) {
	if v.Kind() == reflect.Pointer && v.IsNil() {
		return nil, nil
	}
	ptrVal := reflect.Indirect(v)
	switch ptrVal.Kind() {
	case reflect.Bool:
		return strconv.FormatBool(ptrVal.Bool()), nil
	case reflect.String:
		return strconv.Quote(ptrVal.String()), nil
	case reflect.Int32:
		return strconv.FormatInt(ptrVal.Int(), 10), nil
	case reflect.Slice:
		slice, _ := ptrVal.Interface().([]string)
		if len(slice) > 0 {
			return fmt.Sprintf("[%s]", strings.Join(Map(strconv.Quote, slice), ",")), nil
		}
	case reflect.Struct:
		switch ptrVal.Interface().(type) {
		case metav1.Duration:
			return strconv.Quote(ptrVal.Interface().(metav1.Duration).Duration.String()), nil
		default:
			return nil, fmt.Errorf("%s type is not handle", ptrVal.Interface())
		}
	default:
		return nil, fmt.Errorf("%s type is not handle", ptrVal.Kind())
	}
	return nil, nil
}

type SecretKvMapping map[string]struct {
	Src       string
	Dest      *string
	Mandatory bool
}

func (s *SecretKvMapping) Apply(secret *corev1.Secret, name string) error {
	for k, v := range *s {
		if secret.Type != corev1.SecretTypeOpaque {
			return fmt.Errorf("%s is not Opaque", name)
		}
		if v.Src != "" {
			k = v.Src
		}
		val, ok := secret.Data[k]
		if !ok {
			if v.Mandatory {
				return fmt.Errorf("%s does not contains key `%s`", name, k)
			}
			return nil
		} else {
			*v.Dest = string(val)
		}
	}
	return nil
}

func NotNilCheck(o *any) bool {
	return o != nil
}
