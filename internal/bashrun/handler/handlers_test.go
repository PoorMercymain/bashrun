package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/PoorMercymain/bashrun/internal/bashrun/domain"
	"github.com/PoorMercymain/bashrun/internal/bashrun/domain/mocks"
	"github.com/PoorMercymain/bashrun/internal/bashrun/service"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/semaphore"
)

type testTableElem struct {
	caseName string
	httpMethod string
	route string
	body string
	headers [][2]string
	expectedStatus int
	requireParsing bool
	parsedBody interface{}
}

func testRouter(t *testing.T) *http.ServeMux {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mux := http.NewServeMux()

	var wg sync.WaitGroup

	ar := mocks.NewMockBashrunRepository(ctrl)
	as := service.New(context.Background(), ar, semaphore.NewWeighted(4), &wg)
	ah := New(as)

	ar.EXPECT().Ping(gomock.Any()).Return(errors.New("")).MaxTimes(1)
	ar.EXPECT().Ping(gomock.Any()).Return(nil).AnyTimes()

	//1
	ar.EXPECT().CreateCommand(gomock.Any(), gomock.Any()).Return(0, errors.New("")).MaxTimes(1)

	//2
	ar.EXPECT().CreateCommand(gomock.Any(), gomock.Any()).Return(1, nil).MaxTimes(7)
	ar.EXPECT().ReadStatus(gomock.Any(), 1).Return("stopped", nil).MaxTimes(1)
	ar.EXPECT().UpdateStatus(gomock.Any(), 1, gomock.Any()).Return(nil).MaxTimes(1)

	//3
	ar.EXPECT().ReadStatus(gomock.Any(), 1).Return("", errors.New("")).MaxTimes(1)
	ar.EXPECT().UpdateStatus(gomock.Any(), 1, gomock.Any()).Return(nil).MaxTimes(1)

	//4
	ar.EXPECT().ReadStatus(gomock.Any(), 1).Return("", nil).MaxTimes(1)
	ar.EXPECT().UpdatePID(gomock.Any(), 1, gomock.Any()).Return(errors.New("")).MaxTimes(1)
	ar.EXPECT().UpdateStatus(gomock.Any(), 1, gomock.Any()).Return(nil).MaxTimes(1)

	//5
	ar.EXPECT().ReadStatus(gomock.Any(), 1).Return("", nil).MaxTimes(1)
	ar.EXPECT().UpdatePID(gomock.Any(), 1, gomock.Any()).Return(nil).AnyTimes()
	ar.EXPECT().UpdateStatus(gomock.Any(), 1, gomock.Any()).Return(errors.New("")).MaxTimes(1)
	ar.EXPECT().UpdateStatus(gomock.Any(), 1, gomock.Any()).Return(nil).MaxTimes(1)

	//6
	ar.EXPECT().ReadStatus(gomock.Any(), 1).Return("", nil).MaxTimes(1)
	ar.EXPECT().UpdateStatus(gomock.Any(), 1, gomock.Any()).Return(nil).MaxTimes(1)
	ar.EXPECT().UpdateOutput(gomock.Any(), 1, gomock.Any()).Return(errors.New("")).MaxTimes(1)
	ar.EXPECT().UpdateStatus(gomock.Any(), 1, gomock.Any()).Return(nil).MaxTimes(1)

	//7
	ar.EXPECT().ReadStatus(gomock.Any(), 1).Return("", nil).MaxTimes(1)
	ar.EXPECT().UpdateStatus(gomock.Any(), 1, gomock.Any()).Return(nil).MaxTimes(1)
	ar.EXPECT().UpdateOutput(gomock.Any(), 1, gomock.Any()).Return(nil).AnyTimes()
	ar.EXPECT().UpdateExitStatus(gomock.Any(), 1, gomock.Any()).Return(errors.New("")).MaxTimes(1)
	ar.EXPECT().UpdateStatus(gomock.Any(), 1, gomock.Any()).Return(nil).MaxTimes(1)

	mux.Handle("GET /ping", http.HandlerFunc(ah.Ping))
	mux.Handle("POST /commands", http.HandlerFunc(ah.CreateCommand))
	mux.Handle("GET /commands", http.HandlerFunc(ah.ListCommands))
	mux.Handle("GET /commands/stop/{command_id}", http.HandlerFunc(ah.StopCommand))
	mux.Handle("GET /commands/{command_id}", http.HandlerFunc(ah.ReadCommand))
	mux.Handle("GET /commands/output/{command_id}", http.HandlerFunc(ah.ReadOutput))

	return mux
}

func buildRequest(httpMethod string, route string, body string, headers [][2]string, addr string) (*http.Request, error) {
	req, err := http.NewRequest(httpMethod, fmt.Sprintf("%s%s", addr, route), strings.NewReader(body))
	if err != nil {
		return nil, err
	}

	for _, header := range headers {
		req.Header.Add(header[0], header[1])
	}

	return req, nil
}

func sendReq(t *testing.T, client *http.Client, req *http.Request, expectedStatus int, parsedBody interface{}, requireParsing bool) {
	resp, err := client.Do(req)
	require.NoError(t, err)

	require.Equal(t, expectedStatus, resp.StatusCode)

	if requireParsing {
		err = json.NewDecoder(resp.Body).Decode(parsedBody)
		require.NoError(t, err)
	}

	resp.Body.Close()
}

func Test_bashrunHandlers_Ping(t *testing.T) {
	ts := httptest.NewServer(testRouter(t))
	defer ts.Close()

	client := http.Client{}

	tests := []testTableElem{
		{
			caseName: "server error",
			httpMethod: http.MethodGet,
			route: "/ping",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusInternalServerError,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "ok",
			httpMethod: http.MethodGet,
			route: "/ping",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusNoContent,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "wrong http method",
			httpMethod: http.MethodPost,
			route: "/ping",
			body: "",
			headers: [][2]string{},
			expectedStatus: http.StatusMethodNotAllowed,
			requireParsing: false,
			parsedBody: nil,
		},
	}

	for _, testCase := range tests {
		t.Log(testCase.caseName)

		req, err := buildRequest(testCase.httpMethod, testCase.route, testCase.body, testCase.headers, ts.URL)
		require.NoError(t, err)

		sendReq(t, &client, req, testCase.expectedStatus, testCase.parsedBody, testCase.requireParsing)
	}
}

func Test_bashrunHandlers_CreateCommand(t *testing.T) {
	ts := httptest.NewServer(testRouter(t))
	defer ts.Close()

	client := http.Client{}

	var id domain.ID
	tests := []testTableElem{
		{
			caseName: "unknown key in JSON",
			httpMethod: http.MethodPost,
			route: "/commands",
			body: "{\"cmd\": \"abc\"}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusBadRequest,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "duplicate key in JSON",
			httpMethod: http.MethodPost,
			route: "/commands",
			body: "{\"command\": \"abc\", \"command\": \"abc2\"}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusBadRequest,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "wrong JSON",
			httpMethod: http.MethodPost,
			route: "/commands",
			body: "{\"command\": \"abc\",",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusBadRequest,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "duplicate key in JSON",
			httpMethod: http.MethodPost,
			route: "/commands",
			body: "{\"command\": \"abc\", \"command\": \"abc2\"}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusBadRequest,
			requireParsing: false,
			parsedBody: nil,
		},
		{
			caseName: "empty command",
			httpMethod: http.MethodPost,
			route: "/commands",
			body: "{\"command\": \"\"}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusBadRequest,
			requireParsing: false,
			parsedBody: nil,
		},
		{//1
			caseName: "server error",
			httpMethod: http.MethodPost,
			route: "/commands",
			body: "{\"command\": \"abc\"}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusInternalServerError,
			requireParsing: false,
			parsedBody: nil,
		},
		{//2
			caseName: "ok (but command is stopped before running)",
			httpMethod: http.MethodPost,
			route: "/commands",
			body: "{\"command\": \"exit 0\"}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusAccepted,
			requireParsing: true,
			parsedBody: &id,
		},
		{//3
			caseName: "ok (but status check failed before running)",
			httpMethod: http.MethodPost,
			route: "/commands",
			body: "{\"command\": \"exit 0\"}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusAccepted,
			requireParsing: true,
			parsedBody: &id,
		},
		{//4
			caseName: "ok (but update PID failed)",
			httpMethod: http.MethodPost,
			route: "/commands",
			body: "{\"command\": \"exit 0\"}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusAccepted,
			requireParsing: true,
			parsedBody: &id,
		},
		{//5
			caseName: "ok (but update status failed)",
			httpMethod: http.MethodPost,
			route: "/commands",
			body: "{\"command\": \"exit 0\"}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusAccepted,
			requireParsing: true,
			parsedBody: &id,
		},
		{//6
			caseName: "ok (but update output failed)",
			httpMethod: http.MethodPost,
			route: "/commands",
			body: "{\"command\": \"ls\"}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusAccepted,
			requireParsing: true,
			parsedBody: &id,
		},
		{//7
			caseName: "ok (but update exit status failed)",
			httpMethod: http.MethodPost,
			route: "/commands",
			body: "{\"command\": \"ls\"}",
			headers: [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusAccepted,
			requireParsing: true,
			parsedBody: &id,
		},
	}

	for _, testCase := range tests {
		t.Log(testCase.caseName)

		req, err := buildRequest(testCase.httpMethod, testCase.route, testCase.body, testCase.headers, ts.URL)
		require.NoError(t, err)

		sendReq(t, &client, req, testCase.expectedStatus, testCase.parsedBody, testCase.requireParsing)

		if testCase.caseName == "ok (but command stopped)" {
			require.Equal(t, 2, id.ID)
		} else if testCase.caseName == "ok" {
			require.Equal(t, 3, id.ID)
		} else if strings.HasPrefix(testCase.caseName, "ok (but") {
			require.Equal(t, 1, id.ID)
		}

		<-time.After(time.Millisecond*100)
	}
}
