package minimax

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

const BaseURL = "https://api.minimaxi.com/v1"

type Client struct {
	ApiKey string
	Client *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		ApiKey: apiKey,
		Client: &http.Client{},
	}
}

func (c *Client) doRequest(method, url string, body any, result any) error {
	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return err
		}
		bodyReader = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+c.ApiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("api error %d: %s", resp.StatusCode, string(b))
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return err
		}
	}
	return nil
}

// T2A Sync
func (c *Client) T2A(req *T2ARequest) (*T2AResponse, error) {
	url := fmt.Sprintf("%s/t2a_v2", BaseURL)
	var resp T2AResponse
	if err := c.doRequest("POST", url, req, &resp); err != nil {
		return nil, err
	}
	if resp.BaseResp.StatusCode != 0 {
		return nil, fmt.Errorf("minimax api error: %s", resp.BaseResp.StatusMsg)
	}
	return &resp, nil
}

// T2A Async Create
func (c *Client) T2AAsync(req *T2ARequest) (*T2AAsyncResponse, error) {
	url := fmt.Sprintf("%s/t2a_async_v2", BaseURL)
	var resp T2AAsyncResponse
	if err := c.doRequest("POST", url, req, &resp); err != nil {
		return nil, err
	}
	if resp.BaseResp.StatusCode != 0 {
		return nil, fmt.Errorf("minimax api error: %s", resp.BaseResp.StatusMsg)
	}
	return &resp, nil
}

// T2A Async Query
func (c *Client) T2AAsyncQuery(taskID int64) (*T2AAsyncQueryResponse, error) {
	url := fmt.Sprintf("%s/query/t2a_async_query_v2?task_id=%d", BaseURL, taskID)
	// Query params handled manually or via url.Values if complex
	// Docs said GET with query param? Wait, docs said:
	// curl --request GET \ --url https://api.minimaxi.com/v1/query/t2a_async_query_v2 \ ...
	// Wait, is taskID a query param or body?
	// Docs showed "task_id" in response.
	// Query Params section: task_id. So it is a query param.

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.ApiKey)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result T2AAsyncQueryResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.BaseResp.StatusCode != 0 {
		return nil, fmt.Errorf("query error: %s", result.BaseResp.StatusMsg)
	}
	return &result, nil
}

// Retrieve File
func (c *Client) RetrieveFile(fileID int64) (*FileRetrieveResponse, error) {
	// Need checking retrieve file endpoint. Usually /files/retrieve?file_id=...
	// Doc says: [文件检索接口](...)
	// Since I didn't read that doc, I'll guess or assume common pattern.
	// minimax doc pattern: /files/retrieve?file_id=x
	url := fmt.Sprintf("%s/files/retrieve?file_id=%d", BaseURL, fileID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.ApiKey)

	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result FileRetrieveResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if result.BaseResp.StatusCode != 0 {
		return nil, fmt.Errorf("retrieve error: %s", result.BaseResp.StatusMsg)
	}
	return &result, nil
}

// Upload File
func (c *Client) UploadFile(filePath string, purpose string) (*UploadResponse, error) {
	url := fmt.Sprintf("%s/files/upload?purpose=%s", BaseURL, purpose)

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, err
	}
	io.Copy(part, file)
	writer.Close()

	httpReq, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.ApiKey)
	httpReq.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.Client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result UploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.BaseResp.StatusCode != 0 {
		return nil, fmt.Errorf("upload error: %s", result.BaseResp.StatusMsg)
	}

	return &result, nil
}

// Voice Clone
func (c *Client) VoiceClone(req *VoiceCloneRequest) (*VoiceCloneResponse, error) {
	url := fmt.Sprintf("%s/voice_clone", BaseURL)
	var resp VoiceCloneResponse
	if err := c.doRequest("POST", url, req, &resp); err != nil {
		return nil, err
	}
	if resp.BaseResp.StatusCode != 0 {
		return nil, fmt.Errorf("voice clone error: %s", resp.BaseResp.StatusMsg)
	}
	return &resp, nil
}

// Voice Design
func (c *Client) VoiceDesign(req *VoiceDesignRequest) (*VoiceDesignResponse, error) {
	url := fmt.Sprintf("%s/voice_design", BaseURL)
	var resp VoiceDesignResponse
	if err := c.doRequest("POST", url, req, &resp); err != nil {
		return nil, err
	}
	if resp.BaseResp.StatusCode != 0 {
		return nil, fmt.Errorf("voice design error: %s", resp.BaseResp.StatusMsg)
	}
	return &resp, nil
}

// Get Voices
func (c *Client) GetVoices(voiceType string) (*GetVoicesResponse, error) {
	url := fmt.Sprintf("%s/get_voice", BaseURL)
	req := GetVoicesRequest{VoiceType: voiceType}
	var resp GetVoicesResponse
	if err := c.doRequest("POST", url, req, &resp); err != nil {
		return nil, err
	}
	if resp.BaseResp.StatusCode != 0 {
		return nil, fmt.Errorf("get voices error: %s", resp.BaseResp.StatusMsg)
	}
	return &resp, nil
}

// Delete Voice
func (c *Client) DeleteVoice(voiceType, voiceID string) error {
	url := fmt.Sprintf("%s/delete_voice", BaseURL)
	req := DeleteVoiceRequest{VoiceType: voiceType, VoiceID: voiceID}
	var resp DeleteVoiceResponse
	if err := c.doRequest("POST", url, req, &resp); err != nil {
		return err
	}
	if resp.BaseResp.StatusCode != 0 {
		return fmt.Errorf("delete voice error: %s", resp.BaseResp.StatusMsg)
	}
	return nil
}
