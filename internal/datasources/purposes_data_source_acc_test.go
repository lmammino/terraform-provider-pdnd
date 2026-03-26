package datasources_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/lmammino/terraform-provider-pdnd/internal/testing/fakepdnd"
)

func TestAccPurposesDataSource(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	esID := uuid.New()
	now := time.Now().UTC()

	// Seed two purposes, different states
	fake.SeedStandalonePurpose(fakepdnd.StoredPurpose{
		ID: uuid.New(), EServiceID: esID, ConsumerID: fake.ConsumerID(),
		Title: "Active Purpose", Description: "An active purpose",
		IsFreeOfCharge: false, IsRiskAnalysisValid: true, CreatedAt: now,
		Versions: []fakepdnd.StoredPurposeVersion{
			{ID: uuid.New(), State: "ACTIVE", DailyCalls: 1000, CreatedAt: now},
		},
	})
	fake.SeedStandalonePurpose(fakepdnd.StoredPurpose{
		ID: uuid.New(), EServiceID: esID, ConsumerID: fake.ConsumerID(),
		Title: "Draft Purpose", Description: "A draft purpose",
		IsFreeOfCharge: true, IsRiskAnalysisValid: true, CreatedAt: now,
		Versions: []fakepdnd.StoredPurposeVersion{
			{ID: uuid.New(), State: "DRAFT", DailyCalls: 500, CreatedAt: now},
		},
	})

	ts := fake.Start()
	defer ts.Close()

	// Test: list all (no filters)
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + `
data "pdnd_purposes" "test" {}
`,
				Check: resource.TestCheckResourceAttr("data.pdnd_purposes.test", "purposes.#", "2"),
			},
		},
	})
}

func TestAccPurposesDataSource_FilterByState(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	esID := uuid.New()
	now := time.Now().UTC()

	fake.SeedStandalonePurpose(fakepdnd.StoredPurpose{
		ID: uuid.New(), EServiceID: esID, ConsumerID: fake.ConsumerID(),
		Title: "Active Purpose", Description: "An active purpose",
		IsFreeOfCharge: false, IsRiskAnalysisValid: true, CreatedAt: now,
		Versions: []fakepdnd.StoredPurposeVersion{
			{ID: uuid.New(), State: "ACTIVE", DailyCalls: 1000, CreatedAt: now},
		},
	})
	fake.SeedStandalonePurpose(fakepdnd.StoredPurpose{
		ID: uuid.New(), EServiceID: esID, ConsumerID: fake.ConsumerID(),
		Title: "Draft Purpose", Description: "A draft purpose",
		IsFreeOfCharge: true, IsRiskAnalysisValid: true, CreatedAt: now,
		Versions: []fakepdnd.StoredPurposeVersion{
			{ID: uuid.New(), State: "DRAFT", DailyCalls: 500, CreatedAt: now},
		},
	})

	ts := fake.Start()
	defer ts.Close()

	// Test: filter by ACTIVE state only
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
data "pdnd_purposes" "test" {
  eservice_ids = [%q]
  states       = ["ACTIVE"]
}
`, esID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.pdnd_purposes.test", "purposes.#", "1"),
					resource.TestCheckResourceAttr("data.pdnd_purposes.test", "purposes.0.state", "ACTIVE"),
				),
			},
		},
	})
}
