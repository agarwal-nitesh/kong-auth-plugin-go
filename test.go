package main

// import (
// 	"encoding/json"
// 	"errors"
// 	"fmt"
// 	"net/http"
// 	"strings"
// 	"time"

// 	"github.com/gojektech/heimdall"
// 	"github.com/gojektech/heimdall/hystrix"
// )

// const TIMEOUT = 10 * time.Second
// const RETRY_COUNT = 3
// const BACKOFF_INTERVAL = 2 * time.Second
// const JSON_PARSE_ERR = "1331"
// const JSON_PARSE_ERR_STATUS_CODE = 1331
// const AUTH_ERROR = "401"
// const AUTH_ERROR_CODE = 401

// // Define a maximum jitter interval. It must be more than 1*time.Millisecond
// const MAX_JITTER_INTERVAL = 5 * time.Millisecond
// const AUTH_URL = "https://auth.zkrull.com/auth/user/authorize"
// const X_AUTH_TOKEN = "1234445556"

// func main() {
// 	getAuth("Bearer esILKLDLAADkdow.fgJzdWIiJILKSWJDDFF.fsywlkfolqormc1k_Q")
// }

// type User struct {
// 	Id           int64  `json:"id"`
// 	Name         string `json:"name"`
// 	Email        string `json:"email"`
// 	AuthProvider string `json:"provider"`
// 	Role         string `json:"role"`
// 	ProviderId   string `json:"providerId"`
// 	ImageUrl     string `json:"imageUrl"`
// }

// func getAuth(reqToken string) (*User, error) {
// 	fmt.Println("Inside getAuth")
// 	fallbackFn := func(err error) error {
// 		fmt.Println("Inside fallback hsytrix activated ", err)
// 		return errors.New(fmt.Sprintf("%d", http.StatusTooManyRequests))
// 	}
// 	splitToken := strings.Split(reqToken, "Bearer")
// 	var authToken string
// 	if len(splitToken) > 1 {
// 		authToken = splitToken[1]
// 	} else {
// 		authToken = ""
// 	}

// 	backoff := heimdall.NewConstantBackoff(BACKOFF_INTERVAL, MAX_JITTER_INTERVAL)
// 	retrier := heimdall.NewRetrier(backoff)
// 	client := hystrix.NewClient(
// 		hystrix.WithHTTPTimeout(TIMEOUT),
// 		hystrix.WithCommandName("AuthUser"),
// 		hystrix.WithHystrixTimeout(1100),
// 		hystrix.WithMaxConcurrentRequests(100),
// 		hystrix.WithErrorPercentThreshold(20),
// 		hystrix.WithSleepWindow(10),
// 		hystrix.WithRequestVolumeThreshold(10),
// 		hystrix.WithFallbackFunc(fallbackFn),
// 		hystrix.WithRetryCount(RETRY_COUNT),
// 		hystrix.WithRetrier(retrier),
// 	)
// 	headers := http.Header{}
// 	headers.Set("Content-Type", "application/json")
// 	headers.Set("X_AUTH_TOKEN", X_AUTH_TOKEN)
// 	headers.Set("X_JWT_TOKEN", authToken)
// 	fmt.Println("Calling Auth Service with authToken = ", authToken)
// 	res, err := client.Get(AUTH_URL, headers)
// 	if err != nil {
// 		fmt.Println("Error getting the response", err)
// 		return nil, err
// 	}
// 	if err != nil {
// 		fmt.Println("Error getting the response")
// 		return nil, err
// 	}
// 	if res.StatusCode == 200 {
// 		fmt.Println("Got response")
// 		decoder := json.NewDecoder(res.Body)
// 		var data User
// 		err = decoder.Decode(&data)
// 		if err == nil {
// 			fmt.Println("User Data : ", data.Role == "")
// 			if "ADMIN" != data.Role &&
// 				"" != data.Role &&
// 				"" != data.Role {
// 				fmt.Println("Role not allowed ")
// 			}
// 			return &data, err
// 		} else {
// 			fmt.Println(err)
// 			return nil, errors.New(JSON_PARSE_ERR)
// 		}
// 	} else {
// 		return nil, errors.New(fmt.Sprintf("%d", res.StatusCode))
// 	}
// }
