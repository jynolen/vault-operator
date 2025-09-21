package utils

import (
	"fmt"
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

func MapSecret(secret *corev1.Secret, name, src_key string, dest_key *string, mandatory ...bool) error {
	if secret.Type != corev1.SecretTypeOpaque {
		return fmt.Errorf("%s is not Opaque", name)
	}
	val, ok := secret.Data[src_key]
	if !ok {
		if len(mandatory) > 0 && mandatory[0] {
			return fmt.Errorf("%s does not contains key `%s`", name, src_key)
		}
		return nil
	}
	*dest_key = string(val)
	return nil
}

type KvMapping struct {
	Dest      *string
	Mandatory bool
}
