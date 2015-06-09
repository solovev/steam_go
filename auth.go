package steam_auth

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var (
	steam_login = "https://steamcommunity.com/openid/login"

	openId_mode       = "checkid_setup"
	openId_ns         = "http://specs.openid.net/auth/2.0"
	openId_identifier = "http://specs.openid.net/auth/2.0/identifier_select"

	ErrValidation            = errors.New("Validation error: Unable validate openId.")
	ErrValidationIdPattern   = errors.New("Validation error: Invalid steam id pattern.")
	ErrValidationWrongNS     = errors.New("Validation error: Wrong ns in the response.")
	ErrValidationUnknownMode = errors.New("Validation error: Mode must equal to \"id_res\".")
	ErrValidationReturnUrl   = errors.New("Validation error: The return_to url must match the url of current request.")

	validation_regexp        = regexp.MustCompile("^(http|https)://steamcommunity.com/openid/id/[0-9]{15,25}$")
	digits_extraction_regexp = regexp.MustCompile("\\D+")
)

type OpenId struct {
	root      string
	returnUrl string
	data      url.Values
}

func NewOpenId(r *http.Request) *OpenId {
	id := new(OpenId)

	proto := "http://"
	if r.TLS != nil {
		proto = "https://"
	}
	id.root = proto + r.Host

	uri := r.RequestURI
	if i := strings.Index(uri, "openid"); i != -1 {
		uri = uri[0 : i-1]
	}
	id.returnUrl = id.root + uri

	switch r.Method {
	case "POST":
		id.data = r.Form
	case "GET":
		id.data = r.URL.Query()
	}

	return id
}

func (id *OpenId) Validate() (string, error) {
	if id.Mode() != "id_res" {
		return "", ErrValidationUnknownMode
	}

	if id.data.Get("openid.return_to") != id.returnUrl {
		return "", ErrValidationReturnUrl
	}

	params := make(url.Values)
	params.Set("openid.assoc_handle", id.data.Get("openid.assoc_handle"))
	params.Set("openid.signed", id.data.Get("openid.signed"))
	params.Set("openid.sig", id.data.Get("openid.sig"))
	params.Set("openid.ns", id.data.Get("openid.ns"))

	split := strings.Split(id.data.Get("openid.signed"), ",")
	for _, item := range split {
		params.Set("openid."+item, id.data.Get("openid."+item))
	}
	params.Set("openid.mode", "check_authentication")

	resp, err := http.PostForm(steam_login, params)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	response := strings.Split(string(content), "\n")
	if response[0] != "ns:"+openId_ns {
		return "", ErrValidationWrongNS
	}
	if strings.HasSuffix(response[1], "false") {
		return "", ErrValidation
	}

	openIdUrl := id.data.Get("openid.claimed_id")
	if !validation_regexp.MatchString(openIdUrl) {
		return "", ErrValidationIdPattern
	}

	return digits_extraction_regexp.ReplaceAllString(openIdUrl, ""), nil
}

func (id OpenId) Mode() string {
	return id.data.Get("openid.mode")
}

func (id OpenId) AuthUrl() string {
	data := map[string]string{
		"openid.claimed_id": openId_identifier,
		"openid.identity":   openId_identifier,
		"openid.mode":       openId_mode,
		"openid.ns":         openId_ns,
		"openid.realm":      id.root,
		"openid.return_to":  id.returnUrl,
	}

	i := 0
	url := steam_login + "?"
	for key, value := range data {
		url += key + "=" + value
		if i != len(data)-1 {
			url += "&"
		}
		i++
	}
	return url
}
