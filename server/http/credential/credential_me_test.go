package credential

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"ritchie-server/server"
	"ritchie-server/server/mock"
	"testing"
)

func TestHandler_HandlerMe(t *testing.T) {
	type fields struct {
		v       server.VaultManager
		c       server.Config
		method  string
		org     string
		ctx     string
		payload string
	}
	tests := []struct {
		name   string
		fields fields
		out    http.HandlerFunc
	}{
		{
			name:   "method not found",
			fields: fields{method: http.MethodPatch},
			out: func() http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					http.NotFound(w, r)
				}
			}(),
		},
		{
			name: "get credential success",
			fields: fields{method: http.MethodGet, v: vaultManagerMock{
				Error:      nil,
				ReturnMap:  map[string]interface{}{"a": "b"},
				ReturnList: nil,
			}},
			out: func() http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.Header().Add("Content-Type", "application/json")
				}
			}(),
		},
		{
			name: "post credential success",
			fields: fields{method: http.MethodPost, v: vaultManagerMock{
				Error:      nil,
				ReturnMap:  map[string]interface{}{"a": "b"},
				ReturnList: nil,
			},
				payload: mock.DummyCredential(),
				ctx:     "default",
				org:     "zup",
				c:       mock.DummyConfig(),
			},
			out: func() http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusCreated)
				}
			}(),
		},
		{
			name: "post credential invalid json",
			fields: fields{method: http.MethodPost, v: vaultManagerMock{
				Error:      nil,
				ReturnMap:  map[string]interface{}{"a": "b"},
				ReturnList: nil,
			},
				payload: "failed",
				ctx:     "default",
				org:     "zup",
				c:       mock.DummyConfig(),
			},
			out: func() http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}
			}(),
		},
		{
			name: "post credential bad request",
			fields: fields{method: http.MethodPost, v: vaultManagerMock{
				Error:      nil,
				ReturnMap:  map[string]interface{}{"a": "b"},
				ReturnList: nil,
			},
				payload: mock.DummyCredentialBadRequest(),
				ctx:     "default",
				org:     "zup",
				c:       mock.DummyConfig(),
			},
			out: func() http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusBadRequest)
					w.Header().Add("Content-Type", "application/json")
				}
			}(),
		},
		{
			name: "post credential error write",
			fields: fields{method: http.MethodPost, v: vaultManagerMock{
				Error:      errors.New("error"),
				ReturnMap:  map[string]interface{}{"a": "b"},
				ReturnList: nil,
			},
				payload: mock.DummyCredential(),
				ctx:     "default",
				org:     "zup",
				c:       mock.DummyConfig(),
			},
			out: func() http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}
			}(),
		},
		{
			name: "get credential error",
			fields: fields{method: http.MethodGet, v: vaultManagerMock{
				Error:      errors.New("error"),
				ReturnMap:  nil,
				ReturnList: nil,
			}},
			out: func() http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}
			}(),
		},
		{
			name: "get credential not found",
			fields: fields{method: http.MethodGet, v: vaultManagerMock{
				Error:      nil,
				ReturnMap:  nil,
				ReturnList: nil,
			}},
			out: func() http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNotFound)
				}
			}(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := NewCredentialHandler(tt.fields.v, tt.fields.c)
			b := []byte{}
			if len(tt.fields.payload) > 0 {
				b = append(b, []byte(tt.fields.payload)...)
			}
			r, _ := http.NewRequest(tt.fields.method, "/test", bytes.NewReader(b))
			r.Header.Add(authorizationHeader, "Bearer "+bearerTest)
			r.Header.Add(server.ContextHeader, tt.fields.ctx)
			r.Header.Add(server.OrganizationHeader, tt.fields.org)
			r.Header.Add("Content-Type", "application/json")

			w := httptest.NewRecorder()

			tt.out.ServeHTTP(w, r)

			g := httptest.NewRecorder()

			h.HandlerMe().ServeHTTP(g, r)

			if g.Code != w.Code {
				t.Errorf("Handler returned wrong status code: got %v out %v", g.Code, w.Code)
			}

			if g.Header().Get("Content-Type") != w.Header().Get("Content-Type") {
				t.Errorf("Wrong content type. Got %v out %v", g.Header().Get("Content-Type"), w.Header().Get("Content-Type"))
			}
		})
	}
}