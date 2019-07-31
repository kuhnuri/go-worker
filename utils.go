package kuhnuri

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
)

type Path = string

func getName(in *url.URL) string {
	return filepath.Base(filepath.FromSlash(in.Path))
}

func WithExt(path Path, ext string) Path {
	from := filepath.Ext(path)
	return path[0:len(path)-len(from)] + ext
}

func Unzip(zipFile Path, tempDir Path) error {
	fmt.Printf("INFO: Unzip %s to %s\n", zipFile, tempDir)
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return fmt.Errorf("Failed to open ZIP file reader: %v", err)
	}
	defer r.Close()

	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			continue
		}
		file := filepath.Clean(filepath.ToSlash(filepath.Join(tempDir, f.Name)))
		fmt.Printf("DEBUG: Copy %s to %s\n", f.Name, file)

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("Failed to open ZIP entry: %v", err)
		} else {
			wc, err := Create(file)
			if err != nil {
				return fmt.Errorf("Failed to open ZIP extract writer: %v", err)
			} else {
				_, err = io.Copy(wc, rc)
				if err != nil {
					return fmt.Errorf("Failed to copy ZIP entry to outputy: %v", err)
				}
			}
			wc.Close()
		}
		rc.Close()
	}
	return nil
}

func Create(file string) (*os.File, error) {
	dir := filepath.Dir(file)
	_, err := os.Stat(dir)
	if os.IsNotExist(err) {
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			fmt.Errorf("Failed to create directory %s: %v", dir, err)
		}
	}
	return os.Create(file)
}

func Zip(zipFile Path, tempDir Path) error {
	fmt.Printf("INFO: Zip %s to %s\n", tempDir, zipFile)
	out, err := Create(zipFile)
	if err != nil {
		return err
	}
	defer out.Close()

	w := zip.NewWriter(out)
	defer w.Close()

	filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		fi, err := os.Stat(path)
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(tempDir, path)
		if err != nil {
			log.Fatalf("ERROR: Failed to relativize %s\n", path)
		}
		fmt.Printf("DEBUG: Read %s\n", path)
		src, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("Failed to open file reader: %v", err)
		} else {
			dst, err := w.Create(filepath.ToSlash(rel))
			if err != nil {
				return fmt.Errorf("Failed to open ZIP entry writer: %v", err)
			} else {
				_, err = io.Copy(dst, src)
				if err != nil {
					return fmt.Errorf("Failed to copy file to ZIP entry: %v", err)
				}
			}
			src.Close()
		}
		return nil
	})

	return nil
}
