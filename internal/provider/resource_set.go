// Copyright (c) Persona
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var _ resource.ResourceWithModifyPlan = (*SetResource)(nil)

type PlanOrState interface {
	Set(context.Context, interface{}) diag.Diagnostics
}

func NewSetResource() resource.Resource {
	return &SetResource{}
}

type SetResource struct{}

func (r *SetResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var model setModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	model.ID = types.StringValue("-")

	r.modify(ctx, model, types.SetValueMust(types.StringType, []attr.Value{}), &resp.Diagnostics, &resp.State)
}

// Delete does not need to explicitly call resp.State.RemoveResource() as this is automatically handled by the
// [framework](https://github.com/hashicorp/terraform-plugin-framework/pull/301).
func (r *SetResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (r *SetResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_set"
}

func (r *SetResource) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	// Will be when the resource is being deleted.
	if req.Plan.Raw.IsNull() {
		return
	}

	var model setModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if model.Values.IsUnknown() {
		resp.Diagnostics.Append(resp.Plan.SetAttribute(ctx, path.Root("result"), types.SetUnknown(types.StringType))...)
		return
	}

	// Read existing result field from state, if present.
	existingResult := types.SetValueMust(types.StringType, []attr.Value{})
	if !req.State.Raw.IsNull() {
		resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("result"), &existingResult)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	r.modify(ctx, model, existingResult, &resp.Diagnostics, &resp.Plan)
}

// Read does not need to perform any operations as the state in ReadResourceResponse is already populated.
func (r *SetResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
}

func (r *SetResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Tracks a set of string values permanently. Values added to `values` accumulate in `result` over time and are never removed, even if omitted from future configurations.",
		Attributes: map[string]schema.Attribute{
			"values": schema.SetAttribute{
				Description: "The set of string values to track.",
				ElementType: types.StringType,
				Required:    true,
			},

			// Computed
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "A static value used internally by Terraform, this should not be referenced in configurations.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"result": schema.SetAttribute{
				Computed:    true,
				Description: "The accumulated union of all values ever provided.",
				ElementType: types.StringType,
			},
		},
	}
}

// Update ensures the plan value is copied to the state to complete the update.
func (r *SetResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var model setModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &model)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Read existing result field from state.
	var existingResult types.Set
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("result"), &existingResult)...)

	if resp.Diagnostics.HasError() {
		return
	}

	r.modify(ctx, model, existingResult, &resp.Diagnostics, &resp.State)
}

func (r *SetResource) modify(ctx context.Context, model setModel, existingResult types.Set, diagnostics *diag.Diagnostics, state PlanOrState) {
	seen := make(map[string]bool)
	union := []string{}

	for _, s := range []types.Set{existingResult, model.Values} {
		if s.IsNull() || s.IsUnknown() {
			continue
		}

		var elements []basetypes.StringValue

		diagnostics.Append(s.ElementsAs(ctx, &elements, false)...)
		if diagnostics.HasError() {
			return
		}

		for _, v := range elements {
			str := v.ValueString()
			if _, ok := seen[str]; !ok {
				seen[str] = true
				union = append(union, str)
			}
		}
	}

	var diags diag.Diagnostics
	model.Result, diags = types.SetValueFrom(ctx, types.StringType, union)
	diagnostics.Append(diags...)
	if diagnostics.HasError() {
		return
	}

	diagnostics.Append(state.Set(ctx, model)...)
}

type setModel struct {
	ID     types.String `tfsdk:"id"`
	Result types.Set    `tfsdk:"result"`
	Values types.Set    `tfsdk:"values"`
}
