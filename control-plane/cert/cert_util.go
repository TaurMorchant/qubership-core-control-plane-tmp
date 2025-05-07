package cert

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/netcracker/qubership-core-lib-go/v3/logging"
)

var log = logging.GetLogger("cert")

type CertificateManager struct {
}

type CertificatesValidationResult struct {
	CertificateValidationResult []*CertificateValidationResult
}

type CertificateValidationResult struct {
	CertificateInfo *CertificateInfo
	Valid           bool
	Reason          string
}

type CertificateInfo struct {
	Id             string
	Name           string
	IssuerId       string
	IssuerName     string
	ValidFrom      time.Time
	ValidTill      time.Time
	DaysTillExpiry int
	SANs           []string
}

func (cm *CertificateManager) VerifyCert(certPEM string) (*CertificatesValidationResult, error) {
	certs, err := cm.loadCertificate(certPEM)
	if err != nil {
		return nil, fmt.Errorf("failed to load certificate: %v", err)
	}
	result := &CertificatesValidationResult{}
	for _, cert := range certs {
		err := cm.checkCert(cert, certs)
		certificateValidationResult := cm.prepareResponse(cert)
		if err != nil {
			certificateValidationResult.Reason = fmt.Sprintf("failed to check certificate. Error: %+v", err)
		}
		certificateValidationResult.Valid = err == nil
		result.CertificateValidationResult = append(result.CertificateValidationResult, certificateValidationResult)
	}
	return result, nil
}
func (cm *CertificateManager) prepareResponse(cert *x509.Certificate) *CertificateValidationResult {
	result := &CertificateValidationResult{}
	if cert != nil {
		daysTillExpiration := int(time.Until(cert.NotAfter).Hours() / 24)
		result.CertificateInfo = &CertificateInfo{
			Id:             cert.SerialNumber.String(),
			Name:           cert.Subject.CommonName,
			IssuerId:       cert.Issuer.CommonName,
			IssuerName:     cert.Issuer.CommonName,
			ValidFrom:      cert.NotBefore,
			ValidTill:      cert.NotAfter,
			DaysTillExpiry: daysTillExpiration,
			SANs:           cert.DNSNames, //TODO actually we must add IP addresses + emails here as well
		}
	}
	return result
}

func (cm *CertificateManager) loadCertificate(certPEM string) ([]*x509.Certificate, error) {
	certPEMBlock := []byte(certPEM)
	var certs []*x509.Certificate
	for {
		var block *pem.Block
		block, certPEMBlock = pem.Decode(certPEMBlock)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" {
			continue
		}
		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse certificate: %v", err)
		}
		certs = append(certs, cert)
	}

	if len(certs) == 0 {
		return nil, fmt.Errorf("no certificates found in PEM")
	}
	return certs, nil
}

func (cm *CertificateManager) validateChain(cert *x509.Certificate, certs []*x509.Certificate) error {
	roots := x509.NewCertPool()
	intermediates := x509.NewCertPool()
	for _, certificate := range certs {
		if cm.isRoot(certificate.Issuer, certificate.Subject) {
			roots.AddCert(certificate)
		} else {
			intermediates.AddCert(certificate)
		}
	}

	opts := x509.VerifyOptions{
		Roots:         roots,
		Intermediates: intermediates,
	}
	_, err := cert.Verify(opts)
	return err
}

func (cm *CertificateManager) isRoot(n1, n2 pkix.Name) bool {
	return n1.CommonName == n2.CommonName &&
		n1.SerialNumber == n2.SerialNumber &&
		cm.equalSlices(n1.Organization, n2.Organization) &&
		cm.equalSlices(n1.Country, n2.Country) &&
		cm.equalSlices(n1.OrganizationalUnit, n2.OrganizationalUnit) &&
		cm.equalSlices(n1.Locality, n2.Locality) &&
		cm.equalSlices(n1.Province, n2.Province) &&
		cm.equalSlices(n1.StreetAddress, n2.StreetAddress)
}

func (cm *CertificateManager) equalSlices(s1, s2 []string) bool {
	if len(s1) != len(s2) {
		return false
	}
	for i := range s1 {
		if s1[i] != s2[i] {
			return false
		}
	}
	return true
}

func (cm *CertificateManager) checkCert(cert *x509.Certificate, certs []*x509.Certificate) error {
	now := time.Now()
	if now.Before(cert.NotBefore) {
		return fmt.Errorf("certificate is not yet valid")
	}
	if now.After(cert.NotAfter) {
		return fmt.Errorf("certificate is expired ")
	}
	if err := cm.checkSystemCA(cert); err != nil {
		return err
	}
	if len(certs) > 1 {
		return cm.validateChain(cert, certs)
	}
	return nil
}

func (cm *CertificateManager) checkSystemCA(cert *x509.Certificate) error {
	roots, err := x509.SystemCertPool()
	if err != nil {
		return fmt.Errorf("failed to load system cert pool: %v", err)
	}
	opts := x509.VerifyOptions{
		Roots: roots,
	}

	if _, err := cert.Verify(opts); err == nil {
		//for now just log warning but not throw error if tls cert is in system cert pool
		log.Warnf("certificate found in trust store")
	}
	return nil
}
