package tls

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/go-errors/errors"
	"github.com/netcracker/qubership-core-lib-go/v3/logging"
	"strings"
)

var logger logging.Logger

func GetTlsWithCertificatesValidation(pemTrustedCA string, pemCerts string, pemPrivateKey string) error {
	if _, err := TryToDecodePemAndParseX509Certificates(pemTrustedCA); err != nil {
		return err
	}

	if _, err := TryToValidateX509PrivateKeyClientCertPair(pemPrivateKey, pemCerts); err != nil {
		return err
	}

	return nil
}

func TryToDecodePemAndParseX509Certificates(pemCerts string) (string, error) {
	if len(pemCerts) == 0 {
		return "", nil
	}
	var errs []string
	var validCerts string
	var certStart = "-----BEGIN CERTIFICATE-----\n"
	var certEnd = "\n-----END CERTIFICATE-----"
	slicePemCerts := splitPemCerts(pemCerts)
	for _, pemCert := range slicePemCerts {
		if !strings.HasPrefix(pemCert, certStart) || !strings.HasSuffix(pemCert, certEnd) {
			errs = append(errs, fmt.Errorf("wrong cert format: certificate must start with %s and end with %s \n%s",
				strings.ReplaceAll(certStart, "\n", ""),
				strings.ReplaceAll(certEnd, "\n", ""),
				pemCert).Error())
			continue
		}
		block, _ := pem.Decode([]byte(pemCert))
		if block == nil {
			errs = append(errs, fmt.Errorf("failed to decode PEM format for cert: \n%s", pemCert).Error())
			continue
		}
		_, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			errs = append(errs, fmt.Errorf("failed to parse certificate: %w: \n%s", err, pemCert).Error())
			continue
		}
		validCerts += pemCert + "\n"
	}

	if len(errs) == 0 {
		return pemCerts, nil
	} else {
		return validCerts, fmt.Errorf(strings.Join(errs, "\n"))
	}
}

func splitPemCerts(pemCerts string) []string {
	pemCerts = strings.ReplaceAll(pemCerts, "\r\n", "\n")
	split := strings.Split(pemCerts, "-\n-")
	if len(split) > 1 {
		for i := 0; i < len(split); i++ {
			if i == 0 {
				split[i] = split[i] + "-"
			} else if i == len(split)-1 {
				split[i] = "-" + split[i]
			} else {
				split[i] = "-" + split[i] + "-"
			}
		}
	}
	return split
}
func TryToDecodePemAndParseX509PrivateKey(pemPrivateKey string) error {
	if len(pemPrivateKey) == 0 {
		return nil
	}

	var keyStart = "-----BEGIN"
	var keyEnd = "KEY-----"
	if !strings.HasPrefix(pemPrivateKey, keyStart) || !strings.HasSuffix(pemPrivateKey, keyEnd) {
		return fmt.Errorf("wrong private key format: PK must start with %s and end with %s", keyStart, keyEnd)
	}

	block, rest := pem.Decode([]byte(pemPrivateKey))
	if block == nil || len(rest) != 0 {
		return errors.New("failed to decode PEM format for private key")
	}
	switch strings.ToUpper(block.Type) {
	case "RSA PRIVATE KEY":
		_, err := x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return fmt.Errorf("failed to parse RSA private key: %w", err)
		}
	case "PRIVATE KEY":
		_, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return fmt.Errorf("failed to parse private key: %w", err)
		}
	case "EC PRIVATE KEY":
		_, err := x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return fmt.Errorf("failed to parse EC private key: %w", err)
		}
	default:
		logger.Warn("Type of pem block %s is unknown, parsing is skipped", block.Type)
	}
	return nil
}
func TryToValidateX509PrivateKeyClientCertPair(pemPrivateKey string, pemCerts string) (string, error) {
	if len(pemPrivateKey) == 0 || len(pemCerts) == 0 {
		return "", nil
	}
	var errs []string
	var validated = ""
	validCerts, err := TryToDecodePemAndParseX509Certificates(pemCerts)
	if err != nil || validCerts == "" {
		errs = append(errs, err.Error())
	}
	if err := TryToDecodePemAndParseX509PrivateKey(pemPrivateKey); err != nil {
		errs = append(errs, err.Error())
		return "", fmt.Errorf(strings.Join(errs, "\n"))
	}

	slicePemCerts := splitPemCerts(validCerts)
	for _, pemCert := range slicePemCerts {
		_, err := tls.X509KeyPair([]byte(pemCert), []byte(pemPrivateKey))
		if err != nil {
			errs = append(errs, fmt.Errorf("private key does not match with provided client certificate: %s\nError:\n %w", pemCert, err).Error())
			continue
		}
		validated += pemCert + "\n"
	}
	if len(errs) == 0 {
		return pemCerts, nil
	} else {
		return validated, fmt.Errorf(strings.Join(errs, "\n"))
	}
}
