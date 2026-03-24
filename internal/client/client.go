package client

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	Server string
	Token  string
	http   *http.Client
}

type Item struct {
	Namespace string `json:"namespace"`
	Type      string `json:"type"`
	Name      string `json:"name"`
	Content   string `json:"content"`
	Author    string `json:"author"`
	Version   int    `json:"version"`
}

type Profile struct {
	Namespace string       `json:"namespace"`
	Name      string       `json:"name"`
	Items     []ProfileRef `json:"items"`
	Author    string       `json:"author"`
}

type ProfileRef struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type LoginResponse struct {
	Token string `json:"token"`
	Email string `json:"email"`
}

func New(server, token string) *Client {
	return &Client{
		Server: strings.TrimRight(server, "/"),
		Token:  token,
		http:   &http.Client{},
	}
}

func (c *Client) Login(email string) (*LoginResponse, error) {
	resp, err := c.do("POST", "/login", map[string]string{"email": email})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var lr LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&lr); err != nil {
		return nil, err
	}
	return &lr, nil
}

func (c *Client) ListItems(namespace, itemType string) ([]Item, error) {
	var path string
	if namespace == "" {
		path = "/" + pluralize(itemType)
	} else {
		path = "/" + namespace + "/" + pluralize(itemType)
	}
	resp, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var items []Item
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, err
	}
	return items, nil
}

func (c *Client) GetItem(namespace, itemType, name string) (*Item, error) {
	path := "/" + namespace + "/" + pluralize(itemType) + "/" + name
	resp, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var item Item
	if err := json.NewDecoder(resp.Body).Decode(&item); err != nil {
		return nil, err
	}
	return &item, nil
}

func (c *Client) PushItem(namespace, itemType, name string, content []byte) error {
	path := "/" + namespace + "/" + pluralize(itemType)
	body := map[string]string{
		"name":    name,
		"content": base64.StdEncoding.EncodeToString(content),
	}
	resp, err := c.do("POST", path, body)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (c *Client) DeleteItem(namespace, itemType, name string) error {
	path := "/" + namespace + "/" + pluralize(itemType) + "/" + name
	resp, err := c.do("DELETE", path, nil)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (c *Client) Search(query string) ([]Item, error) {
	path := "/search?q=" + url.QueryEscape(query)
	resp, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var items []Item
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		return nil, err
	}
	return items, nil
}

func (c *Client) GetProfile(namespace, name string) (*Profile, error) {
	path := "/" + namespace + "/profiles/" + name
	resp, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var p Profile
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (c *Client) ListProfiles(namespace string) ([]Profile, error) {
	path := "/" + namespace + "/profiles"
	resp, err := c.do("GET", path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var profiles []Profile
	if err := json.NewDecoder(resp.Body).Decode(&profiles); err != nil {
		return nil, err
	}
	return profiles, nil
}

func (c *Client) CreateProfile(namespace, name string) error {
	path := "/" + namespace + "/profiles"
	resp, err := c.do("POST", path, map[string]string{"name": name})
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

type TokenResponse struct {
	Token string `json:"token"`
	Name  string `json:"name"`
}

type TokenListEntry struct {
	Name      string    `json:"name"`
	Prefix    string    `json:"prefix"`
	CreatedAt time.Time `json:"created_at"`
}

func (c *Client) CreateToken(name string) (*TokenResponse, error) {
	resp, err := c.do("POST", "/tokens", map[string]string{"name": name})
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var tr TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return nil, err
	}
	return &tr, nil
}

func (c *Client) ListTokens() ([]TokenListEntry, error) {
	resp, err := c.do("GET", "/tokens", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var tokens []TokenListEntry
	if err := json.NewDecoder(resp.Body).Decode(&tokens); err != nil {
		return nil, err
	}
	return tokens, nil
}

func (c *Client) DeleteToken(name string) error {
	resp, err := c.do("DELETE", "/tokens/"+name, nil)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func (c *Client) AddProfileItem(namespace, profileName string, ref ProfileRef) error {
	path := "/" + namespace + "/profiles/" + profileName + "/items"
	resp, err := c.do("POST", path, ref)
	if err != nil {
		return err
	}
	resp.Body.Close()
	return nil
}

func pluralize(itemType string) string {
	return itemType + "s"
}

func (c *Client) do(method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.Server+path, bodyReader)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		msg, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, strings.TrimSpace(string(msg)))
	}

	return resp, nil
}
