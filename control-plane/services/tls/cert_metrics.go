package tls

import (
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

const (
	retryDelay            = time.Minute
	certificateCheckDelay = 24 * time.Hour
)

var (
	certDaysToExpiry = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "certificate_days_to_expiry",
			Help: "Number of days until certificate expiry",
		},
		[]string{"tls_def", "cert_common_name"},
	)

	certValid = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "certificate_valid",
			Help: "Certificate validity status (1 for valid, 0 for invalid)",
		},
		[]string{"tls_def", "cert_common_name"},
	)

	trigger = make(chan bool)
)

func RegisterCertificateMetrics(tlsService *Service) {
	prometheus.MustRegister(certDaysToExpiry)
	prometheus.MustRegister(certValid)

	go updateCertificateMetricsPeriodically(tlsService)
	logger.Info("certificate metrics registered")
}

func TriggerCertificateMetricsUpdate() {
	trigger <- true
}

func checkCertificates(tlsService *Service) error {
	certificateValidationResponse, err := tlsService.ValidateCertificates()
	if err != nil {
		logger.Errorf("Certificates metrics can't be exposed. Can't validate certificates: %v.", err)
		return err
	}

	certificateDetails := certificateValidationResponse.TlsDefDetails

	for _, v := range certificateDetails {
		tlsDef := v.Name
		for _, c := range v.Certificates {
			certName := c.CertificateCommonName
			if c.Valid {
				certValid.WithLabelValues(tlsDef, certName).Set(1)
			} else {
				certValid.WithLabelValues(tlsDef, certName).Set(0)
			}
			certDaysToExpiry.WithLabelValues(tlsDef, certName).Set(float64(c.DaysTillExpiry))
		}
	}
	return nil
}

func updateCertificateMetrics(tlsService *Service) {
	logger.Info("certificate metrics updating started")
	err := checkCertificates(tlsService)
	if err != nil {
		for attempt := 1; attempt <= 10; attempt++ {
			err = checkCertificates(tlsService)
			if err == nil {
				break
			}
			time.Sleep(retryDelay)
		}
	}
}

func updateCertificateMetricsPeriodically(tlsService *Service) {
	updateCertificateMetrics(tlsService)

	ticker := time.NewTicker(certificateCheckDelay)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			logger.Info("try to collect certificate metrics")
			updateCertificateMetrics(tlsService)
		case <-trigger:
			logger.Info("certificate metrics updating triggered")
			updateCertificateMetrics(tlsService)
		}
	}
}
