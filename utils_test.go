package kuhnuri

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
)

func setup() string {
	base, err := ioutil.TempDir("", "")
	if err != nil {
		log.Fatalf("Failed to create temporary directory\n")
	}
	src := createDir(filepath.Join(base, "src"))
	createDir(filepath.Join(base, "out"))
	files := []string{
		"foo.txt",
		"bar/baz.txt",
	}
	for _, file := range files {
		abs := filepath.Join(src, filepath.ToSlash(file))
		f, err := Create(abs)
		if err != nil {
			log.Fatalf("Failed to create temporary file %s\n", abs)
		}
		_, err = f.WriteString(file)
		if err != nil {
			log.Fatalf("Failed to write temporary file %s\n", abs)
		}
		f.Close()
	}
	return base
}

func createDir(dir string) string {
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			log.Fatalf("Failed to create temporary directory %s\n", dir)
		}
	}
	return dir
}

func teardown(dir string) {
	fmt.Printf("Shutdown %s\n", dir)
	err := os.RemoveAll(dir)
	if err != nil {
		fmt.Errorf("Failed to clean %s\n", err)
	}
}

func TestZip(t *testing.T) {
	base := setup()

	src := filepath.Join(base, "src")
	out := filepath.Join(base, "out")
	zip := filepath.Join(out, "tmp.zip")

	err := Zip(zip, src)
	if err != nil {
		t.Fatalf("Failed to zip: %s", err)
	}
	err = Unzip(zip, out)
	if err != nil {
		t.Fatalf("Failed to unzip: %s", err)
	}

	files := []string{
		"foo.txt",
		"bar/baz.txt",
	}
	for _, file := range files {
		abs := filepath.Join(out, filepath.ToSlash(file))
		fmt.Printf("Check %s\n", abs)
		_, err := os.Stat(abs)

		if os.IsNotExist(err) {
			t.Fatalf("File not found %s", abs)
		}
	}
	teardown(base)
}
