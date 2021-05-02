package cortex

import (
	"net/url"

	"github.com/ghodss/yaml"
	"github.com/go-logr/logr"

	v1 "github.com/bolindalabs/cortex-alert-operator/api/v1"
)

func (c *Client) SetRuleGroup(log logr.Logger, namespace string, group v1.RuleGroup) error {
	payload, err := yaml.Marshal(&group)
	if err != nil {
		return err
	}

	escapedNamespace := url.PathEscape(namespace)
	path := c.apiPath + "/" + escapedNamespace

	res, err := c.doRequest(log, path, "POST", payload)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	return nil
}

func (c *Client) DeleteRuleGroup(log logr.Logger, namespace string, groupName string) error {
	escapedNamespace := url.PathEscape(namespace)
	escapedGroupName := url.PathEscape(groupName)
	path := c.apiPath + "/" + escapedNamespace + "/" + escapedGroupName

	_, err := c.doRequest(log, path, "DELETE", nil)
	return err
}

func (c *Client) DeleteRuleNamespace(log logr.Logger, namespace string) error {
	escapedNamespace := url.PathEscape(namespace)
	path := c.apiPath + "/" + escapedNamespace

	_, err := c.doRequest(log, path, "DELETE", nil)
	return err
}
