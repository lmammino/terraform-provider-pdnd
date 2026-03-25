package resources_test

import (
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/lmammino/terraform-provider-pdnd/internal/testing/fakepdnd"
)

func TestAccEService_CreateDraft(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + `
resource "pdnd_eservice" "test" {
  name        = "Test E-Service"
  description = "A test e-service for acceptance testing"
  technology  = "REST"
  mode        = "DELIVER"

  initial_descriptor_agreement_approval_policy = "AUTOMATIC"
  initial_descriptor_audience                  = ["api.test.example.com"]
  initial_descriptor_daily_calls_per_consumer  = 1000
  initial_descriptor_daily_calls_total         = 10000
  initial_descriptor_voucher_lifespan          = 3600
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("pdnd_eservice.test", "id"),
					resource.TestCheckResourceAttr("pdnd_eservice.test", "name", "Test E-Service"),
					resource.TestCheckResourceAttr("pdnd_eservice.test", "description", "A test e-service for acceptance testing"),
					resource.TestCheckResourceAttr("pdnd_eservice.test", "technology", "REST"),
					resource.TestCheckResourceAttr("pdnd_eservice.test", "mode", "DELIVER"),
					resource.TestCheckResourceAttrSet("pdnd_eservice.test", "producer_id"),
					resource.TestCheckResourceAttrSet("pdnd_eservice.test", "initial_descriptor_id"),
				),
			},
		},
	})
}

func TestAccEService_UpdateDraft(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + `
resource "pdnd_eservice" "test" {
  name        = "Test E-Service"
  description = "A test e-service for acceptance testing"
  technology  = "REST"
  mode        = "DELIVER"

  initial_descriptor_agreement_approval_policy = "AUTOMATIC"
  initial_descriptor_audience                  = ["api.test.example.com"]
  initial_descriptor_daily_calls_per_consumer  = 1000
  initial_descriptor_daily_calls_total         = 10000
  initial_descriptor_voucher_lifespan          = 3600
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pdnd_eservice.test", "name", "Test E-Service"),
					resource.TestCheckResourceAttr("pdnd_eservice.test", "description", "A test e-service for acceptance testing"),
				),
			},
			{
				Config: testAccProviderConfig(ts.URL) + `
resource "pdnd_eservice" "test" {
  name        = "Updated E-Service"
  description = "An updated e-service description"
  technology  = "REST"
  mode        = "DELIVER"

  initial_descriptor_agreement_approval_policy = "AUTOMATIC"
  initial_descriptor_audience                  = ["api.test.example.com"]
  initial_descriptor_daily_calls_per_consumer  = 1000
  initial_descriptor_daily_calls_total         = 10000
  initial_descriptor_voucher_lifespan          = 3600
}
`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("pdnd_eservice.test", "name", "Updated E-Service"),
					resource.TestCheckResourceAttr("pdnd_eservice.test", "description", "An updated e-service description"),
				),
			},
		},
	})
}

func TestAccEService_Delete(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy: func(s *terraform.State) error {
			for _, rs := range s.RootModule().Resources {
				if rs.Type != "pdnd_eservice" {
					continue
				}
				id, err := uuid.Parse(rs.Primary.ID)
				if err != nil {
					return fmt.Errorf("invalid eservice ID %q: %w", rs.Primary.ID, err)
				}
				if fake.GetEService(id) != nil {
					return fmt.Errorf("eservice %s still exists", rs.Primary.ID)
				}
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + `
resource "pdnd_eservice" "test" {
  name        = "Test E-Service"
  description = "A test e-service for acceptance testing"
  technology  = "REST"
  mode        = "DELIVER"

  initial_descriptor_agreement_approval_policy = "AUTOMATIC"
  initial_descriptor_audience                  = ["api.test.example.com"]
  initial_descriptor_daily_calls_per_consumer  = 1000
  initial_descriptor_daily_calls_total         = 10000
  initial_descriptor_voucher_lifespan          = 3600
}
`,
				Check: resource.TestCheckResourceAttrSet("pdnd_eservice.test", "id"),
			},
		},
	})
}

func TestAccEService_Import(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	ts := fake.Start()
	defer ts.Close()

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + `
resource "pdnd_eservice" "test" {
  name        = "Test E-Service"
  description = "A test e-service for acceptance testing"
  technology  = "REST"
  mode        = "DELIVER"

  initial_descriptor_agreement_approval_policy = "AUTOMATIC"
  initial_descriptor_audience                  = ["api.test.example.com"]
  initial_descriptor_daily_calls_per_consumer  = 1000
  initial_descriptor_daily_calls_total         = 10000
  initial_descriptor_voucher_lifespan          = 3600
}
`,
			},
			{
				ResourceName:      "pdnd_eservice.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"initial_descriptor_agreement_approval_policy",
					"initial_descriptor_audience",
					"initial_descriptor_daily_calls_per_consumer",
					"initial_descriptor_daily_calls_total",
					"initial_descriptor_voucher_lifespan",
					"initial_descriptor_description",
				},
			},
		},
	})
}
