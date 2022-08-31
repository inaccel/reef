package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/inaccel/reef/internal"
	"github.com/urfave/cli/v2"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var initCommand = &cli.Command{
	Name:      "init",
	Usage:     "Update SSL certification/key files",
	ArgsUsage: " ",
	Action: func(context *cli.Context) error {
		now := time.Now()

		kube, err := config.GetConfig()
		if err != nil {
			return err
		}
		api, err := client.New(kube, client.Options{})
		if err != nil {
			return err
		}

		caSerialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 8*20))
		if err != nil {
			return err
		}
		ca := &x509.Certificate{
			SerialNumber: caSerialNumber,
			Subject: pkix.Name{
				Organization: []string{
					"inaccel.com",
				},
			},
			NotBefore: now,
			NotAfter:  now.AddDate(1, 0, 0),
			KeyUsage:  x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
			ExtKeyUsage: []x509.ExtKeyUsage{
				x509.ExtKeyUsageClientAuth,
				x509.ExtKeyUsageServerAuth,
			},
			BasicConstraintsValid: true,
			IsCA:                  true,
		}

		caPriv, err := rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			return err
		}

		caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, caPriv.Public(), caPriv)
		if err != nil {
			return err
		}

		mutatingWebhookConfiguration := &admissionregistrationv1.MutatingWebhookConfiguration{}
		if err := api.Get(context.Context, client.ObjectKey{
			Name: os.Getenv("MUTATING_WEBHOOK_CONFIGURATION_NAME"),
		}, mutatingWebhookConfiguration); err != nil {
			return err
		}

		mutatingWebhookConfiguration.Webhooks[0].ClientConfig.CABundle = pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: caBytes,
		})

		mutatingWebhookConfiguration.Webhooks[0].Rules = internal.Rules

		if err := api.Update(context.Context, mutatingWebhookConfiguration); err != nil {
			return err
		}

		serverSerialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 8*20))
		if err != nil {
			return err
		}
		server := &x509.Certificate{
			SerialNumber: serverSerialNumber,
			Subject: pkix.Name{
				Organization: []string{
					"inaccel.com",
				},
			},
			NotBefore: now,
			NotAfter:  now.AddDate(1, 0, 0),
			KeyUsage:  x509.KeyUsageDigitalSignature,
			ExtKeyUsage: []x509.ExtKeyUsage{
				x509.ExtKeyUsageServerAuth,
			},
			BasicConstraintsValid: true,
			IsCA:                  false,
			DNSNames: []string{
				mutatingWebhookConfiguration.Webhooks[0].ClientConfig.Service.Name,
				strings.Join([]string{
					mutatingWebhookConfiguration.Webhooks[0].ClientConfig.Service.Name,
					mutatingWebhookConfiguration.Webhooks[0].ClientConfig.Service.Namespace,
				}, "."),
				strings.Join([]string{
					mutatingWebhookConfiguration.Webhooks[0].ClientConfig.Service.Name,
					mutatingWebhookConfiguration.Webhooks[0].ClientConfig.Service.Namespace,
					"svc",
				}, "."),
			},
		}

		serverPriv, err := rsa.GenerateKey(rand.Reader, 4096)
		if err != nil {
			return err
		}

		serverBytes, err := x509.CreateCertificate(rand.Reader, server, ca, serverPriv.Public(), caPriv)
		if err != nil {
			return err
		}

		if err := os.MkdirAll(filepath.Dir(context.String("key")), os.ModePerm); err != nil {
			return err
		}
		if err := os.WriteFile(context.String("key"), pem.EncodeToMemory(&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(serverPriv),
		}), os.ModePerm); err != nil {
			return err
		}

		if err := os.MkdirAll(filepath.Dir(context.String("cert")), os.ModePerm); err != nil {
			return err
		}
		if err := os.WriteFile(context.String("cert"), pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: serverBytes,
		}), os.ModePerm); err != nil {
			return err
		}

		return nil
	},
}
