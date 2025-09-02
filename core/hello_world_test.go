package core

import "testing"

func TestHelloWorld(t *testing.T) {
    if 1+1 != 2 {
        t.Errorf("Expected 1 + 1 to equal 2")
    }
}