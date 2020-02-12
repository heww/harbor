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
	"fmt"

	"github.com/garyburd/redigo/redis"
	"github.com/goharbor/harbor/src/common/utils/log"
	util "github.com/goharbor/harbor/src/common/utils/redis"
	ierror "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/internal/orm"
	"github.com/goharbor/harbor/src/pkg/blob"
)

var (
	// Ctl is a global blob controller instance
	Ctl = NewController()
)

// Controller defines the operations related with blobs
type Controller interface {
	// AttachToArtifact attach blobs to manifest.
	AttachToArtifact(ctx context.Context, params AttachToArtifactParams) error

	// AttachToProject attach blob to project,
	// `BlobID` in params will used to attach to the project when it provided,
	// otherwise the blob query by digest will be attached.
	AttachToProject(ctx context.Context, params AttachToProjectParams) error

	// Ensure create blob when it not exist.
	Ensure(ctx context.Context, params EnsureParams) (int64, error)

	// Exist check blob exist by params,
	// returns true when blob exist if only `Digest` provided in params,
	// returns true when blob exist and attached to the artifact if only `Digest` and `ArtifactDigest` provided in params,
	// returns true when blob exist and attached to the project if only `Digest` and `ProjectID` provided in params,
	// returns true when blob exist and attached to both artifact and project if `Digest`, `ArtifactDigest` and `ProjectID` provided in params.
	Exist(ctx context.Context, params ExistParams) (bool, error)

	// FindNotAttachedToProject returns blobs which not attached to project.
	// FindNotAttachedToProject(ctx context.Context)

	// Get get the blob by params,
	// returns blob if only `Digest` provided in params,
	// returns blob when blob attached to the artifact if only `Digest` and `ArtifactDigest` provided in params,
	// returns blob when blob attached to the project if only `Digest` and `ProjectID` provided in params,
	// returns true when blob attached to both artifact and project if `Digest`, `ArtifactDigest` and `ProjectID` provided in params.
	Get(ctx context.Context, params GetParams) (*blob.Blob, error)

	// Sync create blobs from `References` when they are not exist
	// and update the blob content type when they are exist,
	Sync(ctx context.Context, params SyncParams) error

	// SetAcceptedBlobSize update the accepted size of stream upload blob.
	SetAcceptedBlobSize(sessionID string, size int64) error

	// GetAcceptedBlobSize returns the accepted size of stream upload blob.
	GetAcceptedBlobSize(sessionID string) (int64, error)
}

// NewController creates an instance of the default repository controller
func NewController() Controller {
	return &controller{
		blobMgr:   blob.Mgr,
		logPrefix: "[controller][blob]",
	}
}

type controller struct {
	blobMgr   blob.Manager
	logPrefix string
}

func (c *controller) AttachToArtifact(ctx context.Context, params AttachToArtifactParams) error {
	// TODO: add `Exist` method to `artifact.Manager` to check artifact exist
	exist, err := c.blobMgr.IsAttachedToArtifact(ctx, params.ArtifactDigest, params.ArtifactDigest)
	if err != nil {
		return err
	}

	if exist {
		log.Infof("%s: artifact digest %s already exist, skip to attach blobs to the artifact", c.logPrefix, params.ArtifactDigest)
		return nil
	}

	for _, blobDigest := range params.BlobDigests {
		_, err := c.blobMgr.AttachToArtifact(ctx, blobDigest, params.ArtifactDigest)
		if err != nil {
			return err
		}
	}

	// process manifest as blob
	_, err = c.blobMgr.AttachToArtifact(ctx, params.ArtifactDigest, params.ArtifactDigest)
	return err
}

func (c *controller) AttachToProject(ctx context.Context, params AttachToProjectParams) error {
	if params.BlobID != 0 && params.Digest != "" {
		return ierror.BadRequestError(nil).WithMessage("require BlobID or Digest")
	}

	blobID := params.BlobID

	if params.Digest != "" {
		blob, err := c.blobMgr.Get(ctx, params.Digest)
		if err != nil {
			return err
		}

		blobID = blob.ID
	}

	_, err := c.blobMgr.AttachToProject(ctx, blobID, params.ProjectID)
	return err
}

func (c *controller) Get(ctx context.Context, params GetParams) (*blob.Blob, error) {
	if params.Digest == "" {
		return nil, ierror.New(nil).WithCode(ierror.BadRequestCode).WithMessage("require Digest")
	}

	blob, err := c.blobMgr.Get(ctx, params.Digest)
	if err != nil {
		return nil, err
	}

	if params.ProjectID != 0 {
		exist, err := c.blobMgr.IsAttachedToProject(ctx, params.Digest, params.ProjectID)
		if err != nil {
			return nil, err
		}

		if !exist {
			return nil, ierror.NotFoundError(nil).WithMessage("blob %s not found in project %d", params.Digest, params.ProjectID)
		}
	}

	if params.ArtifactDigest != "" {
		exist, err := c.blobMgr.IsAttachedToArtifact(ctx, params.Digest, params.ArtifactDigest)
		if err != nil {
			return nil, err
		}

		if !exist {
			return nil, ierror.NotFoundError(nil).WithMessage("blob %s not found in artifact %s", params.Digest, params.ArtifactDigest)
		}
	}

	return blob, nil
}

func (c *controller) Ensure(ctx context.Context, params EnsureParams) (blobID int64, err error) {
	blob, err := c.blobMgr.Get(ctx, params.Digest)
	if err == nil {
		return blob.ID, nil
	}

	if !ierror.IsNotFoundErr(err) {
		return 0, err
	}

	return c.blobMgr.Create(ctx, params.Digest, params.ContentType, params.Size)
}

func (c *controller) Exist(ctx context.Context, params ExistParams) (bool, error) {
	if params.Digest == "" {
		return false, ierror.BadRequestError(nil).WithMessage("exist blob require digest")
	}

	_, err := c.Get(ctx, GetParams{Digest: params.Digest, ProjectID: params.ProjectID, ArtifactDigest: params.ArtifactDigest})
	if err != nil {
		if ierror.IsNotFoundErr(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (c *controller) Sync(ctx context.Context, params SyncParams) error {
	if len(params.References) == 0 {
		return nil
	}

	var digests []string
	for _, reference := range params.References {
		digests = append(digests, reference.Digest.String())
	}

	blobs, err := c.blobMgr.List(ctx, blob.ListParams{Digests: digests})
	if err != nil {
		return err
	}

	mp := make(map[string]*blob.Blob, len(blobs))
	for _, blob := range blobs {
		mp[blob.Digest] = blob
	}

	var missing, updating []*blob.Blob
	for _, reference := range params.References {
		if exist, found := mp[reference.Digest.String()]; found {
			if exist.ContentType != reference.MediaType {
				exist.ContentType = reference.MediaType
				updating = append(updating, exist)
			}
		} else {
			missing = append(missing, &blob.Blob{
				Digest:      reference.Digest.String(),
				ContentType: reference.MediaType,
				Size:        reference.Size,
			})
		}
	}

	if len(updating) > 0 {
		orm.WithTransaction(func(ctx context.Context) error {
			for _, blob := range updating {
				if err := c.blobMgr.Update(ctx, blob); err != nil {
					log.Warningf("Failed to update blob %s, error: %v", blob.Digest, err)
					return err
				}
			}

			return nil
		})(ctx)
	}

	if len(missing) > 0 {
		for _, blob := range missing {
			if _, err := c.blobMgr.Create(ctx, blob.Digest, blob.ContentType, blob.Size); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *controller) SetAcceptedBlobSize(sessionID string, size int64) error {
	conn := util.DefaultPool().Get()
	defer conn.Close()

	key := fmt.Sprintf("upload:%s:size", sessionID)
	reply, err := redis.String(conn.Do("SET", key, size))
	if err != nil {
		return err
	}

	if reply != "OK" {
		return fmt.Errorf("bad reply value")
	}

	return nil
}

func (c *controller) GetAcceptedBlobSize(sessionID string) (int64, error) {
	conn := util.DefaultPool().Get()
	defer conn.Close()

	key := fmt.Sprintf("upload:%s:size", sessionID)
	size, err := redis.Int64(conn.Do("GET", key))
	if err != nil {
		return 0, err
	}

	return size, nil
}
