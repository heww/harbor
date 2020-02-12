// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dao

import (
	"context"
	"time"

	"github.com/goharbor/harbor/src/internal/orm"
	"github.com/goharbor/harbor/src/pkg/q"
)

// CreateBlob create blob ignore conflict on digest
func CreateBlob(ctx context.Context, blob *Blob) (int64, error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}

	blob.CreationTime = time.Now()

	// ignore conflict on digest
	return o.InsertOrUpdate(blob, "digest")
}

// GetBlobByDigest returns blob by digest
func GetBlobByDigest(ctx context.Context, digest string) (*Blob, error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	blob := &Blob{Digest: digest}
	if err = o.Read(blob, "digest"); err != nil {
		return nil, orm.WrapNotFoundError(err, "blob %s not found", digest)
	}

	return blob, nil
}

// UpdateBlob update blob
func UpdateBlob(ctx context.Context, blob *Blob) error {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return err
	}

	_, err = o.Update(blob)
	return err
}

// ListBlobs list blobs by query
func ListBlobs(ctx context.Context, query *q.Query) ([]*Blob, error) {
	qs, err := orm.QuerySetter(ctx, &Blob{}, query)
	if err != nil {
		return nil, err
	}

	blobs := []*Blob{}
	if _, err = qs.All(&blobs); err != nil {
		return nil, err
	}
	return blobs, nil
}
