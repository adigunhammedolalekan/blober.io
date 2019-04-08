package blober

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type BloberClient struct {
	client *http.Client
	appName string
	hostUrl string
}

type UploadResponse struct {
	Id int64 `json:"ID"`
	DownloadUrl string `json:"download_url"`
	Hash string `json:"hash"`
	Size int64 `json:"size"`
	Filename string `json:"filename"`
	ContentType string `json:"content_type"`
}

func New(host, appName string) (*BloberClient, error) {
	u, err := url.Parse(host)
	if err != nil {
		return nil, err
	}

	if len(appName) == 0 {
		return nil, errors.New("invalid app name")
	}

	httpClient := &http.Client{Timeout: 60 * time.Second}
	return &BloberClient{client:httpClient, hostUrl:u.String(), appName:appName}, nil
}

func (client *BloberClient) Upload(filename string, isPrivate bool, body io.Reader) (*UploadResponse, error) {
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
	file, err := w.CreateFormFile("file_data", filename)
	if err != nil {
		return nil, err
	}

	contents, err := ioutil.ReadAll(body)
	_, err = file.Write(contents)
	if err != nil {
		return nil, err
	}

	err = w.WriteField("private", strconv.FormatBool(isPrivate))
	if err != nil {
		return nil, err
	}

	err = w.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", client.formatHostUrl(), buf)
	if err != nil {
		return nil, err
	}

	resp, err := client.client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			log.Printf("failed to close response body %v", err)
		}
	}()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	uploadResponse := &UploadResponse{}
	err = json.Unmarshal(data, uploadResponse)
	if err != nil {
		return nil, err
	}

	return uploadResponse, nil
}

func (client *BloberClient) formatHostUrl() string {
	return fmt.Sprintf("%s%s%s", client.hostUrl, client.appName, "/upload")
}