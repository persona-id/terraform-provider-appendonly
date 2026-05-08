// Copyright (c) Persona
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccResourceSet(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"appendonly": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			// Create with initial values.
			{
				Config: `
				resource "appendonly_set" "test" {
					values = ["a", "b"]
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("appendonly_set.test", "result.#", "2"),
					resource.TestCheckTypeSetElemAttr("appendonly_set.test", "result.*", "a"),
					resource.TestCheckTypeSetElemAttr("appendonly_set.test", "result.*", "b"),
				),
			},
			// Same config — no changes.
			{
				Config: `
				resource "appendonly_set" "test" {
					values = ["a", "b"]
				}
				`,
				PlanOnly: true,
			},
			// Add a value — result grows.
			{
				Config: `
				resource "appendonly_set" "test" {
					values = ["a", "b", "c"]
				}
				`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue(
							"appendonly_set.test",
							tfjsonpath.New("result"),
							knownvalue.SetExact([]knownvalue.Check{
								knownvalue.StringExact("a"),
								knownvalue.StringExact("b"),
								knownvalue.StringExact("c"),
							}),
						),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("appendonly_set.test", "result.#", "3"),
					resource.TestCheckTypeSetElemAttr("appendonly_set.test", "result.*", "a"),
					resource.TestCheckTypeSetElemAttr("appendonly_set.test", "result.*", "b"),
					resource.TestCheckTypeSetElemAttr("appendonly_set.test", "result.*", "c"),
				),
			},
			// Remove a value — result retains all previously seen values.
			{
				Config: `
				resource "appendonly_set" "test" {
					values = ["b", "c"]
				}
				`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue(
							"appendonly_set.test",
							tfjsonpath.New("result"),
							knownvalue.SetExact([]knownvalue.Check{
								knownvalue.StringExact("a"),
								knownvalue.StringExact("b"),
								knownvalue.StringExact("c"),
							}),
						),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("appendonly_set.test", "result.#", "3"),
					resource.TestCheckTypeSetElemAttr("appendonly_set.test", "result.*", "a"),
					resource.TestCheckTypeSetElemAttr("appendonly_set.test", "result.*", "b"),
					resource.TestCheckTypeSetElemAttr("appendonly_set.test", "result.*", "c"),
				),
			},
			// Same config — no further changes despite result being larger than value.
			{
				Config: `
				resource "appendonly_set" "test" {
					values = ["b", "c"]
				}
				`,
				PlanOnly: true,
			},
			// Re-add a previously removed value — plan shows no change to result.
			{
				Config: `
				resource "appendonly_set" "test" {
					values = ["a", "b", "c"]
				}
				`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue(
							"appendonly_set.test",
							tfjsonpath.New("result"),
							knownvalue.SetExact([]knownvalue.Check{
								knownvalue.StringExact("a"),
								knownvalue.StringExact("b"),
								knownvalue.StringExact("c"),
							}),
						),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("appendonly_set.test", "result.#", "3"),
					resource.TestCheckTypeSetElemAttr("appendonly_set.test", "result.*", "a"),
					resource.TestCheckTypeSetElemAttr("appendonly_set.test", "result.*", "b"),
					resource.TestCheckTypeSetElemAttr("appendonly_set.test", "result.*", "c"),
				),
			},
		},
	})
}

func TestAccResourceSetEmptyValue(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"appendonly": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			// Create with values.
			{
				Config: `
				resource "appendonly_set" "test" {
					values = ["a", "b"]
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("appendonly_set.test", "result.#", "2"),
					resource.TestCheckTypeSetElemAttr("appendonly_set.test", "result.*", "a"),
					resource.TestCheckTypeSetElemAttr("appendonly_set.test", "result.*", "b"),
				),
			},
			// Clear value — result is retained.
			{
				Config: `
				resource "appendonly_set" "test" {
					values = []
				}
				`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue(
							"appendonly_set.test",
							tfjsonpath.New("result"),
							knownvalue.SetExact([]knownvalue.Check{
								knownvalue.StringExact("a"),
								knownvalue.StringExact("b"),
							}),
						),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("appendonly_set.test", "result.#", "2"),
					resource.TestCheckTypeSetElemAttr("appendonly_set.test", "result.*", "a"),
					resource.TestCheckTypeSetElemAttr("appendonly_set.test", "result.*", "b"),
				),
			},
			// Same config — no changes.
			{
				Config: `
				resource "appendonly_set" "test" {
					values = []
				}
				`,
				PlanOnly: true,
			},
		},
	})
}

func TestAccResourceSetUnknownValue(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"appendonly": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			// value is unknown during plan — result should also be unknown.
			{
				Config: `
				resource "terraform_data" "input" {
					input = toset(["a", "b"])
				}
				resource "appendonly_set" "test" {
					values = toset(terraform_data.input.output)
				}
				`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectUnknownValue(
							"appendonly_set.test",
							tfjsonpath.New("result"),
						),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("appendonly_set.test", "result.#", "2"),
					resource.TestCheckTypeSetElemAttr("appendonly_set.test", "result.*", "a"),
					resource.TestCheckTypeSetElemAttr("appendonly_set.test", "result.*", "b"),
				),
			},
		},
	})
}

func TestAccResourceSetEmptyStart(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"appendonly": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			// Create with no values.
			{
				Config: `
				resource "appendonly_set" "test" {
					values = []
				}
				`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("appendonly_set.test", "result.#", "0"),
				),
			},
			// Same config — no changes.
			{
				Config: `
				resource "appendonly_set" "test" {
					values = []
				}
				`,
				PlanOnly: true,
			},
			// Add values — result grows from empty.
			{
				Config: `
				resource "appendonly_set" "test" {
					values = ["a", "b"]
				}
				`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectKnownValue(
							"appendonly_set.test",
							tfjsonpath.New("result"),
							knownvalue.SetExact([]knownvalue.Check{
								knownvalue.StringExact("a"),
								knownvalue.StringExact("b"),
							}),
						),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("appendonly_set.test", "result.#", "2"),
					resource.TestCheckTypeSetElemAttr("appendonly_set.test", "result.*", "a"),
					resource.TestCheckTypeSetElemAttr("appendonly_set.test", "result.*", "b"),
				),
			},
		},
	})
}
