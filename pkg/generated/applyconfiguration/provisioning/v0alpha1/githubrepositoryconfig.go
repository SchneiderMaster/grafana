// SPDX-License-Identifier: AGPL-3.0-only

// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v0alpha1

import (
	provisioningv0alpha1 "github.com/grafana/grafana/pkg/apis/provisioning/v0alpha1"
)

// GitHubRepositoryConfigApplyConfiguration represents a declarative configuration of the GitHubRepositoryConfig type for use
// with apply.
type GitHubRepositoryConfigApplyConfiguration struct {
	Owner                     *string                         `json:"owner,omitempty"`
	Repository                *string                         `json:"repository,omitempty"`
	Branch                    *string                         `json:"branch,omitempty"`
	Token                     *string                         `json:"token,omitempty"`
	Workflows                 []provisioningv0alpha1.Workflow `json:"workflows,omitempty"`
	GenerateDashboardPreviews *bool                           `json:"generateDashboardPreviews,omitempty"`
}

// GitHubRepositoryConfigApplyConfiguration constructs a declarative configuration of the GitHubRepositoryConfig type for use with
// apply.
func GitHubRepositoryConfig() *GitHubRepositoryConfigApplyConfiguration {
	return &GitHubRepositoryConfigApplyConfiguration{}
}

// WithOwner sets the Owner field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Owner field is set to the value of the last call.
func (b *GitHubRepositoryConfigApplyConfiguration) WithOwner(value string) *GitHubRepositoryConfigApplyConfiguration {
	b.Owner = &value
	return b
}

// WithRepository sets the Repository field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Repository field is set to the value of the last call.
func (b *GitHubRepositoryConfigApplyConfiguration) WithRepository(value string) *GitHubRepositoryConfigApplyConfiguration {
	b.Repository = &value
	return b
}

// WithBranch sets the Branch field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Branch field is set to the value of the last call.
func (b *GitHubRepositoryConfigApplyConfiguration) WithBranch(value string) *GitHubRepositoryConfigApplyConfiguration {
	b.Branch = &value
	return b
}

// WithToken sets the Token field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Token field is set to the value of the last call.
func (b *GitHubRepositoryConfigApplyConfiguration) WithToken(value string) *GitHubRepositoryConfigApplyConfiguration {
	b.Token = &value
	return b
}

// WithWorkflows adds the given value to the Workflows field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the Workflows field.
func (b *GitHubRepositoryConfigApplyConfiguration) WithWorkflows(values ...provisioningv0alpha1.Workflow) *GitHubRepositoryConfigApplyConfiguration {
	for i := range values {
		b.Workflows = append(b.Workflows, values[i])
	}
	return b
}

// WithGenerateDashboardPreviews sets the GenerateDashboardPreviews field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the GenerateDashboardPreviews field is set to the value of the last call.
func (b *GitHubRepositoryConfigApplyConfiguration) WithGenerateDashboardPreviews(value bool) *GitHubRepositoryConfigApplyConfiguration {
	b.GenerateDashboardPreviews = &value
	return b
}
