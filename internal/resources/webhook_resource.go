package resources

import (
	"context"
	"fmt"

	"github.com/datamammoth/terraform-provider-datamammoth/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &webhookResource{}
	_ resource.ResourceWithImportState = &webhookResource{}
)

type webhookResource struct {
	client *client.Client
}

type webhookResourceModel struct {
	ID     types.String `tfsdk:"id"`
	URL    types.String `tfsdk:"url"`
	Events types.List   `tfsdk:"events"`
	Secret types.String `tfsdk:"secret"`
	Active types.Bool   `tfsdk:"active"`
}

func NewWebhookResource() resource.Resource {
	return &webhookResource{}
}

func (r *webhookResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook"
}

func (r *webhookResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a DataMammoth webhook subscription for event notifications.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier of the webhook.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"url": schema.StringAttribute{
				Required:    true,
				Description: "HTTPS URL to receive webhook payloads.",
			},
			"events": schema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "List of event types to subscribe to (e.g. server.created, server.deleted, snapshot.completed).",
			},
			"secret": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Secret used to sign webhook payloads (HMAC-SHA256).",
			},
			"active": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Whether the webhook is active. Defaults to true.",
			},
		},
	}
}

func (r *webhookResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *webhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan webhookResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var events []string
	resp.Diagnostics.Append(plan.Events.ElementsAs(ctx, &events, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{
		"url":    plan.URL.ValueString(),
		"events": events,
		"active": plan.Active.ValueBool(),
	}
	if !plan.Secret.IsNull() && !plan.Secret.IsUnknown() {
		body["secret"] = plan.Secret.ValueString()
	}

	result, err := r.client.Post("/webhooks", body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating webhook", err.Error())
		return
	}

	data, _ := result["data"].(map[string]interface{})
	if id, ok := data["id"].(string); ok {
		plan.ID = types.StringValue(id)
	}
	if active, ok := data["active"].(bool); ok {
		plan.Active = types.BoolValue(active)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *webhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state webhookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.Get(fmt.Sprintf("/webhooks/%s", state.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Error reading webhook", err.Error())
		return
	}

	data, _ := result["data"].(map[string]interface{})
	if url, ok := data["url"].(string); ok {
		state.URL = types.StringValue(url)
	}
	if active, ok := data["active"].(bool); ok {
		state.Active = types.BoolValue(active)
	}
	if eventsRaw, ok := data["events"].([]interface{}); ok {
		var events []types.String
		for _, e := range eventsRaw {
			if s, ok := e.(string); ok {
				events = append(events, types.StringValue(s))
			}
		}
		eventsList, diags := types.ListValueFrom(ctx, types.StringType, events)
		resp.Diagnostics.Append(diags...)
		state.Events = eventsList
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *webhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan webhookResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state webhookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var events []string
	resp.Diagnostics.Append(plan.Events.ElementsAs(ctx, &events, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := map[string]interface{}{
		"url":    plan.URL.ValueString(),
		"events": events,
		"active": plan.Active.ValueBool(),
	}
	if !plan.Secret.IsNull() && !plan.Secret.IsUnknown() {
		body["secret"] = plan.Secret.ValueString()
	}

	_, err := r.client.Patch(fmt.Sprintf("/webhooks/%s", state.ID.ValueString()), body)
	if err != nil {
		resp.Diagnostics.AddError("Error updating webhook", err.Error())
		return
	}

	plan.ID = state.ID
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *webhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state webhookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.Delete(fmt.Sprintf("/webhooks/%s", state.ID.ValueString()))
	if err != nil {
		resp.Diagnostics.AddError("Error deleting webhook", err.Error())
		return
	}
}

func (r *webhookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
