package app

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
)

// App is the application we are running.
type App struct {
	S3Client s3iface.S3API
}

func (app *App) getS3Client(region string) s3iface.S3API {
	config := &aws.Config{}

	if region != "" {
		config.Region = aws.String(region)
	}

	if app.S3Client == nil {
		app.S3Client = s3.New(session.Must(session.NewSession(config)))
	}
	return app.S3Client
}

// RunContext contains the context that the build container is run in.
type RunContext struct {
	Bucket        string
	Path          string
	BuildID       string
	Version       string
	MappedCodeDir string
	Params        map[string]interface{}
}

// Run runs the release.
func (app *App) Run(context *RunContext, outputStream, errorStream io.Writer) (map[string]string, error) {
	config, err := getConfig(context.BuildID, context.Params)
	if err != nil {
		return nil, fmt.Errorf("error getting config: %w", err)
	}

	fmt.Fprintf(errorStream, "\ncdflow2-build-lambda: running \n")
	fmt.Fprintf(errorStream, "\ncdflow2-build-lambda: zipping target %q\n\n", config.target)

	tmpfile, err := os.CreateTemp("", "cdflow2-release-lambda-*")

	if err != nil {
		return nil, fmt.Errorf("target_directory '%s' does not exist: %w", config.target, err)
	}

	targetInfo, err := os.Stat(config.target)

	if targetInfo.IsDir() {
		if err := zipDir(tmpfile, config.target); err != nil {
			return nil, fmt.Errorf("error zipping directory: %w", err)
		}
	} else {
		if err := zipFile(tmpfile, config.target); err != nil {
			return nil, fmt.Errorf("error zipping file: %w", err)
		}
	}

	if err := tmpfile.Sync(); err != nil {
		return nil, fmt.Errorf("error syncing write on zipfile: %w", err)
	}

	if _, err := tmpfile.Seek(0, 0); err != nil {
		return nil, fmt.Errorf("error seeking zipfile: %w", err)
	}

	bucket := context.Bucket
	key := context.Path

	if config.region != "" {
		bucket = fmt.Sprintf("%s-%s", bucket, config.region)
	}

	fmt.Fprintf(os.Stderr, "\ncdflow2-build-lambda: uploading zip to s3://%s/%s...", bucket, key)

	s3client := app.getS3Client(config.region)
	if _, err := s3client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   tmpfile,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "\n\n")
		return nil, fmt.Errorf("error uploading to s3: %w", err)
	}

	fmt.Fprintf(os.Stderr, "\ndone.\n\n")

	return map[string]string{
		"bucket": bucket,
		"key":    key,
	}, nil
}

type config struct {
	target string
	region string
}

func getConfig(buildID string, params map[string]interface{}) (*config, error) {
	result := config{
		target: "./target",
	}

	targetDirectoryI, ok := params["target_directory"]
	if ok {
		result.target, ok = targetDirectoryI.(string)
		if !ok {
			return nil, fmt.Errorf("unexpected type for build.%v.params.target_directory: %T (should be string)", buildID, targetDirectoryI)
		}
	}

	regionI, ok := params["region"]
	if ok {
		result.region, ok = regionI.(string)
		if !ok {
			return nil, fmt.Errorf("unexpected type for build.%v.params.region: %T (should be string)", buildID, regionI)
		}
	}

	return &result, nil
}

func zipFile(writer io.Writer, file string) error {
	zipWriter := zip.NewWriter(writer)

	info, err := os.Lstat(file)
	if err != nil {
		return err
	}

	// 3. Create a local file header
	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	// set compression
	header.Method = zip.Deflate

	headerWriter, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	reader, err := os.Open(file)
	if err != nil {
		return err
	}
	defer reader.Close()
	_, err = io.Copy(headerWriter, reader)
	if err != nil {
		return err
	}
	return zipWriter.Close()
}

func zipDir(writer io.Writer, dir string) error {
	zipWriter := zip.NewWriter(writer)
	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 3. Create a local file header
		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		// set compression
		header.Method = zip.Deflate

		// 4. Set relative path of a file as the header name
		header.Name, err = filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		if info.IsDir() {
			header.Name += "/"
		}

		// 5. Create writer for the file header and save content of the file
		headerWriter, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		reader, err := os.Open(path)
		if err != nil {
			return err
		}
		defer reader.Close()

		_, err = io.Copy(headerWriter, reader)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return zipWriter.Close()
}
