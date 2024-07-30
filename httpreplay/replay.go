package httpreplay

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"runtime/debug"
	"strings"
	"testing"
)

type ExpectedCall struct {
	Call           string
	ReqBodyJSONStr string
	Form           map[string]string
	ReqHeaders     map[string]string
	ResponseCode   int
	Response       string
}

func NewTransport(t testing.TB, expectedCalls ...*ExpectedCall) *Transport {
	tr := &Transport{
		T:             t,
		expectedCalls: expectedCalls,
	}
	t.Cleanup(tr.Verify)
	return tr
}

type Transport struct {
	T             testing.TB
	IncludeHost   bool
	expectedCalls []*ExpectedCall
	count         int
}

func (tr *Transport) Expect(call *ExpectedCall) {
	tr.expectedCalls = append(tr.expectedCalls, call)
}

func (tr *Transport) Verify() {
	tr.T.Helper()
	if tr.count != len(tr.expectedCalls) {
		tr.T.Errorf("** call count = %d, wanted %d (missing %d)", tr.count, len(tr.expectedCalls), len(tr.expectedCalls)-tr.count)
	}
	tr.expectedCalls = nil
	tr.count = 0
}

func (tr *Transport) RoundTrip(r *http.Request) (*http.Response, error) {
	if r.Body == nil {
		r.Body = io.NopCloser(bytes.NewReader(nil))
	}
	defer r.Body.Close()

	tr.count++
	var actualCall string
	if tr.IncludeHost {
		u2 := *r.URL
		u2.RawQuery = ""
		actualCall = r.Method + " " + u2.String()
	} else {
		actualCall = r.Method + " " + r.URL.Path
	}
	if tr.count > len(tr.expectedCalls) {
		tr.T.Errorf("*** unexpected call #%d %s:\n%s", tr.count, actualCall, debug.Stack())
		return errorResponse(http.StatusInternalServerError), nil
	}

	expected := tr.expectedCalls[tr.count-1]
	if actualCall != expected.Call {
		tr.T.Errorf("*** unexpected call #%d %s, wanted %s:\n%s", tr.count, actualCall, expected.Call, debug.Stack())
		return errorResponse(http.StatusInternalServerError), nil
	}

	if expected.ReqBodyJSONStr != "" {
		assertJSONRequestBody(tr.T, r.Body, expected.ReqBodyJSONStr)
	}

	var mismatchedFormKeys []string
	for k, v := range expected.Form {
		if r.FormValue(k) != v {
			mismatchedFormKeys = append(mismatchedFormKeys, k)
		}
	}
	if mismatchedFormKeys != nil {
		tr.T.Errorf("*** call #%d %s has incorrect value for form key %s; got %s, wanted %s", tr.count, expected.Call, strings.Join(mismatchedFormKeys, ", "), mustMarshalString(r.Form), mustMarshalString(expected.Form))
		return errorResponse(http.StatusInternalServerError), nil
	}

	var mismatchedHeaders []string
	for k, v := range expected.ReqHeaders {
		if r.Header.Get(k) != v {
			mismatchedHeaders = append(mismatchedHeaders, k)
		}
	}
	if mismatchedHeaders != nil {
		tr.T.Errorf("*** call #%d %s has incorrect value for header %s; got %s, wanted %s", tr.count, expected.Call, strings.Join(mismatchedHeaders, ", "), mustMarshalString(r.Header), mustMarshalString(expected.ReqHeaders))
		return errorResponse(http.StatusInternalServerError), nil
	}

	w := httptest.NewRecorder()
	if expected.ResponseCode != 0 {
		w.WriteHeader(expected.ResponseCode)
	}
	w.Write([]byte(expected.Response))
	return w.Result(), nil
}

func errorResponse(status int) *http.Response {
	w := httptest.NewRecorder()
	w.WriteHeader(http.StatusInternalServerError)
	return w.Result()
}

func assertJSONRequestBody(t testing.TB, actual io.ReadCloser, expected string) {
	t.Helper()
	var actualJSON, expectedJSON interface{}
	err := json.Unmarshal([]byte(expected), &expectedJSON)
	if err != nil {
		panic(err)
	}
	err = json.NewDecoder(actual).Decode(&actualJSON)
	if err != nil {
		panic(err)
	}
	if a, e := actualJSON, expectedJSON; !reflect.DeepEqual(a, e) {
		t.Errorf("*** request body = %v, wanted %v", a, e)
	}
}
