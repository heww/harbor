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

package admiral

import (
	"sync"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/security/admiral/authcontext"
	"github.com/goharbor/harbor/src/core/promgr"
	"github.com/goharbor/harbor/src/pkg/permission/evaluator"
	"github.com/goharbor/harbor/src/pkg/permission/evaluator/admin"
	"github.com/goharbor/harbor/src/pkg/permission/types"
)

// SecurityContext implements security.Context interface based on
// auth context and project manager
type SecurityContext struct {
	ctx       *authcontext.AuthContext
	pm        promgr.ProjectManager
	evaluator evaluator.Evaluator
	once      sync.Once
}

// NewSecurityContext ...
func NewSecurityContext(ctx *authcontext.AuthContext, pm promgr.ProjectManager) *SecurityContext {
	return &SecurityContext{
		ctx: ctx,
		pm:  pm,
	}
}

// IsAuthenticated returns true if the user has been authenticated
func (s *SecurityContext) IsAuthenticated() bool {
	if s.ctx == nil {
		return false
	}
	return len(s.ctx.PrincipalID) > 0
}

// GetUsername returns the username of the authenticated user
// It returns null if the user has not been authenticated
func (s *SecurityContext) GetUsername() string {
	if !s.IsAuthenticated() {
		return ""
	}
	return s.ctx.PrincipalID
}

// IsSysAdmin returns whether the authenticated user is system admin
// It returns false if the user has not been authenticated
func (s *SecurityContext) IsSysAdmin() bool {
	if !s.IsAuthenticated() {
		return false
	}

	return s.ctx.IsSysAdmin()
}

// IsSolutionUser ...
func (s *SecurityContext) IsSolutionUser() bool {
	return false
}

// Can returns whether the user can do action on resource
func (s *SecurityContext) Can(action types.Action, resource types.Resource) bool {
	s.once.Do(func() {
		var evaluators evaluator.Evaluators
		if s.IsSysAdmin() {
			evaluators = evaluators.Add(admin.New(s.GetUsername()))
		}
		evaluators = evaluators.Add(rbac.NewProjectRBACEvaluator(s, s.pm))

		s.evaluator = evaluators
	})

	return s.evaluator != nil && s.evaluator.HasPermission(resource, action)
}

// GetMyProjects ...
func (s *SecurityContext) GetMyProjects() ([]*models.Project, error) {
	return s.ctx.GetMyProjects(), nil
}

// GetProjectRoles ...
func (s *SecurityContext) GetProjectRoles(projectIDOrName interface{}) []int {
	if !s.IsAuthenticated() || projectIDOrName == nil {
		return []int{}
	}

	return s.ctx.GetProjectRoles(projectIDOrName)
}
