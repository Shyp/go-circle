package circle

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

type CircleConfig struct {
	Organizations map[string]organization
}

type organization struct {
	Token string
}

// getCaseInsensitiveOrg finds the key in the list of orgs. This is a case
// insensitive match, so if key is "ShyP" and orgs has a key named "sHYp",
// that will count as a match.
func getCaseInsensitiveOrg(key string, orgs map[string]organization) (organization, error) {
	for k, _ := range orgs {
		lower := strings.ToLower(k)
		if _, ok := orgs[lower]; !ok {
			orgs[lower] = orgs[k]
			delete(orgs, k)
		}
	}
	lowerKey := strings.ToLower(key)
	if o, ok := orgs[lowerKey]; ok {
		return o, nil
	} else {
		return organization{}, fmt.Errorf("Couldn't find organization %s in the config", key)
	}
}

func getToken(orgName string) (string, error) {
	user, err := user.Current()
	if err != nil {
		return "", err
	}
	filename := filepath.Join(user.HomeDir, "cfg", "circleci")
	f, err := os.Open(filename)
	rcFilename := ""
	if err != nil {
		rcFilename = filepath.Join(user.HomeDir, ".circlerc")
		f, err = os.Open(rcFilename)
	}
	if err != nil {
		err = fmt.Errorf(`Couldn't find a config file in %s or %s.

Add a configuration file with your CircleCI token, like this:

[organizations]

    [organizations.Shyp]
    token = "aabbccddeeff00"

Go to https://circleci.com/account/api if you need to create a token.
`, filename, rcFilename)
		return "", err
	}
	var c CircleConfig
	_, err = toml.DecodeReader(f, &c)
	if err != nil {
		return "", err
	}
	org, err := getCaseInsensitiveOrg(orgName, c.Organizations)
	if err != nil {
		return "", err
	}
	return org.Token, nil
}
