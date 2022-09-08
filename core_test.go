package main

import (
	"os"
	"testing"
)

func TestCore(t *testing.T) {
	f, err := os.Open(`messages`)
	if err != nil { t.Fatal(err) }
	core := NewCore()
	core.ReadLog(f)

}
