package app

import (
	"bytes"
	"testing"
)

func TestNetworkFrame(t *testing.T) {

	message := []byte("ABABIL Network Test")

	var buffer bytes.Buffer

	err := WriteFrame(&buffer, message)
	if err != nil {
		t.Fatal(err)
	}

	result, err := ReadFrame(&buffer)
	if err != nil {
		t.Fatal(err)
	}

	if string(result) != string(message) {
		t.Fatal("frame data mismatch")
	}
}
