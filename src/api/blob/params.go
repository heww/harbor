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
	"github.com/docker/distribution"
)

// AttachToArtifactParams attach to artifact params
type AttachToArtifactParams struct {
	ArtifactDigest string   // manifest digest
	BlobDigests    []string // blob digests
}

// AttachToProjectParams attach to project params
type AttachToProjectParams struct {
	BlobID int64  // blob id
	Digest string // blob digest

	ProjectID int64 // attach blob to the project
}

// EnsureParams ensure params
type EnsureParams struct {
	ContentType string // blob content type
	Digest      string // blob digest
	Size        int64  // blob size
}

// ExistParams exist query for blob
type ExistParams struct {
	Digest string // blob digest

	ArtifactDigest string // check blob attached to the artifact

	ProjectID int64 // check blob attached to the project
}

// GetParams get params
type GetParams struct {
	Digest string // blob digest

	ArtifactDigest string // get blob and it must attached to the artifact

	ProjectID int64 // get blob and it must attached to the project
}

// SyncParams sync params
type SyncParams struct {
	References []distribution.Descriptor // the references to sync
}
