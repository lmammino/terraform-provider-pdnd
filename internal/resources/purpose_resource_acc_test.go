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

func TestAccPurpose_CreateDraft(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	ts := fake.Start()
	defer ts.Close()

	esID := uuid.New()
	fake.SeedEService(fakepdnd.StoredEService{
		ID: esID, ProducerID: fake.ProducerID(), Name: "Test EService",
		Description: "test", Technology: "REST", Mode: "DELIVER",
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_purpose" "test" {
  eservice_id      = %q
  title            = "Test Purpose Draft"
  description      = "A test purpose in draft state"
  daily_calls      = 1000
  is_free_of_charge = false
  desired_state    = "DRAFT"
}
`, esID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("pdnd_purpose.test", "id"),
					resource.TestCheckResourceAttr("pdnd_purpose.test", "state", "DRAFT"),
					resource.TestCheckResourceAttr("pdnd_purpose.test", "title", "Test Purpose Draft"),
					resource.TestCheckResourceAttr("pdnd_purpose.test", "daily_calls", "1000"),
				),
			},
		},
	})
}

func TestAccPurpose_CreateActive(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	ts := fake.Start()
	defer ts.Close()

	esID := uuid.New()
	fake.SeedEService(fakepdnd.StoredEService{
		ID: esID, ProducerID: fake.ProducerID(), Name: "Test EService",
		Description: "test", Technology: "REST", Mode: "DELIVER",
	})

	tracker := newPurposeIDTracker(fake)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_purpose" "test" {
  eservice_id      = %q
  title            = "Test Purpose Active"
  description      = "A test purpose in active state"
  daily_calls      = 1000
  is_free_of_charge = false
  desired_state    = "ACTIVE"
}
`, esID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pdnd_purpose.test", "state", "ACTIVE"),
					resource.TestCheckResourceAttrSet("pdnd_purpose.test", "version_id"),
					tracker.track(),
				),
			},
			tracker.cleanupStep(testAccProviderConfig(ts.URL)),
		},
	})
}

func TestAccPurpose_CreateActive_WaitingForApproval_Allowed(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	fake.SetPurposeApprovalThreshold(500) // Low threshold
	ts := fake.Start()
	defer ts.Close()

	esID := uuid.New()
	fake.SeedEService(fakepdnd.StoredEService{
		ID: esID, ProducerID: fake.ProducerID(), Name: "Test EService",
		Description: "test", Technology: "REST", Mode: "DELIVER",
	})

	tracker := newPurposeIDTracker(fake)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_purpose" "test" {
  eservice_id                = %q
  title                      = "High Calls Purpose"
  description                = "A purpose that exceeds threshold"
  daily_calls                = 1000
  is_free_of_charge          = false
  desired_state              = "ACTIVE"
  allow_waiting_for_approval = true
}
`, esID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pdnd_purpose.test", "state", "WAITING_FOR_APPROVAL"),
					tracker.track(),
				),
			},
			tracker.cleanupStep(testAccProviderConfig(ts.URL)),
		},
	})
}

func TestAccPurpose_ActiveToSuspended(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	ts := fake.Start()
	defer ts.Close()

	esID := uuid.New()
	fake.SeedEService(fakepdnd.StoredEService{
		ID: esID, ProducerID: fake.ProducerID(), Name: "Test EService",
		Description: "test", Technology: "REST", Mode: "DELIVER",
	})

	tracker := newPurposeIDTracker(fake)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_purpose" "test" {
  eservice_id      = %q
  title            = "Test Purpose"
  description      = "A test purpose for state transitions"
  daily_calls      = 1000
  is_free_of_charge = false
  desired_state    = "ACTIVE"
}
`, esID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pdnd_purpose.test", "state", "ACTIVE"),
					tracker.track(),
				),
			},
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_purpose" "test" {
  eservice_id      = %q
  title            = "Test Purpose"
  description      = "A test purpose for state transitions"
  daily_calls      = 1000
  is_free_of_charge = false
  desired_state    = "SUSPENDED"
}
`, esID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pdnd_purpose.test", "state", "SUSPENDED"),
					tracker.track(),
				),
			},
			tracker.cleanupStep(testAccProviderConfig(ts.URL)),
		},
	})
}

func TestAccPurpose_SuspendedToActive(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	ts := fake.Start()
	defer ts.Close()

	esID := uuid.New()
	fake.SeedEService(fakepdnd.StoredEService{
		ID: esID, ProducerID: fake.ProducerID(), Name: "Test EService",
		Description: "test", Technology: "REST", Mode: "DELIVER",
	})

	tracker := newPurposeIDTracker(fake)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_purpose" "test" {
  eservice_id      = %q
  title            = "Test Purpose"
  description      = "A test purpose for state transitions"
  daily_calls      = 1000
  is_free_of_charge = false
  desired_state    = "ACTIVE"
}
`, esID),
				Check: resource.TestCheckResourceAttr("pdnd_purpose.test", "state", "ACTIVE"),
			},
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_purpose" "test" {
  eservice_id      = %q
  title            = "Test Purpose"
  description      = "A test purpose for state transitions"
  daily_calls      = 1000
  is_free_of_charge = false
  desired_state    = "SUSPENDED"
}
`, esID),
				Check: resource.TestCheckResourceAttr("pdnd_purpose.test", "state", "SUSPENDED"),
			},
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_purpose" "test" {
  eservice_id      = %q
  title            = "Test Purpose"
  description      = "A test purpose for state transitions"
  daily_calls      = 1000
  is_free_of_charge = false
  desired_state    = "ACTIVE"
}
`, esID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pdnd_purpose.test", "state", "ACTIVE"),
					tracker.track(),
				),
			},
			tracker.cleanupStep(testAccProviderConfig(ts.URL)),
		},
	})
}

func TestAccPurpose_UpdateDraftFields(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	ts := fake.Start()
	defer ts.Close()

	esID := uuid.New()
	fake.SeedEService(fakepdnd.StoredEService{
		ID: esID, ProducerID: fake.ProducerID(), Name: "Test EService",
		Description: "test", Technology: "REST", Mode: "DELIVER",
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_purpose" "test" {
  eservice_id      = %q
  title            = "Original Title"
  description      = "Original description for test"
  daily_calls      = 1000
  is_free_of_charge = false
  desired_state    = "DRAFT"
}
`, esID),
				Check: resource.TestCheckResourceAttr("pdnd_purpose.test", "title", "Original Title"),
			},
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_purpose" "test" {
  eservice_id      = %q
  title            = "Updated Title"
  description      = "Updated description for test"
  daily_calls      = 2000
  is_free_of_charge = false
  desired_state    = "DRAFT"
}
`, esID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pdnd_purpose.test", "title", "Updated Title"),
					resource.TestCheckResourceAttr("pdnd_purpose.test", "daily_calls", "2000"),
				),
			},
		},
	})
}

func TestAccPurpose_Delete(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	ts := fake.Start()
	defer ts.Close()

	esID := uuid.New()
	fake.SeedEService(fakepdnd.StoredEService{
		ID: esID, ProducerID: fake.ProducerID(), Name: "Test EService",
		Description: "test", Technology: "REST", Mode: "DELIVER",
	})

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy: func(s *terraform.State) error {
			for _, rs := range s.RootModule().Resources {
				if rs.Type != "pdnd_purpose" {
					continue
				}
				id, err := uuid.Parse(rs.Primary.ID)
				if err != nil {
					return fmt.Errorf("invalid purpose ID %q: %w", rs.Primary.ID, err)
				}
				if fake.GetStandalonePurpose(id) != nil {
					return fmt.Errorf("purpose %s still exists", rs.Primary.ID)
				}
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_purpose" "test" {
  eservice_id      = %q
  title            = "Delete Test"
  description      = "A purpose to be deleted"
  daily_calls      = 1000
  is_free_of_charge = false
  desired_state    = "DRAFT"
}
`, esID),
				Check: resource.TestCheckResourceAttr("pdnd_purpose.test", "state", "DRAFT"),
			},
		},
	})
}

func TestAccPurpose_Import(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	ts := fake.Start()
	defer ts.Close()

	esID := uuid.New()
	fake.SeedEService(fakepdnd.StoredEService{
		ID: esID, ProducerID: fake.ProducerID(), Name: "Test EService",
		Description: "test", Technology: "REST", Mode: "DELIVER",
	})

	tracker := newPurposeIDTracker(fake)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_purpose" "test" {
  eservice_id      = %q
  title            = "Import Test"
  description      = "A purpose for import testing"
  daily_calls      = 1000
  is_free_of_charge = false
  desired_state    = "ACTIVE"
}
`, esID),
				Check: tracker.track(),
			},
			{
				ResourceName:            "pdnd_purpose.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_waiting_for_approval"},
			},
			tracker.cleanupStep(testAccProviderConfig(ts.URL)),
		},
	})
}

// purposeIDTracker tracks purpose IDs for cleanup.
type purposeIDTracker struct {
	ids  []uuid.UUID
	fake *fakepdnd.FakeServer
}

func newPurposeIDTracker(fake *fakepdnd.FakeServer) *purposeIDTracker {
	return &purposeIDTracker{fake: fake}
}

func (t *purposeIDTracker) track() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "pdnd_purpose" {
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

// cleanupStep forces tracked purposes back to DRAFT so they can be deleted.
func (t *purposeIDTracker) cleanupStep(providerConfig string) resource.TestStep {
	return resource.TestStep{
		RefreshState: true,
		PreConfig: func() {
			now := time.Now().UTC()
			for _, id := range t.ids {
				p := t.fake.GetStandalonePurpose(id)
				if p != nil {
					// Force all versions to DRAFT
					for i := range p.Versions {
						p.Versions[i].State = "DRAFT"
						p.Versions[i].UpdatedAt = &now
					}
					p.UpdatedAt = &now
					t.fake.SeedStandalonePurpose(*p)
				}
			}
		},
		ExpectNonEmptyPlan: true,
	}
}
