package client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/spf13/viper"
	"github.com/ubccr/grendel/logger"
	"github.com/ubccr/grendel/model"
	"github.com/ubccr/grendel/nodeset"
)

var log = logger.GetLogger("CLIENT")

type Client struct {
	endpoint string
	clientID string
	secret   string
	client   *retryablehttp.Client
}

func NewClient() (*Client, error) {
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: viper.GetBool("api.insecure")}}

	cacert := viper.GetString("client.cacert")
	pem, err := ioutil.ReadFile(cacert)
	if err == nil {
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("Failed to read cacert: %s", cacert)
		}

		tr = &http.Transport{TLSClientConfig: &tls.Config{RootCAs: certPool, InsecureSkipVerify: false}}
	}

	endpoint := viper.GetString("client.api_endpoint")

	c := &Client{
		clientID: viper.GetString("client.client_id"),
		secret:   viper.GetString("client.client_secret"),
		endpoint: strings.TrimSuffix(endpoint, "/"),
	}

	// Is endpoint a path to a unix domain socket?
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		tr = &http.Transport{
			DialContext: func(ctx context.Context, _, addr string) (net.Conn, error) {
				dialer := net.Dialer{}
				return dialer.DialContext(ctx, "unix", endpoint)
			},
		}
		c.endpoint = "http://unix"
	}

	c.client = retryablehttp.NewClient()
	c.client.HTTPClient = &http.Client{Timeout: time.Second * 3600, Transport: tr}
	c.client.Logger = log

	return c, nil
}

func (c *Client) RetryMax(max int) {
	c.client.RetryMax = max
}

func (c *Client) URL(resource string) string {
	log.Debugf("Resource: %s", resource)
	return fmt.Sprintf("%s%s", c.endpoint, resource)
}

func (c *Client) getRequest(url string) (*http.Request, error) {
	return c.newRequest(http.MethodGet, url, nil)
}

func (c *Client) postRequest(url string, body io.Reader) (*http.Request, error) {
	return c.newRequest(http.MethodPost, url, body)
}

func (c *Client) newRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func (c *Client) HostFind(ns *nodeset.NodeSet) (model.HostList, error) {
	endpoint := fmt.Sprintf("%s/%s", GRENDEL_API_HOST_FIND, ns.String())
	return c.hostList(endpoint)
}

func (c *Client) HostList() (model.HostList, error) {
	return c.hostList(GRENDEL_API_HOST_LIST)
}

func (c *Client) hostList(endpoint string) (model.HostList, error) {
	req, err := c.getRequest(c.URL(endpoint))
	if err != nil {
		return nil, err
	}

	rreq, err := retryablehttp.FromRequest(req)
	if err != nil {
		return nil, err
	}

	res, err := c.client.Do(rreq)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusInternalServerError {
		return nil, fmt.Errorf("Failed to fetch hosts: %d", res.StatusCode)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed to fetch hosts unknown error code: %d", res.StatusCode)
	}

	rawJson, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	// log.Debugf("JSON response: %s", rawJson)

	var hostList model.HostList
	err = json.Unmarshal(rawJson, &hostList)
	if err != nil {
		return nil, err
	}

	return hostList, nil
}
