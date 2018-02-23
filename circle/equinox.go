package main

import (
	"fmt"

	"github.com/equinox-io/equinox"
)

const appId = "app_n7HhD13kpUR"

var publicKey = []byte(`
-----BEGIN ECDSA PUBLIC KEY-----
MHYwEAYHKoZIzj0CAQYFK4EEACIDYgAEfTU8qJtih8zB1/mx97/MVDgsHPrA0fAf
fmYDrVRyFwaN3t8+TVwccJJCALGIqdszqgNPUla/O9k1bjSoMOVI3neJkvbCoM3T
bQdN4mYh+1j9c6w5EvN+bwCWX2qM+gJS
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
