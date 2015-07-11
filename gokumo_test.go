package main

import (
	"testing"
)

// TODO  write some more tests

type baseReadingTestData struct {
	in   mecabResult
	got  string
	want string
}

func TestGetBaseReading(t *testing.T) {
	cases := []baseReadingTestData{
		baseReadingTestData{
			in: mecabResult{Base: "動く", Surface: "動き", Read: "いごき"}, want: "いごく"},
		baseReadingTestData{
			in: mecabResult{Base: "する", Surface: "し", Read: "し"}, want: "する"},
		baseReadingTestData{
			in: mecabResult{Base: "通行する", Surface: "通行し", Read: "つうこうし"}, want: "つうこうする"}}
	for _, c := range cases {
		c.got = getBaseReading(&c.in)
		if c.got != c.want {
			t.Errorf("getBaseReading(%q) == %q, wanted %q", c.in, c.got, c.want)
		}
	}
}
