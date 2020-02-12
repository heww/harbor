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
	"fmt"
	"testing"

	"github.com/goharbor/harbor/src/pkg/distribution"
	htesting "github.com/goharbor/harbor/src/testing"
	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
)

type ControllerTestSuite struct {
	htesting.Suite
}

func (suite *ControllerTestSuite) prepareBlob() string {

	ctx := suite.Context()
	digest := suite.DigestString()

	_, err := Ctl.Ensure(ctx, EnsureParams{Digest: digest, Size: 100, ContentType: "application/octet-stream"})
	suite.Nil(err)

	return digest
}

func (suite *ControllerTestSuite) TestAttachToArtifact() {
	ctx := suite.Context()

	artifactDigest := suite.DigestString()
	blobDigests := []string{
		suite.prepareBlob(),
		suite.prepareBlob(),
		suite.prepareBlob(),
	}

	suite.Nil(Ctl.AttachToArtifact(ctx, AttachToArtifactParams{ArtifactDigest: artifactDigest, BlobDigests: blobDigests}))

	for _, digest := range blobDigests {
		exist, err := Ctl.Exist(ctx, ExistParams{Digest: digest, ArtifactDigest: artifactDigest})
		suite.Nil(err)
		suite.True(exist)
	}

	suite.Nil(Ctl.AttachToArtifact(ctx, AttachToArtifactParams{ArtifactDigest: artifactDigest, BlobDigests: blobDigests}))
}

func (suite *ControllerTestSuite) TestAttachToProject() {
	suite.WithProject(func(projectID int64, projectName string) {
		ctx := suite.Context()

		digest := suite.prepareBlob()
		suite.Nil(Ctl.AttachToProject(ctx, AttachToProjectParams{Digest: digest, ProjectID: projectID}))

		exist, err := Ctl.Exist(ctx, ExistParams{Digest: digest, ProjectID: projectID})
		suite.Nil(err)
		suite.True(exist)
	})
}

func (suite *ControllerTestSuite) TestEnsure() {
	ctx := suite.Context()

	digest := suite.DigestString()

	params := EnsureParams{Digest: digest, Size: 100, ContentType: "application/octet-stream"}
	_, err := Ctl.Ensure(ctx, params)
	suite.Nil(err)

	exist, err := Ctl.Exist(ctx, ExistParams{Digest: digest})
	suite.Nil(err)
	suite.True(exist)

	_, err = Ctl.Ensure(ctx, params)
	suite.Nil(err)
}

func (suite *ControllerTestSuite) TestExist() {
	ctx := suite.Context()

	exist, err := Ctl.Exist(ctx, ExistParams{Digest: suite.DigestString()})
	suite.Nil(err)
	suite.False(exist)
}

func (suite *ControllerTestSuite) TestGet() {
	ctx := suite.Context()

	{
		digest := suite.prepareBlob()
		blob, err := Ctl.Get(ctx, GetParams{Digest: digest})
		suite.Nil(err)
		suite.Equal(digest, blob.Digest)
		suite.Equal(int64(100), blob.Size)
		suite.Equal("application/octet-stream", blob.ContentType)
	}

	{
		digest := suite.prepareBlob()
		artifactDigest := suite.DigestString()

		_, err := Ctl.Get(ctx, GetParams{Digest: digest, ArtifactDigest: artifactDigest})
		suite.NotNil(err)

		Ctl.AttachToArtifact(ctx, AttachToArtifactParams{BlobDigests: []string{digest}, ArtifactDigest: artifactDigest})

		blob, err := Ctl.Get(ctx, GetParams{Digest: digest, ArtifactDigest: artifactDigest})
		suite.Nil(err)
		suite.Equal(digest, blob.Digest)
		suite.Equal(int64(100), blob.Size)
		suite.Equal("application/octet-stream", blob.ContentType)
	}

	{
		digest := suite.prepareBlob()

		suite.WithProject(func(projectID int64, projectName string) {
			_, err := Ctl.Get(ctx, GetParams{Digest: digest, ProjectID: projectID})
			suite.NotNil(err)

			Ctl.AttachToProject(ctx, AttachToProjectParams{Digest: digest, ProjectID: projectID})

			blob, err := Ctl.Get(ctx, GetParams{Digest: digest, ProjectID: projectID})
			suite.Nil(err)
			suite.Equal(digest, blob.Digest)
			suite.Equal(int64(100), blob.Size)
			suite.Equal("application/octet-stream", blob.ContentType)
		})
	}
}

func (suite *ControllerTestSuite) TestSync() {
	var references []distribution.Descriptor
	for i := 0; i < 5; i++ {
		references = append(references, distribution.Descriptor{
			MediaType: fmt.Sprintf("media type %d", i),
			Digest:    suite.Digest(),
			Size:      int64(100 + i),
		})
	}

	suite.WithProject(func(projectID int64, projectName string) {
		ctx := suite.Context()

		{
			suite.Nil(Ctl.Sync(ctx, SyncParams{References: references}))
			for _, reference := range references {
				blob, err := Ctl.Get(ctx, GetParams{Digest: reference.Digest.String()})
				suite.Nil(err)
				suite.Equal(reference.MediaType, blob.ContentType)
				suite.Equal(reference.Digest.String(), blob.Digest)
				suite.Equal(reference.Size, blob.Size)
			}
		}

		{
			references[0].MediaType = "media type"

			references = append(references, distribution.Descriptor{
				MediaType: "media type",
				Digest:    suite.Digest(),
				Size:      int64(100),
			})

			suite.Nil(Ctl.Sync(ctx, SyncParams{References: references}))
		}
	})
}

func (suite *ControllerTestSuite) TestGetSetAcceptedBlobSize() {
	sessionID := uuid.New().String()

	size, err := Ctl.GetAcceptedBlobSize(sessionID)
	suite.NotNil(err)

	suite.Nil(Ctl.SetAcceptedBlobSize(sessionID, 100))

	size, err = Ctl.GetAcceptedBlobSize(sessionID)
	suite.Nil(err)
	suite.Equal(int64(100), size)
}

func TestControllerTestSuite(t *testing.T) {
	suite.Run(t, &ControllerTestSuite{})
}
