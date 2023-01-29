package gw2api

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
)

//Timeout solution adapted from Volker on stackoverflow
func (gw2 *GW2Api) dialTimeout(network, addr string) (net.Conn, error) {
	return net.DialTimeout(network, addr, gw2.Timeout)
}

func (gw2 *GW2Api) fetchRawEndpoint(url string) (io io.ReadCloser, err error) {
	var resp *http.Response
	if resp, err = gw2.Client.Get(url); err != nil {
		return
	}
	return resp.Body, nil
}

func (gw2 *GW2Api) fetchEndpoint(ver, tag string, params url.Values, result interface{}) (err error) {
	var endpoint *url.URL
	endpoint, _ = url.Parse("https://api.guildwars2.com")
	endpoint.Path += "/" + ver + "/" + tag
	if params != nil {
		endpoint.RawQuery = params.Encode()
	}

	var resp *http.Response
	if resp, err = gw2.Client.Get(endpoint.String()); err != nil {
		return err
	}
	var data []byte
	if data, err = ioutil.ReadAll(resp.Body); err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		if err = json.Unmarshal(data, &result); err != nil {
			var gwerr APIError
			if err = json.Unmarshal(data, &gwerr); err != nil {
				return err
			}
			return fmt.Errorf("Endpoint returned error: %v", gwerr)
		}
	} else {
		err = fmt.Errorf("GW2API response status code: %d. Body: %s", resp.StatusCode, string(data[:]))
	}
	return err
}

func (gw2 *GW2Api) fetchAuthenticatedEndpoint(ver, tag string, perm Permission, params url.Values, result interface{}) (err error) {
	if len(gw2.Auth) < 1 {
		return fmt.Errorf("API requires authentication")
	}

	if perm >= PermAccount && !flagGet(gw2.AuthFlags, uint(perm)) {
		return fmt.Errorf("Missing permissions for authenticated Endpoint")
	}

	if params == nil {
		params = url.Values{}
	}
	params.Add("access_token", gw2.Auth)
	return gw2.fetchEndpoint(ver, tag, params, result)
}
