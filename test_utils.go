package main

import (
	"math/rand"
	"reflect"
	"testing"
	"time"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

func init() {
	rand.Seed(time.Now().UnixNano())
}

// RandString returns a random alphanumeric string on `n` length
func RandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// AssertEquals evaluates whether two values are equals, and fails the test if they are not
func AssertEquals(t *testing.T, a interface{}, b interface{}, msg string) {
	if a == b {
		return
	}
	t.Errorf("%s: Expected: %v (type %v) but was: %v (type %v))", msg, a, reflect.TypeOf(a), b, reflect.TypeOf(b))
}
