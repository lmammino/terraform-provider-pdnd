package datasources_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/lmammino/terraform-provider-pdnd/internal/testing/fakepdnd"
)

func TestAccDeclaredAttributeDataSource(t *testing.T) {
	fake := fakepdnd.NewFakeServer()

	attrID := uuid.New()
	createdAt := time.Date(2025, 2, 20, 14, 0, 0, 0, time.UTC)

	fake.SeedDeclaredAttribute(fakepdnd.StoredDeclaredAttribute{
		ID:          attrID,
		Name:        "ISO 27001",
		Description: "Declared attribute for ISO 27001 certification",
		CreatedAt:   createdAt,
	})

	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
data "pdnd_declared_attribute" "test" {
  id = %q
}
`, attrID.String()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.pdnd_declared_attribute.test", "id", attrID.String()),
					resource.TestCheckResourceAttr("data.pdnd_declared_attribute.test", "name", "ISO 27001"),
					resource.TestCheckResourceAttr("data.pdnd_declared_attribute.test", "description", "Declared attribute for ISO 27001 certification"),
					resource.TestCheckResourceAttr("data.pdnd_declared_attribute.test", "created_at", "2025-02-20T14:00:00Z"),
				),
			},
		},
	})
}
