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
	"io/ioutil"
	"net/http"

	"github.com/docker/distribution/manifest/schema2"
	"github.com/goharbor/harbor/src/api/blob"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/pkg/distribution"
	"github.com/goharbor/harbor/src/server/middleware"
	"github.com/justinas/alice"
)

// PutManifestMiddleware middleware which create Blobs for the foreign layers and attach them to the project,
// update the content type of the Blobs which already exist,
// create Blob for the manifest, attach all Blobs to the manifest after PUT /v2/<name>/manifests/<reference> success.
func PutManifestMiddleware() func(http.Handler) http.Handler {
	skipper := middleware.NegativeSkipper(
		middleware.MethodAndPathSkipper(http.MethodPut, distribution.ManifestURLRegexp),
	)

	before := middleware.BeforeRequest(func(r *http.Request) error {
		// Do nothing, only make the request nopclose
		return nil
	}, skipper)

	after := middleware.AfterResponse(func(w http.ResponseWriter, r *http.Request, statusCode int) error {
		if statusCode != http.StatusCreated {
			return nil
		}

		logPrefix := fmt.Sprintf("[middleware][%s][blob]", r.URL.Path)

		ctx := r.Context()
		p, err := projectController.GetByName(ctx, distribution.ParseName(r.URL.Path))
		if err != nil {
			log.Errorf("%s: get project failed, error: %v", logPrefix, err)
			return err
		}

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return err
		}

		contentType := r.Header.Get("Content-Type")
		manifest, descriptor, err := distribution.UnmarshalManifest(contentType, body)
		if err != nil {
			log.Errorf("%s: unmarshal manifest failed, error: %v", logPrefix, err)
			return err
		}

		// sync blobs
		if err := blobController.Sync(ctx, blob.SyncParams{References: manifest.References()}); err != nil {
			log.Errorf("%s: sync missing blobs from manifest %s failed, error: %c", logPrefix, descriptor.Digest.String(), err)
			return err
		}

		for _, digest := range findForeignBlobDigests(manifest) {
			if err := blobController.AttachToProject(ctx, blob.AttachToProjectParams{Digest: digest, ProjectID: p.ProjectID}); err != nil {
				return err
			}
		}

		artifactDigest := descriptor.Digest.String()

		// ensure Blob for the manifest
		blobID, err := blobController.Ensure(ctx, blob.EnsureParams{ContentType: contentType, Digest: artifactDigest, Size: descriptor.Size})
		if err != nil {
			log.Errorf("%s: ensure blob %s failed, error: %v", logPrefix, descriptor.Digest, err)
			return err
		}

		if err := blobController.AttachToProject(ctx, blob.AttachToProjectParams{BlobID: blobID, ProjectID: p.ProjectID}); err != nil {
			return err
		}

		var blobDigests []string
		for _, reference := range manifest.References() {
			blobDigests = append(blobDigests, reference.Digest.String())
		}

		// attach blobDigests of the manifest to artifact
		if err := blobController.AttachToArtifact(ctx, blob.AttachToArtifactParams{BlobDigests: blobDigests, ArtifactDigest: artifactDigest}); err != nil {
			log.Errorf("%s: attach blobs to artifact %s failed, error: %v", logPrefix, descriptor.Digest, err)
			return err
		}

		return nil
	}, skipper)

	return func(next http.Handler) http.Handler {
		return alice.New(before, after).Then(next)
	}
}

func isForeign(d *distribution.Descriptor) bool {
	return d.MediaType == schema2.MediaTypeForeignLayer
}

func findForeignBlobDigests(manifest distribution.Manifest) []string {
	var digests []string
	for _, reference := range manifest.References() {
		if isForeign(&reference) {
			digests = append(digests, reference.Digest.String())
		}
	}
	return digests
}
