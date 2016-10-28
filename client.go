/******************************************************
# DESC    : pack client hessian http request and parse response
# AUTHOR  : Alex Stocks
# EMAIL   : alexstocks@foxmail.com
# MOD     : 2016-10-22 20:25
# FILE    : client.go
******************************************************/

package hessian

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
)

type hessianRequest struct {
	body []byte
}

//向hessian服务发请求,并将解析结果返回
//url string hessian 服务地址
//method string hessian 公开的方法
//params ...Any 请求参数
func Request(url string, method string, params ...Any) (interface{}, error) {
	r := &hessianRequest{}
	r.packHead(method)
	for _, v := range params {
		r.packParam(v)
	}
	r.packEnd()

	resp, err := httpPost(url, bytes.NewReader(r.body))
	if err != nil {
		return nil, err
	}

	this := NewDecoder(resp)
	v, err := this.Decode()

	if err != nil {
		return nil, err
	}

	return v, nil
}

// http post 请求, 返回body字节数组
func httpPost(url string, body io.Reader) ([]byte, error) {
	var (
		err  error
		rb   []byte
		resp *http.Response
	)

	if resp, err = http.Post(url, "application/binary", body); err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf(resp.Status)
		return nil, err
	}
	rb, err = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	return rb, err
}

// 封装 hessian 请求头
func (this *hessianRequest) packHead(method string) {
	this.body = append(this.body, []byte{99, 0, 1, 109}...)
	this.body = append(this.body, PackUint16(uint16(len(method)))...)
	this.body = append(this.body, []byte(method)...)
}

// 封装参数
func (this *hessianRequest) packParam(p Any) {
	this.body = Encode(p, this.body)
}

// 封装包尾
func (this *hessianRequest) packEnd() {
	this.body = append(this.body, 'z')
}
