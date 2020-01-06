package util

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

type Money struct {
	decimal.Decimal
}

func (d Money) EncodeValues(key string, v *url.Values) error {
	v.Add(key, d.String())
	return nil
}

func DateFromStringOrPanic(t string) Date {
	var d Date
	err := d.UnmarshalJSON([]byte("\"" + t + "\""))
	if err != nil {
		panic(err)
	}
	return d
}

func (d *Date) Time() time.Time {
	return d.t
}

func (d *Date) String() string {
	return d.t.Format("2006-01-02")
}

func (d *Date) After(o Date) bool {
	return d.t.After(o.t)
}

func (d *Date) Before(o Date) bool {
	return d.t.Before(o.t)
}

func (d *Date) Equal(o Date) bool {
	return d.t.Equal(o.t)
}

type Date struct {
	t time.Time
}

func (d *Date) UnmarshalJSON(data []byte) error {
	// Ignore null, like in the main JSON package.
	if string(data) == "null" {
		return nil
	}
	// Fractional seconds are handled implicitly by Parse.
	var err error
	noQuote := data[1 : len(data)-2]
	date := strings.Split(string(noQuote), "T")
	part := strings.Split(date[0], "-")
	var intPart []int
	for _, p := range part {
		i, err := strconv.Atoi(p)
		if err != nil {
			return err
		}
		intPart = append(intPart, i)
	}

	d.t = time.Date(intPart[0], time.Month(intPart[1]), intPart[2], 0, 0, 0, 0, time.UTC)
	return err
}

func (d Date) EncodeValues(key string, v *url.Values) error {
	b, err := d.MarshalJSON()
	if err != nil {
		return err
	}
	if len(b) != 0 {
		v.Add(key, string(b[1:len(b)-1]))
	}
	return nil
}

func (d Date) MarshalJSON() ([]byte, error) {
	if y := d.t.Year(); y < 0 || y >= 10000 {
		// RFC 3339 is clear that years are 4 digits exactly.
		// See golang.org/issue/4556#c15 for more discussion.
		return nil, errors.New("Time.MarshalJSON: year outside of range [0,9999]")
	}

	if d.t.Year() < 1000 {
		return []byte{}, nil
	}

	b := make([]byte, 0, len(`"2019-12-12"`))
	b = append(b, '"')
	b = append(b, []byte(d.String())...)
	b = append(b, '"')
	return b, nil
}
