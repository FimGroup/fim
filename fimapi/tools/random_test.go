package tools

import "testing"

func TestGenerateRandomString(t *testing.T) {
	t.Log(RandomString())
	if len(RandomString()) != 32 {
		t.Fatal("length not match")
	}
}
