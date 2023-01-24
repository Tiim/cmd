package main

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"storj.io/uplink"
)

func SaveMediaFile(file io.Reader, mime *mimetype.MIME, accessGrant, bucketName, prefix, formatUrl string) (string, error) {

	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	access, err := uplink.ParseAccess(accessGrant)
	if err != nil {
		return "", fmt.Errorf("could not request access grant: %v", err)
	}

	// Open up the Project we will be working with.
	project, err := uplink.OpenProject(context.Background(), access)
	if err != nil {
		return "", fmt.Errorf("could not open project: %v", err)
	}
	defer project.Close()

	// Ensure the desired Bucket within the Project is created.
	_, err = project.EnsureBucket(context.Background(), bucketName)
	if err != nil {
		return "", fmt.Errorf("could not ensure bucket: %v", err)
	}

	// Intitiate the upload of our Object to the specified bucket and key.
	key, name := uploadKey(prefix, mime)
	upload, err := project.UploadObject(context.Background(), bucketName, key, nil)
	if err != nil {
		return "", fmt.Errorf("could not initiate upload: %v", err)
	}

	// Copy the data to the upload.
	_, err = io.Copy(upload, file)
	if err != nil {
		_ = upload.Abort()
		return "", fmt.Errorf("could not upload data: %v", err)
	}
	// Commit the uploaded object.
	err = upload.Commit()
	if err != nil {
		return "", fmt.Errorf("could not commit uploaded object: %v", err)
	}
	url := strings.ReplaceAll(formatUrl, "{{filename}}", name)
	url = strings.ReplaceAll(url, "{{prefix}}", prefix)
	url = strings.ReplaceAll(url, "{{bucket}}", bucketName)

	return url, nil
}

func uploadKey(prefix string, mimeType *mimetype.MIME) (string, string) {
	key := prefix
	extension := mimeType.Extension()
	id := uuid.New()
	name := id.String() + extension
	key += name

	return key, name
}
