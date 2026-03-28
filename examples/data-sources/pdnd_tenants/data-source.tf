data "pdnd_tenants" "all" {}

data "pdnd_tenants" "by_ipa" {
  ipa_code = "c_h501"
}
