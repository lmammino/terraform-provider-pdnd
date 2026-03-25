package resources_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/lmammino/terraform-provider-pdnd/internal/testing/fakepdnd"
)

func TestAccDescriptorDeclaredAttributes_CreateAndUpdate(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	eserviceID, descriptorID := seedEServiceWithDraftDescriptor(fake)

	attrID1 := uuid.New()
	attrID2 := uuid.New()
	fake.SeedDeclaredAttribute(fakepdnd.StoredDeclaredAttribute{
		ID: attrID1, Name: "Declared 1", Description: "Test 1", CreatedAt: time.Now().UTC(),
	})
	fake.SeedDeclaredAttribute(fakepdnd.StoredDeclaredAttribute{
		ID: attrID2, Name: "Declared 2", Description: "Test 2", CreatedAt: time.Now().UTC(),
	})

	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_eservice_descriptor_declared_attributes" "test" {
  eservice_id   = %q
  descriptor_id = %q

  group {
    attribute_ids = [%q]
  }
}
`, eserviceID, descriptorID, attrID1),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("pdnd_eservice_descriptor_declared_attributes.test", "id"),
					resource.TestCheckResourceAttr("pdnd_eservice_descriptor_declared_attributes.test", "group.#", "1"),
				),
			},
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_eservice_descriptor_declared_attributes" "test" {
  eservice_id   = %q
  descriptor_id = %q

  group {
    attribute_ids = [%q, %q]
  }
}
`, eserviceID, descriptorID, attrID1, attrID2),
				Check: resource.TestCheckResourceAttr("pdnd_eservice_descriptor_declared_attributes.test", "group.0.attribute_ids.#", "2"),
			},
		},
	})
}

func TestAccDescriptorDeclaredAttributes_Import(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	eserviceID, descriptorID := seedEServiceWithDraftDescriptor(fake)

	attrID := uuid.New()
	fake.SeedDeclaredAttribute(fakepdnd.StoredDeclaredAttribute{
		ID: attrID, Name: "Declared 1", Description: "Test 1", CreatedAt: time.Now().UTC(),
	})

	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_eservice_descriptor_declared_attributes" "test" {
  eservice_id   = %q
  descriptor_id = %q

  group {
    attribute_ids = [%q]
  }
}
`, eserviceID, descriptorID, attrID),
			},
			{
				ResourceName:      "pdnd_eservice_descriptor_declared_attributes.test",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/%s", eserviceID, descriptorID),
				ImportStateVerify: true,
			},
		},
	})
}
