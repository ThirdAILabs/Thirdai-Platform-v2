package licensing

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"thirdai_platform/search/ndb"
	"time"
)

var (
	ErrInvalidLicense   = errors.New("license is invalid")
	ErrExpiredLicense   = errors.New("license is expired")
	ErrCpuLimitExceeded = errors.New("maximum cpu limit for license exceeded")
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
		slog.Error("unable to parse license expiry", "error", err)
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
	// public key is stored as a variable rather than in a .pem file to make it more
	// difficult for someone to change the public key their own and then sign their
	// own licenses.
	block, _ := pem.Decode([]byte(publicKey))
	if block == nil {
		log.Panicf("pem file is corrupted")
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		log.Panicf("licensing error: parsing public key: %v", err)
	}

	rsaKey, ok := key.(*rsa.PublicKey)
	if !ok {
		log.Panicf("licensing error: key must be valid rsa key")
	}

	v := &LicenseVerifier{publicKey: rsaKey, licensePath: licensePath}

	if _, err := v.Verify(0); err != nil {
		log.Panicf("must have valid license for initialization: %v", err)
	}

	return v
}

func (v *LicenseVerifier) LoadLicense() (PlatformLicense, error) {
	var license PlatformLicense

	file, err := os.Open(v.licensePath)
	if err != nil {
		slog.Error("error opening license file", "error", err)
		return license, fmt.Errorf("unable to access platform license: %v", err)
	}
	defer file.Close()

	err = json.NewDecoder(file).Decode(&license)
	if err != nil {
		slog.Error("unable to parse license file", "error", err)
		return license, fmt.Errorf("unable to parse platform license: %v", err)
	}

	return license, nil
}

func (v *LicenseVerifier) Verify(currCpuUsage int) (LicensePayload, error) {
	// License is loaded each time so it can be swapped without restarting the service
	license, err := v.LoadLicense()
	if err != nil {
		return LicensePayload{}, err
	}

	signature, err := base64.StdEncoding.DecodeString(license.Signature)
	if err != nil {
		slog.Error("error decoding license signature", "error", err)
		return LicensePayload{}, fmt.Errorf("unable to decode license signature: %v", err)
	}

	message, err := json.Marshal(license.License)
	if err != nil {
		return LicensePayload{}, fmt.Errorf("error encoding message: %v", err)
	}

	hash := sha256.New()
	hash.Write(message)

	matchErr := rsa.VerifyPKCS1v15(v.publicKey, crypto.SHA256, hash.Sum(nil), signature)
	if matchErr != nil {
		slog.Error("platform license signature doesn't match", "error", err)
		return LicensePayload{}, ErrInvalidLicense
	}

	expiry, err := license.License.Expiry()
	if err != nil {
		return LicensePayload{}, err
	}

	if expiry.Before(time.Now().UTC()) {
		slog.Error("platform license is expired", "error", err)
		return LicensePayload{}, ErrExpiredLicense
	}

	// TODO(Anyone): why is this not just stored as an integer in the license?
	cpuLimit, err := strconv.Atoi(license.License.CpuMhzLimit)
	if err != nil {
		slog.Error("platform license has invalid cpu limit", "error", err)
		return LicensePayload{}, fmt.Errorf("invalid cpu limit: %v", err)
	}

	if cpuLimit < currCpuUsage {
		slog.Error("platform license cpu limit exceeded", "error", err)
		return LicensePayload{}, ErrCpuLimitExceeded
	}

	return license.License, nil
}

func ActivateThirdAILicense(thirdaiLicense string) error {
	if strings.HasPrefix(thirdaiLicense, "file ") {
		slog.Info("activating file-based license")
		licenseData := thirdaiLicense[len("file "):] // Extract the base64 encoded data
		decodedData, err := base64.StdEncoding.DecodeString(licenseData)
		if err != nil {
			return err
		}

		licensePath := "./thirdai.license"
		if err := os.WriteFile(licensePath, decodedData, 0644); err != nil {
			return err
		}

		return ndb.SetLicensePath(licensePath)
	} else {
		slog.Info("activating key-based license")
		return ndb.SetLicenseKey(thirdaiLicense)
	}
}
