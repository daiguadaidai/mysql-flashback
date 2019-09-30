package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

type responseData struct {
	Status  bool        `form:"status" json:"status"`
	Message string      `form:"message" json:"message"`
	Data    interface{} `form:"data" json:data`
}

// post访问url
func PostURL(url string, data interface{}) (interface{}, error) {
	bytesData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("序列化json出错. %v", err)
	}

	resp, err := http.Post(url, "application/json;charset=utf-8",
		bytes.NewBuffer(bytesData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return visitOk(resp)
}

// get请求
func GetURL(url string, query string) (interface{}, error) {
	url += query
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return visitOk(resp)
}

// 下载文件
func DownloadFile(url string, std string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, err := ioutil.ReadAll(resp.Body)
		errMSG := fmt.Sprintf("http状态码: %d. %s", resp.StatusCode, string(b))
		if err != nil {
			errMSG = fmt.Sprintf("%s. %v", errMSG, err)
		}
		return fmt.Errorf("%v", errMSG)
	}

	out, err := os.Create(std)
	if err != nil {
		return fmt.Errorf("创建命令文件失败. %s", std)
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		return fmt.Errorf("拷贝数据出错. %v", err)
	}

	return nil
}

func PutURL(url string, data interface{}) (interface{}, error) {
	d, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	body := bytes.NewBuffer(d)
	req, err := http.NewRequest(http.MethodPut, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json; charset=utf-8")
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return visitOk(resp)
}

func GetURLQuery(data map[string]interface{}) string {
	params := make([]string, 0, 1)
	for key, value := range data {
		d := fmt.Sprintf("%v=%v", key, value)
		params = append(params, d)
	}

	if len(params) == 0 {
		return ""
	}

	return strings.Replace(
		fmt.Sprintf("?%s", strings.Join(params, "&")),
		" ", "%20", -1)
}

// get请求
func GetURLRaw(url string, query string) ([]byte, error) {
	url += query
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("获取返回数据失败. %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("code: %d. 访问pili失败. %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// 判断访问数据是否成功
func visitOk(resp *http.Response) (interface{}, error) {
	d, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("获取返回数据失败. %v", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("code: %d. 访问失败. %s", resp.StatusCode, string(d))
	}

	result := new(responseData)
	err = json.Unmarshal(d, result)
	if err != nil {
		return nil, err
	}

	if !result.Status {
		return nil, fmt.Errorf("%v", result.Message)
	}

	return result.Data, nil
}
