package main_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestPersonStorageIntegration(testCase *testing.T) {
	projectRoot, err := os.Getwd()
	if err != nil {
		testCase.Fatalf("failed to get project root: %v", err)
	}

	tempDir := testCase.TempDir()
	binaryPath := filepath.Join(tempDir, "personstorage")
	buildBinary(testCase, projectRoot, binaryPath)

	listenAddr := reserveListenAddr(testCase)
	databasePath := filepath.Join(tempDir, "integration.db")
	baseURL := "http://" + listenAddr

	serverProcess, waitForExit, output := startServerProcess(testCase, projectRoot, binaryPath, listenAddr, databasePath)
	testCase.Cleanup(func() {
		if serverProcess.Process != nil {
			_ = serverProcess.Process.Kill()
		}
		select {
		case <-time.After(3 * time.Second):
		case <-waitForExit:
		}
	})

	waitForServer(testCase, baseURL, waitForExit, output)

	personPayload := map[string]string{
		"external_id":   "integration-person-1",
		"name":          "Grace Hopper",
		"email":         "grace.hopper@example.com",
		"date_of_birth": "1906-12-09",
	}

	saveResponse := doRequest(
		testCase,
		http.MethodPost,
		baseURL+"/people",
		personPayload,
	)
	assertStatusCode(testCase, saveResponse, http.StatusCreated)
	assertJSONBody(testCase, saveResponse, map[string]string{
		"message": "Successfully saved",
	})

	loadResponse := doRequest(
		testCase,
		http.MethodGet,
		baseURL+"/people/integration-person-1",
		nil,
	)
	assertStatusCode(testCase, loadResponse, http.StatusOK)
	assertJSONBody(testCase, loadResponse, personPayload)
}

func buildBinary(testCase *testing.T, projectRoot string, binaryPath string) {
	testCase.Helper()

	buildCommand := exec.Command("go", "build", "-o", binaryPath, ".")
	buildCommand.Dir = projectRoot
	buildCommand.Env = os.Environ()
	output, err := buildCommand.CombinedOutput()
	if err != nil {
		testCase.Fatalf("failed to build integration test binary: %v\n%s", err, output)
	}
}

func reserveListenAddr(testCase *testing.T) string {
	testCase.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		testCase.Fatalf("failed to reserve listen address: %v", err)
	}
	defer listener.Close()

	return listener.Addr().String()
}

func startServerProcess(testCase *testing.T, projectRoot string, binaryPath string, listenAddr string, databasePath string) (*exec.Cmd, <-chan error, *bytes.Buffer) {
	testCase.Helper()

	serverContext, cancel := context.WithCancel(context.Background())
	testCase.Cleanup(cancel)

	serverCommand := exec.CommandContext(serverContext, binaryPath)
	serverCommand.Dir = projectRoot
	serverCommand.Env = append(
		os.Environ(),
		"LISTEN_ADDR="+listenAddr,
		"DATABASE_PATH="+databasePath,
	)

	var output bytes.Buffer
	serverCommand.Stdout = &output
	serverCommand.Stderr = &output

	if err := serverCommand.Start(); err != nil {
		testCase.Fatalf("failed to start server process: %v", err)
	}

	waitResult := make(chan error, 1)
	go func() {
		waitResult <- serverCommand.Wait()
	}()

	return serverCommand, waitResult, &output
}

func waitForServer(testCase *testing.T, baseURL string, waitForExit <-chan error, output *bytes.Buffer) {
	testCase.Helper()

	client := &http.Client{Timeout: 200 * time.Millisecond}
	deadline := time.Now().Add(5 * time.Second)

	for time.Now().Before(deadline) {
		response, err := client.Get(baseURL + "/people/missing-id")
		if err == nil {
			closeBody(testCase, response.Body)
			if response.StatusCode == http.StatusNotFound {
				return
			}
		}

		select {
		case exitErr := <-waitForExit:
			testCase.Fatalf("server process exited before becoming ready: %v\n%s", exitErr, output.String())
		default:
		}

		time.Sleep(100 * time.Millisecond)
	}

	testCase.Fatalf("server did not become ready:\n%s", output.String())/
}

func doRequest(testCase *testing.T, method string, url string, payload any) *http.Response {
	testCase.Helper()

	var requestBody io.Reader
	if payload != nil {
		encodedPayload, err := json.Marshal(payload)
		if err != nil {
			testCase.Fatalf("failed to encode request payload: %v", err)
		}
		requestBody = bytes.NewReader(encodedPayload)
	}

	request, err := http.NewRequest(method, url, requestBody)
	if err != nil {
		testCase.Fatalf("failed to create request: %v", err)
	}
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	response, err := (&http.Client{Timeout: 2 * time.Second}).Do(request)
	if err != nil {
		testCase.Fatalf("request failed: %v", err)
	}

	testCase.Cleanup(func() {
		closeBody(testCase, response.Body)
	})

	return response
}

func assertStatusCode(testCase *testing.T, response *http.Response, want int) {
	testCase.Helper()

	if response.StatusCode != want {
		body, _ := io.ReadAll(response.Body)
		testCase.Fatalf("expected status %d, got %d: %s", want, response.StatusCode, string(body))
	}
}

func assertJSONBody(testCase *testing.T, response *http.Response, want map[string]string) {
	testCase.Helper()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		testCase.Fatalf("failed to read response body: %v", err)
	}

	var got map[string]string
	if err := json.Unmarshal(body, &got); err != nil {
		testCase.Fatalf("failed to decode response body %q: %v", string(body), err)
	}

	if len(got) != len(want) {
		testCase.Fatalf("expected JSON body %v, got %v", want, got)
	}

	for key, wantValue := range want {
		if got[key] != wantValue {
			testCase.Fatalf("expected JSON body %v, got %v", want, got)
		}
	}
}

func closeBody(testCase *testing.T, body io.ReadCloser) {
	testCase.Helper()

	if err := body.Close(); err != nil {
		testCase.Fatalf("failed to close response body: %v", err)
	}
}
