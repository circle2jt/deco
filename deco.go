package deco

import (
	"bytes"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/circle2jt/deco/utils"
)

var (
	// {"operation": "read"}
	readBody []byte = []byte{123, 34, 111, 112, 101, 114, 97, 116, 105, 111, 110, 34, 58, 34, 114, 101, 97, 100, 34, 125}
)

type passwordKeyResponse struct {
	Result struct {
		Username string   `json:"username"`
		Password []string `json:"password"`
	} `json:"result"`
	ErrCode int `json:"error_code"`
}

type sessionKeyResponse struct {
	Result struct {
		Seq uint     `json:"seq"`
		Key []string `json:"key"`
	} `json:"result"`
	ErrCode int `json:"error_code"`
}

type loginParams struct {
	Password string `json:"password"`
}

type loginRequest struct {
	Params    loginParams `json:"params"`
	Operation string      `json:"operation"`
}

type loginResponse struct {
	Result struct {
		Stok string `json:"stok"`
	} `json:"result"`
	ErrCode int `json:"error_code"`
}

type response struct {
	Data string `json:"data"`
}

type request struct {
	Operation string            `json:"operation,omitempty"`
	Params    map[string]string `json:"params,omitempty"`
}

// EndpointArgs holds the url params to be sent
type EndpointArgs struct {
	form string
}

func (e *EndpointArgs) queryParams() url.Values {
	q := make(url.Values)

	q.Add("form", e.form)
	return q
}

func (c *Client) getPasswordKey() (*rsa.PublicKey, error) {
	args := EndpointArgs{
		form: "keys",
	}
	var passKey passwordKeyResponse
	_, err := c.doPost(";stok=/login", args, readBody, &passKey)
	if err != nil {
		return nil, err
	}

	key, err := utils.GenerateRsaKey(passKey.Result.Password)
	if err != nil {
		return nil, err
	}

	return key, nil
}

func (c *Client) getSessionKey() (*rsa.PublicKey, uint, error) {
	args := EndpointArgs{
		form: "auth",
	}
	var passKey sessionKeyResponse
	_, err := c.doPost(";stok=/login", args, readBody, &passKey)
	if err != nil {
		return nil, 0, err
	}

	key, err := utils.GenerateRsaKey(passKey.Result.Key)
	if err != nil {
		return nil, 0, err
	}

	return key, passKey.Result.Seq, nil
}

func (c *Client) doEncryptedPost(path string, params EndpointArgs, body []byte, isLogin bool, result interface{}) error {
	encryptedData, err := utils.AES256Encrypt(string(body), *c.aes)
	if err != nil {
		return err
	}

	length := int(c.sequence) + len(encryptedData)
	var sign string

	if isLogin {
		sign = fmt.Sprintf("k=%s&i=%s&h=%s&s=%v", c.aes.Key, c.aes.Iv, c.hash, length)
	} else {
		sign = fmt.Sprintf("h=%s&s=%v", c.hash, length)
	}

	if len(sign) > 53 {
		first, _ := utils.EncryptRsa(sign[:53], c.rsa)
		second, _ := utils.EncryptRsa(sign[53:], c.rsa)
		sign = fmt.Sprintf("%s%s", first, second)
	} else {
		sign, _ = utils.EncryptRsa(sign, c.rsa)
	}

	postData := fmt.Sprintf("sign=%s&data=%s", url.QueryEscape(sign), url.QueryEscape(encryptedData))
	var req response
	resData, err := c.doPost(path, params, []byte(postData), &req)
	var decoded string
	if err == nil {
		if req.Data != "" {
			decoded, err = utils.AES256Decrypt(req.Data, *c.aes)
			if err == nil {
				err = json.Unmarshal([]byte(decoded), &result)
			}
		} else {
			err = fmt.Errorf("empty: %s", "response data is empty")
		}
	}
	if err != nil {
		err = fmt.Errorf(">>> error: %s, request: %s, response: %s, decoded: %s", err.Error(), req.Data, resData, decoded)
	}
	return err
}

func (c *Client) doPost(path string, params EndpointArgs, body []byte, result interface{}) (string, error) {
	var content string
	endpt := baseURL.ResolveReference(&url.URL{Path: path, RawQuery: params.queryParams().Encode()})

	req, err := http.NewRequest("POST", endpt.String(), bytes.NewBuffer(body))

	if err == nil {
		req.Header.Set("Connection", "keep-alive")
		req.Header.Set("Content-Length", strconv.Itoa(len(body)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Host", baseURL.Host)
		req.Header.Set("Origin", "http://"+baseURL.Host)
		res, _err := c.c.Do(req)
		if _err == nil {
			defer res.Body.Close()

			buf := new(bytes.Buffer)
			buf.ReadFrom(res.Body)

			content = buf.String()

			err = json.NewDecoder(buf).Decode(&result)
		} else {
			err = _err
		}
	}
	return content, err
}
