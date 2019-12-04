package main

import (
    "strings"
    "testing"
)

func TestSay(t *testing.T) {
    word := say("world")
    if !strings.Contains(word, "world") {
        t.Errorf("world is not in word")
    }
}
