package datasources_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/lmammino/terraform-provider-pdnd/internal/testing/fakepdnd"
)

func TestAccDeclaredAttributesDataSource(t *testing.T) {
	fake := fakepdnd.NewFakeServer()

	createdAt := time.Date(2025, 2, 20, 14, 0, 0, 0, time.UTC)

	fake.SeedDeclaredAttribute(fakepdnd.StoredDeclaredAttribute{
		ID:          uuid.New(),
		Name:        "ISO 27001",
		Description: "Declared attribute for ISO 27001",
		CreatedAt:   createdAt,
	})

	fake.SeedDeclaredAttribute(fakepdnd.StoredDeclaredAttribute{
		ID:          uuid.New(),
		Name:        "ISO 9001",
		Description: "Declared attribute for ISO 9001",
		CreatedAt:   createdAt,
	})

	fake.SeedDeclaredAttribute(fakepdnd.StoredDeclaredAttribute{
		ID:          uuid.New(),
		Name:        "GDPR Compliance",
		Description: "Declared attribute for GDPR compliance",
		CreatedAt:   createdAt,
	})

	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + `
data "pdnd_declared_attributes" "test" {
  name = "ISO"
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.pdnd_declared_attributes.test", "attributes.#", "2"),
				),
			},
		},
	})
}
