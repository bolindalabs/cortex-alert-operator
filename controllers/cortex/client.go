package cortex

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"

	"github.com/go-logr/logr"
)

const (
	rulerAPIPath = "/api/v1/rules"
)

var (
	ErrNoConfig         = errors.New("no config exists for this user")
	ErrResourceNotFound = errors.New("requested resource not found")
)

// Config is used to configure a Ruler Client
type Config struct {
	Key             string `yaml:"key"`
	Address         string `yaml:"address"`
	ID              string `yaml:"id"`
	UseLegacyRoutes bool   `yaml:"use_legacy_routes"`
}

type Client struct {
	Client http.Client

	key      string
	id       string
	endpoint *url.URL
	apiPath  string
}

func New(cfg Config) (*Client, error) {
	endpoint, err := url.Parse(cfg.Address)
	if err != nil {
		return nil, err
	}

	client := http.Client{}

	c := &Client{
		key:      cfg.Key,
		id:       cfg.ID,
		endpoint: endpoint,
		Client:   client,
		apiPath:  rulerAPIPath,
	}
	return c, nil
}

func (c *Client) doRequest(log logr.Logger, path, method string, payload []byte) (*http.Response, error) {
	req, err := buildRequest(path, method, *c.endpoint, payload)
	if err != nil {
		return nil, err
	}

	if c.key != "" {
		req.SetBasicAuth(c.id, c.key)
	}

	req.Header.Add("X-Scope-OrgID", c.id)

	log.WithValues(
		"url", req.URL.String(),
		"method", req.Method,
	).Info("sending request to cortex api")

	resp, err := c.Client.Do(req)
	if err != nil {
		// log.WithFields(log.Fields{
		// 	"url":    req.URL.String(),
		// 	"method": req.Method,
		// 	"error":  err.Error(),
		// }).Errorln("error during request to cortex api")
		return nil, err
	}

	err = checkResponse(log, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// checkResponse checks the API response for errors
func checkResponse(log logr.Logger, r *http.Response) error {
	log.WithValues(
		"status", r.Status,
	).Info("checking response")
	if 200 <= r.StatusCode && r.StatusCode <= 299 {
		return nil
	}

	var msg, errMsg string
	scanner := bufio.NewScanner(io.LimitReader(r.Body, 512))
	if scanner.Scan() {
		msg = scanner.Text()
	}

	if msg == "" {
		errMsg = fmt.Sprintf("server returned HTTP status %s", r.Status)
	} else {
		errMsg = fmt.Sprintf("server returned HTTP status %s: %s", r.Status, msg)
	}

	if r.StatusCode == http.StatusNotFound {
		log.WithValues(
			"status", r.Status,
			"msg", msg,
		).Info(errMsg)
		return ErrResourceNotFound
	}

	err := errors.New(errMsg)
	log.WithValues(
		"status", r.Status,
		"msg", msg,
	).Error(err, "cortex request failed")

	return err
}

func buildRequest(p, m string, endpoint url.URL, payload []byte) (*http.Request, error) {
	// parse path parameter again (as it already contains escaped path information
	pURL, err := url.Parse(p)
	if err != nil {
		return nil, err
	}

	// if path or endpoint contains escaping that requires RawPath to be populated, also join rawPath
	if pURL.RawPath != "" || endpoint.RawPath != "" {
		endpoint.RawPath = path.Join(endpoint.EscapedPath(), pURL.EscapedPath())
	}
	endpoint.Path = path.Join(endpoint.Path, pURL.Path)
	return http.NewRequest(m, endpoint.String(), bytes.NewBuffer(payload))
}
