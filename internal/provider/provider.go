// Copyright (c) Persona
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure AppendOnly satisfies various provider interfaces.
var _ provider.Provider = &AppendOnly{}

// AppendOnly defines the provider implementation.
type AppendOnly struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &AppendOnly{
			version: version,
		}
	}
}

func (p *AppendOnly) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
}

func (p *AppendOnly) DataSources(ctx context.Context) []func() datasource.DataSource {
	return nil
}

func (p *AppendOnly) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "appendonly"
	resp.Version = p.version
}

func (p *AppendOnly) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewSetResource,
	}
}

func (p *AppendOnly) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "This Terraform provider provides resources for tracking values permanently. Once a value is recorded, it persists in state forever even if removed from the input.",
	}
}
