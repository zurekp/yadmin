package client

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type ClientOptions struct {
	BaseUrl                   string
	AdminLogin                string
	AdminPassword             string
	SkipCertificateValidation bool
}

type Client struct {
	httpClient    http.Client
	baseUrl       *url.URL
	adminLogin    string
	adminPassword string
}

type Status struct {
	LoggedIn, Initialized bool
}

func New(options ClientOptions) (*Client, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: options.SkipCertificateValidation},
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	url, err := url.Parse(options.BaseUrl)
	if err != nil {
		return nil, err
	}

	login := strings.TrimSpace(options.AdminLogin)
	if len(login) == 0 {
		return nil, errors.New("Admin login musn't be empty.")
	}

	password := strings.TrimSpace(options.AdminPassword)
	if len(password) == 0 {
		return nil, errors.New("Admin password musn't be empty.")
	}

	return &Client{
		httpClient:    http.Client{Transport: transport, Jar: jar},
		adminLogin:    options.AdminLogin,
		adminPassword: options.AdminPassword,
		baseUrl:       url,
	}, nil
}

func (client *Client) Status() (*Status, error) {
	doc, err := client.readHtml(*client.baseUrl)
	if err != nil {
		return nil, err
	}

	//walk(doc)

	passwordNodes := findMatchingNodes(doc, loginPagePasswordInputMatcher)
	if len(passwordNodes) > 1 {
		return nil, errors.New("More than one passworrd input detected.")
	}
	if len(passwordNodes) == 1 {
		return &Status{Initialized: true, LoggedIn: false}, nil
	}

	return &Status{false, false}, nil
}

func (client *Client) String() string {
	return fmt.Sprintf("{user: \"%s\", baseUrl: \"%s\"}", client.adminLogin, client.baseUrl)
}

func (client *Client) readHtml(url url.URL) (*html.Node, error) {
	response, err := client.httpClient.Get(url.String())
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return nil, errors.New(fmt.Sprintf("Expectd 200/OK status code, but %d was received.", response.StatusCode))
	}

	document, err := html.Parse(response.Body)
	if err != nil {
		return nil, err
	}

	return document, nil
}

func findAttribute(attributes []html.Attribute, key string) (value string, ok bool) {
	for _, a := range attributes {
		if a.Key == key {
			value, ok = a.Val, true
			return
		}
	}
	return
}

func loginPagePasswordInputMatcher(node *html.Node) bool {
	if node == nil || node.DataAtom != atom.Input {
		return false
	}
	if v, ok := findAttribute(node.Attr, "type"); !ok || v != "password" {
		return false
	}
	if _, ok := findAttribute(node.Attr, "name"); !ok {
		return false
	}
	return true
}

func findMatchingNodes(start *html.Node, matcher func(*html.Node) bool) (matchingElements []*html.Node) {
	if start == nil || matcher == nil {
		return []*html.Node{}
	}
	toVisit := []*html.Node{start}
	node := start
	for len(toVisit) > 0 {
		node, toVisit = toVisit[0], toVisit[1:]
		if node == nil {
			continue
		}
		if matcher(node) {
			matchingElements = append(matchingElements, node)
		}
		for n := node.FirstChild; n != nil; n = n.NextSibling {
			toVisit = append(toVisit, n)
		}
	}
	return matchingElements
}

func walk(node *html.Node) {
	if html.ElementNode == node.Type {
		pathToParent := node.Data
		for p := node.Parent; p != nil; p = p.Parent {
			if html.ElementNode == p.Type {
				pathToParent = p.Data + "." + pathToParent
			}
		}
		fmt.Println(pathToParent, "->", node.Attr)
	}
	for n := node.FirstChild; n != nil; n = n.NextSibling {
		walk(n)
	}
}
