package resources_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/lmammino/terraform-provider-pdnd/internal/testing/fakepdnd"
)

func TestAccTenantDeclaredAttribute_Create(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	tenantID := uuid.New()
	attrID := uuid.New()
	delegationID := uuid.New()
	fake.SeedTenant(fakepdnd.StoredTenant{
		ID: tenantID, Name: "Test Tenant", Kind: "PA",
		ExternalOrigin: "IPA", ExternalValue: "test123",
		CreatedAt: time.Now().UTC(),
	})
	fake.SeedDeclaredAttribute(fakepdnd.StoredDeclaredAttribute{
		ID: attrID, Name: "Test Declared Attr", Description: "test",
		CreatedAt: time.Now().UTC(),
	})

	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_tenant_declared_attribute" "test" {
  tenant_id     = %q
  attribute_id  = %q
  delegation_id = %q
}
`, tenantID, attrID, delegationID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("pdnd_tenant_declared_attribute.test", "assigned_at"),
					resource.TestCheckResourceAttr("pdnd_tenant_declared_attribute.test", "id", fmt.Sprintf("%s/%s", tenantID, attrID)),
					resource.TestCheckResourceAttr("pdnd_tenant_declared_attribute.test", "delegation_id", delegationID.String()),
				),
			},
		},
	})
}

func TestAccTenantDeclaredAttribute_Delete(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	tenantID := uuid.New()
	attrID := uuid.New()
	fake.SeedTenant(fakepdnd.StoredTenant{
		ID: tenantID, Name: "Test Tenant", Kind: "PA",
		ExternalOrigin: "IPA", ExternalValue: "test123",
		CreatedAt: time.Now().UTC(),
	})
	fake.SeedDeclaredAttribute(fakepdnd.StoredDeclaredAttribute{
		ID: attrID, Name: "Test Declared Attr", Description: "test",
		CreatedAt: time.Now().UTC(),
	})

	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy: func(s *terraform.State) error {
			attrs := fake.GetTenantDeclaredAttrs(tenantID)
			for _, a := range attrs {
				if a.ID == attrID && a.RevokedAt == nil {
					return fmt.Errorf("expected declared attribute %s to be revoked", attrID)
				}
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_tenant_declared_attribute" "test" {
  tenant_id    = %q
  attribute_id = %q
}
`, tenantID, attrID),
				Check: resource.TestCheckResourceAttrSet("pdnd_tenant_declared_attribute.test", "assigned_at"),
			},
		},
	})
}
