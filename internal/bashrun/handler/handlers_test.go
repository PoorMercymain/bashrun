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

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/semaphore"

	appErrors "github.com/PoorMercymain/bashrun/errors"
	"github.com/PoorMercymain/bashrun/internal/bashrun/domain"
	"github.com/PoorMercymain/bashrun/internal/bashrun/domain/mocks"
	"github.com/PoorMercymain/bashrun/internal/bashrun/service"
)

type testTableElem struct {
	caseName       string
	httpMethod     string
	route          string
	body           string
	headers        [][2]string
	expectedStatus int
	requireParsing bool
	parsedBody     interface{}
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

	//8
	ar.EXPECT().ListCommands(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("")).MaxTimes(1)

	//9
	ar.EXPECT().ListCommands(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, appErrors.ErrNoRows).MaxTimes(1)

	var exitStatus int
	//10
	ar.EXPECT().ListCommands(gomock.Any(), gomock.Any(), gomock.Any()).Return([]domain.CommandFromDB{{ID: 1, Command: "ls", PID: 5, Output: "", Status: "done", ExitStatus: &exitStatus}}, nil).AnyTimes()

	//11
	ar.EXPECT().ReadStatus(gomock.Any(), 5).Return("", errors.New("")).MaxTimes(1)

	//12
	ar.EXPECT().ReadStatus(gomock.Any(), 5).Return("", appErrors.ErrNoRows).MaxTimes(1)

	//13
	ar.EXPECT().ReadStatus(gomock.Any(), 5).Return("created", nil).MaxTimes(2)
	ar.EXPECT().UpdateStatus(gomock.Any(), 5, gomock.Any()).Return(errors.New("")).MaxTimes(1)

	//14
	ar.EXPECT().UpdateStatus(gomock.Any(), 5, gomock.Any()).Return(nil).MaxTimes(1)

	//15
	ar.EXPECT().ReadStatus(gomock.Any(), 5).Return("", nil).MaxTimes(1)

	//16
	ar.EXPECT().ReadStatus(gomock.Any(), 5).Return("started", nil).MaxTimes(1)
	ar.EXPECT().ReadPID(gomock.Any(), 5).Return(5, errors.New("")).MaxTimes(1)

	//17
	ar.EXPECT().ReadCommand(gomock.Any(), gomock.Any()).Return(domain.CommandFromDB{}, errors.New("")).MaxTimes(1)

	//18
	ar.EXPECT().ReadCommand(gomock.Any(), gomock.Any()).Return(domain.CommandFromDB{}, appErrors.ErrNoRows).MaxTimes(1)

	//19
	ar.EXPECT().ReadCommand(gomock.Any(), gomock.Any()).Return(domain.CommandFromDB{ID: 1, Command: "ls", PID: 5, Output: "", Status: "done", ExitStatus: &exitStatus}, nil).AnyTimes()

	//20
	ar.EXPECT().ReadOutput(gomock.Any(), gomock.Any()).Return("", errors.New("")).MaxTimes(1)

	//21
	ar.EXPECT().ReadOutput(gomock.Any(), gomock.Any()).Return("", appErrors.ErrNoRows).MaxTimes(1)

	//22
	ar.EXPECT().ReadOutput(gomock.Any(), gomock.Any()).Return("", nil).MaxTimes(1)

	//23
	ar.EXPECT().ReadOutput(gomock.Any(), gomock.Any()).Return("a", nil).AnyTimes()

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
			caseName:       "server error",
			httpMethod:     http.MethodGet,
			route:          "/ping",
			body:           "",
			headers:        [][2]string{},
			expectedStatus: http.StatusInternalServerError,
			requireParsing: false,
			parsedBody:     nil,
		},
		{
			caseName:       "ok",
			httpMethod:     http.MethodGet,
			route:          "/ping",
			body:           "",
			headers:        [][2]string{},
			expectedStatus: http.StatusNoContent,
			requireParsing: false,
			parsedBody:     nil,
		},
		{
			caseName:       "wrong http method",
			httpMethod:     http.MethodPost,
			route:          "/ping",
			body:           "",
			headers:        [][2]string{},
			expectedStatus: http.StatusMethodNotAllowed,
			requireParsing: false,
			parsedBody:     nil,
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
			caseName:       "unknown key in JSON",
			httpMethod:     http.MethodPost,
			route:          "/commands",
			body:           "{\"cmd\": \"abc\"}",
			headers:        [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusBadRequest,
			requireParsing: false,
			parsedBody:     nil,
		},
		{
			caseName:       "duplicate key in JSON",
			httpMethod:     http.MethodPost,
			route:          "/commands",
			body:           "{\"command\": \"abc\", \"command\": \"abc2\"}",
			headers:        [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusBadRequest,
			requireParsing: false,
			parsedBody:     nil,
		},
		{
			caseName:       "wrong JSON",
			httpMethod:     http.MethodPost,
			route:          "/commands",
			body:           "{\"command\": \"abc\",",
			headers:        [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusBadRequest,
			requireParsing: false,
			parsedBody:     nil,
		},
		{
			caseName:       "duplicate key in JSON",
			httpMethod:     http.MethodPost,
			route:          "/commands",
			body:           "{\"command\": \"abc\", \"command\": \"abc2\"}",
			headers:        [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusBadRequest,
			requireParsing: false,
			parsedBody:     nil,
		},
		{
			caseName:       "empty command",
			httpMethod:     http.MethodPost,
			route:          "/commands",
			body:           "{\"command\": \"\"}",
			headers:        [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusBadRequest,
			requireParsing: false,
			parsedBody:     nil,
		},
		{ //1
			caseName:       "server error",
			httpMethod:     http.MethodPost,
			route:          "/commands",
			body:           "{\"command\": \"abc\"}",
			headers:        [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusInternalServerError,
			requireParsing: false,
			parsedBody:     nil,
		},
		{ //2
			caseName:       "ok (but command is stopped before running)",
			httpMethod:     http.MethodPost,
			route:          "/commands",
			body:           "{\"command\": \"exit 0\"}",
			headers:        [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusAccepted,
			requireParsing: true,
			parsedBody:     &id,
		},
		{ //3
			caseName:       "ok (but status check failed before running)",
			httpMethod:     http.MethodPost,
			route:          "/commands",
			body:           "{\"command\": \"exit 0\"}",
			headers:        [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusAccepted,
			requireParsing: true,
			parsedBody:     &id,
		},
		{ //4
			caseName:       "ok (but update PID failed)",
			httpMethod:     http.MethodPost,
			route:          "/commands",
			body:           "{\"command\": \"exit 0\"}",
			headers:        [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusAccepted,
			requireParsing: true,
			parsedBody:     &id,
		},
		{ //5
			caseName:       "ok (but update status failed)",
			httpMethod:     http.MethodPost,
			route:          "/commands",
			body:           "{\"command\": \"exit 0\"}",
			headers:        [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusAccepted,
			requireParsing: true,
			parsedBody:     &id,
		},
		{ //6
			caseName:       "ok (but update output failed)",
			httpMethod:     http.MethodPost,
			route:          "/commands",
			body:           "{\"command\": \"ls\"}",
			headers:        [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusAccepted,
			requireParsing: true,
			parsedBody:     &id,
		},
		{ //7
			caseName:       "ok (but update exit status failed)",
			httpMethod:     http.MethodPost,
			route:          "/commands",
			body:           "{\"command\": \"ls\"}",
			headers:        [][2]string{{"Content-Type", "application/json"}},
			expectedStatus: http.StatusAccepted,
			requireParsing: true,
			parsedBody:     &id,
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

		<-time.After(time.Millisecond * 100)
	}
}

func Test_bashrunHandlers_ListCommands(t *testing.T) {
	ts := httptest.NewServer(testRouter(t))
	defer ts.Close()

	client := http.Client{}

	var commands []domain.CommandFromDB
	tests := []testTableElem{
		{
			caseName:       "non-numeric limit",
			httpMethod:     http.MethodGet,
			route:          "/commands?limit=a",
			body:           "",
			headers:        [][2]string{},
			expectedStatus: http.StatusBadRequest,
			requireParsing: false,
			parsedBody:     nil,
		},
		{
			caseName:       "non-numeric offset",
			httpMethod:     http.MethodGet,
			route:          "/commands?offset=a",
			body:           "",
			headers:        [][2]string{},
			expectedStatus: http.StatusBadRequest,
			requireParsing: false,
			parsedBody:     nil,
		},
		{ //8
			caseName:       "server error",
			httpMethod:     http.MethodGet,
			route:          "/commands",
			body:           "",
			headers:        [][2]string{},
			expectedStatus: http.StatusInternalServerError,
			requireParsing: false,
			parsedBody:     nil,
		},
		{ //9
			caseName:       "rows not found",
			httpMethod:     http.MethodGet,
			route:          "/commands",
			body:           "",
			headers:        [][2]string{},
			expectedStatus: http.StatusNoContent,
			requireParsing: false,
			parsedBody:     nil,
		},
		{ //10
			caseName:       "ok",
			httpMethod:     http.MethodGet,
			route:          "/commands",
			body:           "",
			headers:        [][2]string{},
			expectedStatus: http.StatusOK,
			requireParsing: true,
			parsedBody:     &commands,
		},
	}

	for _, testCase := range tests {
		t.Log(testCase.caseName)

		req, err := buildRequest(testCase.httpMethod, testCase.route, testCase.body, testCase.headers, ts.URL)
		require.NoError(t, err)

		sendReq(t, &client, req, testCase.expectedStatus, testCase.parsedBody, testCase.requireParsing)

		if testCase.caseName == "ok" {
			require.Equal(t, 1, len(commands))
		}
	}
}

func Test_bashrunHandlers_StopCommand(t *testing.T) {
	ts := httptest.NewServer(testRouter(t))
	defer ts.Close()

	client := http.Client{}

	tests := []testTableElem{
		{
			caseName:       "non-numeric id",
			httpMethod:     http.MethodGet,
			route:          "/commands/stop/a",
			body:           "",
			headers:        [][2]string{},
			expectedStatus: http.StatusBadRequest,
			requireParsing: false,
			parsedBody:     nil,
		},
		{ //11
			caseName:       "server error",
			httpMethod:     http.MethodGet,
			route:          "/commands/stop/5",
			body:           "",
			headers:        [][2]string{},
			expectedStatus: http.StatusInternalServerError,
			requireParsing: false,
			parsedBody:     nil,
		},
		{ //12
			caseName:       "rows not found",
			httpMethod:     http.MethodGet,
			route:          "/commands/stop/5",
			body:           "",
			headers:        [][2]string{},
			expectedStatus: http.StatusNotFound,
			requireParsing: false,
			parsedBody:     nil,
		},
		{ //13
			caseName:       "failed to update status",
			httpMethod:     http.MethodGet,
			route:          "/commands/stop/5",
			body:           "",
			headers:        [][2]string{},
			expectedStatus: http.StatusInternalServerError,
			requireParsing: false,
			parsedBody:     nil,
		},
		{ //14
			caseName:       "command status is \"created\"",
			httpMethod:     http.MethodGet,
			route:          "/commands/stop/5",
			body:           "",
			headers:        [][2]string{},
			expectedStatus: http.StatusNoContent,
			requireParsing: false,
			parsedBody:     nil,
		},
		{ //15
			caseName:       "command not started",
			httpMethod:     http.MethodGet,
			route:          "/commands/stop/5",
			body:           "",
			headers:        [][2]string{},
			expectedStatus: http.StatusBadRequest,
			requireParsing: false,
			parsedBody:     nil,
		},
		{ //16
			caseName:       "read PID failed",
			httpMethod:     http.MethodGet,
			route:          "/commands/stop/5",
			body:           "",
			headers:        [][2]string{},
			expectedStatus: http.StatusInternalServerError,
			requireParsing: false,
			parsedBody:     nil,
		},
	}

	for _, testCase := range tests {
		t.Log(testCase.caseName)

		req, err := buildRequest(testCase.httpMethod, testCase.route, testCase.body, testCase.headers, ts.URL)
		require.NoError(t, err)

		sendReq(t, &client, req, testCase.expectedStatus, testCase.parsedBody, testCase.requireParsing)
	}
}

func Test_bashrunHandlers_ReadCommand(t *testing.T) {
	ts := httptest.NewServer(testRouter(t))
	defer ts.Close()

	client := http.Client{}

	var command domain.CommandFromDB
	tests := []testTableElem{
		{
			caseName:       "non-numeric id",
			httpMethod:     http.MethodGet,
			route:          "/commands/a",
			body:           "",
			headers:        [][2]string{},
			expectedStatus: http.StatusBadRequest,
			requireParsing: false,
			parsedBody:     nil,
		},
		{ //17
			caseName:       "server error",
			httpMethod:     http.MethodGet,
			route:          "/commands/1",
			body:           "",
			headers:        [][2]string{},
			expectedStatus: http.StatusInternalServerError,
			requireParsing: false,
			parsedBody:     nil,
		},
		{ //18
			caseName:       "rows not found",
			httpMethod:     http.MethodGet,
			route:          "/commands/1",
			body:           "",
			headers:        [][2]string{},
			expectedStatus: http.StatusNotFound,
			requireParsing: false,
			parsedBody:     nil,
		},
		{ //19
			caseName:       "ok",
			httpMethod:     http.MethodGet,
			route:          "/commands/1",
			body:           "",
			headers:        [][2]string{},
			expectedStatus: http.StatusOK,
			requireParsing: true,
			parsedBody:     &command,
		},
	}

	for _, testCase := range tests {
		t.Log(testCase.caseName)

		req, err := buildRequest(testCase.httpMethod, testCase.route, testCase.body, testCase.headers, ts.URL)
		require.NoError(t, err)

		sendReq(t, &client, req, testCase.expectedStatus, testCase.parsedBody, testCase.requireParsing)

		if testCase.caseName == "ok" {
			require.Equal(t, 1, command.ID)
		}
	}
}

func Test_bashrunHandlers_ReadOutput(t *testing.T) {
	ts := httptest.NewServer(testRouter(t))
	defer ts.Close()

	client := http.Client{}

	tests := []testTableElem{
		{
			caseName:       "non-numeric id",
			httpMethod:     http.MethodGet,
			route:          "/commands/output/a",
			body:           "",
			headers:        [][2]string{},
			expectedStatus: http.StatusBadRequest,
			requireParsing: false,
			parsedBody:     nil,
		},
		{ //20
			caseName:       "server error",
			httpMethod:     http.MethodGet,
			route:          "/commands/output/1",
			body:           "",
			headers:        [][2]string{},
			expectedStatus: http.StatusInternalServerError,
			requireParsing: false,
			parsedBody:     nil,
		},
		{ //21
			caseName:       "rows not found",
			httpMethod:     http.MethodGet,
			route:          "/commands/output/1",
			body:           "",
			headers:        [][2]string{},
			expectedStatus: http.StatusNotFound,
			requireParsing: false,
			parsedBody:     nil,
		},
		{ //22
			caseName:       "empty output",
			httpMethod:     http.MethodGet,
			route:          "/commands/output/1",
			body:           "",
			headers:        [][2]string{},
			expectedStatus: http.StatusNoContent,
			requireParsing: false,
			parsedBody:     nil,
		},
		{ //23
			caseName:       "ok",
			httpMethod:     http.MethodGet,
			route:          "/commands/output/1",
			body:           "",
			headers:        [][2]string{},
			expectedStatus: http.StatusOK,
			requireParsing: false,
			parsedBody:     nil,
		},
	}

	for _, testCase := range tests {
		t.Log(testCase.caseName)

		req, err := buildRequest(testCase.httpMethod, testCase.route, testCase.body, testCase.headers, ts.URL)
		require.NoError(t, err)

		sendReq(t, &client, req, testCase.expectedStatus, testCase.parsedBody, testCase.requireParsing)
	}
}
