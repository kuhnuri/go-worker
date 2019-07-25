package main

import (
	"archive/zip"
	"errors"
	"fmt"
	"github.com/kuhnuri/jar"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

type Path = string

func run(src *url.URL, tmp Path) error {
	_, err := download(src, tmp)
	if err != nil {
		return err
	}
	return nil
}

func download(input *url.URL, tmp Path) (*url.URL, error) {
	switch input.Scheme {
	case "s3":
		fallthrough
	case "jar":
		path, err := downloadFile(input, tmp)
		if err != nil {
			return nil, err
		}
		res := url.URL{Scheme: "file"}
		res.Path = path
		return &res, nil
	default:
		return input, nil
	}
}

func downloadFile(input *url.URL, tempDir Path) (Path, error) {
	switch input.Scheme {
	case "s3":
		// FIXME
		return "", errors.New("S3 not supported")
		//return downloadFromS3(input, tempDir);
	case "jar":
		jarUri, err := jar.Parse(input.String())
		if err != nil {
			return "", err
		}
		jarFile, err := downloadFile(&jarUri.Url, tempDir)
		if err != nil {
			return "", err
		}
		err = unzip(jarFile, tempDir)
		if err != nil {
			return "", err
		}
		os.Remove(jarFile)

		if jarUri.Entry != "" {
			return filepath.Clean(filepath.Join(tempDir, jarUri.Entry)), nil
		} else {
			return tempDir, nil
		}
	case "http":
		fallthrough
	case "https":
		return downloadFromHttp(input, tempDir)
	default:
		log.Panic(errors.New(fmt.Sprintf("Unsupported scheme %s", input.Scheme)))
	}
	return "", nil
}

func unzip(zipFile Path, tempDir Path) error {
	fmt.Printf("INFO: Unzip %s to %s\n", zipFile, tempDir)
	r, err := zip.OpenReader(zipFile)
	if err != nil {
		return err
		//log.Fatal(err)
	}
	defer r.Close()

	for _, f := range r.File {
		file := filepath.Clean(filepath.ToSlash(filepath.Join(tempDir, f.Name)))
		fmt.Printf("DEBUG: Copy %s to %s\n", f.Name, file)

		rc, err := f.Open()
		if err != nil {
			return err
			//log.Fatal(err)
		} else {
			wc, err := os.Create(file)
			if err != nil {
				return err
				//log.Fatal(err)
			} else {
				_, err = io.Copy(wc, rc)
				if err != nil {
					return err
					//log.Fatal(err)
				}
			}
			wc.Close()
		}
		rc.Close()
	}
	return nil
}

func downloadFromHttp(in *url.URL, tempDir Path) (Path, error) {
	dst := filepath.Join(tempDir, getName(in))
	fmt.Printf("INFO: Download %s to %s", in, dst)

	// Get the data
	resp, err := http.Get(in.String())
	if err != nil {
		return dst, err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(dst)
	if err != nil {
		return dst, err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return dst, err
}

//private Path downloadFromS3(in url.URL, tempDir Path) throws InterruptedException {
//s3Uri AmazonS3url.URL = new AmazonS3url.URL(in);
//var fileName String = getName(in);
//var file Path = tempDir.resolve(fileName);
//System.out.println(String.format("INFO: Download %s to %s", in, file));
//var transfer Transfer = transferManager.download(s3Uri.getBucket(), s3Uri.getKey(), file.toFile());
//transfer.waitForCompletion();
//return file;
//}
//
func getName(in *url.URL) string {
	return filepath.Base(filepath.FromSlash(in.Path))
}

func main() {
	if len(os.Args) != 3 {
		log.Fatal("Usage: dowload IN-URL OUT-PATH")
		os.Exit(1)
	}
	src, err := url.Parse(os.Args[1])
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	tmp, err := filepath.Abs(filepath.Clean(os.Args[2]))
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	if err := run(src, tmp); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
