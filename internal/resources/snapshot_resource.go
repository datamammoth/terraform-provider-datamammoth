package resources

import (
	"context"
	"fmt"

	"github.com/datamammoth/terraform-provider-datamammoth/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &snapshotResource{}

type snapshotResource struct {
	client *client.Client
}

type snapshotResourceModel struct {
	ID        types.String `tfsdk:"id"`
	ServerID  types.String `tfsdk:"server_id"`
	Name      types.String `tfsdk:"name"`
	SizeGB    types.Int64  `tfsdk:"size_gb"`
	Status    types.String `tfsdk:"status"`
	CreatedAt types.String `tfsdk:"created_at"`
}

func NewSnapshotResource() resource.Resource {
	return &snapshotResource{}
}

func (r *snapshotResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_snapshot"
}

func (r *snapshotResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a snapshot of a DataMammoth server.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier of the snapshot.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"server_id": schema.StringAttribute{
				Required:    true,
				Description: "ID of the server to snapshot.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the snapshot.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"size_gb": schema.Int64Attribute{
				Computed:    true,
				Description: "Size of the snapshot in GB.",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Current snapshot status (creating, available, error).",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "ISO 8601 timestamp of snapshot creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *snapshotResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", "Expected *client.Client")
		return
	}
	r.client = c
}

func (r *snapshotResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan snapshotResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{
		"name": plan.Name.ValueString(),
	}

	result, err := r.client.Post(fmt.Sprintf("/servers/%s/snapshots", plan.ServerID.ValueString()), body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating snapshot", err.Error())
		return
	}

	data, _ := result["data"].(map[string]interface{})
	if id, ok := data["id"].(string); ok {
		plan.ID = types.StringValue(id)
	}
	if sizeGB, ok := data["size_gb"].(float64); ok {
		plan.SizeGB = types.Int64Value(int64(sizeGB))
	}
	if status, ok := data["status"].(string); ok {
		plan.Status = types.StringValue(status)
	}
	if createdAt, ok := data["created_at"].(string); ok {
		plan.CreatedAt = types.StringValue(createdAt)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *snapshotResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state snapshotResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Get(fmt.Sprintf("/servers/%s/snapshots/%s", state.ServerID.ValueString(), state.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Error reading snapshot", err.Error())
		return
	}

	data, _ := result["data"].(map[string]interface{})
	if sizeGB, ok := data["size_gb"].(float64); ok {
		state.SizeGB = types.Int64Value(int64(sizeGB))
	}
	if status, ok := data["status"].(string); ok {
		state.Status = types.StringValue(status)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *snapshotResource) Update(_ context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("Snapshots are immutable", "Snapshots cannot be updated. Delete and recreate instead.")
}

func (r *snapshotResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state snapshotResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Delete(fmt.Sprintf("/servers/%s/snapshots/%s", state.ServerID.ValueString(), state.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Error deleting snapshot", err.Error())
		return
	}
}
