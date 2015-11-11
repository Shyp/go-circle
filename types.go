package circle

// Circle has some types that we want to decode to normal Go types, define/do
// that here.

import (
	"encoding/json"
	"net/url"
	"time"
)

// Unmarshallable URL
type URL url.URL

func (oururl *URL) UnmarshalJSON(b []byte) error {
	// extra hop here to strip the leading/trailing quotes
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return err
	}
	u, err := url.Parse(s)
	if err != nil {
		return err
	}
	*oururl = URL(*u)
	return nil
}

type CircleNullTime struct {
	Time  time.Time
	Valid bool
}

type CircleDuration time.Duration

func (cd *CircleDuration) UnmarshalJSON(b []byte) error {
	var d time.Duration
	err := json.Unmarshal(b, &d)
	if err != nil {
		return err
	}
	*cd = CircleDuration(d * time.Millisecond)
	return nil
}

// Necessary because Circle returns "null" for some time instances
func (nt *CircleNullTime) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		nt.Valid = false
		return nil
	}
	var t time.Time
	err := json.Unmarshal(b, &t)
	if err != nil {
		return err
	}
	nt.Valid = true
	nt.Time = t
	return nil
}
