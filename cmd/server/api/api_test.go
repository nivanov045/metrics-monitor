package api

import (
	"encoding/json"
	"github.com/nivanov045/silver-octo-train/cmd/server/config"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nivanov045/silver-octo-train/cmd/server/service"
	"github.com/nivanov045/silver-octo-train/cmd/server/storage"
	"github.com/nivanov045/silver-octo-train/internal/metrics"
)

func Test_api_updateMetricsHandler(t *testing.T) {
	type args struct {
		name       string
		valueInt   int64
		valueFloat float64
		mType      string
	}
	type want struct {
		statusCode int
	}
	tests := []struct {
		name string
		args args
		want
	}{
		{
			name: "correct counter request",
			args: args{
				name:       "testCounter",
				valueInt:   100,
				valueFloat: 0,
				mType:      "counter",
			},
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name: "correct gauge request",
			args: args{
				name:       "testGauge",
				valueInt:   100.0,
				valueFloat: 0,
				mType:      "gauge",
			},
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name: "unknown metrics type",
			args: args{
				name:       "testCounter",
				valueInt:   100,
				valueFloat: 0,
				mType:      "unknown",
			},
			want: want{
				statusCode: http.StatusNotImplemented,
			},
		},
	}
	myStorage, err := storage.New(config.Config{
		Address:       "",
		StoreInterval: 0 * time.Second,
		StoreFile:     "/tmp/devops-metrics-db.json",
		Restore:       false,
		Key:           "",
		Database:      "",
	})
	assert.NoError(t, err)
	serv := service.New("", myStorage)
	a := api{serv}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			marshal, err := json.Marshal(metrics.Interface{
				ID:    tt.args.name,
				MType: tt.args.mType,
				Delta: &tt.args.valueInt,
				Value: &tt.args.valueFloat,
			})
			assert.NoError(t, err)
			request := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/update/", strings.NewReader(string(marshal)))
			w := httptest.NewRecorder()
			h := http.HandlerFunc(a.updateMetricsHandler)
			h.ServeHTTP(w, request)
			result := w.Result()
			defer result.Body.Close()
			assert.Equal(t, tt.want.statusCode, result.StatusCode)
		})
	}
}

func Test_api_getMetricsHandler(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "correct",
		},
	}
	myStorage, err := storage.New(config.Config{
		Address:       "",
		StoreInterval: 0 * time.Second,
		StoreFile:     "/tmp/devops-metrics-db.json",
		Restore:       false,
		Key:           "",
		Database:      "",
	})
	assert.NoError(t, err)
	serv := service.New("", myStorage)
	a := api{serv}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := int64(100)
			marshal, err := json.Marshal(metrics.Interface{
				ID:    "TestMetrics",
				MType: "counter",
				Delta: &value,
				Value: nil,
			})
			assert.NoError(t, err)
			requestSend := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/update/", strings.NewReader(string(marshal)))
			wSend := httptest.NewRecorder()
			hSend := http.HandlerFunc(a.updateMetricsHandler)
			hSend.ServeHTTP(wSend, requestSend)
			request := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/value/", strings.NewReader(string(marshal)))
			w := httptest.NewRecorder()
			h := http.HandlerFunc(a.getMetricsHandler)
			h.ServeHTTP(w, request)
			result := w.Result()

			respBody, err := ioutil.ReadAll(result.Body)
			require.NoError(t, err)
			defer result.Body.Close()
			assert.Equal(t, http.StatusOK, result.StatusCode)
			var mi metrics.Interface
			err = json.Unmarshal(respBody, &mi)
			require.NoError(t, err)
			assert.Equal(t, int64(100), *mi.Delta)
			assert.Equal(t, "counter", mi.MType)
			assert.Equal(t, "TestMetrics", mi.ID)
		})
	}
}

func Test_api_rootHandler(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "correct",
		},
	}
	myStorage, err := storage.New(config.Config{
		Address:       "",
		StoreInterval: 0 * time.Second,
		StoreFile:     "/tmp/devops-metrics-db.json",
		Restore:       false,
		Key:           "",
		Database:      "",
	})
	assert.NoError(t, err)
	serv := service.New("", myStorage)
	a := api{serv}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := int64(100)
			marshal, err := json.Marshal(metrics.Interface{
				ID:    "TestMetrics",
				MType: "counter",
				Delta: &value,
				Value: nil,
			})
			assert.NoError(t, err)
			requestSend := httptest.NewRequest(http.MethodPost, "http://127.0.0.1/update/", strings.NewReader(string(marshal)))
			wSend := httptest.NewRecorder()
			hSend := http.HandlerFunc(a.updateMetricsHandler)
			hSend.ServeHTTP(wSend, requestSend)
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()
			h := http.HandlerFunc(a.rootHandler)
			h.ServeHTTP(w, request)
			result := w.Result()

			respBody, err := ioutil.ReadAll(result.Body)
			require.NoError(t, err)
			defer result.Body.Close()

			assert.Equal(t, http.StatusOK, result.StatusCode)
			assert.Equal(t, "TestMetrics", strings.Trim(string(respBody), "\n"))
		})
	}
}
