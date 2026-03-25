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

// seedEServiceWithDraftDescriptor creates an e-service with both a PUBLISHED and a DRAFT descriptor
// so that attribute groups can be managed on the DRAFT one.
func seedEServiceWithDraftDescriptor(fake *fakepdnd.FakeServer) (eserviceID, draftDescriptorID uuid.UUID) {
	eserviceID = uuid.New()
	publishedDescriptorID := uuid.New()
	draftDescriptorID = uuid.New()
	now := time.Now().UTC()

	fake.SeedEService(fakepdnd.StoredEService{
		ID:          eserviceID,
		ProducerID:  fake.ProducerID(),
		Name:        "Seeded E-Service",
		Description: "An e-service for descriptor attribute tests",
		Technology:  "REST",
		Mode:        "DELIVER",
	})

	fake.SeedDescriptor(fakepdnd.StoredDescriptor{
		ID:                      publishedDescriptorID,
		EServiceID:              eserviceID,
		Version:                 "1",
		State:                   "PUBLISHED",
		AgreementApprovalPolicy: "AUTOMATIC",
		Audience:                []string{"api.test.example.com"},
		DailyCallsPerConsumer:   1000,
		DailyCallsTotal:         10000,
		VoucherLifespan:         3600,
		CreatedAt:               now,
		PublishedAt:             &now,
	})

	fake.SeedDescriptor(fakepdnd.StoredDescriptor{
		ID:                      draftDescriptorID,
		EServiceID:              eserviceID,
		Version:                 "2",
		State:                   "DRAFT",
		AgreementApprovalPolicy: "AUTOMATIC",
		Audience:                []string{"api.test.example.com"},
		DailyCallsPerConsumer:   1000,
		DailyCallsTotal:         10000,
		VoucherLifespan:         3600,
		CreatedAt:               now,
	})

	return eserviceID, draftDescriptorID
}

func TestAccDescriptorCertifiedAttributes_CreateSingleGroup(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	eserviceID, descriptorID := seedEServiceWithDraftDescriptor(fake)

	attrID1 := uuid.New()
	attrID2 := uuid.New()
	fake.SeedCertifiedAttribute(fakepdnd.StoredCertifiedAttribute{
		ID: attrID1, Name: "Attr 1", Description: "Test 1", Code: "A1", Origin: "IPA",
		CreatedAt: time.Now().UTC(),
	})
	fake.SeedCertifiedAttribute(fakepdnd.StoredCertifiedAttribute{
		ID: attrID2, Name: "Attr 2", Description: "Test 2", Code: "A2", Origin: "IPA",
		CreatedAt: time.Now().UTC(),
	})

	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_eservice_descriptor_certified_attributes" "test" {
  eservice_id   = %q
  descriptor_id = %q

  group {
    attribute_ids = [%q, %q]
  }
}
`, eserviceID, descriptorID, attrID1, attrID2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("pdnd_eservice_descriptor_certified_attributes.test", "id"),
					resource.TestCheckResourceAttr("pdnd_eservice_descriptor_certified_attributes.test", "group.#", "1"),
					resource.TestCheckResourceAttr("pdnd_eservice_descriptor_certified_attributes.test", "group.0.attribute_ids.#", "2"),
				),
			},
		},
	})
}

func TestAccDescriptorCertifiedAttributes_CreateMultipleGroups(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	eserviceID, descriptorID := seedEServiceWithDraftDescriptor(fake)

	attrID1 := uuid.New()
	attrID2 := uuid.New()
	attrID3 := uuid.New()
	fake.SeedCertifiedAttribute(fakepdnd.StoredCertifiedAttribute{
		ID: attrID1, Name: "Attr 1", Description: "Test 1", Code: "A1", Origin: "IPA",
		CreatedAt: time.Now().UTC(),
	})
	fake.SeedCertifiedAttribute(fakepdnd.StoredCertifiedAttribute{
		ID: attrID2, Name: "Attr 2", Description: "Test 2", Code: "A2", Origin: "IPA",
		CreatedAt: time.Now().UTC(),
	})
	fake.SeedCertifiedAttribute(fakepdnd.StoredCertifiedAttribute{
		ID: attrID3, Name: "Attr 3", Description: "Test 3", Code: "A3", Origin: "IPA",
		CreatedAt: time.Now().UTC(),
	})

	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_eservice_descriptor_certified_attributes" "test" {
  eservice_id   = %q
  descriptor_id = %q

  group {
    attribute_ids = [%q]
  }
  group {
    attribute_ids = [%q, %q]
  }
}
`, eserviceID, descriptorID, attrID1, attrID2, attrID3),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pdnd_eservice_descriptor_certified_attributes.test", "group.#", "2"),
					resource.TestCheckResourceAttr("pdnd_eservice_descriptor_certified_attributes.test", "group.0.attribute_ids.#", "1"),
					resource.TestCheckResourceAttr("pdnd_eservice_descriptor_certified_attributes.test", "group.1.attribute_ids.#", "2"),
				),
			},
		},
	})
}

func TestAccDescriptorCertifiedAttributes_Update(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	eserviceID, descriptorID := seedEServiceWithDraftDescriptor(fake)

	attrID1 := uuid.New()
	attrID2 := uuid.New()
	attrID3 := uuid.New()
	for i, id := range []uuid.UUID{attrID1, attrID2, attrID3} {
		fake.SeedCertifiedAttribute(fakepdnd.StoredCertifiedAttribute{
			ID: id, Name: fmt.Sprintf("Attr %d", i+1), Description: fmt.Sprintf("Test %d", i+1),
			Code: fmt.Sprintf("A%d", i+1), Origin: "IPA", CreatedAt: time.Now().UTC(),
		})
	}

	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_eservice_descriptor_certified_attributes" "test" {
  eservice_id   = %q
  descriptor_id = %q

  group {
    attribute_ids = [%q]
  }
}
`, eserviceID, descriptorID, attrID1),
				Check: resource.TestCheckResourceAttr("pdnd_eservice_descriptor_certified_attributes.test", "group.#", "1"),
			},
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_eservice_descriptor_certified_attributes" "test" {
  eservice_id   = %q
  descriptor_id = %q

  group {
    attribute_ids = [%q]
  }
  group {
    attribute_ids = [%q, %q]
  }
}
`, eserviceID, descriptorID, attrID1, attrID2, attrID3),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pdnd_eservice_descriptor_certified_attributes.test", "group.#", "2"),
					resource.TestCheckResourceAttr("pdnd_eservice_descriptor_certified_attributes.test", "group.1.attribute_ids.#", "2"),
				),
			},
		},
	})
}

func TestAccDescriptorCertifiedAttributes_Delete(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	eserviceID, descriptorID := seedEServiceWithDraftDescriptor(fake)

	attrID := uuid.New()
	fake.SeedCertifiedAttribute(fakepdnd.StoredCertifiedAttribute{
		ID: attrID, Name: "Attr 1", Description: "Test 1", Code: "A1", Origin: "IPA",
		CreatedAt: time.Now().UTC(),
	})

	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy: func(s *terraform.State) error {
			groups := fake.GetDescriptorAttributeGroups(eserviceID, descriptorID, "certified")
			for _, g := range groups {
				if len(g.Attributes) > 0 {
					return fmt.Errorf("expected all attributes to be removed, but group still has %d", len(g.Attributes))
				}
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_eservice_descriptor_certified_attributes" "test" {
  eservice_id   = %q
  descriptor_id = %q

  group {
    attribute_ids = [%q]
  }
}
`, eserviceID, descriptorID, attrID),
				Check: resource.TestCheckResourceAttr("pdnd_eservice_descriptor_certified_attributes.test", "group.#", "1"),
			},
		},
	})
}

func TestAccDescriptorCertifiedAttributes_Import(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	eserviceID, descriptorID := seedEServiceWithDraftDescriptor(fake)

	attrID := uuid.New()
	fake.SeedCertifiedAttribute(fakepdnd.StoredCertifiedAttribute{
		ID: attrID, Name: "Attr 1", Description: "Test 1", Code: "A1", Origin: "IPA",
		CreatedAt: time.Now().UTC(),
	})

	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_eservice_descriptor_certified_attributes" "test" {
  eservice_id   = %q
  descriptor_id = %q

  group {
    attribute_ids = [%q]
  }
}
`, eserviceID, descriptorID, attrID),
			},
			{
				ResourceName:      "pdnd_eservice_descriptor_certified_attributes.test",
				ImportState:       true,
				ImportStateId:     fmt.Sprintf("%s/%s", eserviceID, descriptorID),
				ImportStateVerify: true,
			},
		},
	})
}
