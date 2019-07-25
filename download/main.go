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
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type Path = string

//func run(input *url.URL, tempDir Path) error {
//	switch input.Scheme {
//	case "http":
//		fallthrough
//	case "https":
//	case "s3":
//		fallthrough
//	case "jar":
//		path, err := downloadFile(input, tempDir)
//		if err != nil {
//			return err
//		}
//		res := url.URL{Scheme: "file"}
//		res.Path = path
//		return nil
//	default:
//		return nil
//	}
//}

func downloadFile(input *url.URL, tempDir Path) (Path, error) {
	switch input.Scheme {
	case "s3":
		return downloadFromS3(input, tempDir)
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
	case "file":
		return filepath.FromSlash(input.Path), nil
	default:
		errors.New(fmt.Sprintf("Unsupported scheme %s", input.Scheme))
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

	resp, err := http.Get(in.String())
	if err != nil {
		return dst, err
	}
	defer resp.Body.Close()

	out, err := os.Create(dst)
	if err != nil {
		return dst, err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return dst, err
}

func downloadFromS3(in *url.URL, tempDir Path) (Path, error) {
	dst := filepath.Join(tempDir, getName(in))
	fmt.Printf("INFO: Download %s to %s", in, dst)
	out, err := os.Create(dst)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Failed to create destination %s for S3 download", dst))
	}
	defer out.Close()

	sess := session.Must(session.NewSession())
	downloader := s3manager.NewDownloader(sess)
	_, err = downloader.Download(out, &s3.GetObjectInput{
		Bucket: aws.String(in.Host),
		Key:    aws.String(in.Path),
	})
	if err != nil {
		return "", errors.New(fmt.Sprintf("Failed to download %s", in.String()))
	}

	return dst, nil
}

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

	if _, err := downloadFile(src, tmp); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
