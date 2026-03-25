package datasources_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/lmammino/terraform-provider-pdnd/internal/testing/fakepdnd"
)

// agreementIDTracker tracks agreement IDs seen in terraform state.
type agreementIDTracker struct {
	ids  []uuid.UUID
	fake *fakepdnd.FakeServer
}

func newAgreementIDTracker(fake *fakepdnd.FakeServer) *agreementIDTracker {
	return &agreementIDTracker{fake: fake}
}

// track returns a Check function that records all agreement IDs in the state.
func (t *agreementIDTracker) track() resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "pdnd_agreement" {
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

// cleanupStep returns a RefreshState step that first forces tracked agreements
// to DRAFT, then refreshes the terraform state.
func (t *agreementIDTracker) cleanupStep() resource.TestStep {
	return resource.TestStep{
		RefreshState: true,
		PreConfig: func() {
			for _, id := range t.ids {
				a := t.fake.GetAgreement(id)
				if a != nil {
					t.fake.SeedAgreement(fakepdnd.StoredAgreement{
						ID:           a.ID,
						EServiceID:   a.EServiceID,
						DescriptorID: a.DescriptorID,
						ProducerID:   a.ProducerID,
						ConsumerID:   a.ConsumerID,
						State:        "DRAFT",
						CreatedAt:    a.CreatedAt,
					})
				}
			}
		},
		ExpectNonEmptyPlan: true,
	}
}

func TestAccAgreementsDataSource(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	ts := fake.Start()
	defer ts.Close()

	eserviceID1 := uuid.New().String()
	descriptorID1 := uuid.New().String()
	eserviceID2 := uuid.New().String()
	descriptorID2 := uuid.New().String()
	tracker := newAgreementIDTracker(fake)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_agreement" "active" {
  eservice_id   = %q
  descriptor_id = %q
  desired_state = "ACTIVE"
}

resource "pdnd_agreement" "draft" {
  eservice_id   = %q
  descriptor_id = %q
  desired_state = "DRAFT"
}

data "pdnd_agreements" "active_only" {
  states = ["ACTIVE"]
  depends_on = [pdnd_agreement.active, pdnd_agreement.draft]
}
`, eserviceID1, descriptorID1, eserviceID2, descriptorID2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.pdnd_agreements.active_only", "agreements.#", "1"),
					tracker.track(),
				),
			},
			tracker.cleanupStep(),
		},
	})
}
