package utils

import (
	"strconv"
	"strings"

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
