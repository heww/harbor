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
)

// CreateArtifactAndBlob create ArtifactAndBlob by artifact digest and blob digest
func CreateArtifactAndBlob(ctx context.Context, md *ArtifactAndBlob) (int64, error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}

	md.CreationTime = time.Now()

	return o.InsertOrUpdate(md, "digest_af, digest_blob")
}

// GetArtifactAndBlob get ArtifactAndBlob by artifact digest and blob digest
func GetArtifactAndBlob(ctx context.Context, artifactDigest, blobDigest string) (*ArtifactAndBlob, error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return nil, err
	}

	md := &ArtifactAndBlob{
		DigestAF:   artifactDigest,
		DigestBlob: blobDigest,
	}

	if err := o.Read(md, "digest_af", "digest_blob"); err != nil {
		return nil, orm.WrapNotFoundError(err, "")
	}

	return md, nil
}
