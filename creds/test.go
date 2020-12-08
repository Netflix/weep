/*
 * Copyright 2020 Netflix, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package creds

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

type ClientMock struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (c *ClientMock) Do(req *http.Request) (*http.Response, error) {
	return c.DoFunc(req)
}

func GetTestClient(responseBody interface{}) (*Client, error) {
	resp, err := json.Marshal(responseBody)
	if err != nil {
		return nil, err
	}
	client := &Client{
		Httpc: &ClientMock{
			DoFunc: func(*http.Request) (*http.Response, error) {
				r := ioutil.NopCloser(bytes.NewReader(resp))
				return &http.Response{
					StatusCode: 200,
					Body:       r,
				}, nil
			},
		},
	}
	return client, nil
}
