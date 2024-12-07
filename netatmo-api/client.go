package netatmo_api

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/oauth2"
)

const (
	authURL  = "https://api.netatmo.com/oauth2/authorize"
	tokenURL = "https://api.netatmo.com/oauth2/token"
)

type Config struct {
	Username     string
	Password     string
	RefreshToken string
	ClientID     string
	ClientSecret string
	Scopes       []string
}

type Client struct {
	httpClient *http.Client
	ctx        context.Context
	Config     *Config
}

func NewClient(ctx context.Context, cnf *Config) (*Client, error) {
	httpClient, err := getOauthClient(ctx, cnf)
	if err != nil {
		return nil, err
	}

	return &Client{
		httpClient: httpClient,
		ctx:        ctx,
		Config:     cnf,
	}, nil
}

func getOauthToken(ctx context.Context, oauth *oauth2.Config, cnf *Config) (*oauth2.Token, error) {
	const tokenFile = "netatmo_token.json"

	token := &oauth2.Token{}
	if file, err := ioutil.ReadFile(tokenFile); err == nil {
		if err := json.Unmarshal(file, token); err == nil && token.Valid() && token.Expiry.After(time.Now()) {
			log.Println("Loaded valid token from file.")
			return token, nil
		}
	}

	if cnf.RefreshToken != "" {
		data := url.Values{}
		data.Set("grant_type", "refresh_token")
		data.Set("refresh_token", cnf.RefreshToken)
		data.Set("client_id", cnf.ClientID)
		data.Set("client_secret", cnf.ClientSecret)

		req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(data.Encode()))
		if err != nil {
			return nil, fmt.Errorf("failed to create refresh token request: %w", err)
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to refresh token: %w", err)
		}
		defer closeBody(resp)

		if resp.StatusCode == http.StatusOK {
			var response struct {
				AccessToken  string `json:"access_token"`
				RefreshToken string `json:"refresh_token"`
				ExpiresIn    int    `json:"expires_in"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&response); err == nil {
				log.Println("Token refreshed successfully.")

				expiryTime := time.Now().Add(time.Duration(response.ExpiresIn) * time.Second)
				token = &oauth2.Token{
					AccessToken:  response.AccessToken,
					RefreshToken: response.RefreshToken,
					Expiry:       expiryTime,
				}

				if file, err := json.Marshal(token); err == nil {
					_ = ioutil.WriteFile(tokenFile, file, 0644)
					log.Println("Token saved to file with updated expiry.")
				} else {
					log.Printf("failed to save token to file: %v", err)
				}

				return token, nil
			}
			log.Printf("failed to decode refreshed token response")
		} else {
			body, _ := ioutil.ReadAll(resp.Body)
			return nil, fmt.Errorf("failed to refresh token, status: %d, body: %s", resp.StatusCode, string(body))
		}
	}

	log.Println("Falling back to password authentication.")
	token, err := oauth.PasswordCredentialsToken(ctx, cnf.Username, cnf.Password)
	if err != nil {
		return nil, fmt.Errorf("could not get token for %v: %w", cnf.Username, err)
	}

	if file, err := json.Marshal(token); err == nil {
		_ = ioutil.WriteFile(tokenFile, file, 0644)
		log.Println("Token saved to file.")
	} else {
		log.Printf("failed to save token to file: %v", err)
	}

	return token, nil
}

func getOauthClient(ctx context.Context, cnf *Config) (*http.Client, error) {
	oauth := &oauth2.Config{
		ClientID:     cnf.ClientID,
		ClientSecret: cnf.ClientSecret,
		Scopes:       cnf.Scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		},
	}

	token, err := getOauthToken(ctx, oauth, cnf)
	if err != nil {
		return nil, err
	}

	return oauth.Client(ctx, token), nil
}

func closeBody(res *http.Response) {
	err := res.Body.Close()
	if err != nil {
		log.Printf("Error closing response body: %v\n", err)
	}
}

func (c *Client) get(u *url.URL, v interface{}) error {
	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}

	return c.request(req, v)
}

func (c *Client) request(req *http.Request, v interface{}) error {
	res, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("error during HTTP request: %w", err)
	}
	defer closeBody(res)

	
	if res.StatusCode == http.StatusForbidden {
		log.Println("Access token expired. Attempting to refresh token.")
		oauthConfig := &oauth2.Config{
			ClientID:     c.Config.ClientID,
			ClientSecret: c.Config.ClientSecret,
			Scopes:       c.Config.Scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  authURL,
				TokenURL: tokenURL,
			},
		}

		token, err := getOauthToken(c.ctx, oauthConfig, c.Config)
		if err != nil {
			return fmt.Errorf("could not refresh token: %w", err)
		}

		c.httpClient = oauthConfig.Client(c.ctx, token)
		log.Println("Token refreshed successfully.")

		req.Body = ioutil.NopCloser(strings.NewReader("")) 
				return c.request(req, v)
	}

	switch res.StatusCode {
	case http.StatusOK:
		var objmap map[string]json.RawMessage
		if err := json.NewDecoder(res.Body).Decode(&objmap); err != nil {
			return fmt.Errorf("could not decode JSON: %w", err)
		}
		if body, ok := objmap["body"]; ok {
			if err := json.Unmarshal(body, &v); err != nil {
				return fmt.Errorf("could not decode body: %w", err)
			}
			return nil
		}
		return fmt.Errorf("could not find body: %v", objmap)
	default:
		bodyString, _ := readString(res)
		return fmt.Errorf("invalid request: status_code = %d content=%v", res.StatusCode, bodyString)
	}
}

func readString(resp *http.Response) (string, error) {
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(bodyBytes), nil
}
