package main

import (
	"fmt"
	"github.com/imroc/req/v3"
	"testing"
)

func TestSendRegisterEmailRequest(t *testing.T) {
	testCases := []struct {
		testName string
		email    string
	}{
		{"case1", "your@qq.com"},
		{"case2", "your@163.com"},
	}
	for _, tt := range testCases {
		tt := tt
		t.Run(tt.testName, func(t *testing.T) {
			t.Parallel() // 将每个测试用例标记为能够彼此并行运行
			fmt.Println(tt.testName)
			fmt.Println(req.DevMode().R().SetBody(tt.email).MustPost("api/register/email").TraceInfo())
		})
	}
}

func TestSendDoRegisterRequest(t *testing.T) {
	type testFrom struct {
		username string `json:"username"`
		phone    string `json:"phone"`
		password string `json:"password"`
		email    string `json:"email"`
		code     string `json:"code"`
	}
	testCases := []struct {
		testName string
		testFrom *testFrom
	}{
		{"case1", &testFrom{"张三", "1234564550", "123456", "test1@qq.com", "123456"}},
		{"case2", &testFrom{"李四", "1235355252", "123456", "test1@qq.com", "123456"}},
		{"case3", &testFrom{"王五", "12565355252", "123456", "test1@qq.com", "123456"}},
	}
	for _, tt := range testCases {
		tt := tt
		t.Run(tt.testName, func(t *testing.T) {
			t.Parallel() // 将每个测试用例标记为能够彼此并行运行
			fmt.Println(tt.testName)
			fmt.Println(req.DevMode().R().SetBody(tt.testFrom).MustPost("api/register").TraceInfo())
		})
	}
}

func TestLoginRequest(t *testing.T) {
	type testFrom struct {
		phone     string `json:"phone"`
		password  string `json:"password"`
		currentIp string
	}
	testCases := []struct {
		testName string
		testFrom *testFrom
	}{
		{"case1", &testFrom{"1236814310", "123456", "157.152.22.2"}},
		{"case2", &testFrom{"1236814310", "123456", "157.152.22.2"}},
		{"case3", &testFrom{"1236815310", "123456", "157.152.22.2"}},
	}
	for _, tt := range testCases {
		tt := tt
		t.Run(tt.testName, func(t *testing.T) {
			t.Parallel() // 将每个测试用例标记为能够彼此并行运行
			fmt.Println(tt.testName)
			fmt.Println(req.DevMode().R().SetBody(tt.testFrom).MustPost("api/login").TraceInfo())
		})
	}
}

func TestReSetEmailRequest(t *testing.T) {

	testCases := []struct {
		testName   string
		resetEmail string
	}{
		{"case1", "rest@email"},
		{"case2", "rest@email"},
		{"case3", "rest@email"},
	}
	for _, tt := range testCases {
		tt := tt
		t.Run(tt.testName, func(t *testing.T) {
			t.Parallel() // 将每个测试用例标记为能够彼此并行运行
			fmt.Println(tt.testName)
			fmt.Println(req.DevMode().R().SetBody(tt.resetEmail).MustPost("api/password/reset/email").TraceInfo())
		})
	}
}
