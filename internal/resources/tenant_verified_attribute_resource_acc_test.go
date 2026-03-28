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

func TestAccTenantVerifiedAttribute_Create(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	tenantID := uuid.New()
	attrID := uuid.New()
	agreementID := uuid.New()
	fake.SeedTenant(fakepdnd.StoredTenant{
		ID: tenantID, Name: "Test Tenant", Kind: "PA",
		ExternalOrigin: "IPA", ExternalValue: "test123",
		CreatedAt: time.Now().UTC(),
	})
	fake.SeedVerifiedAttribute(fakepdnd.StoredVerifiedAttribute{
		ID: attrID, Name: "Test Verified Attr", Description: "test",
		CreatedAt: time.Now().UTC(),
	})

	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_tenant_verified_attribute" "test" {
  tenant_id    = %q
  attribute_id = %q
  agreement_id = %q
}
`, tenantID, attrID, agreementID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("pdnd_tenant_verified_attribute.test", "assigned_at"),
					resource.TestCheckResourceAttr("pdnd_tenant_verified_attribute.test", "id", fmt.Sprintf("%s/%s", tenantID, attrID)),
				),
			},
		},
	})
}

func TestAccTenantVerifiedAttribute_Delete(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	tenantID := uuid.New()
	attrID := uuid.New()
	agreementID := uuid.New()
	fake.SeedTenant(fakepdnd.StoredTenant{
		ID: tenantID, Name: "Test Tenant", Kind: "PA",
		ExternalOrigin: "IPA", ExternalValue: "test123",
		CreatedAt: time.Now().UTC(),
	})
	fake.SeedVerifiedAttribute(fakepdnd.StoredVerifiedAttribute{
		ID: attrID, Name: "Test Verified Attr", Description: "test",
		CreatedAt: time.Now().UTC(),
	})

	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy: func(s *terraform.State) error {
			attrs := fake.GetTenantVerifiedAttrs(tenantID)
			for _, a := range attrs {
				if a.ID == attrID {
					return fmt.Errorf("expected verified attribute %s to be removed", attrID)
				}
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_tenant_verified_attribute" "test" {
  tenant_id    = %q
  attribute_id = %q
  agreement_id = %q
}
`, tenantID, attrID, agreementID),
				Check: resource.TestCheckResourceAttrSet("pdnd_tenant_verified_attribute.test", "assigned_at"),
			},
		},
	})
}

func TestAccTenantVerifiedAttribute_Import(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	tenantID := uuid.New()
	attrID := uuid.New()
	agreementID := uuid.New()
	fake.SeedTenant(fakepdnd.StoredTenant{
		ID: tenantID, Name: "Test Tenant", Kind: "PA",
		ExternalOrigin: "IPA", ExternalValue: "test123",
		CreatedAt: time.Now().UTC(),
	})
	fake.SeedVerifiedAttribute(fakepdnd.StoredVerifiedAttribute{
		ID: attrID, Name: "Test Verified Attr", Description: "test",
		CreatedAt: time.Now().UTC(),
	})

	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_tenant_verified_attribute" "test" {
  tenant_id    = %q
  attribute_id = %q
  agreement_id = %q
}
`, tenantID, attrID, agreementID),
			},
			{
				ResourceName: "pdnd_tenant_verified_attribute.test",
				ImportState:  true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					for _, rs := range s.RootModule().Resources {
						if rs.Type == "pdnd_tenant_verified_attribute" {
							return fmt.Sprintf("%s/%s",
								rs.Primary.Attributes["tenant_id"],
								rs.Primary.Attributes["attribute_id"]), nil
						}
					}
					return "", fmt.Errorf("resource not found")
				},
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"agreement_id", "expiration_date"},
			},
		},
	})
}
