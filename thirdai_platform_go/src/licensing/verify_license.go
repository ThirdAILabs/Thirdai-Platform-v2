package licensing

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"time"
)

type LicensePayload struct {
	CpuMhzLimit    string `json:"cpuMhzLimit"`
	ExpiryDate     string `json:"expiryDate"`
	BoltLicenseKey string `json:"boltLicenseKey"`
}

func (l *LicensePayload) Expiry() (time.Time, error) {
	layout := "2006-01-02T15:04:05-07:00"
	expiry, err := time.Parse(layout, l.ExpiryDate)
	if err != nil {
		return time.Time{}, fmt.Errorf("unable to parse expiry in license: %v", err)
	}
	return expiry, nil
}

type PlatformLicense struct {
	License   LicensePayload `json:"license"`
	Signature string         `json:"signature"`
}

type LicenseVerifier struct {
	publicKey   *rsa.PublicKey
	licensePath string
}

func NewVerifier(licensePath string) *LicenseVerifier {
	block, _ := pem.Decode([]byte(publicKey))
	if block == nil {
		log.Fatalf("pem file is corrupted")
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		log.Fatalf("licensing error: parsing public key: %v", err)
	}

	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		log.Fatalf("licensing error: key must be valid rsa key")
	}

	return &LicenseVerifier{publicKey: rsaKey, licensePath: licensePath}
}

func (v *LicenseVerifier) loadLicense() (PlatformLicense, error) {
	var license PlatformLicense

	file, err := os.Open(v.licensePath)
	defer file.Close()
	if err != nil {
		return license, fmt.Errorf("unable to access platform license: %v", err)
	}
	err = json.NewDecoder(file).Decode(&license)
	if err != nil {
		return license, fmt.Errorf("unable to parse platform license: %v", err)
	}

	return license, nil
}

func (v *LicenseVerifier) Verify() error {
	license, err := v.loadLicense()
	if err != nil {
		return err
	}

	signature, err := base64.StdEncoding.DecodeString(license.Signature)
	if err != nil {
		return fmt.Errorf("unable to decode license signature: %v", err)
	}

	message, err := json.Marshal(license.License)
	if err != nil {
		return fmt.Errorf("error encoding message: %v", err)
	}

	hash := sha256.New()
	hash.Write(message)

	matchErr := rsa.VerifyPKCS1v15(v.publicKey, crypto.SHA256, hash.Sum(nil), signature)
	if matchErr != nil {
		return fmt.Errorf("license is invalid")
	}

	expiry, err := license.License.Expiry()
	if err != nil {
		return err
	}

	if expiry.Before(time.Now()) {
		return fmt.Errorf("license is expired")
	}

	return nil
}
