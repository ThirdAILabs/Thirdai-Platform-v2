package tests

import "thirdai_platform/model_bazaar/licensing"

func verifyTestLicense() error {
	v := licensing.NewVerifier("platform_test_license.json")
	license, err := v.LoadLicense()
	if err != nil {
		return err
	}
	licensing.ActivateThirdAILicense(license.License.BoltLicenseKey)
	return nil
}
