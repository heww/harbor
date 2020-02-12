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

package blob

import (
	"context"

	ierror "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/pkg/blob/dao"
	"github.com/goharbor/harbor/src/pkg/q"
)

var (
	// Mgr default blob manager
	Mgr = NewManager()
)

// Manager interface provide the management functions for blobs
type Manager interface {
	// AttachToArtifact attach blob to artifact
	AttachToArtifact(ctx context.Context, blobDigest, artifactDigest string) (int64, error)

	// AttachToProject attach blob to project
	AttachToProject(ctx context.Context, blobID, projectID int64) (int64, error)

	// Create create blob
	Create(ctx context.Context, digest string, contentType string, size int64) (int64, error)

	// Get get blob by digest
	Get(ctx context.Context, digest string) (*Blob, error)

	// Update the blob
	Update(ctx context.Context, blob *Blob) error

	// List returns blobs by params
	List(ctx context.Context, params ListParams) ([]*Blob, error)

	// IsAttachedToArtifact returns true when blob attached to artifact
	IsAttachedToArtifact(ctx context.Context, blobDigest string, artifactDigest string) (bool, error)

	// IsAttachedToProject returns true when blob attached to project
	IsAttachedToProject(ctx context.Context, digest string, projectID int64) (bool, error)
}

type manager struct{}

func (m *manager) toBlob(blob *dao.Blob) *Blob {
	return &Blob{
		ID:          blob.ID,
		Digest:      blob.Digest,
		ContentType: blob.ContentType,
		Size:        blob.Size,
	}
}

func (m *manager) AttachToArtifact(ctx context.Context, blobDigest, artifactDigest string) (int64, error) {
	return dao.CreateArtifactAndBlob(ctx, &dao.ArtifactAndBlob{DigestAF: artifactDigest, DigestBlob: blobDigest})
}

func (m *manager) AttachToProject(ctx context.Context, blobID, projectID int64) (int64, error) {
	return dao.CreateProjectBlob(ctx, &dao.ProjectBlob{BlobID: blobID, ProjectID: projectID})
}

func (m *manager) Create(ctx context.Context, digest string, contentType string, size int64) (int64, error) {
	return dao.CreateBlob(ctx, &dao.Blob{Digest: digest, ContentType: contentType, Size: size})
}

func (m *manager) Get(ctx context.Context, digest string) (*Blob, error) {
	blob, err := dao.GetBlobByDigest(ctx, digest)
	if err != nil {
		return nil, err
	}

	return m.toBlob(blob), nil
}

func (m *manager) Update(ctx context.Context, blob *Blob) error {
	return dao.UpdateBlob(ctx, &dao.Blob{
		ID:           blob.ID,
		Digest:       blob.Digest,
		ContentType:  blob.ContentType,
		Size:         blob.Size,
		CreationTime: blob.CreationTime,
	})
}

func (m *manager) List(ctx context.Context, params ListParams) ([]*Blob, error) {
	kw := q.KeyWords{}
	if len(params.Digests) > 0 {
		kw["digest__in"] = params.Digests
	}

	blobs, err := dao.ListBlobs(ctx, q.New(kw))
	if err != nil {
		return nil, err
	}

	var results []*Blob
	for _, blob := range blobs {
		results = append(results, m.toBlob(blob))
	}

	return results, nil
}

func (m *manager) IsAttachedToArtifact(ctx context.Context, blobDigest string, manifestDigest string) (bool, error) {
	md, err := dao.GetArtifactAndBlob(ctx, manifestDigest, blobDigest)
	if err != nil && !ierror.IsNotFoundErr(err) {
		return false, err
	}

	return md != nil, nil
}

func (m *manager) IsAttachedToProject(ctx context.Context, digest string, projectID int64) (bool, error) {
	return dao.ExistProjectBlob(ctx, digest, projectID)
}

// NewManager returns blob manager
func NewManager() Manager {
	return &manager{}
}
