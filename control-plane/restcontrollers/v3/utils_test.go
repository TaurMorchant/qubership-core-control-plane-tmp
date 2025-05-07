package v3

import (
	"github.com/stretchr/testify/assert"
	"io"
	"net/http"
	"testing"
)

type serviceMock struct {
	invocations    map[string][]invocation
	responsesQueue map[string][]responseMock
}

type responseMock struct {
	specs any
	err   error
}

func newServiceMock(functionNames ...string) *serviceMock {
	invocations := make(map[string][]invocation)
	responsesQueue := make(map[string][]responseMock)
	for _, functionName := range functionNames {
		invocations[functionName] = make([]invocation, 0)
		responsesQueue[functionName] = make([]responseMock, 0)
	}
	return &serviceMock{invocations: invocations, responsesQueue: responsesQueue}
}

type invocation struct {
	FuncName string
	Args     []any
}

func (m *serviceMock) respond(funcName string) responseMock {
	responsesQueue := m.responsesQueue[funcName]
	if len(responsesQueue) == 0 {
		return responseMock{}
	}
	var response responseMock
	response, m.responsesQueue[funcName] = responsesQueue[0], responsesQueue[1:]
	return response
}

func (m *serviceMock) invoke(funcName string, args ...any) responseMock {
	invocation := invocation{FuncName: funcName, Args: args}
	m.invocations[funcName] = append(m.invocations[funcName], invocation)
	return m.respond(funcName)
}

func (m *serviceMock) EnqueueResponse(funcName string, response responseMock) {
	m.responsesQueue[funcName] = append(m.responsesQueue[funcName], response)
}

func (m *serviceMock) GetInvocation(funcName string) *invocation {
	invocations := m.invocations[funcName]
	var invoke invocation
	if len(invocations) == 0 {
		return nil
	} else if len(invocations) == 1 {
		invoke, m.invocations[funcName] = invocations[0], make([]invocation, 0, 4)
	} else {
		invoke, m.invocations[funcName] = invocations[0], invocations[1:]
	}
	return &invoke
}

func verifyResponseBody(t *testing.T, expectedBody []byte, response *http.Response) {
	actualBody, err := io.ReadAll(response.Body)
	assert.Nil(t, err)
	assert.Equal(t, expectedBody, actualBody)
}
