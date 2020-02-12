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

// CreateProjectBlob add blob to project and ignore conflict
func CreateProjectBlob(ctx context.Context, md *ProjectBlob) (int64, error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return 0, err
	}

	md.CreationTime = time.Now()

	// ignore conflict error on (blob_id, project_id)
	return o.InsertOrUpdate(md, "blob_id, project_id")
}

// ExistProjectBlob returns true when ProjectBlob exist
func ExistProjectBlob(ctx context.Context, blobDigest string, projectID int64) (bool, error) {
	o, err := orm.FromContext(ctx)
	if err != nil {
		return false, err
	}

	sql := `SELECT COUNT(*) FROM project_blob JOIN blob ON project_blob.blob_id = blob.id AND project_id = ? AND digest = ?`

	var count int64
	if err := o.Raw(sql, projectID, blobDigest).QueryRow(&count); err != nil {
		return false, err
	}

	return count > 0, nil
}
