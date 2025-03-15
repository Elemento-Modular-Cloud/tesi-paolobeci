package ecloud


// ------------------------------ API CALLS FUNCTIONS -------------------------

func (c *Client) GetVm(url string, resType interface{}) error {
	return c.CallAPI("GET", url, nil, resType, true)
}

func (c *Client) GetVmById(url string, id string, resType interface{}) error {
	return c.CallAPI("GET", url, nil, resType, true)
}

// ...


// ------------------------------ UTILS FUNCTIONS -----------------------------

func (c *Client) CallAPI(method, path string, reqBody, resType interface{}, needAuth bool) error {
	req, err := c.NewRequest(method, path, reqBody, needAuth)
	if err != nil {
		return err
	}
	req = req.WithContext(ctx)
	response, err := c.Do(req)
	if err != nil {
		return err
	}
	return c.UnmarshalResponse(response, resType)
}

// NewRequest returns a new HTTP request
func (c *Client) NewRequest(method, path string, reqBody interface{}, needAuth bool) (*http.Request, error) {
	var body []byte
	var err error

	if reqBody != nil {
		body, err = json.Marshal(reqBody)
		if err != nil {
			return nil, err
		}
	}

	target := getTarget(c.endpoint, path)
	req, err := http.NewRequest(method, target, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	// Inject headers
	if body != nil {
		req.Header.Add("Content-Type", "application/json;charset=utf-8")
	}
	if c.AppKey != "" {
		req.Header.Add("X-Ovh-Application", c.AppKey)
	}
	req.Header.Add("Accept", "application/json")

	// Inject signature. Some methods do not need authentication, especially /time,
	// /auth and some /order methods are actually broken if authenticated.
	if needAuth {
		if c.AppKey != "" {
			timeDelta, err := c.TimeDelta()
			if err != nil {
				return nil, err
			}

			timestamp := getLocalTime().Add(-timeDelta).Unix()

			req.Header.Add("X-Ovh-Timestamp", strconv.FormatInt(timestamp, 10))
			req.Header.Add("X-Ovh-Consumer", c.ConsumerKey)

			h := sha1.New()
			h.Write([]byte(fmt.Sprintf("%s+%s+%s+%s+%s+%d",
				c.AppSecret,
				c.ConsumerKey,
				method,
				target,
				body,
				timestamp,
			)))
			req.Header.Add("X-Ovh-Signature", fmt.Sprintf("$1$%x", h.Sum(nil)))
		} else if c.ClientID != "" {
			token, err := c.oauth2TokenSource.Token()
			if err != nil {
				return nil, fmt.Errorf("failed to retrieve OAuth2 Access Token: %w", err)
			}

			req.Header.Set("Authorization", "Bearer "+token.AccessToken)
		} else if c.AccessToken != "" {
			req.Header.Set("Authorization", "Bearer "+c.AccessToken)
		}
	}

	// Send the request with requested timeout
	c.Client.Timeout = c.Timeout

	if c.UserAgent != "" {
		req.Header.Set("User-Agent", "github.com/ovh/go-ovh ("+c.UserAgent+")")
	} else {
		req.Header.Set("User-Agent", "github.com/ovh/go-ovh")
	}

	return req, nil
}

// Do sends an HTTP request and returns an HTTP response
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	if c.Logger != nil {
		c.Logger.LogRequest(req)
	}
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}
	if c.Logger != nil {
		c.Logger.LogResponse(resp)
	}
	return resp, nil
}


// UnmarshalResponse checks the response and unmarshals it into the response
// type if needed Helper function, called from CallAPI
func (c *Client) UnmarshalResponse(response *http.Response, resType interface{}) error {
	// Read all the response body
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return err
	}

	// < 200 && >= 300 : API error
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		apiError := &APIError{Code: response.StatusCode}
		if err = json.Unmarshal(body, apiError); err != nil {
			apiError.Message = string(body)
		}
		apiError.QueryID = response.Header.Get("X-Ovh-QueryID")

		return apiError
	}

	// Nothing to unmarshal
	if len(body) == 0 || resType == nil {
		return nil
	}

	d := json.NewDecoder(bytes.NewReader(body))
	d.UseNumber()
	return d.Decode(&resType)
}