package main

import (
    "testing"
)

func testGetFirstParam(t *testing.T) {
    if getFirstParam("http://localhist:8080/meetings/123") != "123" {
        t.Error("Expected 123 - not mttching")
    }
}