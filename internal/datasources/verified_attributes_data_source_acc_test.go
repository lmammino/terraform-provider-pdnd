package datasources_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/lmammino/terraform-provider-pdnd/internal/testing/fakepdnd"
)

func TestAccVerifiedAttributesDataSource(t *testing.T) {
	fake := fakepdnd.NewFakeServer()

	createdAt := time.Date(2025, 3, 10, 8, 45, 0, 0, time.UTC)

	fake.SeedVerifiedAttribute(fakepdnd.StoredVerifiedAttribute{
		ID:          uuid.New(),
		Name:        "PagoPA Verified",
		Description: "Verified attribute for PagoPA",
		CreatedAt:   createdAt,
	})

	fake.SeedVerifiedAttribute(fakepdnd.StoredVerifiedAttribute{
		ID:          uuid.New(),
		Name:        "PagoPA Advanced",
		Description: "Verified attribute for PagoPA advanced",
		CreatedAt:   createdAt,
	})

	fake.SeedVerifiedAttribute(fakepdnd.StoredVerifiedAttribute{
		ID:          uuid.New(),
		Name:        "SPID Level 2",
		Description: "Verified attribute for SPID Level 2",
		CreatedAt:   createdAt,
	})

	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + `
data "pdnd_verified_attributes" "test" {
  name = "PagoPA"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.pdnd_verified_attributes.test", "attributes.#", "2"),
				),
			},
		},
	})
}
