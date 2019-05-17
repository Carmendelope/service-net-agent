/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package client

import (
	"io/ioutil"
	"os"
	"time"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"

	"google.golang.org/grpc/metadata"
)

var _ = ginkgo.Describe("client", func() {

	var address string = "localhost:12345"
	var certfile string

	ginkgo.BeforeSuite(func() {
		f, err := ioutil.TempFile("", "*.crt")
		gomega.Expect(err).To(gomega.Succeed())
		// Let's Encrypt certificate
		_, err = f.Write([]byte(`
-----BEGIN CERTIFICATE-----
MIIFjTCCA3WgAwIBAgIRANOxciY0IzLc9AUoUSrsnGowDQYJKoZIhvcNAQELBQAw
TzELMAkGA1UEBhMCVVMxKTAnBgNVBAoTIEludGVybmV0IFNlY3VyaXR5IFJlc2Vh
cmNoIEdyb3VwMRUwEwYDVQQDEwxJU1JHIFJvb3QgWDEwHhcNMTYxMDA2MTU0MzU1
WhcNMjExMDA2MTU0MzU1WjBKMQswCQYDVQQGEwJVUzEWMBQGA1UEChMNTGV0J3Mg
RW5jcnlwdDEjMCEGA1UEAxMaTGV0J3MgRW5jcnlwdCBBdXRob3JpdHkgWDMwggEi
MA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQCc0wzwWuUuR7dyXTeDs2hjMOrX
NSYZJeG9vjXxcJIvt7hLQQWrqZ41CFjssSrEaIcLo+N15Obzp2JxunmBYB/XkZqf
89B4Z3HIaQ6Vkc/+5pnpYDxIzH7KTXcSJJ1HG1rrueweNwAcnKx7pwXqzkrrvUHl
Npi5y/1tPJZo3yMqQpAMhnRnyH+lmrhSYRQTP2XpgofL2/oOVvaGifOFP5eGr7Dc
Gu9rDZUWfcQroGWymQQ2dYBrrErzG5BJeC+ilk8qICUpBMZ0wNAxzY8xOJUWuqgz
uEPxsR/DMH+ieTETPS02+OP88jNquTkxxa/EjQ0dZBYzqvqEKbbUC8DYfcOTAgMB
AAGjggFnMIIBYzAOBgNVHQ8BAf8EBAMCAYYwEgYDVR0TAQH/BAgwBgEB/wIBADBU
BgNVHSAETTBLMAgGBmeBDAECATA/BgsrBgEEAYLfEwEBATAwMC4GCCsGAQUFBwIB
FiJodHRwOi8vY3BzLnJvb3QteDEubGV0c2VuY3J5cHQub3JnMB0GA1UdDgQWBBSo
SmpjBH3duubRObemRWXv86jsoTAzBgNVHR8ELDAqMCigJqAkhiJodHRwOi8vY3Js
LnJvb3QteDEubGV0c2VuY3J5cHQub3JnMHIGCCsGAQUFBwEBBGYwZDAwBggrBgEF
BQcwAYYkaHR0cDovL29jc3Aucm9vdC14MS5sZXRzZW5jcnlwdC5vcmcvMDAGCCsG
AQUFBzAChiRodHRwOi8vY2VydC5yb290LXgxLmxldHNlbmNyeXB0Lm9yZy8wHwYD
VR0jBBgwFoAUebRZ5nu25eQBc4AIiMgaWPbpm24wDQYJKoZIhvcNAQELBQADggIB
ABnPdSA0LTqmRf/Q1eaM2jLonG4bQdEnqOJQ8nCqxOeTRrToEKtwT++36gTSlBGx
A/5dut82jJQ2jxN8RI8L9QFXrWi4xXnA2EqA10yjHiR6H9cj6MFiOnb5In1eWsRM
UM2v3e9tNsCAgBukPHAg1lQh07rvFKm/Bz9BCjaxorALINUfZ9DD64j2igLIxle2
DPxW8dI/F2loHMjXZjqG8RkqZUdoxtID5+90FgsGIfkMpqgRS05f4zPbCEHqCXl1
eO5HyELTgcVlLXXQDgAWnRzut1hFJeczY1tjQQno6f6s+nMydLN26WuU4s3UYvOu
OsUxRlJu7TSRHqDC3lSE5XggVkzdaPkuKGQbGpny+01/47hfXXNB7HntWNZ6N2Vw
p7G6OfY+YQrZwIaQmhrIqJZuigsrbe3W+gdn5ykE9+Ky0VgVUsfxo52mwFYs1JKY
2PGDuWx8M6DlS6qQkvHaRUo0FMd8TsSlbF0/v965qGFKhSDeQoMpYnwcmQilRh/0
ayLThlHLN81gSkJjVrPI0Y8xCVPB4twb1PFUd2fPM3sA1tJ83sZ5v8vgFv2yofKR
PB0t6JzUA81mSqM3kxl5e+IZwhYAyO0OTg3/fs8HqGTNKd9BqoUwSRBzp06JMg5b
rUCGwbCUDI0mxadJ3Bz4WxR6fyNpBK2yAinWEsikxqEt
-----END CERTIFICATE-----
		`))
		gomega.Expect(err).To(gomega.Succeed())
		certfile = f.Name()
		f.Close()
	})

	ginkgo.AfterSuite(func() {
		os.Remove(certfile)
	})
	// We can't actually verify that the connection is
	// set up right, as we can't inspect the final dial
	// options. DialOption is an instance with a function
	// call, which we cannot compare. We can only create
	// the connection and see if errors are returned
	ginkgo.It("should create a non-tls client", func() {
		opts := &ConnectionOptions{}

		client, err := NewAgentClient(address, opts)
		gomega.Expect(err).To(gomega.Succeed())
		gomega.Expect(client).ToNot(gomega.BeNil())
		gomega.Expect(client.address).To(gomega.Equal(address))
		gomega.Expect(client.opts).To(gomega.Equal(opts))
	})

	ginkgo.It("should create a tls client", func() {
		opts := &ConnectionOptions{
			UseTLS: true,
		}

		_, err := NewAgentClient(address, opts)
		gomega.Expect(err).To(gomega.Succeed())
	})

	ginkgo.It("should create an insecure tls client", func() {
		opts := &ConnectionOptions{
			UseTLS: true,
			Insecure: true,
		}

		_, err := NewAgentClient(address, opts)
		gomega.Expect(err).To(gomega.Succeed())
	})

	ginkgo.It("should create a tls client with a specific cert", func() {
		opts := &ConnectionOptions{
			UseTLS: true,
			CACert: certfile,
		}

		_, err := NewAgentClient(address, opts)
		gomega.Expect(err).To(gomega.Succeed())
	})

	ginkgo.It("should create a context with a token", func() {
		opts := &ConnectionOptions{
			Token: "testtoken",
		}
		client, err := NewAgentClient(address, opts)
		gomega.Expect(err).To(gomega.Succeed())

		md, found := metadata.FromOutgoingContext(client.GetContext())
		gomega.Expect(found).To(gomega.BeTrue())
		gomega.Expect(md.Get("authorization")).To(gomega.ConsistOf("testtoken"))

	})

	ginkgo.It("should create a context with a timeout", func() {
		opts := &ConnectionOptions{
			Timeout: 10 * time.Second,
		}
		client, err := NewAgentClient(address, opts)
		gomega.Expect(err).To(gomega.Succeed())

		deadline, set := client.GetContext().Deadline()
		gomega.Expect(set).To(gomega.BeTrue())

		// The deadline should be now + timeout, within one
		// second (due to execution time of the test)
		gomega.Expect(deadline).To(gomega.BeTemporally("~", time.Now().Add(10 * time.Second), time.Second))
	})
})
