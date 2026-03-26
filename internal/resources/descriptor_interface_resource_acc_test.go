package resources_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/lmammino/terraform-provider-pdnd/internal/testing/fakepdnd"
)

func TestAccDescriptorInterface_Create(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	eserviceID, descriptorID := seedEServiceWithDraftDescriptor(fake)

	ts := fake.Start()
	defer ts.Close()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test-doc.yaml")
	if err := os.WriteFile(filePath, []byte("openapi: 3.0.0\ninfo:\n  title: Test"), 0644); err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_eservice_descriptor_interface" "test" {
  eservice_id   = %q
  descriptor_id = %q
  pretty_name   = "OpenAPI Spec"
  file_path     = %q
  content_type  = "application/yaml"
}
`, eserviceID, descriptorID, filePath),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("pdnd_eservice_descriptor_interface.test", "document_id"),
					resource.TestCheckResourceAttrSet("pdnd_eservice_descriptor_interface.test", "name"),
					resource.TestCheckResourceAttrSet("pdnd_eservice_descriptor_interface.test", "pretty_name"),
					resource.TestCheckResourceAttrSet("pdnd_eservice_descriptor_interface.test", "content_type"),
					resource.TestCheckResourceAttrSet("pdnd_eservice_descriptor_interface.test", "created_at"),
					resource.TestCheckResourceAttrSet("pdnd_eservice_descriptor_interface.test", "file_hash"),
				),
			},
		},
	})
}

func TestAccDescriptorInterface_Delete(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	eserviceID, descriptorID := seedEServiceWithDraftDescriptor(fake)

	ts := fake.Start()
	defer ts.Close()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test-doc.yaml")
	if err := os.WriteFile(filePath, []byte("openapi: 3.0.0\ninfo:\n  title: Test"), 0644); err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		CheckDestroy: func(s *terraform.State) error {
			iface := fake.GetInterface(eserviceID, descriptorID)
			if iface != nil {
				return fmt.Errorf("expected interface to be deleted, but it still exists")
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_eservice_descriptor_interface" "test" {
  eservice_id   = %q
  descriptor_id = %q
  pretty_name   = "OpenAPI Spec"
  file_path     = %q
  content_type  = "application/yaml"
}
`, eserviceID, descriptorID, filePath),
				Check: resource.TestCheckResourceAttrSet("pdnd_eservice_descriptor_interface.test", "document_id"),
			},
		},
	})
}

func TestAccDescriptorInterface_Import(t *testing.T) {
	fake := fakepdnd.NewFakeServer()
	eserviceID, descriptorID := seedEServiceWithDraftDescriptor(fake)

	ts := fake.Start()
	defer ts.Close()

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test-doc.yaml")
	if err := os.WriteFile(filePath, []byte("openapi: 3.0.0\ninfo:\n  title: Test"), 0644); err != nil {
		t.Fatal(err)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfig(ts.URL) + fmt.Sprintf(`
resource "pdnd_eservice_descriptor_interface" "test" {
  eservice_id   = %q
  descriptor_id = %q
  pretty_name   = "OpenAPI Spec"
  file_path     = %q
  content_type  = "application/yaml"
}
`, eserviceID, descriptorID, filePath),
			},
			{
				ResourceName:            "pdnd_eservice_descriptor_interface.test",
				ImportState:             true,
				ImportStateId:           fmt.Sprintf("%s/%s", eserviceID, descriptorID),
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"file_path", "file_hash", "document_id", "name", "pretty_name", "content_type", "created_at"},
			},
		},
	})
}
