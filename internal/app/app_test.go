package app_test

import (
	"archive/zip"
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/mergermarket/cdflow2-build-lambda/internal/app"
)

// Copy source to destination.
func Copy(source, destination string) error {
	in, err := os.Open(source)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(destination)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}
	return out.Close()
}

type mockedS3 struct {
	s3iface.S3API
	contents map[string]string
	uploaded bytes.Buffer
}

func (m *mockedS3) PutObject(input *s3.PutObjectInput) (*s3.PutObjectOutput, error) {
	data, err := ioutil.ReadAll(input.Body)
	if err != nil {
		return nil, err
	}
	zipReader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}
	m.contents = make(map[string]string)
	for _, f := range zipReader.File {
		if f.FileInfo().IsDir() {
			continue
		}
		r, err := f.Open()
		if err != nil {
			return nil, err
		}
		d, err := ioutil.ReadAll(r)
		if err != nil {
			return nil, err
		}
		m.contents[f.Name] = string(d)
	}
	return &s3.PutObjectOutput{}, nil
}

func TestRun(t *testing.T) {
	// Given
	manifestConfig := map[string]interface{}{
		
	}
	s3Client := &mockedS3{}
	application := &app.App{
		S3Client: s3Client,
	}
	var outputBuffer bytes.Buffer
	var errorBuffer bytes.Buffer

	bucket := "test-bucket"
	path := "foo/bar"
	version := "2"

	// When
	metadata, err := application.Run(&app.RunContext{
		Bucket:        bucket,
		BuildID:       "lambda",
		Version:       version,
		Path:          path,
		Params:        manifestConfig,
	}, &outputBuffer, &errorBuffer)
	if err != nil {
		t.Fatalf("error in Run: %s\n  output: %q", err, errorBuffer.String())
	}

	// Then
	if metadata["bucket"] != bucket {
		t.Fatalf("expected %s, got %s", bucket, metadata["bucket"])
	}
	if metadata["key"] != path {
		t.Fatalf("expected %s, got %s", path, metadata["key"])
	}
	expected := map[string]string{"app": "default test content"}
	if !reflect.DeepEqual(s3Client.contents, expected) {
		t.Fatalf("got %#v, expected %#v", s3Client.contents, expected)
	}
}

func TestRunWithTestingDirectory(t *testing.T) {
	// Given
	manifestConfig := map[string]interface{}{
		"target_directory":  "../../test/target",
	}
	s3Client := &mockedS3{}
	application := &app.App{
		S3Client: s3Client,
	}
	var outputBuffer bytes.Buffer
	var errorBuffer bytes.Buffer

	bucket := "test-bucket"
	path := "foo/bar"
	version := "2"

	// When
	metadata, err := application.Run(&app.RunContext{
		Bucket:        bucket,
		BuildID:       "lambda",
		Version:       version,
		Path:          path,
		Params:        manifestConfig,
	}, &outputBuffer, &errorBuffer)
	if err != nil {
		t.Fatalf("error in Run: %s\n  output: %q", err, errorBuffer.String())
	}

	// Then
	if metadata["bucket"] != bucket {
		t.Fatalf("expected %s, got %s", bucket, metadata["bucket"])
	}
	if metadata["key"] != path {
		t.Fatalf("expected %s, got %s", path, metadata["key"])
	}
	expected := map[string]string{"app": "test content"}
	if !reflect.DeepEqual(s3Client.contents, expected) {
		t.Fatalf("got %#v, expected %#v", s3Client.contents, expected)
	}
}