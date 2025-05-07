package cert

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	mathRand "math/rand"
	"testing"
	"time"
)

func generateCertificate(commonName string, notBefore, notAfter *time.Time, isCA bool, parent *x509.Certificate, parentPriv *rsa.PrivateKey) (string, *x509.Certificate, *rsa.PrivateKey, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", nil, nil, err
	}
	r := mathRand.New(mathRand.NewSource(time.Now().UnixNano()))
	id := big.NewInt(r.Int63())
	certTemplate := x509.Certificate{
		SerialNumber: id,
		Subject:      pkix.Name{CommonName: commonName},
		NotBefore:    *notBefore,
		NotAfter:     *notAfter,
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IsCA:         isCA,
	}

	if isCA {
		certTemplate.KeyUsage |= x509.KeyUsageCertSign | x509.KeyUsageCRLSign
		certTemplate.ExtKeyUsage = nil
		certTemplate.BasicConstraintsValid = true
		certTemplate.MaxPathLen = -1
	}

	var certDER []byte
	if parent == nil {
		certDER, err = x509.CreateCertificate(rand.Reader, &certTemplate, &certTemplate, &priv.PublicKey, priv)
	} else {
		certDER, err = x509.CreateCertificate(rand.Reader, &certTemplate, parent, &priv.PublicKey, parentPriv)
	}

	if err != nil {
		return "", nil, nil, err
	}
	cert, _ := x509.ParseCertificate(certDER)
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	return string(certPEM), cert, priv, nil
}

func getCertResult(cert *x509.Certificate, certrResults []*CertificateValidationResult) *CertificateValidationResult {
	for _, result := range certrResults {
		if cert.SerialNumber.String() == result.CertificateInfo.Id {
			return result
		}
	}
	fmt.Errorf("cert=%v not found in %v", cert, certrResults)
	panic("failed")
}

func TestNoCertificate(t *testing.T) {
	certManager := &CertificateManager{}
	if _, err := certManager.VerifyCert("some incorrect data"); err == nil {
		t.Errorf("certificate check failed without error")
	}
}

func TestValidCertificate(t *testing.T) {
	notBefore := time.Now()
	notAfter := time.Now().Add(2 * 24 * time.Hour)
	certPEM, rootCert, _, err := generateCertificate("root ca", &notBefore, &notAfter, true, nil, nil)
	if err != nil {
		t.Fatalf("failed to generate root ca: %v", err)
	}
	certManager := &CertificateManager{}
	if result, err := certManager.VerifyCert(certPEM); err != nil {
		t.Errorf("certificate check failed with error: %v", err)
	} else if cert := getCertResult(rootCert, result.CertificateValidationResult); !cert.Valid {
		t.Errorf("certificate check failed: %v", cert.Reason)
	} else if cert.CertificateInfo.DaysTillExpiry >= 2 {
		t.Errorf("incorrect expiration day")
	}
}

func TestNotValidYetCertificate(t *testing.T) {
	notBefore := time.Now().Add(1 * time.Hour)
	notAfter := notBefore.Add(365 * 24 * time.Hour)
	certPEM, rootCert, _, err := generateCertificate("root ca", &notBefore, &notAfter, true, nil, nil)
	if err != nil {
		t.Fatalf("failed to generate root ca: %v", err)
	}
	certManager := &CertificateManager{}
	if result, err := certManager.VerifyCert(certPEM); err != nil {
		t.Errorf("certificate check failed with error: %v", err)
	} else if cert := getCertResult(rootCert, result.CertificateValidationResult); cert.Valid {
		t.Errorf("certificate check failed")
	}
}

func TestExpiredCertificate(t *testing.T) {
	notBefore := time.Now().Add(-2 * 24 * time.Hour)
	notAfter := time.Now().Add(-1 * 24 * time.Hour)

	certPEM, rootCert, _, err := generateCertificate("root ca", &notBefore, &notAfter, true, nil, nil)
	if err != nil {
		t.Fatalf("failed to generate root ca: %v", err)
	}
	certManager := &CertificateManager{}
	if result, err := certManager.VerifyCert(certPEM); err != nil {
		t.Errorf("certificate check failed with error: %v", err)
	} else if cert := getCertResult(rootCert, result.CertificateValidationResult); cert.Valid {
		t.Errorf("certificate check failed")
	} else if cert.CertificateInfo.DaysTillExpiry != -1 {
		t.Errorf("incorrect expiration day")
	}
}

func TestCertificateChainWithExpiredRoot(t *testing.T) {
	notBefore := time.Now().Add(-2 * 24 * time.Hour)
	notAfter := time.Now().Add(-1 * time.Hour)

	rootCertPEM, rootCert, rootPriv, err := generateCertificate("root ca", &notBefore, &notAfter, true, nil, nil)
	if err != nil {
		t.Fatalf("failed to generate root ca: %v", err)
	}

	notBefore = time.Now()
	notAfter = notBefore.Add(365 * 24 * time.Hour)
	intermediateCertPEM, intCert, intPriv, err := generateCertificate("intermidiate ca", &notBefore, &notAfter, true, rootCert, rootPriv)
	if err != nil {
		t.Fatalf("failed to generate intermidiate ca: %v", err)
	}

	leafCert, _, _, err := generateCertificate("leaf cert", &notBefore, &notAfter, false, intCert, intPriv)
	if err != nil {
		t.Fatalf("failed to generate leaf cert: %v", err)
	}
	certManager := &CertificateManager{}
	certPEM := rootCertPEM + "\n" + intermediateCertPEM + "\n" + leafCert
	if result, err := certManager.VerifyCert(certPEM); err != nil {
		t.Errorf("certificate check failed with error: %v", err)
	} else if cert := getCertResult(rootCert, result.CertificateValidationResult); cert.Valid {
		t.Errorf("certificate check failed")
	}
}

func TestCertificateChainWithValidRoot(t *testing.T) {
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)

	rootCertPEM, rootCert, rootPriv, err := generateCertificate("root ca", &notBefore, &notAfter, true, nil, nil)
	if err != nil {
		t.Fatalf("failed to generate root ca: %v", err)
	}

	intermediateCertPEM, intCert, intPriv, err := generateCertificate("intermidiate ca", &notBefore, &notAfter, true, rootCert, rootPriv)
	if err != nil {
		t.Fatalf("failed to generate intermidiate ca: %v", err)
	}

	leafCertPEM, leafCert, _, err := generateCertificate("leaf cert", &notBefore, &notAfter, false, intCert, intPriv)
	if err != nil {
		t.Fatalf("failed to generate leaf cert: %v", err)
	}
	certManager := &CertificateManager{}
	certPEM := rootCertPEM + intermediateCertPEM + leafCertPEM
	if result, err := certManager.VerifyCert(certPEM); err != nil {
		t.Errorf("certificate check failed with error: %v", err)
	} else if cert := getCertResult(leafCert, result.CertificateValidationResult); !cert.Valid {
		t.Errorf("certificate check failed: %v", cert.Reason)
	}
}

func TestCertificateChainWithoutRoot(t *testing.T) {
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)

	_, rootCert, rootPriv, err := generateCertificate("root ca", &notBefore, &notAfter, true, nil, nil)
	if err != nil {
		t.Fatalf("failed to generate root ca: %v", err)
	}

	intermediateCertPEM, intCert, intPriv, err := generateCertificate("intermidiate ca", &notBefore, &notAfter, true, rootCert, rootPriv)
	if err != nil {
		t.Fatalf("failed to generate intermidiate ca: %v", err)
	}

	leafCertPEM, leafCert, _, err := generateCertificate("leaf cert", &notBefore, &notAfter, false, intCert, intPriv)
	if err != nil {
		t.Fatalf("failed to generate leaf cert: %v", err)
	}
	certManager := &CertificateManager{}
	certPEM := intermediateCertPEM + leafCertPEM
	if result, err := certManager.VerifyCert(certPEM); err != nil {
		t.Errorf("certificate check failed with error: %v", err)
	} else if cert := getCertResult(leafCert, result.CertificateValidationResult); cert.Valid {
		t.Errorf("certificate check failed")
	}
}

func TestCertificateChainWithoutIntermidiate(t *testing.T) {
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)

	rootCertPEM, rootCert, rootPriv, err := generateCertificate("root ca", &notBefore, &notAfter, true, nil, nil)
	if err != nil {
		t.Fatalf("failed to generate root ca: %v", err)
	}

	_, intCert, intPriv, err := generateCertificate("intermidiate ca", &notBefore, &notAfter, true, rootCert, rootPriv)
	if err != nil {
		t.Fatalf("failed to generate intermidiate ca: %v", err)
	}

	leafCertPEM, leafCert, _, err := generateCertificate("leaf cert", &notBefore, &notAfter, false, intCert, intPriv)
	if err != nil {
		t.Fatalf("failed to generate leaf cert: %v", err)
	}
	certManager := &CertificateManager{}
	certPEM := rootCertPEM + leafCertPEM
	if result, err := certManager.VerifyCert(certPEM); err != nil {
		t.Errorf("certificate check failed with error: %v", err)
	} else if cert := getCertResult(leafCert, result.CertificateValidationResult); cert.Valid {
		t.Errorf("certificate check failed")
	}
}
