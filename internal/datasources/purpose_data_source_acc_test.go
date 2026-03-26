package datasources_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/lmammino/terraform-provider-pdnd/internal/testing/fakepdnd"
)

func TestAccPurposeDataSource(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	purposeID := uuid.New()
	versionID := uuid.New()
	now := time.Now().UTC()

	fake.SeedStandalonePurpose(fakepdnd.StoredPurpose{
		ID:                  purposeID,
		EServiceID:          uuid.New(),
		ConsumerID:          fake.ConsumerID(),
		Title:               "Test Purpose",
		Description:         "A test purpose for data source",
		IsFreeOfCharge:      false,
		IsRiskAnalysisValid: true,
		CreatedAt:           now,
		Versions: []fakepdnd.StoredPurposeVersion{
			{
				ID:                versionID,
				State:             "ACTIVE",
				DailyCalls:        1000,
				CreatedAt:         now,
				FirstActivationAt: &now,
			},
		},
	})

	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
data "pdnd_purpose" "test" {
  id = %q
}
`, purposeID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.pdnd_purpose.test", "title", "Test Purpose"),
					resource.TestCheckResourceAttr("data.pdnd_purpose.test", "state", "ACTIVE"),
					resource.TestCheckResourceAttr("data.pdnd_purpose.test", "daily_calls", "1000"),
				),
			},
		},
	})
}
