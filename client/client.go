package client

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/ubccr/grendel/api"
	"github.com/ubccr/grendel/bmc"
	"github.com/ubccr/grendel/model"
	"github.com/ubccr/grendel/nodeset"
)

type Client struct {
	endpoint string
	clientID string
	secret   string
	client   *http.Client
}

type NetbootResult map[string]string
type BMCResult map[string]*bmc.System

func NewClient(endpoint, clientID, secret, cacert string, insecure bool) (*Client, error) {
	tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure}}

	pem, err := ioutil.ReadFile(cacert)
	if err == nil {
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("Failed to read cacert: %s", cacert)
		}

		tr = &http.Transport{TLSClientConfig: &tls.Config{RootCAs: certPool, InsecureSkipVerify: false}}
	}

	c := &Client{
		clientID: clientID,
		secret:   secret,
		endpoint: strings.TrimSuffix(endpoint, "/"),
		client:   &http.Client{Timeout: time.Second * 3600, Transport: tr},
	}

	return c, nil
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

func (c *Client) Netboot(params api.NetbootParams) (NetbootResult, error) {
	data, err := json.Marshal(&params)
	if err != nil {
		return nil, err
	}

	req, err := c.postRequest(c.URL(GRENDEL_API_NETBOOT), bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusInternalServerError {
		return nil, fmt.Errorf("Failed to netboot hosts: %d", res.StatusCode)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed to netboot hosts unknown error code: %d", res.StatusCode)
	}

	rawJson, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	log.Debugf("netboot json response: %s", rawJson)

	var netbootResult NetbootResult
	err = json.Unmarshal(rawJson, &netbootResult)
	if err != nil {
		return nil, err
	}

	return netbootResult, nil
}

func (c *Client) BMCStatus(params api.NetbootParams) (BMCResult, error) {
	data, err := json.Marshal(&params)
	if err != nil {
		return nil, err
	}

	req, err := c.postRequest(c.URL(GRENDEL_API_POWER), bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusInternalServerError {
		return nil, fmt.Errorf("Failed to fetch power: %d", res.StatusCode)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed to fetch power unknown error code: %d", res.StatusCode)
	}

	rawJson, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	log.Debugf("power json response: %s", rawJson)

	var bmcResult BMCResult
	err = json.Unmarshal(rawJson, &bmcResult)
	if err != nil {
		return nil, err
	}

	return bmcResult, nil
}

func (c *Client) HostFind(ns *nodeset.NodeSet) (model.HostList, error) {
	endpoint := fmt.Sprintf("%s/%s", GRENDEL_API_HOST_FIND, ns.String())
	req, err := c.getRequest(c.URL(endpoint))
	if err != nil {
		return nil, err
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusInternalServerError {
		return nil, fmt.Errorf("Failed to find hosts: %d", res.StatusCode)
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed to find hosts unknown error code: %d", res.StatusCode)
	}

	rawJson, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	log.Debugf("find response: %s", rawJson)

	var hostList model.HostList
	err = json.Unmarshal(rawJson, &hostList)
	if err != nil {
		return nil, err
	}

	return hostList, nil
}
