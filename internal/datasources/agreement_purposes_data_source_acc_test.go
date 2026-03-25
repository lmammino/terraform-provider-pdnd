package datasources_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/lmammino/terraform-provider-pdnd/internal/testing/fakepdnd"
)

func TestAccAgreementPurposesDataSource(t *testing.T) {
	fake := fakepdnd.NewFakeServer()

	agreementID := uuid.New()
	fake.SeedAgreement(fakepdnd.StoredAgreement{
		ID:           agreementID,
		EServiceID:   uuid.New(),
		DescriptorID: uuid.New(),
		ProducerID:   fake.ProducerID(),
		ConsumerID:   fake.ConsumerID(),
		State:        "ACTIVE",
		CreatedAt:    time.Now(),
	})
	fake.SeedPurpose(agreementID, fakepdnd.StoredPurpose{
		ID:                  uuid.New(),
		EServiceID:          uuid.New(),
		ConsumerID:          fake.ConsumerID(),
		Title:               "Test Purpose",
		Description:         "A test purpose",
		CreatedAt:           time.Now(),
		IsRiskAnalysisValid: true,
		IsFreeOfCharge:      false,
	})

	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
data "pdnd_agreement_purposes" "test" {
  agreement_id = %q
}
`, agreementID.String()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.pdnd_agreement_purposes.test", "purposes.#", "1"),
					resource.TestCheckResourceAttr("data.pdnd_agreement_purposes.test", "purposes.0.title", "Test Purpose"),
				),
			},
		},
	})
}
