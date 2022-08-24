// Copyright (c) 2022 Contributors to the Eclipse Foundation
//
// See the NOTICE file(s) distributed with this work for additional
// information regarding copyright ownership.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// http://www.eclipse.org/legal/epl-2.0
//
// SPDX-License-Identifier: EPL-2.0

package testing

import (
	"io"
	"strings"
)

const (
	cert = `-----BEGIN CERTIFICATE-----
MIIEWzCCAkMCFEhFajH0fSSgRaqg8mEnozLvxQHFMA0GCSqGSIb3DQEBCwUAMGwx
CzAJBgNVBAYTAkJHMQ4wDAYDVQQIDAVTb2ZpYTEOMAwGA1UEBwwFU29maWExETAP
BgNVBAoMCEJvc2NoLklPMQwwCgYDVQQLDANQQVAxHDAaBgNVBAMME1NEVl9Cb3Nj
aF9UZXN0X0NlcnQwHhcNMjExMDA2MTgzMjUzWhcNMjcwNzA3MTgzMjUzWjBoMQsw
CQYDVQQGEwJCRzEOMAwGA1UECAwFU29maWExDjAMBgNVBAcMBVNvZmlhMREwDwYD
VQQKDAhCb3NjaC5JTzEMMAoGA1UECwwDUEFQMRgwFgYDVQQDDA9wYXAtZGVtby1k
ZXZpY2UwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDHA4S1sLVf7NKD
xz2AmwO7xDMr+avRvni3Pay31yGIx3F3D2lRCoSNk6lpoxlMSWFcasR6CEODylA/
mmcSgfQyKGVOOuRir7asSOlyahpeMI0cAfcf6FNX03s55gS4Tmjn7TYpdwTMSWYK
ie4Q0XnyO7uJhK6UcPL9s2nYV564geMwnbBGqMPOBnLjjSWDG1Cik3PHcLmsmg0a
2cYYmzglr4qRd23n/gLV19t6XoNV7zTAwD2jMuvnHASR9ITO2rywur9yeQmzx5XH
Pg3Dgi49NuRKHeuTEoI1N5YuzcnEZNA6UIH3yaP0bfZj6mFymlzETkmjPi+yzb6j
/gH9D+51AgMBAAEwDQYJKoZIhvcNAQELBQADggIBAIIrIluWwIeB/PJun2hAxW+C
X5wVcU1eSpn5IN25jRyAvRTpTVlxFpUPXA3W5sb+TWTI1HsRToHHAseBO1c9PnlB
/zQCgWllpkNUNXvyyhDSSInRdlFpg4Ju0A5SidJZjaMwfEA9xUGxt5iiwHBG+uHA
4LvPEJLhXJv7Jyr1oYaI6WbByLF6/7GWbna0ckRZ53mSCojuG/ldWObLKpAw0mrD
/QQPvfA89EG/r6iSeRvdK+mdYMexEmCIJPYz4p3pRcvLIBL6mVE8kFFOmcL5Kpit
BR11Y/+99HiQRFXZTqfQsE2P40wvkKbg4q6cK4D4cWIv+DjhIrsUbsrIdX1VsoQW
C5Lnd84U/simYPGKjLJIILKNLHKdvgOZnZIaMPvFvmG4Cge/icawwyREuGzSXAEK
HDCTv/fMj+xj7vNw61j0q7qdQfwJPoHkYlbkIijl1k3J49VPtmQSDiMKgg2gPVoF
qy13v40DS89APVXL90iyfxTguKGn0Ocbd89WnC7A/JjXBODKNz3EOeKmnpxBbPx0
cUb7TPRTNUEhX+wuLtnRWOGbYaWuK1Pg/6mEswLH3v0sMVcKgiovDs1PkyJ1l8ag
Fa05mI5Z1RfwhcN5b06z3GhgJYM1BoimyLLlt0+0LEYZ7/6Xt2ImueYk5bsRcGSQ
XuqPEyIaaxYwg0+et0CB
-----END CERTIFICATE-----
`

	malformedCert = `-----BEGIN CERTIFICATE-----
MIIEWzCCAkMCFEhFajH0fSSgRaqg8mEnozLvxQHFMA0GCSqGSIb3DQEBCwUAMGwx
CzAJBgNVBAYTAkJHMQ4wDAYDVQQIDAVTb2ZpYTEOMAwGA1UEBwwFU29maWExETAP
BgNVBAoMCEJvc2NoLklPMQwwCgYDVQQLDANQQVAxHDAaBgNVBAMME1NEVl9Cb3Nj
-----END CERTIFICATE-----
`
	certKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAxwOEtbC1X+zSg8c9gJsDu8QzK/mr0b54tz2st9chiMdxdw9p
UQqEjZOpaaMZTElhXGrEeghDg8pQP5pnEoH0MihlTjrkYq+2rEjpcmoaXjCNHAH3
H+hTV9N7OeYEuE5o5+02KXcEzElmConuENF58ju7iYSulHDy/bNp2FeeuIHjMJ2w
RqjDzgZy440lgxtQopNzx3C5rJoNGtnGGJs4Ja+KkXdt5/4C1dfbel6DVe80wMA9
ozLr5xwEkfSEztq8sLq/cnkJs8eVxz4Nw4IuPTbkSh3rkxKCNTeWLs3JxGTQOlCB
98mj9G32Y+phcppcxE5Joz4vss2+o/4B/Q/udQIDAQABAoIBAFkk7lEkclohjrqQ
iLAOv8FfxTwxfhFZrGEIM1G1/8Nw8xZNxPMULwPr3LsA39gYFpB7Er9G7FcgTInw
87KKm4PMLHS6VIsQAldx4X/qnx0Jymt9ReD5BDwW8t+gdQTJupwI2XYBZhjL1/Vo
i0blTiZ/MyYKVNkRLwcNUqAhv2sNl6ni8m1jjNRHyVphrPVmt1VQPii0T3InltRg
hbZPcrPDVbyvtLqNZ6PXec2D1/sZoKaHEmum568QM2P/+6EPcH0nRAP2avhNfikG
VHk20qIKcPneEMjNYxV8ArK4rlb/O6O3fz8spkfaKiXM00v7bUUwr3r9B5qB1HB0
S9Y51QECgYEA/wdUZeH8PvtnfH+ViO250B/PGBFmnNgFg6eBuNQoL+Np3HqjhJTi
Bb0JxumIqGeLvJNZiODCbQfnrxrF3Z9ETxKpoWAPwuMlgccHECEmQpjd4+bo3HLh
LQp7tMkELe15/c6LxLluI/eSIe6QT4faqeF05mfkXOwEHe8AVdJaNvECgYEAx8WS
BGHC7b/rM1h27+zMX9xiB7Tm2HCdzpZWHXZ+vmuW+iq7nwGJSke6q+kkgFn8ZRhw
f+yZsnr0RNM072IAG4jcVKJiK4NvVoi0Tyhh4vvSm/WL0p14bGiEoCTphEMStEHQ
FcBLVjq8WUo2Kq0fi0EwNBLz8eU6Kq5JgPX0F8UCgYA4vnLC8JNlmB6gjurAutRb
QJidrFF+mHoxnvW4IEyIyzrkuczkVRQtXrBsN84WWmO3I7oKQKhCBj5Asd5Qv309
ctOXen5HSK8xvw0NQ7L1onnMmbY6Rr1ffjOkOA3cAjjghjKHJRMioZU8Q46Mg5fd
sLKICZnAKyuHVYRnlBRKoQKBgQDC3G6fo5R2QDv165aoVTzNTLS6e7So7sCfYHlD
Z/AdYej0wHYelWsLb4ggY9vc7umI2xvxTCJnvBNEhxgdYGRmd0sjqvlDJIOXzuTC
Scuhkq1Ov2bR3BQ4+oJTi23UO3ClL4T/koBp7gUGu6K2YgRg2wdf5BTboRLpyvOb
vU2JWQKBgFVVOXnSd8x57r4Na0IFD7E0FwdeWzFUNBY+gQxXMB/J8i3rR7170UOJ
1zgkqTSVQJ8H7zvStUAbidm31jaF3PxsMKDvEK/nMR/q5X6uumJVedZ0l8qt54AM
1bq1vawGJUyJAdDClLMIcH8XGVz0+UVv+hz7qICvMSXM8CDWEQ02
-----END RSA PRIVATE KEY-----`

	malformedCertKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAxwOEtbC1X+zSg8c9gJsDu8QzK/mr0b54tz2st9chiMdxdw9p
UQqEjZOpaaMZTElhXGrEeghDg8pQP5pnEoH0MihlTjrkYq+2rEjpcmoaXjCNHAH3
-----END RSA PRIVATE KEY-----`

	mismatchCertKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAznGmrf7dESN2zoMuiTuJJOjvudlBZaQbdtyXvZYIN9t7FCQM
uVfBDiQmlLyE9Di1j11Zqd2Sr19QFBIHw2HQIpGtEH/6BG2Oi8tGoftZtGryoOyA
NzvRunl1lnt78+Ve+osL4+uZpY8SMQIJbxR591GCYHBO3P1/DKzSl5A9r0NyhpmD
Rn04JYiDCPkLZZzrT57QKmcSfELH6/jlBRGgDObufVFMFNS8MgUnK1XgpuhstP2N
zaBDWxMWbQvwHzQrA3AJMblWXZ7wbKLbV+m0vExhX0H4eOSYaJuCOPPGWAsDkgw1
BU3b8MOirxmawHj4u4HgoPk4KOIu8B/FnuJYhwIDAQABAoIBAFXpM+0Kt+Ku+H0e
WFphvUPv7/tObwmmToubZ0ZNTmQ4YTLTgbwLydphrvCMt2OOyfe8aFjpTWbP6lo/
2p0zclNAfl30dA4trXl9gYpdOEp9izTu1rilmzTX4Nhb0QyBcpIfFTanUAx2yqI8
b8KbKdqDQBd0BU2v7JRQw16xdwoc7h9K1frr+J6wuvq8iqC4mn6FM3EEghaOLmCH
7LIETUy3MBpXGuWsDDPSLE741akVmuAhaUUkH12a+oASRs0USUkPv7f9kVZjdi+H
NyDKC1HqJGdWGzvmxHwZNoqnBZrU9WkiAQjnXj5IMuPOEf0vp8DVuxhZd1kmlJ1W
wzEPb6ECgYEA/doT2BInFprnmho/i0vom8U4QHyeUyvSsHhVe9alToL6jX9Wrqum
4Qc66xT4V0eMbqeMsiAQlf05LsgtMY7U+qtcr/ZPTvhbFNzhaEyKxOyE0/5VAXzj
MeDJXOY0/ULZYriRqhdg73xg6qx/wefd7vgbxXSPF9pKGPPWcHTHZvECgYEA0DDf
iuq6NwfSyKqewS83WGGxpRGcYEF7hmwnhlYqd/7RtDVgZpE0OK1rXQhqY+lUgJWs
xNNsSPrR0nB1IyNC8PQIfYhMAe7YkFGhrdeEjp+ab/myW+sLPk9yraeypW8X2bs2
ZQSUtQ7w1tq6KFauFYUiF3l0S3Egwa94+7saZvcCgYEA+Fz9PVHFXKCCKIu10Buc
oYr71lwWq1kc8ftJ57fCVGZhrT8BGDRpOZFRW99QelROWZUkWsJ0d8sgv1yqmuoc
BoTSUnaycZkbw/W3s8vvmWuvKZqUoLgHsS001eeFwKQ+/A+ItNnaxXTzfab3+Edb
JAsrYK0Bs1ynUnJ/Q9d9oIECgYBGSmn/LhcnI1YQeELXeMiX54wh7ls8yH8bOILz
wT3fe8JztJ3So23dQPgB1iiNiScFrwNBBR0HWt/izCNQdMRSNCJ1t8Hp2Sl3OIh8
+EoCGXL8IXMNw8LtC8ftR7RyVJrZ4XKREsXeh6fa8shtfC6Uh3mmMVSJcC2eF0+i
tl5IqwKBgC/mrKCco9rp4ErNXroweVJE3Fzh/0E4gc2PHZoFxFkT5OaOe4CyEjDj
Oml3d0i8e7boxI72Z1ORw6Imld68s+bSx/PvEfQtSbM6I/VYlvP9EKlSmCudFEi+
Xp7U5RoJZr9D3dDjktCfT7PTP6MpGmtDNOHN/jjZYCI1xlV9SMpQ
-----END RSA PRIVATE KEY-----`
)

// CreateDeviceCertificateReader returns an IO reader for the test device certificate.
func CreateDeviceCertificateReader() io.Reader {
	return strings.NewReader(cert)
}

// CreateMalformedDeviceCertificateReader returns an IO reader for the malformed, test device certificate.
func CreateMalformedDeviceCertificateReader() io.Reader {
	return strings.NewReader(malformedCert)
}

// CreateErrorCertificateReader returns an IO reader for a test device certificate which fails with an error.
func CreateErrorCertificateReader() io.Reader {
	return errorReadWriterCloser{}
}

// CreateMismatchCertificateKeyReader returns an IO reader for a mismatch, test certificate private key.
func CreateMismatchCertificateKeyReader() io.Reader {
	return strings.NewReader(mismatchCertKey)
}

// CreateMalformedCertificateKeyReader returns an IO reader for the malformed, test certificate private key.
func CreateMalformedCertificateKeyReader() io.Reader {
	return strings.NewReader(malformedCertKey)
}

// CreateCertificateKeyReader returns an IO reader for a test certificate private key.
func CreateCertificateKeyReader() io.Reader {
	return strings.NewReader(certKey)
}

// CreateErrorCertificateKeyReader returns an IO reader for a test certificate private key which fails with an error.
func CreateErrorCertificateKeyReader() io.Reader {
	return errorReadWriterCloser{}
}

// DeviceCertificate is getter for the test device certificate.
func DeviceCertificate() string {
	return cert
}

// CertificateKey is getter for the test certificate private key.
func CertificateKey() string {
	return certKey
}

// MalformedCertificateKey is getter for the malformed, test certificate private key.
func MalformedCertificateKey() string {
	return malformedCertKey
}
