package resources

import (
	"context"
	"fmt"
	"time"

	"github.com/datamammoth/terraform-provider-datamammoth/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &serverResource{}
	_ resource.ResourceWithImportState = &serverResource{}
)

type serverResource struct {
	client *client.Client
}

type serverResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Hostname  types.String `tfsdk:"hostname"`
	ProductID types.String `tfsdk:"product_id"`
	ImageID   types.String `tfsdk:"image_id"`
	ZoneID    types.String `tfsdk:"zone_id"`
	IPAddress types.String `tfsdk:"ip_address"`
	Status    types.String `tfsdk:"status"`
	CPU       types.Int64  `tfsdk:"cpu"`
	Memory    types.Int64  `tfsdk:"memory"`
	Disk      types.Int64  `tfsdk:"disk"`
	CreatedAt types.String `tfsdk:"created_at"`
}

func NewServerResource() resource.Resource {
	return &serverResource{}
}

func (r *serverResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (r *serverResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a DataMammoth cloud server instance.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier of the server.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"hostname": schema.StringAttribute{
				Required:    true,
				Description: "Hostname for the server.",
			},
			"product_id": schema.StringAttribute{
				Required:    true,
				Description: "Product/plan ID determining server specs (e.g. prod_vps_medium).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"image_id": schema.StringAttribute{
				Required:    true,
				Description: "OS image ID (e.g. ubuntu-22.04, debian-12, rocky-9).",
			},
			"zone_id": schema.StringAttribute{
				Required:    true,
				Description: "Availability zone ID (e.g. us-east, eu-west).",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ip_address": schema.StringAttribute{
				Computed:    true,
				Description: "Primary IPv4 address assigned to the server.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Current server status (provisioning, running, stopped, error).",
			},
			"cpu": schema.Int64Attribute{
				Computed:    true,
				Description: "Number of CPU cores.",
			},
			"memory": schema.Int64Attribute{
				Computed:    true,
				Description: "Memory in MB.",
			},
			"disk": schema.Int64Attribute{
				Computed:    true,
				Description: "Disk size in GB.",
			},
			"created_at": schema.StringAttribute{
				Computed:    true,
				Description: "ISO 8601 timestamp of server creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *serverResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *serverResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serverResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{
		"hostname":   plan.Hostname.ValueString(),
		"product_id": plan.ProductID.ValueString(),
		"image_id":   plan.ImageID.ValueString(),
		"zone_id":    plan.ZoneID.ValueString(),
	}

	result, err := r.client.Post("/servers", body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating server", err.Error())
		return
	}

	data, _ := result["data"].(map[string]interface{})
	taskID, _ := data["task_id"].(string)
	serverID, _ := data["id"].(string)

	// Poll task until completion
	if taskID != "" {
		for i := 0; i < 120; i++ {
			time.Sleep(5 * time.Second)
			taskResult, err := r.client.Get(fmt.Sprintf("/tasks/%s", taskID))
			if err != nil {
				resp.Diagnostics.AddError("Error polling task", err.Error())
				return
			}
			taskData, _ := taskResult["data"].(map[string]interface{})
			status, _ := taskData["status"].(string)
			if status == "completed" {
				break
			}
			if status == "failed" {
				errMsg, _ := taskData["error"].(string)
				resp.Diagnostics.AddError("Server provisioning failed", errMsg)
				return
			}
		}
	}

	// Fetch full server details
	serverResult, err := r.client.Get(fmt.Sprintf("/servers/%s", serverID))
	if err != nil {
		resp.Diagnostics.AddError("Error reading server after creation", err.Error())
		return
	}

	r.mapResponseToModel(serverResult, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *serverResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serverResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Get(fmt.Sprintf("/servers/%s", state.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Error reading server", err.Error())
		return
	}

	r.mapResponseToModel(result, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *serverResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan serverResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state serverResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{
		"hostname": plan.Hostname.ValueString(),
	}

	// If image changed, trigger a rebuild
	if !plan.ImageID.Equal(state.ImageID) {
		body["image_id"] = plan.ImageID.ValueString()
	}

	result, err := r.client.Patch(fmt.Sprintf("/servers/%s", state.ID.ValueString()), body)
	if err != nil {
		resp.Diagnostics.AddError("Error updating server", err.Error())
		return
	}

	data, _ := result["data"].(map[string]interface{})
	taskID, _ := data["task_id"].(string)

	// Poll task if async update
	if taskID != "" {
		for i := 0; i < 120; i++ {
			time.Sleep(5 * time.Second)
			taskResult, err := r.client.Get(fmt.Sprintf("/tasks/%s", taskID))
			if err != nil {
				resp.Diagnostics.AddError("Error polling task", err.Error())
				return
			}
			taskData, _ := taskResult["data"].(map[string]interface{})
			status, _ := taskData["status"].(string)
			if status == "completed" {
				break
			}
			if status == "failed" {
				errMsg, _ := taskData["error"].(string)
				resp.Diagnostics.AddError("Server update failed", errMsg)
				return
			}
		}
	}

	// Refresh state
	serverResult, err := r.client.Get(fmt.Sprintf("/servers/%s", state.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Error reading server after update", err.Error())
		return
	}

	r.mapResponseToModel(serverResult, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *serverResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serverResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Delete(fmt.Sprintf("/servers/%s", state.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Error deleting server", err.Error())
		return
	}

	data, _ := result["data"].(map[string]interface{})
	taskID, _ := data["task_id"].(string)

	// Poll deletion task
	if taskID != "" {
		for i := 0; i < 60; i++ {
			time.Sleep(5 * time.Second)
			taskResult, err := r.client.Get(fmt.Sprintf("/tasks/%s", taskID))
			if err != nil {
				// Task endpoint may 404 after completion
				break
			}
			taskData, _ := taskResult["data"].(map[string]interface{})
			status, _ := taskData["status"].(string)
			if status == "completed" {
				break
			}
			if status == "failed" {
				errMsg, _ := taskData["error"].(string)
				resp.Diagnostics.AddError("Server deletion failed", errMsg)
				return
			}
		}
	}
}

func (r *serverResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *serverResource) mapResponseToModel(result map[string]interface{}, model *serverResourceModel) {
	data, _ := result["data"].(map[string]interface{})
	if data == nil {
		return
	}

	if id, ok := data["id"].(string); ok {
		model.ID = types.StringValue(id)
	}
	if hostname, ok := data["hostname"].(string); ok {
		model.Hostname = types.StringValue(hostname)
	}
	if productID, ok := data["product_id"].(string); ok {
		model.ProductID = types.StringValue(productID)
	}
	if imageID, ok := data["image_id"].(string); ok {
		model.ImageID = types.StringValue(imageID)
	}
	if zoneID, ok := data["zone_id"].(string); ok {
		model.ZoneID = types.StringValue(zoneID)
	}
	if ip, ok := data["ip_address"].(string); ok {
		model.IPAddress = types.StringValue(ip)
	}
	if status, ok := data["status"].(string); ok {
		model.Status = types.StringValue(status)
	}
	if cpu, ok := data["cpu"].(float64); ok {
		model.CPU = types.Int64Value(int64(cpu))
	}
	if memory, ok := data["memory"].(float64); ok {
		model.Memory = types.Int64Value(int64(memory))
	}
	if disk, ok := data["disk"].(float64); ok {
		model.Disk = types.Int64Value(int64(disk))
	}
	if createdAt, ok := data["created_at"].(string); ok {
		model.CreatedAt = types.StringValue(createdAt)
	}
}
