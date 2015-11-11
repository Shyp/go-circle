package main

import (
	"fmt"

	"github.com/equinox-io/equinox"
)

const appId = "app_n7HhD13kpUR"

var publicKey = []byte(`
-----BEGIN ECDSA PUBLIC KEY-----
MHYwEAYHKoZIzj0CAQYFK4EEACIDYgAEHDC84WuSrsogYANW2FKygG179sPPhKCz
ZyFhws7n1Ocyxk4sUg8+eX+G+Fo8GNhsXMcDu06MEAGdKKX6HCBtG4qjpunkJjM5
Ktdq2DCS+H93c1fYG6KxFqxuzLPsJqF2
-----END ECDSA PUBLIC KEY-----
`)

func equinoxUpdate() error {
	var opts equinox.Options
	if err := opts.SetPublicKeyPEM(publicKey); err != nil {
		return err
	}

	// check for the update
	resp, err := equinox.Check(appId, opts)
	switch {
	case err == equinox.NotAvailableErr:
		fmt.Println("No update available, already at the latest version!")
		return nil
	case err != nil:
		fmt.Println("Update failed:", err)
		return err
	}

	// fetch the update and apply it
	err = resp.Apply()
	if err != nil {
		return err
	}

	fmt.Printf("Updated to new version: %s!\n", resp.ReleaseVersion)
	return nil
}
