package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

type Portal struct {
	Endpoint string
	Token    string
}

type EnvCheckInfo struct {
	AMI string `json:"ami_id"`
	AZ  string `json:"az_id"`
}

func LoadPortalCredentials() (*Portal, error) {
	var credentials struct {
		Dev   bool   `json:"dev"`
		Token string `json:"token"`
		Host  string `json:"host"`
	}
	f, err := os.Open("/opt/isucon-env-checker/portal_credentials.json")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if err := json.NewDecoder(f).Decode(&credentials); err != nil {
		return nil, err
	}
	return &Portal{
		Endpoint: "https://" + credentials.Host + "/envcheck",
		Token:    credentials.Token,
	}, nil
}

func (p *Portal) NewHttpRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", "Bearer "+p.Token)
	return req, nil
}

func (p *Portal) GetInfo(name string) (EnvCheckInfo, error) {
	q := make(url.Values)
	q.Set("name", name)
	req, err := p.NewHttpRequest("GET", p.Endpoint+"/info/?"+q.Encode(), nil)
	if err != nil {
		return EnvCheckInfo{}, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return EnvCheckInfo{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		msg, _ := io.ReadAll(res.Body)
		return EnvCheckInfo{}, fmt.Errorf("http status error: %d (%s)", res.StatusCode, string(msg))
	}

	var info EnvCheckInfo
	if err := json.NewDecoder(res.Body).Decode(&info); err != nil {
		return EnvCheckInfo{}, err
	}

	return info, nil
}

func (p *Portal) SendResult(r Result) error {
	body, err := json.Marshal(struct {
		Name         string `json:"name"`
		Passed       bool   `json:"passed"`
		IPAddress    string `json:"ip_address"`
		Message      string `json:"message"`
		AdminMessage string `json:"admin_message"`
		RawData      string `json:"raw_data"`
	}{
		Name:         r.Name,
		Passed:       r.Passed,
		IPAddress:    r.LocalIPAddress,
		Message:      r.Message,
		AdminMessage: r.AdminMessage,
		RawData:      r.RawData,
	})
	if err != nil {
		return err
	}
	req, err := p.NewHttpRequest("POST", p.Endpoint+"/result/", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusNoContent {
		msg, _ := io.ReadAll(res.Body)
		return fmt.Errorf("http status error: %d (%s)", res.StatusCode, string(msg))
	}

	io.Copy(io.Discard, res.Body)
	return nil
}
