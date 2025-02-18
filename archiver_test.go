package arc

import (
	"context"
	"os"
	"testing"
)

func TestArchiveFiles(t *testing.T) {
	DEBUG = true
	ctx := context.Background()
	files, err := MakeFilesMap(ctx, []string{`go.mod`, `go.sum`}, `.`)
	if err != nil {
		t.Fatal(err)
	}
	logging("Successfully mapped files: %#v", files)
	os.MkdirAll(`testdata`, os.ModePerm)
	err = ArchiveFiles(ctx, files, `testdata/test.zip`, nil, ArchivalMap[`zip`])
	if err != nil {
		t.Fatal(err)
	}
	err = Unarchive(ctx, `testdata/test.zip`, `testdata/unarchive`)
	if err != nil {
		t.Fatal(err)
	}
}
