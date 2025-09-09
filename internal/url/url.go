package url

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// URL wraps url.URL.
// It has custom json marshal methods that enable it to be used in K8s CRDs
// such that the CRD resource will have the URL but operator code can can work with url.URL struct
type URL struct {
	url.URL
}

func (u *URL) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", u.String())), nil
}

func (u *URL) UnmarshalJSON(b []byte) error {
	var ref string
	if err := json.Unmarshal(b, &ref); err != nil {
		return err
	}
	if ref == "" {
		*u = URL{}
		return nil
	}

	r, err := url.Parse(ref)
	if err != nil {
		return err
	} else if r != nil {
		*u = URL{*r}
	} else {
		*u = URL{}
	}
	return nil
}

func (u *URL) String() string {
	if u == nil {
		return ""
	}
	return u.URL.String()
}

// DeepCopyInto copy the receiver, write into out. in must be non-nil.
func (in *URL) DeepCopyInto(out *URL) {
	*out = *in
}

// DeepCopy copy the receiver, create a new URL.
func (in *URL) DeepCopy() *URL {
	if in == nil {
		return nil
	}
	out := new(URL)
	in.DeepCopyInto(out)
	return out
}
