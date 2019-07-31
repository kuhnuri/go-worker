package kuhnuri

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

func UploadFile(outDirOrFile Path, output *url.URL) (error) {
	switch output.Scheme {
	case "s3":
		return uploadFromS3(outDirOrFile, output)
	case "jar":
		jarUri, err := Parse(output.String())
		if err != nil {
			return err
		}
		jarFile, err := ioutil.TempFile("", "*.zip")
		if err != nil {
			return err
		}
		defer os.Remove(jarFile.Name())

		err = Zip(jarFile.Name(), outDirOrFile)
		if err != nil {
			return err
		}

		err = UploadFile(jarFile.Name(), &jarUri.Url)
		if err != nil {
			return err
		}

		return nil
	case "http":
		fallthrough
	case "https":
		return uploadFromHttp(outDirOrFile, output)
	case "file":
		return nil
	default:
		fmt.Errorf("Unsupported scheme %s", output.Scheme)
	}
	return nil
}

func uploadFromHttp(tempFile Path, output *url.URL) error {
	fmt.Printf("INFO: Upload %s to %s\n", tempFile, output)

	src, err := os.Open(tempFile)
	if err != nil {
		return err
	}
	defer src.Close()

	resp, err := http.Post(output.String(), "application/binary", src)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// TODO upload directory recursively
func uploadFromS3(tempFile Path, output *url.URL) error {
	fmt.Printf("INFO: Upload %s to %s\n", tempFile, output)

	src, err := os.Open(tempFile)
	if err != nil {
		return err
	}
	defer src.Close()

	sess := session.Must(session.NewSession())
	uploader := s3manager.NewUploader(sess)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(output.Host),
		Key:    aws.String(output.Path),
		Body:   src,
	})
	if err != nil {
		return fmt.Errorf("Failed to upload %s: %v", output.String(), err)
	}

	return nil
}
