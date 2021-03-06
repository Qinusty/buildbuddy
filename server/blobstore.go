package blobstore

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"cloud.google.com/go/storage"
	"github.com/buildbuddy-io/buildbuddy/server/config"
	"github.com/buildbuddy-io/buildbuddy/server/interfaces"
	"google.golang.org/api/option"
)

// Returns whatever blobstore is specified in the config.
func GetConfiguredBlobstore(c *config.Configurator) (interfaces.Blobstore, error) {
	if c.GetStorageDiskRootDir() != "" {
		return NewDiskBlobStore(c.GetStorageDiskRootDir()), nil
	}
	if gcsConfig := c.GetStorageGCSConfig(); gcsConfig != nil {
		opts := make([]option.ClientOption, 0)
		if gcsConfig.CredentialsFile != "" {
			opts = append(opts, option.WithCredentialsFile(gcsConfig.CredentialsFile))
		}
		return NewGCSBlobStore(gcsConfig.Bucket, gcsConfig.ProjectID, opts...)
	}
	return nil, fmt.Errorf("No storage backend configured -- please specify at least one in the config")
}

func ensureDirectoryExists(dir string) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return os.MkdirAll(dir, 0755)
	}
	return nil
}

// A Disk-based blob storage implementation that reads and writes blobs to/from
// files.
type DiskBlobStore struct {
	rootDir string
}

func NewDiskBlobStore(rootDir string) *DiskBlobStore {
	return &DiskBlobStore{
		rootDir: rootDir,
	}
}

func decompress(in []byte, err error) ([]byte, error) {
	if err != nil {
		return in, err
	}

	var buf bytes.Buffer
	// Write instead of using NewBuffer because if this is not a gzip file
	// we want to return "in" directly later, and NewBuffer takes ownership
	// of it.
	if _, err := buf.Write(in); err != nil {
		return nil, err
	}
	zr, err := gzip.NewReader(&buf)
	if err == gzip.ErrHeader {
		// Compatibility hack: if we got a header error it means this
		// is probably an uncompressed record written before we were
		// compressing. Just read it as-is.
		return in, nil
	}
	if err != nil {
		log.Printf("zr err: %s", err)
		return nil, err
	}
	rsp, err := ioutil.ReadAll(zr)
	if err != nil {
		log.Printf("readall err: %s", err)
		return nil, err
	}
	if err := zr.Close(); err != nil {
		log.Printf("close err: %s", err)
		return nil, err
	}
	return rsp, nil
}

func compress(in []byte) ([]byte, error) {
	var buf bytes.Buffer
	zr := gzip.NewWriter(&buf)
	if _, err := zr.Write(in); err != nil {
		return nil, err
	}
	if err := zr.Close(); err != nil {
		return nil, err
	}
	return ioutil.ReadAll(&buf)
}

func (d *DiskBlobStore) WriteBlob(ctx context.Context, blobName string, data []byte) error {
	// Probably could be more careful here but we are generating these ourselves
	// for now.
	if strings.Contains(blobName, "..") {
		return fmt.Errorf("blobName (%s) must not contain ../", blobName)
	}
	fullPath := filepath.Join(d.rootDir, blobName)
	if err := ensureDirectoryExists(filepath.Dir(fullPath)); err != nil {
		return err
	}
	tmpFileName := fullPath + ".tmp"

	compressedData, err := compress(data)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(tmpFileName, compressedData, 0644); err != nil {
		return err
	}
	return os.Rename(tmpFileName, fullPath)
}

func (d *DiskBlobStore) ReadBlob(ctx context.Context, blobName string) ([]byte, error) {
	return decompress(ioutil.ReadFile(filepath.Join(d.rootDir, blobName)))
}

func (d *DiskBlobStore) DeleteBlob(ctx context.Context, blobName string) error {
	return os.Remove(filepath.Join(d.rootDir, blobName))
}

// GCSBlobStore implements the blobstore API on top of the google cloud storage API.
type GCSBlobStore struct {
	gcsClient    *storage.Client
	bucketHandle *storage.BucketHandle
	projectID    string
}

func NewGCSBlobStore(bucketName, projectID string, opts ...option.ClientOption) (*GCSBlobStore, error) {
	ctx := context.Background()
	gcsClient, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, err
	}
	g := &GCSBlobStore{
		gcsClient: gcsClient,
		projectID: projectID,
	}
	err = g.createBucketIfNotExists(ctx, bucketName)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (g *GCSBlobStore) createBucketIfNotExists(ctx context.Context, bucketName string) error {
	if _, err := g.gcsClient.Bucket(bucketName).Attrs(ctx); err != nil {
		log.Printf("Creating storage bucket: %s", bucketName)
		g.bucketHandle = g.gcsClient.Bucket(bucketName)
		return g.bucketHandle.Create(ctx, g.projectID, nil)
	}
	g.bucketHandle = g.gcsClient.Bucket(bucketName)
	return nil
}

func (g *GCSBlobStore) ReadBlob(ctx context.Context, blobName string) ([]byte, error) {
	reader, err := g.bucketHandle.Object(blobName).NewReader(ctx)
	if err != nil {
		return nil, err
	}
	return decompress(ioutil.ReadAll(reader))
}

func (g *GCSBlobStore) WriteBlob(ctx context.Context, blobName string, data []byte) error {
	writer := g.bucketHandle.Object(blobName).NewWriter(ctx)
	defer writer.Close()
	compressedData, err := compress(data)
	if err != nil {
		return err
	}
	_, err = writer.Write(compressedData)
	return err
}

func (g *GCSBlobStore) DeleteBlob(ctx context.Context, blobName string) error {
	return g.bucketHandle.Object(blobName).Delete(ctx)
}
