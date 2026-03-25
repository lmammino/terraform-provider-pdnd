package datasources_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/lmammino/terraform-provider-pdnd/internal/testing/fakepdnd"
)

func TestAccCertifiedAttributeDataSource(t *testing.T) {
	fake := fakepdnd.NewFakeServer()

	attrID := uuid.New()
	createdAt := time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC)

	fake.SeedCertifiedAttribute(fakepdnd.StoredCertifiedAttribute{
		ID:          attrID,
		Name:        "Comuni",
		Description: "Certified attribute for Comuni",
		Code:        "L6_COMUNI",
		Origin:      "IPA",
		CreatedAt:   createdAt,
	})

	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
data "pdnd_certified_attribute" "test" {
  id = %q
}
`, attrID.String()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.pdnd_certified_attribute.test", "id", attrID.String()),
					resource.TestCheckResourceAttr("data.pdnd_certified_attribute.test", "name", "Comuni"),
					resource.TestCheckResourceAttr("data.pdnd_certified_attribute.test", "description", "Certified attribute for Comuni"),
					resource.TestCheckResourceAttr("data.pdnd_certified_attribute.test", "code", "L6_COMUNI"),
					resource.TestCheckResourceAttr("data.pdnd_certified_attribute.test", "origin", "IPA"),
					resource.TestCheckResourceAttr("data.pdnd_certified_attribute.test", "created_at", "2025-01-15T10:30:00Z"),
				),
			},
		},
	})
}
