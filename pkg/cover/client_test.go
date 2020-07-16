/*
 Copyright 2020 Qiniu Cloud (qiniu.com)

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package cover

import (
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"net/http"
)

func TestClientAction(t *testing.T) {
	// mock goc server
	ts := httptest.NewServer(GocServer(os.Stdout))
	defer ts.Close()
	var client = NewWorker(ts.URL)

	// mock profile server
	profileMockResponse := "mode: count\nmockService/main.go:30.13,48.33 13 1"
	profileSuccessMockSvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(profileMockResponse))
	}))
	defer profileSuccessMockSvr.Close()

	profileErrMockSvr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("error"))
	}))
	defer profileErrMockSvr.Close()

	// regsiter service into goc server
	var src Service
	src.Name = "serviceSuccess"
	src.Address = profileSuccessMockSvr.URL
	res, err := client.RegisterService(src)
	assert.NoError(t, err)
	assert.Contains(t, string(res), "success")

	// do list and check service
	res, err = client.ListServices()
	assert.NoError(t, err)
	assert.Contains(t, string(res), src.Address)
	assert.Contains(t, string(res), src.Name)

	// get porfile from goc server
	profileItems := []struct {
		service Service
		param   ProfileParam
		res     string
	}{
		{
			service: Service{Name: "serviceOK", Address: profileSuccessMockSvr.URL},
			param:   ProfileParam{Force: false, Service: []string{"serviceOK"}, Address: []string{profileSuccessMockSvr.URL}},
			res:     "use 'service' and 'address' flag at the same time is illegal",
		},
		{
			service: Service{Name: "serviceOK", Address: profileSuccessMockSvr.URL},
			param:   ProfileParam{},
			res:     profileMockResponse,
		},
		{
			service: Service{Name: "serviceOK", Address: profileSuccessMockSvr.URL},
			param:   ProfileParam{Service: []string{"serviceOK"}},
			res:     profileMockResponse,
		},
		{
			service: Service{Name: "serviceOK", Address: profileSuccessMockSvr.URL},
			param:   ProfileParam{Address: []string{profileSuccessMockSvr.URL}},
			res:     profileMockResponse,
		},
		{
			service: Service{Name: "serviceOK", Address: profileSuccessMockSvr.URL},
			param:   ProfileParam{Service: []string{"unknown"}},
			res:     "service [unknown] not found",
		},
		{
			service: Service{Name: "serviceErr", Address: profileErrMockSvr.URL},
			res:     "bad mode line: error",
		},
		{
			service: Service{Name: "serviceErr", Address: profileErrMockSvr.URL},
			param:   ProfileParam{Force: true},
			res:     "no profiles",
		},
		{
			service: Service{Name: "serviceNotExist", Address: "http://172.0.0.2:7777"},
			res:     "connection refused",
		},
		{
			service: Service{Name: "serviceNotExist", Address: "http://172.0.0.2:7777"},
			param:   ProfileParam{Force: true},
			res:     "no profiles",
		},
	}
	for _, item := range profileItems {
		// init server
		res, err := client.InitSystem()
		assert.NoError(t, err)
		// register server
		res, err = client.RegisterService(item.service)
		assert.NoError(t, err)
		assert.Contains(t, string(res), "success")
		res, err = client.Profile(item.param)
		if err != nil {
			assert.Equal(t, err.Error(), item.res)
		} else {
			assert.Contains(t, string(res), item.res)
		}
	}

	// init system and check service again
	res, err = client.InitSystem()
	assert.NoError(t, err)
	res, err = client.ListServices()
	assert.NoError(t, err)
	assert.Equal(t, "{}", string(res))
}

func TestE2E(t *testing.T) {
	// FIXME: start goc server
	// FIXME: call goc build to cover goc server
	// FIXME: do some tests again goc server
	// FIXME: goc profile and checkout coverage
}
