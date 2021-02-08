package cortex

import (
	"net/url"

	"github.com/ghodss/yaml"

	v1 "github.com/bolindalabs/cortex-alert-operator/api/v1"
)

func (c *Client) SetRuleGroup(namespace string, group v1.RuleGroup) error {
	payload, err := yaml.Marshal(&group)
	if err != nil {
		return err
	}

	escapedNamespace := url.PathEscape(namespace)
	path := c.apiPath + "/" + escapedNamespace

	res, err := c.doRequest(path, "POST", payload)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	return nil
}

func (c *Client) DeleteRuleGroup(namespace string, groupName string) error {
	escapedNamespace := url.PathEscape(namespace)
	escapedGroupName := url.PathEscape(groupName)
	path := c.apiPath + "/" + escapedNamespace + "/" + escapedGroupName

	_, err := c.doRequest(path, "DELETE", nil)
	return err
}

func (c *Client) DeleteRuleNamespace(namespace string) error {
	escapedNamespace := url.PathEscape(namespace)
	path := c.apiPath + "/" + escapedNamespace

	_, err := c.doRequest(path, "DELETE", nil)
	return err
}
