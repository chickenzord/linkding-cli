package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client is the linkding API client.
type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

// New creates a new Client.
func New(baseURL, token string) *Client {
	// Normalize base URL: strip trailing slash
	baseURL = strings.TrimRight(baseURL, "/")
	return &Client{
		BaseURL: baseURL,
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// APIError represents an error returned by the API.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error %d: %s", e.StatusCode, e.Body)
}

func (c *Client) request(method, path string, body interface{}) (*http.Response, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, c.BaseURL+path, reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Token "+c.Token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (c *Client) do(method, path string, body interface{}, out interface{}) error {
	resp, err := c.request(method, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &APIError{StatusCode: resp.StatusCode, Body: string(data)}
	}

	if out != nil && len(data) > 0 {
		return json.Unmarshal(data, out)
	}
	return nil
}

// ---- Bookmarks ----

type Bookmark struct {
	ID               int      `json:"id"`
	URL              string   `json:"url"`
	Title            string   `json:"title"`
	Description      string   `json:"description"`
	Notes            string   `json:"notes"`
	WebsiteTitle     string   `json:"website_title"`
	WebsiteDesc      string   `json:"website_description"`
	IsArchived       bool     `json:"is_archived"`
	Unread           bool     `json:"unread"`
	Shared           bool     `json:"shared"`
	TagNames         []string `json:"tag_names"`
	DateAdded        string   `json:"date_added"`
	DateModified     string   `json:"date_modified"`
	FaviconURL       string   `json:"favicon_url"`
	PreviewImageURL  string   `json:"preview_image_url"`
}

type BookmarkList struct {
	Count    int        `json:"count"`
	Next     string     `json:"next"`
	Previous string     `json:"previous"`
	Results  []Bookmark `json:"results"`
}

type BookmarkListOptions struct {
	Query    string
	Archived bool
	Unread   bool
	Shared   bool
	Limit    int
	Offset   int
}

func (c *Client) ListBookmarks(opts BookmarkListOptions) (*BookmarkList, error) {
	q := url.Values{}
	if opts.Query != "" {
		q.Set("q", opts.Query)
	}
	if opts.Archived {
		q.Set("is_archived", "true")
	}
	if opts.Unread {
		q.Set("unread", "true")
	}
	if opts.Shared {
		q.Set("shared", "true")
	}
	if opts.Limit > 0 {
		q.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		q.Set("offset", strconv.Itoa(opts.Offset))
	}

	path := "/api/bookmarks/?" + q.Encode()
	var result BookmarkList
	if err := c.do("GET", path, nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetBookmark(id int) (*Bookmark, error) {
	var b Bookmark
	if err := c.do("GET", fmt.Sprintf("/api/bookmarks/%d/", id), nil, &b); err != nil {
		return nil, err
	}
	return &b, nil
}

type BookmarkInput struct {
	URL         string   `json:"url"`
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	Notes       string   `json:"notes,omitempty"`
	IsArchived  bool     `json:"is_archived,omitempty"`
	Unread      bool     `json:"unread,omitempty"`
	Shared      bool     `json:"shared,omitempty"`
	TagNames    []string `json:"tag_names,omitempty"`
}

func (c *Client) CreateBookmark(input BookmarkInput) (*Bookmark, error) {
	var b Bookmark
	if err := c.do("POST", "/api/bookmarks/", input, &b); err != nil {
		return nil, err
	}
	return &b, nil
}

func (c *Client) UpdateBookmark(id int, input BookmarkInput) (*Bookmark, error) {
	var b Bookmark
	if err := c.do("PATCH", fmt.Sprintf("/api/bookmarks/%d/", id), input, &b); err != nil {
		return nil, err
	}
	return &b, nil
}

func (c *Client) DeleteBookmark(id int) error {
	return c.do("DELETE", fmt.Sprintf("/api/bookmarks/%d/", id), nil, nil)
}

type CheckResult struct {
	Bookmark     *Bookmark `json:"bookmark"`
	Metadata     Metadata  `json:"metadata"`
	AutoTags     []string  `json:"auto_tags"`
}

type Metadata struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (c *Client) CheckBookmark(rawURL string) (*CheckResult, error) {
	q := url.Values{}
	q.Set("url", rawURL)
	var result CheckResult
	if err := c.do("GET", "/api/bookmarks/check/?"+q.Encode(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) ArchiveBookmark(id int) error {
	return c.do("POST", fmt.Sprintf("/api/bookmarks/%d/archive/", id), nil, nil)
}

func (c *Client) UnarchiveBookmark(id int) error {
	return c.do("POST", fmt.Sprintf("/api/bookmarks/%d/unarchive/", id), nil, nil)
}

// ---- Tags ----

type Tag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type TagList struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []Tag  `json:"results"`
}

type TagListOptions struct {
	Query  string
	Limit  int
	Offset int
}

func (c *Client) ListTags(opts TagListOptions) (*TagList, error) {
	q := url.Values{}
	if opts.Query != "" {
		q.Set("q", opts.Query)
	}
	if opts.Limit > 0 {
		q.Set("limit", strconv.Itoa(opts.Limit))
	}
	if opts.Offset > 0 {
		q.Set("offset", strconv.Itoa(opts.Offset))
	}
	var result TagList
	if err := c.do("GET", "/api/tags/?"+q.Encode(), nil, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) GetTag(id int) (*Tag, error) {
	var t Tag
	if err := c.do("GET", fmt.Sprintf("/api/tags/%d/", id), nil, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

type TagInput struct {
	Name string `json:"name"`
}

func (c *Client) CreateTag(input TagInput) (*Tag, error) {
	var t Tag
	if err := c.do("POST", "/api/tags/", input, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

// ---- User Profile ----

type UserProfile struct {
	Theme                    string `json:"theme"`
	CustomCSS                string `json:"custom_css"`
	DefaultMarkAsUnread      bool   `json:"default_mark_as_unread"`
	DefaultSharedBookmarks   bool   `json:"default_share_bookmarks"`
	EnableSharing            bool   `json:"enable_sharing"`
	EnablePublicSharing      bool   `json:"enable_public_sharing"`
	EnableFavicons           bool   `json:"enable_favicons"`
	DisplayURL               bool   `json:"display_url"`
	PermanentNotes           bool   `json:"permanent_notes"`
	SearchPreferences        struct {
		Sort   string `json:"sort"`
		Shared string `json:"shared"`
		Unread string `json:"unread"`
	} `json:"search_preferences"`
}

func (c *Client) GetUserProfile() (*UserProfile, error) {
	var p UserProfile
	if err := c.do("GET", "/api/user/profile/", nil, &p); err != nil {
		return nil, err
	}
	return &p, nil
}
