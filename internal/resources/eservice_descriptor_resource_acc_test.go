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

// seedEServiceWithPublishedDescriptor creates an e-service with a PUBLISHED descriptor
// so that new DRAFT descriptors can be created (the API forbids having two DRAFT descriptors).
func seedEServiceWithPublishedDescriptor(fake *fakepdnd.FakeServer) (eserviceID uuid.UUID, publishedDescriptorID uuid.UUID) {
	eserviceID = uuid.New()
	publishedDescriptorID = uuid.New()
	now := time.Now().UTC()

	fake.SeedEService(fakepdnd.StoredEService{
		ID:          eserviceID,
		ProducerID:  fake.ProducerID(),
		Name:        "Seeded E-Service",
		Description: "A seeded e-service for descriptor tests",
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

	return eserviceID, publishedDescriptorID
}

func TestAccEServiceDescriptor_CreateDraft(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	eserviceID, _ := seedEServiceWithPublishedDescriptor(fake)
	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_eservice_descriptor" "test" {
  eservice_id               = %q
  desired_state             = "DRAFT"
  agreement_approval_policy = "AUTOMATIC"
  audience                  = ["api.test.example.com"]
  daily_calls_per_consumer  = 1000
  daily_calls_total         = 10000
  voucher_lifespan          = 3600
}
`, eserviceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("pdnd_eservice_descriptor.test", "id"),
					resource.TestCheckResourceAttr("pdnd_eservice_descriptor.test", "state", "DRAFT"),
					resource.TestCheckResourceAttrSet("pdnd_eservice_descriptor.test", "version"),
				),
			},
		},
	})
}

func TestAccEServiceDescriptor_CreatePublished(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	eserviceID, _ := seedEServiceWithPublishedDescriptor(fake)
	ts := fake.Start()
	defer ts.Close()

	// descriptorIDTracker to force cleanup
	tracker := newDescriptorIDTracker(fake, eserviceID)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_eservice_descriptor" "test" {
  eservice_id               = %q
  desired_state             = "PUBLISHED"
  agreement_approval_policy = "AUTOMATIC"
  audience                  = ["api.test.example.com"]
  daily_calls_per_consumer  = 1000
  daily_calls_total         = 10000
  voucher_lifespan          = 3600
}
`, eserviceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pdnd_eservice_descriptor.test", "state", "PUBLISHED"),
					tracker.track(),
				),
			},
			tracker.cleanupStep(testAccProviderConfig(ts.URL)),
		},
	})
}

func TestAccEServiceDescriptor_PublishToSuspend(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	eserviceID, _ := seedEServiceWithPublishedDescriptor(fake)
	ts := fake.Start()
	defer ts.Close()

	tracker := newDescriptorIDTracker(fake, eserviceID)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_eservice_descriptor" "test" {
  eservice_id               = %q
  desired_state             = "PUBLISHED"
  agreement_approval_policy = "AUTOMATIC"
  audience                  = ["api.test.example.com"]
  daily_calls_per_consumer  = 1000
  daily_calls_total         = 10000
  voucher_lifespan          = 3600
}
`, eserviceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pdnd_eservice_descriptor.test", "state", "PUBLISHED"),
					tracker.track(),
				),
			},
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_eservice_descriptor" "test" {
  eservice_id               = %q
  desired_state             = "SUSPENDED"
  agreement_approval_policy = "AUTOMATIC"
  audience                  = ["api.test.example.com"]
  daily_calls_per_consumer  = 1000
  daily_calls_total         = 10000
  voucher_lifespan          = 3600
}
`, eserviceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pdnd_eservice_descriptor.test", "state", "SUSPENDED"),
					tracker.track(),
				),
			},
			tracker.cleanupStep(testAccProviderConfig(ts.URL)),
		},
	})
}

func TestAccEServiceDescriptor_SuspendToPublish(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	eserviceID, _ := seedEServiceWithPublishedDescriptor(fake)
	ts := fake.Start()
	defer ts.Close()

	tracker := newDescriptorIDTracker(fake, eserviceID)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_eservice_descriptor" "test" {
  eservice_id               = %q
  desired_state             = "PUBLISHED"
  agreement_approval_policy = "AUTOMATIC"
  audience                  = ["api.test.example.com"]
  daily_calls_per_consumer  = 1000
  daily_calls_total         = 10000
  voucher_lifespan          = 3600
}
`, eserviceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pdnd_eservice_descriptor.test", "state", "PUBLISHED"),
					tracker.track(),
				),
			},
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_eservice_descriptor" "test" {
  eservice_id               = %q
  desired_state             = "SUSPENDED"
  agreement_approval_policy = "AUTOMATIC"
  audience                  = ["api.test.example.com"]
  daily_calls_per_consumer  = 1000
  daily_calls_total         = 10000
  voucher_lifespan          = 3600
}
`, eserviceID),
				Check: resource.TestCheckResourceAttr("pdnd_eservice_descriptor.test", "state", "SUSPENDED"),
			},
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_eservice_descriptor" "test" {
  eservice_id               = %q
  desired_state             = "PUBLISHED"
  agreement_approval_policy = "AUTOMATIC"
  audience                  = ["api.test.example.com"]
  daily_calls_per_consumer  = 1000
  daily_calls_total         = 10000
  voucher_lifespan          = 3600
}
`, eserviceID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pdnd_eservice_descriptor.test", "state", "PUBLISHED"),
					tracker.track(),
				),
			},
			tracker.cleanupStep(testAccProviderConfig(ts.URL)),
		},
	})
}

func TestAccEServiceDescriptor_Delete(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	eserviceID, _ := seedEServiceWithPublishedDescriptor(fake)
	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy: func(s *terraform.State) error {
			for _, rs := range s.RootModule().Resources {
				if rs.Type != "pdnd_eservice_descriptor" {
					continue
				}
				descID, err := uuid.Parse(rs.Primary.ID)
				if err != nil {
					return fmt.Errorf("invalid descriptor ID %q: %w", rs.Primary.ID, err)
				}
				if fake.GetDescriptor(eserviceID, descID) != nil {
					return fmt.Errorf("descriptor %s still exists", rs.Primary.ID)
				}
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_eservice_descriptor" "test" {
  eservice_id               = %q
  desired_state             = "DRAFT"
  agreement_approval_policy = "AUTOMATIC"
  audience                  = ["api.test.example.com"]
  daily_calls_per_consumer  = 1000
  daily_calls_total         = 10000
  voucher_lifespan          = 3600
}
`, eserviceID),
				Check: resource.TestCheckResourceAttr("pdnd_eservice_descriptor.test", "state", "DRAFT"),
			},
		},
	})
}

func TestAccEServiceDescriptor_Import(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	eserviceID, _ := seedEServiceWithPublishedDescriptor(fake)
	ts := fake.Start()
	defer ts.Close()

	tracker := newDescriptorIDTracker(fake, eserviceID)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_eservice_descriptor" "test" {
  eservice_id               = %q
  desired_state             = "PUBLISHED"
  agreement_approval_policy = "AUTOMATIC"
  audience                  = ["api.test.example.com"]
  daily_calls_per_consumer  = 1000
  daily_calls_total         = 10000
  voucher_lifespan          = 3600
}
`, eserviceID),
				Check: tracker.track(),
			},
			{
				ResourceName:  "pdnd_eservice_descriptor.test",
				ImportState:   true,
				ImportStateIdFunc: func(s *terraform.State) (string, error) {
					for _, rs := range s.RootModule().Resources {
						if rs.Type == "pdnd_eservice_descriptor" {
							return fmt.Sprintf("%s/%s", rs.Primary.Attributes["eservice_id"], rs.Primary.ID), nil
						}
					}
					return "", fmt.Errorf("pdnd_eservice_descriptor not found in state")
				},
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"desired_state", "allow_waiting_for_approval"},
			},
			tracker.cleanupStep(testAccProviderConfig(ts.URL)),
		},
	})
}

// descriptorIDTracker tracks descriptor IDs for cleanup.
type descriptorIDTracker struct {
	ids        []uuid.UUID
	fake       *fakepdnd.FakeServer
	eserviceID uuid.UUID
}

func newDescriptorIDTracker(fake *fakepdnd.FakeServer, eserviceID uuid.UUID) *descriptorIDTracker {
	return &descriptorIDTracker{fake: fake, eserviceID: eserviceID}
}

func (t *descriptorIDTracker) track() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "pdnd_eservice_descriptor" {
				continue
			}
			id, err := uuid.Parse(rs.Primary.ID)
			if err != nil {
				continue
			}
			t.ids = append(t.ids, id)
		}
		return nil
	}
}

// cleanupStep forces tracked descriptors back to DRAFT, then refreshes state.
func (t *descriptorIDTracker) cleanupStep(providerConfig string) resource.TestStep {
	return resource.TestStep{
		RefreshState: true,
		PreConfig: func() {
			for _, id := range t.ids {
				d := t.fake.GetDescriptor(t.eserviceID, id)
				if d != nil {
					t.fake.SeedDescriptor(fakepdnd.StoredDescriptor{
						ID:                      d.ID,
						EServiceID:              d.EServiceID,
						Version:                 d.Version,
						State:                   "DRAFT",
						AgreementApprovalPolicy: d.AgreementApprovalPolicy,
						Audience:                d.Audience,
						DailyCallsPerConsumer:   d.DailyCallsPerConsumer,
						DailyCallsTotal:         d.DailyCallsTotal,
						VoucherLifespan:         d.VoucherLifespan,
						ServerUrls:              d.ServerUrls,
						Description:             d.Description,
						CreatedAt:               d.CreatedAt,
					})
				}
			}
		},
		ExpectNonEmptyPlan: true,
	}
}
