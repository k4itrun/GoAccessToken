//Use this command to build your code run <go build -o myprogram.exe>
// Code by k4itrun :=

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

var client = &http.Client{}

func makeRequest(method, url string, body []byte) (*http.Response, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return client.Do(req)
}

func login(mail, password string) (string, string, string) {
	reqBody := map[string]string{
		"email":    mail,
		"password": password,
	}
	reqJSON, _ := json.Marshal(reqBody)
	res, err := makeRequest("POST", "https://discord.com/api/v9/auth/login", reqJSON)
	if err != nil {
		fmt.Printf("Error making login: %s\n", err)
		return "", "", ""
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("Error reading login: %s\n", err)
		return "", "", ""
	}

	var response map[string]interface{}

	err = json.Unmarshal(body, &response)

	if err != nil {
		fmt.Printf("Error unmarshalling login: %s\n", err)
		return "", "", ""
	}
	if token, ok := response["token"].(string); ok {
		return token, "", ""
	} else if captchaKey, ok := response["captcha_key"].(string); ok {
		return "", captchaKey, ""
	} else if errors, ok := response["errors"].(string); ok {
		return "", "", errors
	} else if ticket, ok := response["ticket"].(string); ok {
		token := TwoFA(ticket)
		return token, "", ""
	}

	return "", "", ""
}

func TwoFA(ticket string) string {
	for {
		fmt.Print("> 2FA Code: ")
		var code string
		fmt.Scanln(&code)

		reqBody := map[string]string{
			"code":   code,
			"ticket": ticket,
		}
		reqJSON, _ := json.Marshal(reqBody)

		res, err := makeRequest("POST", "https://discord.com/api/v9/auth/mfa/totp", reqJSON)
		if err != nil {
			fmt.Printf("Error making 2FA: %s\n", err)
			return ""
		}
		defer res.Body.Close()

		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			fmt.Printf("Error reading 2FA: %s\n", err)
			return ""
		}

		var response map[string]interface{}
		err = json.Unmarshal(body, &response)
		if err != nil {
			fmt.Printf("Error unmarshalling 2FA: %s\n", err)
			return ""
		}

		if token, ok := response["token"].(string); ok {
			return token
		}

		if message, ok := response["message"].(string); ok {
			fmt.Println("An Invalid Code Was Provided!")
			fmt.Println(message)
		}
	}
}

func getUserInfo(token string) (string, map[string]interface{}) {
	url := "https://discord.com/api/v9/users/@me"
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		fmt.Printf("Error creating GET: %s\n", err)
		return "", nil
	}

	req.Header.Set("Authorization", token)

	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making GET: %s\n", err)
		return "", nil
	}

	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("Error reading GET: %s\n", err)
		return "", nil
	}

	var userInfo map[string]interface{}

	err = json.Unmarshal(body, &userInfo)

	if err != nil {
		fmt.Printf("Error unmarshalling GET: %s\n", err)
		return "", nil
	}

	userId, ok := userInfo["id"].(string)
	if !ok {
		fmt.Println("Error: User ID not found")
		return "", nil
	}

	return userId, userInfo
}

func saveUserInfoToFile(userId string, userInfo map[string]interface{}, filename string) error {
	userInfos := make(map[string]map[string]interface{})
	if _, err := os.Stat(filename); err == nil {
		fileData, err := ioutil.ReadFile(filename)
		if err != nil {
			return err
		}
		err = json.Unmarshal(fileData, &userInfos)
		if err != nil {
			return err
		}
	}
	userInfos[userId] = userInfo
	fileData, err := json.MarshalIndent(userInfos, "", "  ")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filename, fileData, 0644)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	fmt.Print("> Mail: ")
	var mail string
	fmt.Scanln(&mail)

	fmt.Print("> Password: ")
	var password string
	fmt.Scanln(&password)

	token, captchaKey, errors := login(mail, password)

	if token != "" {
		fmt.Printf("Your Token Is %s\n", token)

		userId, userInfo := getUserInfo(token)
		if userInfo != nil {
			//fmt.Printf("User Info: %v\n", userInfo)
			err := saveUserInfoToFile(userId, userInfo, "view.json")
			if err != nil {
				fmt.Printf("Error saving user info to file: %s\n", err)
			} else {
				fmt.Println("User info saved")
			}
		}
	} else if captchaKey != "" {
		fmt.Println("Cannot Get Token, There Is A Captcha To Do!")
		fmt.Printf("Captcha Key: %s\n", captchaKey)
	} else if errors != "" {
		fmt.Println("Email Or Password Is Invalid!")
		fmt.Printf("Errors: %s\n", errors)
	}

	var m string
	fmt.Scanln(&m)
}
