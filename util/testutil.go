package util

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http/httptest"

	"github.com/gofiber/fiber/v2"
)

func GetStringSliceRequestResponse(app *fiber.App, method string, url string, reqBody any) (code int, respBody []string, err error) {
	bodyJson := []byte("")
	if reqBody != nil {
		bodyJson, _ = json.Marshal(reqBody)
	}
	req := httptest.NewRequest(method, url, bytes.NewReader(bodyJson))
	req.Header.Set("Content-Type", fiber.MIMEApplicationJSON)

	resp, err := app.Test(req, 10)
	if resp != nil {
		code = resp.StatusCode
	}
	// If error we're done
	if err != nil {
		return
	}
	// If no body content, we're done
	if resp.ContentLength == 0 {
		return
	}
	bodyData := make([]byte, resp.ContentLength)
	n, err := resp.Body.Read(bodyData)
	if n == 0 {
		return
	}
	err = json.Unmarshal(bodyData, &respBody)
	if err != nil {
		log.Printf("Error parsing json: %v for '%s'\n", err, string(bodyData))
	}
	return
}

func GetStringRequestResponse(app *fiber.App, method string, url string, reqBody string) (code int, respBody string, err error) {
	req := httptest.NewRequest(method, url, bytes.NewReader([]byte(reqBody)))
	req.Header.Set("Content-Type", fiber.MIMETextPlain)

	resp, err := app.Test(req, 10)
	// If error we're done
	if resp != nil {
		code = resp.StatusCode
	}
	if err != nil {
		return
	}
	// If no body content, we're done
	if resp.ContentLength == 0 {
		return
	}
	bodyData := make([]byte, resp.ContentLength)
	_, _ = resp.Body.Read(bodyData)
	respBody = string(bodyData)
	return
}

func GetJsonSliceRequestResponse(app *fiber.App, method string, url string, reqBody any) (code int, respBody []map[string]any, err error) {
	bodyJson := []byte("")
	if reqBody != nil {
		bodyJson, _ = json.Marshal(reqBody)
	}
	req := httptest.NewRequest(method, url, bytes.NewReader(bodyJson))
	req.Header.Set("Content-Type", fiber.MIMEApplicationJSON)
	resp, err := app.Test(req, 10)
	// If error we're done
	if resp != nil {
		code = resp.StatusCode
	}
	if err != nil {
		return
	}
	// If no body content, we're done
	if resp.ContentLength == 0 {
		return
	}
	bodyData := make([]byte, resp.ContentLength)
	n, err := resp.Body.Read(bodyData)
	if n == 0 {
		return
	}
	err = json.Unmarshal(bodyData, &respBody)
	if err != nil {
		log.Printf("Error parsing json: %v for '%s'\n", err, string(bodyData))
	}
	return
}

func GetJsonRequestResponse(app *fiber.App, method string, url string, reqBody any) (code int, respBody map[string]any, err error) {
	bodyJson := []byte("")
	if reqBody != nil {
		bodyJson, _ = json.Marshal(reqBody)
	}
	req := httptest.NewRequest(method, url, bytes.NewReader(bodyJson))
	req.Header.Set("Content-Type", fiber.MIMEApplicationJSON)
	resp, err := app.Test(req, 10)
	if resp != nil {
		code = resp.StatusCode
	}
	// If error we're done
	if err != nil {
		return
	}
	// If no body content, we're done
	if resp.ContentLength == 0 {
		return
	}
	bodyData := make([]byte, resp.ContentLength)
	_, _ = resp.Body.Read(bodyData)
	err = json.Unmarshal(bodyData, &respBody)
	if err != nil {
		log.Printf("Error parsing json: %v for '%s'\n", err, string(bodyData))
	}
	return
}
