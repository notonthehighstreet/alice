package newrelic

import "github.com/stretchr/testify/mock"

type MockNewRelicClient struct {
	mock.Mock
}

func (c *MockNewRelicClient) Get(url string, apiKey string) (NewRelicResponse, error) {
	args := c.Mock.Called(url, apiKey)
	return args.Get(0).(NewRelicResponse), args.Error(1)
}
