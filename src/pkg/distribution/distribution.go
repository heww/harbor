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

package distribution

import (
	"fmt"
	"regexp"

	"github.com/docker/distribution"
	// docker schema2 manifest
	_ "github.com/docker/distribution/manifest/schema2"
	// manifestlist
	_ "github.com/docker/distribution/manifest/manifestlist"
	ref "github.com/docker/distribution/reference"
)

// Descriptor alias type of github.com/docker/distribution.Descriptor
type Descriptor = distribution.Descriptor

// Manifest alias type of github.com/docker/distribution.Manifest
type Manifest = distribution.Manifest

var (
	// UnmarshalManifest alias func from `github.com/docker/distribution`
	UnmarshalManifest = distribution.UnmarshalManifest
)

var (
	name      = fmt.Sprintf("(?P<name>%s)", ref.NameRegexp)
	reference = fmt.Sprintf("(?P<reference>(%s|%s))", ref.TagRegexp, ref.DigestRegexp)
	sessionID = "(?P<session_id>[a-zA-Z0-9-_.=]+)"

	// BlobUploadURLRegexp regexp which match blob upload url
	BlobUploadURLRegexp = regexp.MustCompile(`^/v2/` + name + `/blobs/uploads/` + sessionID)

	// InitiateBlobUploadRegexp regexp which match initiate blob upload url
	InitiateBlobUploadRegexp = regexp.MustCompile(`^/v2/` + name + `/blobs/uploads`)

	// ManifestURLRegexp regexp which match manifest url
	ManifestURLRegexp = regexp.MustCompile(`^/v2/` + name + `/manifests/` + reference)
)

var (
	extractNameRegexp      = regexp.MustCompile(`^/v2/` + name + `/(manifests|blobs|tags)`)
	extractSessionIDRegexp = regexp.MustCompile(`^/v2/` + name + `/blobs/uploads/` + sessionID)
)

// ParseName returns name value from distribution API URL path
func ParseName(path string) string {
	m := findNamedMatches(extractNameRegexp, path)
	if len(m) > 0 {
		return m["name"]
	}

	return ""
}

// ParseSessionID returns session id value from distribution API URL path
func ParseSessionID(path string) string {
	m := findNamedMatches(extractSessionIDRegexp, path)
	if len(m) > 0 {
		return m["session_id"]
	}

	return ""
}

func findNamedMatches(regex *regexp.Regexp, str string) map[string]string {
	match := regex.FindStringSubmatch(str)

	results := map[string]string{}
	for i, name := range match {
		results[regex.SubexpNames()[i]] = name
	}
	return results
}
