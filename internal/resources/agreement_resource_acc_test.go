package resources_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/lmammino/terraform-provider-pdnd/internal/provider"
	"github.com/lmammino/terraform-provider-pdnd/internal/testing/fakepdnd"
)

var testPrivateKeyPEM string

func init() {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(fmt.Sprintf("failed to generate test RSA key: %v", err))
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	testPrivateKeyPEM = string(pemBytes)
}

func testAccProtoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"pdnd": providerserver.NewProtocol6WithError(provider.New("test")()),
	}
}

func testAccProviderConfig(serverURL string) string {
	return fmt.Sprintf(`
provider "pdnd" {
  base_url         = %q
  access_token     = "test-access-token"
  dpop_private_key = <<-EOT
%sEOT
  dpop_key_id      = "test-key-id"
}
`, serverURL, testPrivateKeyPEM)
}

// forceAgreementsDraft changes all agreements in the fake server to DRAFT state
// so they can be cleaned up by the test framework's post-test destroy.
// This must be used in combination with a subsequent RefreshState step so that
// the terraform state file picks up the DRAFT state from the fake server.
func forceAgreementsDraft(fake *fakepdnd.FakeServer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "pdnd_agreement" {
				continue
			}
			id, err := uuid.Parse(rs.Primary.ID)
			if err != nil {
				continue
			}
			a := fake.GetAgreement(id)
			if a != nil {
				fake.SeedAgreement(fakepdnd.StoredAgreement{
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
		return nil
	}
}

// cleanupStep returns a RefreshState step that first forces all agreements to
// DRAFT in the fake server, then refreshes state so Terraform sees them as
// deletable during post-test destroy.
func cleanupStep(fake *fakepdnd.FakeServer) resource.TestStep {
	return resource.TestStep{
		RefreshState: true,
		PreConfig: func() {
			// We can't access terraform state here, but we can force ALL
			// agreements in the fake server to DRAFT.
			forceAllAgreementsDraft(fake)
		},
		// After this step, expected_state will show DRAFT instead of ACTIVE,
		// which means desired_state != state. That causes a non-empty plan.
		ExpectNonEmptyPlan: true,
	}
}

// forceAllAgreementsDraft forces ALL agreements in the fake server to DRAFT.
// This is a brute-force cleanup: since each test creates its own fake server,
// there's no risk of cross-test interference.
func forceAllAgreementsDraft(fake *fakepdnd.FakeServer) {
	// We need to iterate all agreements. The FakeServer API only exposes
	// GetAgreement(id), but since each test uses its own server and creates
	// at most a few agreements, we track IDs in the test and clean up.
	// However, for simplicity we'll use a different approach: we'll track IDs
	// via the Check function in the previous step.
	//
	// Actually, for the PreConfig approach, we need agreement IDs before the step.
	// We'll use a tracking slice that gets populated by a Check function.
}

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

func TestAccAgreement_CreateDraft(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	ts := fake.Start()
	defer ts.Close()

	eserviceID := uuid.New().String()
	descriptorID := uuid.New().String()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_agreement" "test" {
  eservice_id   = %q
  descriptor_id = %q
  desired_state = "DRAFT"
}
`, eserviceID, descriptorID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pdnd_agreement.test", "state", "DRAFT"),
					resource.TestCheckResourceAttr("pdnd_agreement.test", "desired_state", "DRAFT"),
					resource.TestCheckResourceAttrSet("pdnd_agreement.test", "id"),
					resource.TestCheckResourceAttrSet("pdnd_agreement.test", "producer_id"),
					resource.TestCheckResourceAttrSet("pdnd_agreement.test", "consumer_id"),
					resource.TestCheckResourceAttrSet("pdnd_agreement.test", "created_at"),
				),
			},
		},
	})
}

func TestAccAgreement_CreateActive_HappyPath(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	ts := fake.Start()
	defer ts.Close()

	eserviceID := uuid.New().String()
	descriptorID := uuid.New().String()
	tracker := newAgreementIDTracker(fake)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_agreement" "test" {
  eservice_id   = %q
  descriptor_id = %q
  desired_state = "ACTIVE"
}
`, eserviceID, descriptorID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pdnd_agreement.test", "state", "ACTIVE"),
					resource.TestCheckResourceAttr("pdnd_agreement.test", "desired_state", "ACTIVE"),
					resource.TestCheckResourceAttrSet("pdnd_agreement.test", "id"),
					tracker.track(),
				),
			},
			tracker.cleanupStep(),
		},
	})
}

func TestAccAgreement_CreateActive_Pending_Allowed(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	fake.SetApprovalPolicy("MANUAL")
	ts := fake.Start()
	defer ts.Close()

	eserviceID := uuid.New().String()
	descriptorID := uuid.New().String()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_agreement" "test" {
  eservice_id   = %q
  descriptor_id = %q
  desired_state = "ACTIVE"
  allow_pending = true
}
`, eserviceID, descriptorID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pdnd_agreement.test", "state", "PENDING"),
					resource.TestCheckResourceAttr("pdnd_agreement.test", "desired_state", "ACTIVE"),
				),
			},
		},
	})
}

func TestAccAgreement_CreateActive_Pending_Forbidden(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	fake.SetApprovalPolicy("MANUAL")
	ts := fake.Start()
	defer ts.Close()

	eserviceID := uuid.New().String()
	descriptorID := uuid.New().String()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_agreement" "test" {
  eservice_id   = %q
  descriptor_id = %q
  desired_state = "ACTIVE"
}
`, eserviceID, descriptorID),
				ExpectError: regexp.MustCompile(`PENDING`),
			},
		},
	})
}

func TestAccAgreement_UpdateActiveToSuspended(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	ts := fake.Start()
	defer ts.Close()

	eserviceID := uuid.New().String()
	descriptorID := uuid.New().String()
	tracker := newAgreementIDTracker(fake)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_agreement" "test" {
  eservice_id   = %q
  descriptor_id = %q
  desired_state = "ACTIVE"
}
`, eserviceID, descriptorID),
				Check: resource.TestCheckResourceAttr("pdnd_agreement.test", "state", "ACTIVE"),
			},
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_agreement" "test" {
  eservice_id   = %q
  descriptor_id = %q
  desired_state = "SUSPENDED"
}
`, eserviceID, descriptorID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pdnd_agreement.test", "state", "SUSPENDED"),
					tracker.track(),
				),
			},
			tracker.cleanupStep(),
		},
	})
}

func TestAccAgreement_UpdateSuspendedToActive(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	ts := fake.Start()
	defer ts.Close()

	eserviceID := uuid.New().String()
	descriptorID := uuid.New().String()
	tracker := newAgreementIDTracker(fake)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_agreement" "test" {
  eservice_id   = %q
  descriptor_id = %q
  desired_state = "ACTIVE"
}
`, eserviceID, descriptorID),
				Check: resource.TestCheckResourceAttr("pdnd_agreement.test", "state", "ACTIVE"),
			},
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_agreement" "test" {
  eservice_id   = %q
  descriptor_id = %q
  desired_state = "SUSPENDED"
}
`, eserviceID, descriptorID),
				Check: resource.TestCheckResourceAttr("pdnd_agreement.test", "state", "SUSPENDED"),
			},
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_agreement" "test" {
  eservice_id   = %q
  descriptor_id = %q
  desired_state = "ACTIVE"
}
`, eserviceID, descriptorID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pdnd_agreement.test", "state", "ACTIVE"),
					tracker.track(),
				),
			},
			tracker.cleanupStep(),
		},
	})
}

func TestAccAgreement_DestroyDraft(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	ts := fake.Start()
	defer ts.Close()

	eserviceID := uuid.New().String()
	descriptorID := uuid.New().String()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy: func(s *terraform.State) error {
			for _, rs := range s.RootModule().Resources {
				if rs.Type != "pdnd_agreement" {
					continue
				}
				id, err := uuid.Parse(rs.Primary.ID)
				if err != nil {
					return fmt.Errorf("invalid agreement ID %q: %w", rs.Primary.ID, err)
				}
				if fake.GetAgreement(id) != nil {
					return fmt.Errorf("agreement %s still exists", rs.Primary.ID)
				}
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_agreement" "test" {
  eservice_id   = %q
  descriptor_id = %q
  desired_state = "DRAFT"
}
`, eserviceID, descriptorID),
				Check: resource.TestCheckResourceAttr("pdnd_agreement.test", "state", "DRAFT"),
			},
		},
	})
}

func TestAccAgreement_DestroyActiveFails(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	ts := fake.Start()
	defer ts.Close()

	eserviceID := uuid.New().String()
	descriptorID := uuid.New().String()
	tracker := newAgreementIDTracker(fake)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_agreement" "test" {
  eservice_id   = %q
  descriptor_id = %q
  desired_state = "ACTIVE"
}
`, eserviceID, descriptorID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pdnd_agreement.test", "state", "ACTIVE"),
					tracker.track(),
				),
			},
			{
				Config:      testAccProviderConfig(ts.URL),
				ExpectError: regexp.MustCompile(`Cannot [Dd]elete [Aa]greement`),
			},
			// After the error step, re-apply the config so the resource is back in state,
			// then clean up for post-test destroy.
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_agreement" "test" {
  eservice_id   = %q
  descriptor_id = %q
  desired_state = "ACTIVE"
}
`, eserviceID, descriptorID),
				Check: tracker.track(),
			},
			tracker.cleanupStep(),
		},
	})
}

func TestAccAgreement_Import(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	ts := fake.Start()
	defer ts.Close()

	eserviceID := uuid.New().String()
	descriptorID := uuid.New().String()
	tracker := newAgreementIDTracker(fake)

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_agreement" "test" {
  eservice_id   = %q
  descriptor_id = %q
  desired_state = "ACTIVE"
}
`, eserviceID, descriptorID),
				Check: tracker.track(),
			},
			{
				ResourceName:            "pdnd_agreement.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"allow_pending", "consumer_notes"},
			},
			tracker.cleanupStep(),
		},
	})
}
