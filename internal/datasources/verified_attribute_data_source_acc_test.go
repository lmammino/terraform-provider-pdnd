package datasources_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/lmammino/terraform-provider-pdnd/internal/testing/fakepdnd"
)

func TestAccVerifiedAttributeDataSource(t *testing.T) {
	fake := fakepdnd.NewFakeServer()

	attrID := uuid.New()
	createdAt := time.Date(2025, 3, 10, 8, 45, 0, 0, time.UTC)

	fake.SeedVerifiedAttribute(fakepdnd.StoredVerifiedAttribute{
		ID:          attrID,
		Name:        "PagoPA Verified",
		Description: "Verified attribute for PagoPA integration",
		CreatedAt:   createdAt,
	})

	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
data "pdnd_verified_attribute" "test" {
  id = %q
}
`, attrID.String()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.pdnd_verified_attribute.test", "id", attrID.String()),
					resource.TestCheckResourceAttr("data.pdnd_verified_attribute.test", "name", "PagoPA Verified"),
					resource.TestCheckResourceAttr("data.pdnd_verified_attribute.test", "description", "Verified attribute for PagoPA integration"),
					resource.TestCheckResourceAttr("data.pdnd_verified_attribute.test", "created_at", "2025-03-10T08:45:00Z"),
				),
			},
		},
	})
}
