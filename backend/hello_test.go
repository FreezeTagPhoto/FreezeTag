package main

import (
	"regexp"
	"testing"
)

func TestHelloName(t *testing.T) {
	name := "Gladys"
	want := regexp.MustCompile(`\b` + name + `\b`)
	msg := Hello(name)
	if !want.MatchString(msg) {
		t.Errorf(`Hello("Gladys") = %q, want match for %#q`, msg, want)
	}
}

func TestHelloEmpty(t *testing.T) {
	name := ""
	want := ""
	msg := Hello(name)
	if want != msg {
		t.Errorf(`Hello("") = %q, want %q`, msg, want)
	}
}
