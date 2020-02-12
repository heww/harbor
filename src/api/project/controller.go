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

package project

import (
	"context"

	"github.com/goharbor/harbor/src/common/utils"
	ierror "github.com/goharbor/harbor/src/internal/error"
	"github.com/goharbor/harbor/src/pkg/project"
)

var (
	// Ctl is a global blob controller instance
	Ctl = NewController()
)

// Controller defines the operations related with blobs
type Controller interface {
	// GetByName get the project by name
	GetByName(ctx context.Context, projectName string) (*Project, error)
}

// NewController creates an instance of the default repository controller
func NewController() Controller {
	return &controller{
		projectMgr: project.Mgr,
	}
}

type controller struct {
	projectMgr project.Manager
}

func (c *controller) GetByName(ctx context.Context, name string) (*Project, error) {
	projectName, rest := utils.ParseRepository(name)
	if projectName == "" {
		projectName = rest
	}
	if projectName == "" {
		return nil, ierror.BadRequestError(nil).WithMessage("invalid name %s", name)
	}

	p, err := c.projectMgr.Get(projectName)
	if err != nil {
		return nil, err
	}
	if p == nil {
		return nil, ierror.NotFoundError(nil).WithMessage("project %s not found", projectName)
	}

	return &Project{
		ProjectID: p.ProjectID,
		Name:      p.Name,
	}, nil
}
