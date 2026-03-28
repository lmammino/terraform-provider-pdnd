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

func TestAccTenantCertifiedAttribute_Create(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	tenantID := uuid.New()
	attrID := uuid.New()
	fake.SeedTenant(fakepdnd.StoredTenant{
		ID: tenantID, Name: "Test Tenant", Kind: "PA",
		ExternalOrigin: "IPA", ExternalValue: "test123",
		CreatedAt: time.Now().UTC(),
	})
	fake.SeedCertifiedAttribute(fakepdnd.StoredCertifiedAttribute{
		ID: attrID, Name: "Test Attr", Description: "test",
		Code: "TA1", Origin: "IPA", CreatedAt: time.Now().UTC(),
	})

	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_tenant_certified_attribute" "test" {
  tenant_id    = %q
  attribute_id = %q
}
`, tenantID, attrID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("pdnd_tenant_certified_attribute.test", "assigned_at"),
					resource.TestCheckResourceAttr("pdnd_tenant_certified_attribute.test", "id", fmt.Sprintf("%s/%s", tenantID, attrID)),
				),
			},
		},
	})
}

func TestAccTenantCertifiedAttribute_Delete(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	tenantID := uuid.New()
	attrID := uuid.New()
	fake.SeedTenant(fakepdnd.StoredTenant{
		ID: tenantID, Name: "Test Tenant", Kind: "PA",
		ExternalOrigin: "IPA", ExternalValue: "test123",
		CreatedAt: time.Now().UTC(),
	})
	fake.SeedCertifiedAttribute(fakepdnd.StoredCertifiedAttribute{
		ID: attrID, Name: "Test Attr", Description: "test",
		Code: "TA1", Origin: "IPA", CreatedAt: time.Now().UTC(),
	})

	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy: func(s *terraform.State) error {
			attrs := fake.GetTenantCertifiedAttrs(tenantID)
			for _, a := range attrs {
				if a.ID == attrID && a.RevokedAt == nil {
					return fmt.Errorf("expected certified attribute %s to be revoked", attrID)
				}
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_tenant_certified_attribute" "test" {
  tenant_id    = %q
  attribute_id = %q
}
`, tenantID, attrID),
				Check: resource.TestCheckResourceAttrSet("pdnd_tenant_certified_attribute.test", "assigned_at"),
			},
		},
	})
}

func TestAccTenantCertifiedAttribute_Import(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	tenantID := uuid.New()
	attrID := uuid.New()
	fake.SeedTenant(fakepdnd.StoredTenant{
		ID: tenantID, Name: "Test Tenant", Kind: "PA",
		ExternalOrigin: "IPA", ExternalValue: "test123",
		CreatedAt: time.Now().UTC(),
	})
	fake.SeedCertifiedAttribute(fakepdnd.StoredCertifiedAttribute{
		ID: attrID, Name: "Test Attr", Description: "test",
		Code: "TA1", Origin: "IPA", CreatedAt: time.Now().UTC(),
	})

	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_tenant_certified_attribute" "test" {
  tenant_id    = %q
  attribute_id = %q
}
`, tenantID, attrID),
			},
			{
				ResourceName: "pdnd_tenant_certified_attribute.test",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					for _, rs := range s.RootModule().Resources {
						if rs.Type == "pdnd_tenant_certified_attribute" {
							return fmt.Sprintf("%s/%s",
								rs.Primary.Attributes["tenant_id"],
								rs.Primary.Attributes["attribute_id"]), nil
						}
					}
					return "", fmt.Errorf("resource not found")
				},
				ImportStateVerify: true,
			},
		},
	})
}
