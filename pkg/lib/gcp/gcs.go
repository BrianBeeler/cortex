/*
Copyright 2020 Cortex Labs, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gcp

import (
	"bytes"
	"context"
	"encoding/json"
	"path/filepath"

	"cloud.google.com/go/storage"
	"github.com/cortexlabs/cortex/pkg/lib/errors"
	s "github.com/cortexlabs/cortex/pkg/lib/strings"
	"google.golang.org/api/iterator"
)

func GCSPath(bucket string, key string) string {
	return "gs://" + filepath.Join(bucket, key)
}

func (c *Client) ReadJSONFromGCP(objPtr interface{}, bucket string, key string) error {
	jsonBytes, err := c.ReadBytesFromGCP(bucket, key)
	if err != nil {
		return err
	}
	return errors.Wrap(json.Unmarshal(jsonBytes, objPtr), GCSPath(bucket, key))
}

func (c *Client) UploadJSONToGCP(obj interface{}, bucket string, key string) error {
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	return c.UploadBytesToGCP(jsonBytes, bucket, key)
}

func (c *Client) DeleteGCSFile(bucket string, key string) error {
	gcsClient, err := c.GCS()
	if err != nil {
		return err
	}
	if err := gcsClient.Bucket(bucket).Object(key).Delete(context.Background()); err != nil {
		return err
	}
	return nil
}

func (c *Client) DeleteGCSDir(bucket string, gcsDir string, continueIfFailure bool) error {
	prefix := s.EnsureSuffix(gcsDir, "/")
	return c.DeleteGCSPrefix(bucket, prefix, continueIfFailure)
}

func (c *Client) DeleteGCSPrefix(bucket string, gcsDir string, continueIfFailure bool) error {
	gcsClient, err := c.GCS()
	if err != nil {
		return err
	}

	bkt := gcsClient.Bucket(bucket)
	objectIterator := bkt.Objects(context.Background(), &storage.Query{
		Prefix: gcsDir,
	})

	var attrs *storage.ObjectAttrs
	for {
		attrs, err = objectIterator.Next()
		if err == iterator.Done {
			break
		}
		if err != nil && !continueIfFailure {
			break
		}
		err = bkt.Object(attrs.Name).Delete(context.Background())
		if err != nil && !continueIfFailure {
			break
		}
	}

	if err != nil && err != iterator.Done && !continueIfFailure {
		return err
	}
	return nil
}

func (c *Client) UploadBytesToGCP(data []byte, bucket string, key string) error {
	gcsClient, err := c.GCS()
	if err != nil {
		return err
	}
	objectWriter := gcsClient.Bucket(bucket).Object(key).NewWriter(context.Background())
	defer objectWriter.Close()
	if _, err := objectWriter.Write(data); err != nil {
		return err
	}
	return nil
}

func (c *Client) ReadBytesFromGCP(bucket string, key string) ([]byte, error) {
	gcsClient, err := c.GCS()
	if err != nil {
		return nil, err
	}
	objectReader, err := gcsClient.Bucket(bucket).Object(key).NewReader(context.Background())
	if err != nil {
		return nil, err
	}
	defer objectReader.Close()
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(objectReader); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
