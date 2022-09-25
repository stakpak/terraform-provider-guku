package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/devopzilla/guku-client-go"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces
var _ resource.Resource = &PlatformBindingResource{}
var _ resource.ResourceWithImportState = &PlatformBindingResource{}

func NewPlatformBindingResource() resource.Resource {
	return &PlatformBindingResource{}
}

// PlatformBindingResource defines the resource implementation.
type PlatformBindingResource struct {
	client *guku.Client
}

// PlatformBindingResourceModel describes the resource data model.
type PlatformBindingResourceModel struct {
	PlatformBindingID types.String `tfsdk:"id"`
	ClusterID         types.String `tfsdk:"cluster_id"`
	PlatformConfigID  types.String `tfsdk:"platform_config_id"`
	PlatformID        types.String `tfsdk:"platform_id"`
	PlatformVersion   types.String `tfsdk:"platform_version"`
	Status            types.String `tfsdk:"status"`
}

func (r *PlatformBindingResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_platform_binding"
}

func (r *PlatformBindingResource) GetSchema(ctx context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "PlatformBinding resource",

		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Computed:            true,
				MarkdownDescription: "Platform Binding id",
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
				Type: types.StringType,
			},
			"cluster_id": {
				MarkdownDescription: "Cluster id",
				Required:            true,
				Type:                types.StringType,
			},
			"platform_config_id": {
				MarkdownDescription: "Platform Config id",
				Required:            true,
				Type:                types.StringType,
			},
			"platform_id": {
				MarkdownDescription: "Platform id",
				Required:            true,
				Type:                types.StringType,
			},
			"platform_version": {
				MarkdownDescription: "Platform Version",
				Required:            true,
				Type:                types.StringType,
			},
			"status": {
				MarkdownDescription: "Platform Binding status, one of `Pending`, `Succeeded`, `Failed`, `Error`",
				Computed:            true,
				Type:                types.StringType,
			},
		},
	}, nil
}

func (r *PlatformBindingResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*guku.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *guku.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *PlatformBindingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *PlatformBindingResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	platformBinding, err := r.client.CreatePlatformBinding(
		data.ClusterID.Value,
		data.PlatformID.Value,
		data.PlatformVersion.Value,
		data.PlatformConfigID.Value,
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create platform binding, got error: %s", err))
		return
	}

	data.PlatformBindingID = types.String{Value: platformBinding.GetPlatformBindingID()}

	// poll till status is not pending
	attempts := 1
	status := string(platformBinding.GetStatus())
	for status == string(guku.PlatformBindingStatusPending) && attempts <= 20 {
		tflog.Trace(ctx, fmt.Sprintf("Polling platform binding %s attempt number %d", data.PlatformBindingID.Value, attempts))

		pb, err := r.client.GetPlatformBinding(data.ClusterID.Value, data.PlatformBindingID.Value)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to poll platform binding, got error: %s", err))
			return
		}

		status = string(pb.GetStatus())

		attempts++
		time.Sleep(time.Second * 30)
	}

	// fail if status is not succeeded
	if status != string(guku.PlatformBindingStatusSucceeded) {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create platform binding, got status: %s", status))
		return
	}

	data.Status = types.String{Value: status}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a platform binding")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PlatformBindingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *PlatformBindingResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	platformBinding, err := r.client.GetPlatformBinding(
		data.ClusterID.Value,
		data.PlatformBindingID.Value,
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read platform binding, got error: %s", err))
		return
	}

	data.PlatformConfigID = types.String{Value: platformBinding.GetPlatformConfigID()}
	data.PlatformID = types.String{Value: platformBinding.GetPlatformID()}
	data.PlatformVersion = types.String{Value: platformBinding.GetPlatformVersion()}
	data.Status = types.String{Value: string(platformBinding.GetStatus())}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PlatformBindingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data *PlatformBindingResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	platformBinding, err := r.client.UpdatePlatformBinding(
		data.ClusterID.Value,
		data.PlatformBindingID.Value,
		ValueStringOrNull(data.PlatformConfigID),
		ValueStringOrNull(data.PlatformVersion),
	)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update platform binding, got error: %s", err))
		return
	}

	// poll till status is not pending
	attempts := 1
	status := string(platformBinding.GetStatus())
	for status == string(guku.PlatformBindingStatusPending) && attempts <= 20 {
		tflog.Trace(ctx, fmt.Sprintf("Polling platform binding %s attempt number %d", data.PlatformBindingID.Value, attempts))

		pb, err := r.client.GetPlatformBinding(data.ClusterID.Value, data.PlatformBindingID.Value)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to poll platform binding, got error: %s", err))
			return
		}

		status = string(pb.GetStatus())

		attempts++
		time.Sleep(time.Second * 30)
	}

	// fail if status is not succeeded
	if status != string(guku.PlatformBindingStatusSucceeded) {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update platform binding, got status: %s", status))
		return
	}

	data.Status = types.String{Value: status}

	tflog.Trace(ctx, "updated a platform binding")

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PlatformBindingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data *PlatformBindingResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.DeletePlatformBinding(
		data.ClusterID.Value,
		data.PlatformBindingID.Value,
	)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete platform binding, got error: %s", err))
		return
	}

	time.Sleep(time.Minute * 2)

	tflog.Trace(ctx, "deleted a platform binding")
}

func (r *PlatformBindingResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
