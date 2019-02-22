package main

import (
	"os"
	"testing"
)

func TestDeleteDirectory(t *testing.T) {
	tmpPath := "tmp-dir"
	tmpFilePath := tmpPath + "/test.tsv"

	os.MkdirAll(tmpPath, os.ModePerm)

	f, err := os.Create(tmpFilePath)
	if err != nil {
		t.Fatalf("Could not create file:%s\n", tmpFilePath)
	}
	f.Write([]byte("1234567890123456789012345678901234567890123456789012345678901234\tusername\tpassword\n"))
	f.Close()

	deleteTmpDir(tmpPath)

	if _, err := os.Stat(tmpPath); os.IsNotExist(err) == false {
		t.Fatalf("Directory \"%s\" still exists.", tmpPath)
	}
}
