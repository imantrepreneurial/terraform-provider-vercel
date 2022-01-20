package vercel

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/vercel/terraform-provider-vercel/client"
)

type resourceProjectDomainType struct{}

func (r resourceProjectDomainType) GetSchema(context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"project_id": {
				Required:      true,
				PlanModifiers: tfsdk.AttributePlanModifiers{tfsdk.RequiresReplace()},
				Type:          types.StringType,
			},
			"team_id": {
				Optional:      true,
				PlanModifiers: tfsdk.AttributePlanModifiers{tfsdk.RequiresReplace()},
				Type:          types.StringType,
				Description:   "The ID of the team the project should be created under",
			},
			"id": {
				Computed: true,
				Type:     types.StringType,
			},
			"domain": {
				Required:      true,
				PlanModifiers: tfsdk.AttributePlanModifiers{tfsdk.RequiresReplace()},
				Type:          types.StringType,
			},
			"redirect": {
				Optional: true,
				Type:     types.StringType,
			},
			"redirect_status_code": {
				Optional: true,
				Type:     types.Int64Type,
			},
			"git_branch": {
				Optional: true,
				Type:     types.StringType,
			},
		},
	}, nil
}

func (r resourceProjectDomainType) NewResource(_ context.Context, p tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	return resourceProjectDomain{
		p: *(p.(*provider)),
	}, nil
}

type resourceProjectDomain struct {
	p provider
}

func (r resourceProjectDomain) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.p.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply. This leads to weird stuff happening, so we'd prefer if you didn't do that. Thanks!",
		)
		return
	}

	var plan ProjectDomain
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := r.p.client.CreateProjectDomain(ctx, plan.ProjectID.Value, plan.TeamID.Value, plan.toCreateRequest())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error adding domain to project",
			fmt.Sprintf(
				"Could not add domain %s to project %s, unexpected error: %s",
				plan.Domain.Value,
				plan.ProjectID.Value,
				err,
			),
		)
		return
	}

	result := convertResponseToProjectDomain(out, plan.TeamID)
	tflog.Trace(
		ctx, "added domain to project",
		"project_id", result.ProjectID.Value,
		"domain", result.Domain.Value,
		"team_id", result.TeamID.Value,
	)

	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceProjectDomain) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	var state ProjectDomain
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := r.p.client.GetProjectDomain(ctx, state.ProjectID.Value, state.Domain.Value, state.TeamID.Value)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading project domain",
			fmt.Sprintf("Could not get domain %s for project %s, unexpected error: %s",
				state.Domain.Value,
				state.ProjectID.Value,
				err,
			),
		)
		return
	}

	result := convertResponseToProjectDomain(out, state.TeamID)
	tflog.Trace(
		ctx, "read project domain",
		"project_id", result.ProjectID.Value,
		"domain", result.Domain.Value,
		"team_id", result.TeamID.Value,
	)

	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceProjectDomain) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	var plan ProjectDomain
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state ProjectDomain
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := r.p.client.UpdateProjectDomain(
		ctx,
		plan.ProjectID.Value,
		plan.Domain.Value,
		plan.TeamID.Value,
		plan.toUpdateRequest(),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating project domain",
			fmt.Sprintf("Could not update domain %s for project %s, unexpected error: %s",
				state.Domain.Value,
				state.ProjectID.Value,
				err,
			),
		)
		return
	}

	result := convertResponseToProjectDomain(out, state.TeamID)
	tflog.Trace(
		ctx, "update project domain",
		"project_id", result.ProjectID.Value,
		"domain", result.Domain.Value,
		"team_id", result.TeamID.Value,
	)

	diags = resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r resourceProjectDomain) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state ProjectDomain
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.p.client.DeleteProjectDomain(ctx, state.ProjectID.Value, state.Domain.Value, state.TeamID.Value)
	var apiErr client.APIError
	if err != nil && errors.As(err, &apiErr) && apiErr.StatusCode == 404 {
		// The domain is already gone - do nothing.
		return
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting project",
			fmt.Sprintf(
				"Could not delete domain %s for project %s, unexpected error: %s",
				state.Domain.Value,
				state.ProjectID.Value,
				err,
			),
		)
		return
	}

	tflog.Trace(
		ctx, "delete project domain",
		"project_id", state.ProjectID.Value,
		"domain", state.Domain.Value,
		"team_id", state.TeamID.Value,
	)
	resp.State.RemoveResource(ctx)
}

func splitProjectDomainID(id string) (teamID, projectID, domain string, ok bool) {
	attributes := strings.Split(id, "/")
	if len(attributes) == 2 {
		// we have project_id/domain
		return "", attributes[0], attributes[1], true
	}
	if len(attributes) == 3 {
		// we have team_id/project_id/domain
		return attributes[0], attributes[1], attributes[2], true
	}
	return "", "", "", false
}

func (r resourceProjectDomain) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	teamID, projectID, domain, ok := splitProjectDomainID(req.ID)
	if !ok {
		resp.Diagnostics.AddError(
			"Error importing project domain",
			fmt.Sprintf("Invalid id '%s' specified. should be in format \"team_id/project_id/domain\" or \"project_id/domain\"", req.ID),
		)
	}

	out, err := r.p.client.GetProjectDomain(ctx, projectID, domain, teamID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading project domain",
			fmt.Sprintf("Could not get domain %s for project %s, unexpected error: %s",
				domain,
				projectID,
				err,
			),
		)
		return
	}

	stringTypeTeamID := types.String{Value: teamID}
	if teamID == "" {
		stringTypeTeamID.Null = true
	}
	result := convertResponseToProjectDomain(out, stringTypeTeamID)
	tflog.Trace(
		ctx, "imported project domain",
		"project_id", result.ProjectID.Value,
		"domain", result.Domain.Value,
		"team_id", result.TeamID.Value,
	)

	diags := resp.State.Set(ctx, result)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}