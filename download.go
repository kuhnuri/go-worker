package kuhnuri

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

func DownloadFile(input *url.URL, tempDir Path) (Path, error) {
	switch input.Scheme {
	case "s3":
		return downloadFromS3(input, tempDir)
	case "jar":
		jarUri, err := Parse(input.String())
		if err != nil {
			return "", err
		}
		jarFile, err := DownloadFile(&jarUri.Url, tempDir)
		if err != nil {
			return "", err
		}
		err = Unzip(jarFile, tempDir)
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

func downloadFromHttp(in *url.URL, tempDir Path) (Path, error) {
	dst := filepath.Join(tempDir, getName(in))
	fmt.Printf("INFO: Download %s to %s", in, dst)

	resp, err := http.Get(in.String())
	if err != nil {
		return dst, err
	}
	defer resp.Body.Close()

	out, err := Create(dst)
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
	out, err := Create(dst)
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
