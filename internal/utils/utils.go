package utils

import (
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetFullCrdFQDN(obj *metav1.TypeMeta) string {
	return strings.ToLower(obj.GetObjectKind().GroupVersionKind().GroupKind().String())
}
