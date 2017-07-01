package zip

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

func ExampleArchiveFile() {
	tmpDir, err := ioutil.TempDir("", "test_zip")
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	outFilePath := filepath.Join(tmpDir, "foo.zip")

	progress := func(archivePath string) {
		fmt.Println(archivePath)
	}

	err = ArchiveFile("testdata/foo", outFilePath, progress)
	if err != nil {
		panic(err)
	}

	// Output:
	// foo/bar
	// foo/baz/aaa
}

func ExampleUnarchiveFile() {
	tmpDir, err := ioutil.TempDir("", "test_zip")
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = os.RemoveAll(tmpDir)
	}()

	progress := func(archivePath string) {
		fmt.Println(archivePath)
	}

	err = UnarchiveFile("testdata/foo.zip", tmpDir, progress)
	if err != nil {
		panic(err)
	}

	// Output:
	// foo/bar
	// foo/baz/aaa
}
