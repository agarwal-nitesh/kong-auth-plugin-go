package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"encoding/json"

	"github.com/Kong/go-pdk"
	"github.com/gojektech/heimdall"
	"github.com/gojektech/heimdall/hystrix"
)

const TIMEOUT = 5 * time.Second
const RETRY_COUNT = 3
const BACKOFF_INTERVAL = 2 * time.Second
const JSON_PARSE_ERR = "500"
const JSON_PARSE_ERR_STATUS_CODE = 500
const AUTH_ERROR = "401"
const AUTH_ERROR_CODE = 401

// Define a maximum jitter interval. It must be more than 1*time.Millisecond
const MAX_JITTER_INTERVAL = 5 * time.Millisecond
const AUTH_URL = "https://auth.zkrull.com/auth/user/authorize"
const X_AUTH_TOKEN = "fafafafafafa"

var RoleNotAllowedResponse = `{"message": "Role not allowed", "status": "FAILED"}`
var AuthHeaderAbsentResponse = `{"message": "No Auth header with bearer token", "status": "FAILED"}`
var JsonParseErrorResponse = `{"message": "Error Parsing User details Json (kong)", "status": "FAILED"}`
var AuthFailedResponse = `{"message": "Authorization Failed!", "status": "FAILED"}`
var UnExpectedErrorResponse = `{"message": "Unexpected Failure!", "status": "FAILED"}`

type Config struct {
	RequestHeader  string `json:"request_header"`
	EnforceRole1   string `json:"enforce_roles_1"`
	EnforceRole2   string `json:"enforce_roles_2"`
	EnforceRole3   string `json:"enforce_roles_3"`
	ResponseHeader string `json:"response_header"`
}

func New() interface{} {
	return &Config{RequestHeader: "authorization",
		ResponseHeader: "authorization",
		EnforceRole1:   "USER",
	}
}

func (conf Config) Access(kong *pdk.PDK) {
	auth, err := kong.Request.GetHeader(conf.RequestHeader)
	responseHeaders := make(map[string][]string)
	responseHeaders["Content-Type"] = append(responseHeaders["Content-Type"], "application/json")
	if err != nil {
		kong.Log.Err("No auth header found ! ", err.Error())
		kong.Response.SetStatus(AUTH_ERROR_CODE)
		kong.Response.Exit(JSON_PARSE_ERR_STATUS_CODE, AuthHeaderAbsentResponse, responseHeaders)
		return
	}
	user, err := conf.getAuth(auth, kong)
	if err != nil {
		if err.Error() == JSON_PARSE_ERR {
			kong.Log.Err("Json Parse Err : ", err)
			kong.Response.SetStatus(AUTH_ERROR_CODE)
			kong.Response.Exit(JSON_PARSE_ERR_STATUS_CODE, JsonParseErrorResponse, responseHeaders)
			return
		} else if err.Error() == AUTH_ERROR {
			kong.Log.Err("Auth Err : ", err)
			kong.Response.SetStatus(AUTH_ERROR_CODE)
			kong.Response.Exit(AUTH_ERROR_CODE, AuthFailedResponse, responseHeaders)
			return
		}
		kong.Log.Err("UnExpected Err: ", err)
		kong.Response.SetStatus(AUTH_ERROR_CODE)
		kong.Response.Exit(AUTH_ERROR_CODE, UnExpectedErrorResponse, responseHeaders)
		return
	}
	kong.Log.Err("r1, r2, r3", conf.EnforceRole1, conf.EnforceRole2, conf.EnforceRole3)
	if user.Role == "" {
		user.Role = "USER"
	}
	if conf.EnforceRole1 != user.Role &&
		conf.EnforceRole2 != user.Role &&
		conf.EnforceRole3 != user.Role {
		kong.Log.Err("Role not allowed ")
		kong.Response.SetStatus(AUTH_ERROR_CODE)
		kong.Response.Exit(AUTH_ERROR_CODE, RoleNotAllowedResponse, responseHeaders)
		return
	}
	kong.Response.SetHeader("authorization", auth)
	kong.Response.SetHeader("user_id", fmt.Sprintf("%v", user.Id))
	kong.Response.SetHeader("user_name", user.Name)
	kong.Response.SetHeader("user_email", user.Email)
	kong.Response.SetHeader("image_url", user.ImageUrl)
	kong.Log.Err("userId, userName, userEmail, userRole ", user.Id,
		user.Name, user.Email, user.Role)
}

type User struct {
	Id           int64  `json:"id"`
	Name         string `json:"name"`
	Email        string `json:"email"`
	AuthProvider string `json:"provider"`
	Role         string `json:"role"`
	ProviderId   string `json:"providerId"`
	ImageUrl     string `json:"imageUrl"`
}

func (conf Config) getAuth(reqToken string, kong *pdk.PDK) (*User, error) {
	fallbackFn := func(err error) error {
		kong.Log.Err("Inside fallback hsytrix activated", err)
		return errors.New(fmt.Sprintf("%d", http.StatusTooManyRequests))
	}
	splitToken := strings.Split(reqToken, "Bearer")
	var authToken string
	if len(splitToken) > 1 {
		authToken = splitToken[1]
	} else {
		authToken = ""
	}
	backoff := heimdall.NewConstantBackoff(BACKOFF_INTERVAL, MAX_JITTER_INTERVAL)
	retrier := heimdall.NewRetrier(backoff)
	client := hystrix.NewClient(
		hystrix.WithHTTPTimeout(TIMEOUT),
		hystrix.WithCommandName("AuthUser"),
		hystrix.WithHystrixTimeout(1100),
		hystrix.WithMaxConcurrentRequests(100),
		hystrix.WithErrorPercentThreshold(20),
		hystrix.WithSleepWindow(10),
		hystrix.WithRequestVolumeThreshold(10),
		hystrix.WithFallbackFunc(fallbackFn),
		hystrix.WithRetryCount(RETRY_COUNT),
		hystrix.WithRetrier(retrier),
	)
	headers := http.Header{}
	headers.Set("Content-Type", "application/json")
	headers.Set("X_AUTH_TOKEN", X_AUTH_TOKEN)
	headers.Set("X_JWT_TOKEN", authToken)
	res, err := client.Get(AUTH_URL, headers)
	if err != nil {
		return nil, err
	}
	kong.Log.Info("Response Status Code = ", res.StatusCode)
	if res.StatusCode == 200 {
		decoder := json.NewDecoder(res.Body)
		var data User
		err = decoder.Decode(&data)
		if err == nil {
			kong.Log.Info("User Data : ", data)
			return &data, err
		} else {
			return nil, errors.New(JSON_PARSE_ERR)
		}
	} else {
		return nil, errors.New(fmt.Sprintf("%d", res.StatusCode))
	}
}
