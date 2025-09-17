package ecloud

import (
	"bytes"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"regexp"
	"strings"

	"github.com/Elemento-Modular-Cloud/tesi-paolobeci/ecloud/schema"
)

//go:embed version.txt
var version string

// ------------------------------ API CALLS FUNCTIONS -------------------------

// Login to the API
func (c *Client) Login(reqBody *schema.LoginRequest) (*schema.LoginResponse, error) {
	var res schema.LoginResponse
	err := c.CallAPI("POST", "47777", "/api/v1/authenticate/login", reqBody, &res, false)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Status login
func (c *Client) StatusLogin() (*schema.StatusLoginResponse, error) {
	var res schema.StatusLoginResponse
	err := c.CallAPI("GET", "47777", "/api/v1/authenticate/status", nil, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Logout from the API
func (c *Client) Logout() (*schema.LogoutResponse, error) {
	var res schema.LogoutResponse
	err := c.CallAPI("POST", "47777", "/api/v1/authenticate/logout", nil, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Health check Compute
func (c *Client) HealthCheckCompute() (*schema.HealthCheckComputeResponse, error) {
	var res schema.HealthCheckComputeResponse
	err := c.CallAPI("GET", "17777", "/", nil, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Can allocate a new compute instance
func (c *Client) CanAllocateCompute(reqBody schema.CanAllocateComputeRequest) (*schema.CanAllocateComputeResponse, error) {
	// Original API call code
	// var res schema.CanAllocateComputeResponse
	// err := c.CallAPI("POST", "17777", "/api/v1.0/client/vm/canallocate", reqBody, &res, true)
	// if err != nil {
	// 	return nil, err
	// }
	// return &res, nil

	// Mock response that always returns true since the API is not implemented yet
	res := &schema.CanAllocateComputeResponse{
		Mesos: []schema.ProviderInfo{
			{
				Price: schema.Price{
					Hour:  0.0,
					Month: 0.0,
					Unit:  "EUR",
				},
				Provider: "mock-provider",
				Region:   "eu-south",
			},
		},
	}
	return res, nil
}

// Create a new compute instance
func (c *Client) CreateCompute(reqBody schema.CreateComputeRequest) (*schema.CreateComputeResponse, error) {
	var res schema.CreateComputeResponse
	err := c.CallAPI("POST", "17777", "/api/v1.0/client/vm/register", reqBody, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Compute instances status
func (c *Client) GetCompute() (*schema.GetComputeResponse, error) {
	var res schema.GetComputeResponse
	err := c.CallAPI("GET", "17777", "/api/v1.0/client/vm/status", nil, &res, true)
	if err != nil {
		return nil, err
	}

	// Convert RAM size from GB to MB for each server
	for i := range res {
		if res[i].ReqJSON.RamSize > 0 {
			res[i].ReqJSON.RamSize = res[i].ReqJSON.RamSize * 1024
		}
	}

	return &res, nil
}

// Compute templates
func (c *Client) ComputeTemplates() (*schema.ComputeTemplatesResponse, error) {
	var res schema.ComputeTemplatesResponse
	err := c.CallAPI("GET", "17777", "/api/v1.0/client/vm/templates", nil, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Compute instance delete
func (c *Client) DeleteCompute(reqBody schema.DeleteComputeRequest) (*schema.DeleteComputeResponse, error) {
	var res schema.DeleteComputeResponse
	err := c.CallAPI("POST", "17777", "/api/v1.0/client/vm/delete", reqBody, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Health check Storage
func (c *Client) HealthCheckStorage() (*schema.HealthCheckStorageResponse, error) {
	var res schema.HealthCheckStorageResponse
	err := c.CallAPI("GET", "27777", "/", nil, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Can create a new storage volume
func (c *Client) CanCreateStorage(reqBody schema.CanCreateStorageRequest) (*schema.CanCreateStorageResponse, error) {
	var res schema.CanCreateStorageResponse
	err := c.CallAPI("POST", "27777", "/api/v1.0/client/volume/cancreate", reqBody, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Create a new storage volume
func (c *Client) CreateStorage(reqBody schema.CreateStorageRequest) (*schema.CreateStorageResponse, error) {
	var res schema.CreateStorageResponse
	err := c.CallAPI("POST", "27777", "/api/v1.0/client/volume/create", reqBody, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Create new storage volume with specified image
func (c *Client) CreateStorageImage(reqBody schema.CreateStorageImageRequest) (*schema.CreateStorageImageResponse, error) {
	var res schema.CreateStorageImageResponse
	err := c.CallAPI("POST", "27777", "/api/v1.0/client/volume/cloudinit/create", reqBody, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Create new cloudinit volume
func (c *Client) CreateStorageCloudInit(reqBody schema.CreateStorageCloudInitRequest, userData string) (*schema.CreateStorageCloudInitResponse, error) {
	var res schema.CreateStorageCloudInitResponse

	fmt.Printf("Marshalling request body: %+v\n", reqBody)
	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal reqBody: %w", err)
	}
	encodedPayload := base64.StdEncoding.EncodeToString(jsonBytes)
	fmt.Printf("Encoded payload (base64): %s\n", encodedPayload)

	// -------------- CLOUD-INIT USER-DATA MODIFICATION --------------
	// Use the embedded user-data template from cloudconfig package
	modifiedContent := CloudinitTemplate
	if reqBody.Name != "" {
		hostname := strings.TrimSuffix(reqBody.Name, "-cloudinit")
		modifiedContent = strings.Replace(modifiedContent, "hostname: myhost", "hostname: "+hostname, 1)
		fmt.Printf("Replaced hostname with: %s\n", hostname)
	}

	// Inject userData into the template by replacing the "data" placeholder
	if userData != "" {
		modifiedContent = injectUserDataIntoTemplate(modifiedContent, userData)
		fmt.Printf("Injected userData into template\n")
	}

	// Create multipart form
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add file field - using the modified content
	fileWriter, err := writer.CreateFormFile("file", "user-data")
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}

	// Write the modified content instead of copying from file
	_, err = fileWriter.Write([]byte(modifiedContent))
	if err != nil {
		return nil, fmt.Errorf("failed to write modified content: %w", err)
	}

	writer.Close()
	// ------------- END cloud-init user-data modification -------------

	// Build URL for the request
	url := c.endpoint + ":27777/api/v1.0/client/volume/cloudinit/metadata/" + encodedPayload

	// Create request with multipart body
	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set the correct Content-Type
	req.Header.Set("Content-Type", writer.FormDataContentType())

	fmt.Printf("Sending HTTP request to: %s\n", url)
	fmt.Printf("Content-Type: %s\n", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		fmt.Println("Closing response body.")
		resp.Body.Close()
	}()

	fmt.Printf("Received response with status code: %d\n", resp.StatusCode)
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("API error response body: %s\n", string(body))
		return nil, fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}

	fmt.Println("Decoding response body into struct...")
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	fmt.Printf("Decoded response: %+v\n", res)
	return &res, nil
}

func (c *Client) FeedFileIntoCloudInitStorage(reqBody schema.FeedFileIntoCloudInitStorageRequest) (string, error) {
	// Marshal reqBody to JSON and encode as base64
	jsonBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal reqBody: %w", err)
	}
	encodedPayload := base64.StdEncoding.EncodeToString(jsonBytes)
	fmt.Printf("Encoded payload (base64): %s\n", encodedPayload)

	// Use the embedded meta-data template from cloudconfig package
	metaDataContent := MetaDataTemplate

	// Create multipart form
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add file field - using just the filename, not the full path
	fileWriter, err := writer.CreateFormFile("file", "meta-data")
	if err != nil {
		return "", fmt.Errorf("failed to create form file: %w", err)
	}

	// Write the meta-data content from string
	_, err = fileWriter.Write([]byte(metaDataContent))
	if err != nil {
		return "", fmt.Errorf("failed to write meta-data content: %w", err)
	}

	writer.Close()

	// Build URL with base64-encoded payload as last path segment
	url := c.endpoint + ":27777/api/v1.0/client/volume/cloudinit/metadata/" + encodedPayload

	// Create request with multipart body
	req, err := http.NewRequest("POST", url, &requestBody)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set the correct Content-Type
	req.Header.Set("Content-Type", writer.FormDataContentType())

	fmt.Printf("Sending HTTP request to: %s\n", url)
	fmt.Printf("Content-Type: %s\n", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer func() {
		fmt.Println("Closing response body.")
		resp.Body.Close()
	}()

	fmt.Printf("Received response with status code: %d\n", resp.StatusCode)
	switch resp.StatusCode {
	case 206:
		return "CONTINUE", nil
	case 200:
		return "OK", nil
	default:
		body, _ := io.ReadAll(resp.Body)
		fmt.Printf("API error response body: %s\n", string(body))
		return "", fmt.Errorf("API error: %d - %s", resp.StatusCode, string(body))
	}
}

// Get storages
func (c *Client) GetStorage() (*schema.GetStorageResponse, error) {
	var res schema.GetStorageResponse
	err := c.CallAPI("GET", "27777", "/api/v1.0/client/volume/accessible", nil, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Get storage by ID
func (c *Client) GetStorageByID(reqBody schema.GetStorageByIDRequest) (*schema.GetStorageByIDResponse, error) {
	var res schema.GetStorageByIDResponse
	err := c.CallAPI("POST", "27777", "/api/v1.0/client/volume/info", reqBody, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Delete a storage volume
func (c *Client) DeleteStorage(reqBody schema.DeleteStorageRequest) (*schema.DeleteStorageResponse, error) {
	var res schema.DeleteStorageResponse
	err := c.CallAPI("POST", "27777", "/api/v1.0/client/volume/delete", reqBody, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// ------------------------------ MOCKED ENDPOINTS -----------------------------

// Get a Network by Id
func (c *Client) GetNetworkByID(reqBody schema.GetNetworkByIDRequest) (*schema.GetNetworkByIDResponse, error) {
	var res schema.GetNetworkByIDResponse

	err := c.CallAPI("POST", "37777", "/api/v1.0/client/network/info", reqBody, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// List all networks
func (c *Client) ListNetwork() (*schema.ListNetworkResponse, error) {
	var res schema.ListNetworkResponse

	err := c.CallAPI("GET", "37777", "/api/v1.0/client/network/list", nil, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Delete a network
func (c *Client) DeleteNetwork(reqBody schema.DeleteNetworkRequest) (*schema.DeleteNetworkResponse, error) {
	var res schema.DeleteNetworkResponse

	err := c.CallAPI("DELETE", "37777", "/api/v1.0/client/network/delete", reqBody, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// Create a network
func (c *Client) CreateNetwork(network schema.CreateNetworkRequest) (*schema.CreateNetworkResponse, error) {
	var res schema.CreateNetworkResponse

	err := c.CallAPI("POST", "37777", "/api/v1.0/client/network/create", network, &res, true)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

// SSH KEYS...

// ------------------------------ UTILS FUNCTIONS -----------------------------

// Base function to perform API calls
func (c *Client) CallAPI(method, port, path string, reqBody, resType interface{}, needAuth bool) error {
	req, err := c.NewRequest(method, port, path, reqBody, needAuth)
	if err != nil {
		return err
	}
	response, err := c.Do(req)
	if err != nil {
		return err
	}
	return c.UnmarshalResponse(response, resType)
}

// NewRequest returns a new HTTP request
func (c *Client) NewRequest(method, port, path string, reqBody interface{}, needAuth bool) (*http.Request, error) {
	var body []byte
	var err error

	if reqBody != nil {
		body, err = json.Marshal(reqBody)
		if err != nil {
			return nil, err
		}
		fmt.Printf("Request body for %s %s:\n%s\n", method, path, string(body)) // TEST
	}

	target := c.endpoint + ":" + port + path
	req, err := http.NewRequest(method, target, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Add("Content-Type", "application/json;charset=utf-8")
	}
	req.Header.Add("Accept", "application/json")

	if needAuth {
		// TODO: add auth
	}

	c.httpClient.Timeout = c.timeout
	return req, nil
}

// Do sends an HTTP request and returns an HTTP response
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	if c.logger != nil {
		c.logger.LogRequest(req)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if c.logger != nil {
		c.logger.LogResponse(resp)
	}
	return resp, nil
}

// UnmarshalResponse checks the response and unmarshals it into the response type if needed
func (c *Client) UnmarshalResponse(response *http.Response, resType interface{}) error {
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		apiErr := &APIError{
			StatusCode: response.StatusCode,
			Message:    string(body),
		}
		return apiErr
	}

	if len(body) == 0 || resType == nil {
		return nil
	}

	d := json.NewDecoder(bytes.NewReader(body))
	d.UseNumber()
	return d.Decode(resType)
}

// ------------------------------ HELPER FUNCTIONS -----------------------------

// injectUserDataIntoTemplate takes a template and userData, indents the userData properly,
// and replaces the "data" placeholder in the template with the indented userData
func injectUserDataIntoTemplate(template, userData string) string {
	// Replace NODEUP_URL_AMD64 references with the actual Elemento kOps release URL - ARM not supported yet
	elementoURL := fmt.Sprintf("https://github.com/Elemento-Modular-Cloud/kops/releases/download/%s/nodeup-linux-amd64", strings.TrimSpace(version))

	// Use regex to replace the value part of NODEUP_URL_AMD64=<existing_value> with the new URL
	nodeupPattern := regexp.MustCompile(`NODEUP_URL_AMD64=([^\s]+)`)
	userData = nodeupPattern.ReplaceAllString(userData, fmt.Sprintf("NODEUP_URL_AMD64=%s", elementoURL))

	// Clean up bash script validation logic - remove hash validation blocks (since the release is not officially signed)
	userData = removeHashValidationBlocks(userData)

	// Indent userData with 6 spaces (needed by yaml format)
	userDataLines := strings.Split(userData, "\n")
	for i, line := range userDataLines {
		if strings.TrimSpace(line) != "" { // Don't indent empty lines
			userDataLines[i] = "      " + line
		}
	}
	indentedUserData := strings.Join(userDataLines, "\n")

	// Replace the "data" placeholder in the template with the actual userData
	return strings.Replace(template, "      data", indentedUserData, 1)
}

// removeHashValidationBlocks removes hash validation logic from bash scripts
func removeHashValidationBlocks(script string) string {
	// Remove the first hash validation block
	hashBlock1 := `  if [[ -f "${file}" ]]; then
    if ! validate-hash "${file}" "${hash}"; then
      rm -f "${file}"
    else
      return 0
    fi
  fi`
	script = strings.ReplaceAll(script, hashBlock1, "")

	// Remove the second hash validation block
	hashBlock2 := `if ! validate-hash "${file}" "${hash}"; then
          echo "== Failed to validate hash for ${url} =="
          rm -f "${file}"
        else
          echo "== Downloaded ${url} with hash ${hash} =="
          return 0
        fi`
	script = strings.ReplaceAll(script, hashBlock2, "")

	// Remove the validate-hash function definition
	validateHashFunc := `validate-hash() {
  local -r file="$1"
  local -r expected="$2"
  local actual

  actual=$(sha256sum "${file}" | awk '{ print $1 }') || true
  if [[ "${actual}" != "${expected}" ]]; then
    echo "== File ${file} is corrupted; hash ${actual} doesn't match expected ${expected} =="
    return 1
  fi
}`
	script = strings.ReplaceAll(script, validateHashFunc, "")

	// Replace the download failure block with success logic
	oldDownloadBlock := `if ! (${cmd} "${url}"); then
          echo "== Failed to download ${url} using ${cmd} =="
          continue
        fi`

	newDownloadBlock := `if ! (${cmd} "${url}"); then
          echo "== Failed to download ${url} using ${cmd} =="
          continue
        else  
          echo "== Successfully downloaded ${url} using ${cmd} =="
          return 0
        fi`

	script = strings.ReplaceAll(script, oldDownloadBlock, newDownloadBlock)

	return script
}

// ------------------------------ ERROR HANDLING -----------------------------

type APIError struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API Error: %d - %s", e.StatusCode, e.Message)
}
