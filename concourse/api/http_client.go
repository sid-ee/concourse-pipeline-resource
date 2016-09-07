package api

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/concourse/fly/rc"

	"golang.org/x/oauth2"
)

type TargetToken struct {
	Type  string
	Value string
}

func LoginWithBasicAuth(
	url string,
	username string,
	password string,
	insecure bool,
) (TargetToken, error) {
	var unusedTarget rc.TargetName
	teamName := "main"
	caCert := ""

	target, err := rc.NewBasicAuthTarget(
		unusedTarget,
		url,
		teamName,
		insecure,
		username,
		password,
		caCert,
	)
	if err != nil {
		panic(err)
	}

	token, err := target.Team().AuthToken()
	if err != nil {
		panic(err)
	}

	return TargetToken{
		Type:  token.Type,
		Value: token.Value,
	}, nil
}

func OAuthHTTPClient(token TargetToken, insecure bool) *http.Client {
	var transport http.RoundTripper

	transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecure,
		},
		Dial: (&net.Dialer{
			Timeout: 10 * time.Second,
		}).Dial,
		Proxy: http.ProxyFromEnvironment,
	}

	oAuthToken := &oauth2.Token{
		TokenType:   token.Type,
		AccessToken: token.Value,
	}

	transport = &oauth2.Transport{
		Source: oauth2.StaticTokenSource(oAuthToken),
		Base:   transport,
	}

	return &http.Client{Transport: transport}
}
