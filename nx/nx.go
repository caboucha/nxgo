package nx

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

// ClientOptions is used to specify options for the Client.
type ClientOptions struct {
	Hosts []string // List of apic hostnames. If unspecified, env var NEXUS_HOSTS is used.
	User  string   // Username. If unspecified, env var NEXUS_USER is used.
	Pass  string   // Password. If unspecified, env var NEXUS_PASS is used.
	Debug bool     // Debug enables verbose debugging messages to console.
}

// Client is an instance for interacting with Nexus switch using API calls.
type Client struct {
	Opt                 ClientOptions   // Options for the Nexus client
	host                int             // Index for current host
	cli                 *http.Client    // Client context for HTTP
	loginToken          string          // Save Nexus login token
	loginRefreshTimeout time.Duration   // Save Nexus refresh period
	socket              *websocket.Conn // websocket for receiving notifications
}

// Environment variables used as default parameters.
const (
	NexusHosts = "NEXUS_HOSTS" // Env var. List of Nexus hostnames or ip addresses. 
                                   // Example: "1.1.1.1" or "1.1.1.1,2.2.2.2,3.3.3.3" or
                                   // "hostnamea,4.4.4.4"
	NexusUser  = "NEXUS_USER"  // Env var. Username. Example: "joe"
	NexusPass  = "NEXUS_PASS"  // Env var. Password. Example: "joesecret"
	NexusDebug = "NEXUS_DEBUG" // Env var. Debug.
)

const (
	contentTypeJSON = "application/json" // Nexus API ignores Content-Type, but we set it rightly anyway
)

// New creates a new Client instance for interacting with Nexus switch using API calls.
func New(o ClientOptions) (*Client, error) {
	if len(o.Hosts) < 1 {
		hosts := os.Getenv(NexusHosts)
		if hosts == "" {
			return nil, fmt.Errorf("missing Nexus hosts: %s=%s", NexusHosts, o.Hosts)
		}
		o.Hosts = strings.Split(hosts, ",")
		if len(o.Hosts) < 1 {
			return nil, fmt.Errorf("missing Nexus hosts: %s=%s", NexusHosts, o.Hosts)
		}
		for _, h := range o.Hosts {
			if strings.TrimSpace(h) == "" {
				return nil, fmt.Errorf("blank Nexus hostname '%s' in %s=%s", 
                                                       h, NexusHosts, o.Hosts)
			}
		}
	}

	if o.User == "" {
		o.User = os.Getenv(NexusUser)
		if o.User == "" {
			return nil, fmt.Errorf("missing Nexus user: %s=%s", NexusUser, o.User)
		}
	}

	if o.Pass == "" {
		o.Pass = os.Getenv(NexusPass)
		if o.Pass == "" {
			return nil, fmt.Errorf("missing Nexus pass: %s=%s", NexusPass, o.Pass)
		}
	}

        if !o.Debug {
            _, o.Debug = os.LookupEnv(NexusDebug)
        }

	c := &Client{Opt: o}

	c.newHTTPClient()

	c.debugf("new client: hosts=%s user=%s pass=%s", c.Opt.Hosts, c.Opt.User, c.Opt.Pass)

	return c, nil
}

func (c *Client) getFuncName(level int) string {
    pc, _, _, _ := runtime.Caller(level)
    f := runtime.FuncForPC(pc)
    x := strings.SplitAfter(f.Name(), ".")
    fmt.Println(x, x[len(x)-1])
    if len(x) > 0 {
        return x[len(x)-1]
    }   
    
    return ""
}

func (c *Client) debugf(fmt string, v ...interface{}) {
	if c.Opt.Debug {
		c.logf("debug "+fmt, v...)
	}
}

func (c *Client) logf(fmt string, v ...interface{}) {
	log.Printf("nxsclient: "+fmt, v...)
}

func (c *Client) jsonAaaUser() string {
	return fmt.Sprintf(`{"aaaUser": {"attributes": {"name": "%s", "pwd": "%s"}}}`, c.Opt.User, c.Opt.Pass)
}

// Logout closes a session to Nexus Switch using the API aaaLogout.
func (c *Client) Logout() {

	api := "/api/aaaLogout.json"

	aaaUser := c.jsonAaaUser()

	//url := c.getURL(api)

	//c.debugf("logout: url=%s json=%s", url, aaaUser)

	body, errPost := c.post(api, contentTypeJSON, bytes.NewBufferString(aaaUser))
	if errPost != nil {
                c.logf("Failed to logout User: Error: %s", errPost)
		return
	}

	c.debugf("logout: reply: %s", string(body))

	return
}

// Login opens a new session into Nexus Switch using the API aaaLogin.
func (c *Client) Login() error {

	api := "/api/aaaLogin.json"

	aaaUser := c.jsonAaaUser()

	c.debugf("login: api=%s json=%s", api, aaaUser)

	body, errPost := c.postScan(api, contentTypeJSON, bytes.NewBufferString(aaaUser))
	if errPost != nil {
		return errPost
	}

	var reply interface{}
	errJSON := json.Unmarshal(body, &reply)
	if errJSON != nil {
		return errJSON
	}

	imdata, imdataError := mapGet(reply, "imdata")
	if imdataError != nil {
		return fmt.Errorf("login: json imdata error: %s", string(body))
	}

	first, firstError := sliceGet(imdata, 0)
	if firstError != nil {
		return fmt.Errorf("login: imdata first error: %s", string(body))
	}

	mm, mmMap := first.(map[string]interface{})
	if !mmMap {
		return fmt.Errorf("login: imdata slice first member not map: %s", string(body))
	}

	for k, v := range mm {
		switch k {
		case "error":
			attr := mapSimple(v, "attributes")
			code := mapString(attr, "code")
			text := mapString(attr, "text")
			return fmt.Errorf("login: error: code=%s text=%s", code, text)
		case "aaaLogin":
			attr := mapSimple(v, "attributes")
			token := mapString(attr, "token")
			refresh := mapString(attr, "refreshTimeoutSeconds")

			c.refresh(token, refresh)

			return nil // ok
		}
	}

	return fmt.Errorf("login: could not find aaaLogin response: %s", string(body))
}

// Refresh resets the session timer on Nexus Switch using the API aaaRefresh.
// In order to keep the session active, Refresh() must be called at a period lower than the timeout reported by RefreshTimeout().
func (c *Client) Refresh() error {

	api := "/api/aaaRefresh.json"

	url := c.getURL(api)

	body, errGet := c.get(url)
	if errGet != nil {
		return errGet
	}

	var reply interface{}
	errJSON := json.Unmarshal(body, &reply)
	if errJSON != nil {
		return errJSON
	}

	imdata, imdataError := mapGet(reply, "imdata")
	if imdataError != nil {
		return fmt.Errorf("refresh: json imdata error: %s", string(body))
	}

	first, firstError := sliceGet(imdata, 0)
	if firstError != nil {
		return fmt.Errorf("refresh: imdata first error: %s", string(body))
	}

	mm, mmMap := first.(map[string]interface{})
	if !mmMap {
		return fmt.Errorf("refresh: imdata slice first member not map: %s", string(body))
	}

	for k, v := range mm {
		switch k {
		case "error":
			attr := mapSimple(v, "attributes")
			code := mapString(attr, "code")
			text := mapString(attr, "text")
			return fmt.Errorf("refresh: error: code=%s text=%s", code, text)
		case "aaaLogin":
			attr := mapSimple(v, "attributes")
			token := mapString(attr, "token")
			refresh := mapString(attr, "refreshTimeoutSeconds")

			c.refresh(token, refresh)

			return nil // ok
		}
	}

	return fmt.Errorf("refresh: could not find aaaLogin response: %s", string(body))
}

func (c *Client) refresh(token, refreshTimeout string) {
	c.loginToken = token // save token

	timeout, timeoutErr := strconv.Atoi(refreshTimeout)
	if timeoutErr != nil {
		c.logf("refreshUpdate: bad refresh timeout '%s': %v", refreshTimeout, timeoutErr)
		timeout = 60 // defaults to 60 seconds
	}
	c.loginRefreshTimeout = time.Duration(timeout) * time.Second // save timeout

	c.debugf("refresh: timeout=%v token=%s", c.RefreshTimeout(), token)
}

// RefreshTimeout gets the session timeout reported by last API call to Nexus Switch.
// In order to keep the session active, Refresh() must be called at a period lower than the timeout reported by RefreshTimeout().
func (c *Client) RefreshTimeout() time.Duration {
	return c.loginRefreshTimeout
}

func tlsConfig() *tls.Config {
	return &tls.Config{
		CipherSuites:             []uint16{tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA, tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA},
		PreferServerCipherSuites: true,
		InsecureSkipVerify:       true,
		MaxVersion:               tls.VersionTLS11,
		MinVersion:               tls.VersionTLS11,
	}
}

func (c *Client) newHTTPClient() {
	tr := &http.Transport{
		TLSClientConfig:    tlsConfig(),
		DisableCompression: true,
		DisableKeepAlives:  true,
		Dial: (&net.Dialer{
			Timeout:   5 * time.Second,
			KeepAlive: 10 * time.Second,
		}).Dial,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	c.cli = &http.Client{
		Transport: tr,
		Timeout:   15 * time.Second,
	}
}

// getURL builds HTTPS URL for API access.
func (c *Client) getURL(api string) string {
	return makeURL("https", c.Opt.Hosts[c.host], api)
}

// getURLws builds websocket URL for notifications.
func (c *Client) getURLws(api string) string {
	return makeURL("wss", c.Opt.Hosts[c.host], api)
}

// url builds URL from protocol, host, path.
func makeURL(proto, host, path string) string {
	return proto + "://" + host + path
}

// postScan scans multiple Nexus Switch hosts.
func (c *Client) postScan(api string, contentType string, r io.Reader) ([]byte, error) {
	var last error

	if isURL(api) {
		return nil, fmt.Errorf("bad api=%s", api)
	}

	for ; c.host < len(c.Opt.Hosts); c.host++ {

		//url := c.getURL(api)

                url := api
		body, errPost := c.post(url, contentType, r)
		if errPost != nil {
			c.debugf("postScan: error: apic: %s: %v", url, errPost)
			last = errPost
			continue
		}

		return body, nil
	}

	return nil, fmt.Errorf("no more apic hosts to try - last: %v", last)
}

func (c *Client) showCookies(urlStr string) {
	if c.cli.Jar == nil {
		c.debugf("no cookies to send")
		return
	}

	u, errURL := url.Parse(urlStr)
	if errURL != nil {
		c.debugf("showCookies: %s: %v", urlStr, errURL)
		return
	}

	cookies := c.cli.Jar.Cookies(u)
	if len(cookies) < 1 {
		c.debugf("no cookies to send url=%s", u)
		return
	}

	for _, ck := range cookies {
		c.debugf("cookie to send: %s", ck.Name)
	}
}

func (c *Client) learnCookies(resp *http.Response) error {
	cookies := resp.Cookies()
	for _, ck := range cookies {
		c.debugf("learnCookies: seen: url=%s cookie=%s", resp.Request.URL, ck.Name)
                // CB_TBD What's Nexus substitute for APIC-cookie? s/b "Set-Cookie" ??
		if ck.Name == "APIC-cookie" {
			if c.cli.Jar == nil {
				var errNew error
				c.cli.Jar, errNew = cookiejar.New(nil) // new jar
				if errNew != nil {
					return errNew
				}
			}
			c.cli.Jar.SetCookies(resp.Request.URL, []*http.Cookie{ck}) // add single cookie to jar
			c.debugf("learnCookies: learnt: url=%s cookie=%s value=%s", resp.Request.URL, ck.Name, ck.Value)
			break
		}
	}
	return nil
}

func (c *Client) post(uri string, contentType string, r io.Reader) ([]byte, error) {

        url := c.getURL(uri)
	if !isURL(url) {
		return nil, fmt.Errorf("bad URL=%s", url)
	}

        callerFuncName := c.getFuncName(2)
	c.debugf("post: Caller %s apic endpoint: %s",
                 callerFuncName, url)

	c.showCookies(url)

	resp, errPost := c.cli.Post(url, contentType, r)
	if errPost != nil {
		return nil, errPost
	}
	defer resp.Body.Close()

	if errLearn := c.learnCookies(resp); errLearn != nil {
		return nil, errLearn
	}

	body, errBody := ioutil.ReadAll(resp.Body)
        if (errBody !=nil) {
            return nil, errBody
        }

        c.debugf("%s: reply: %s", callerFuncName, string(body))

	return body, errBody
}

func (c *Client) get(uri string) ([]byte, error) {

        url := c.getURL(uri)
	if !isURL(url) {
		return nil, fmt.Errorf("bad URL=%s", url)
	}

        callerFuncName := c.getFuncName(2)
	c.debugf("get: Caller %s apic endpoint: %s", 
                 callerFuncName, url)

	c.showCookies(url)

	resp, errPost := c.cli.Get(url)
	if errPost != nil {
		return nil, errPost
	}
	defer resp.Body.Close()

	if errLearn := c.learnCookies(resp); errLearn != nil {
		return nil, errLearn
	}

	body, errBody := ioutil.ReadAll(resp.Body)
        if (errBody !=nil) {
            return nil, errBody
        }

        //c.debugf("%s: reply: %s", callerFuncName, string(body))

        return body, nil
}

func isURL(url string) bool {
	return strings.HasPrefix(url, "https://")
}

func (c *Client) delete(uri string) ([]byte, error) {

        url := c.getURL(uri)
	if !isURL(url) {
		return nil, fmt.Errorf("bad URL=%s", url)
	}
        callerFuncName := c.getFuncName(2)
	c.debugf("delete: Caller %s apic endpoint: %s",
                 callerFuncName, url)

	c.showCookies(url)

	req, errNew := http.NewRequest("DELETE", url, nil)
	if errNew != nil {
		return nil, errNew
	}

	resp, errDel := c.cli.Do(req)
	if errDel != nil {
		return nil, errDel
	}
	defer resp.Body.Close()

	if errLearn := c.learnCookies(resp); errLearn != nil {
		return nil, errLearn
	}

	body, errBody := ioutil.ReadAll(resp.Body)
        if (errBody !=nil) {
            return nil, errBody
        }

        c.debugf("%s: reply: %s", callerFuncName, string(body))

	return body, errBody
}
