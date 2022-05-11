package httplib

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/valyala/fasthttp"
	"mime/multipart"
	"net/url"
	"strings"
	"time"
)

var (
	strGet    = []byte("GET")
	strPost   = []byte("POST")
	strDelete = []byte("DELETE")
	strPut    = []byte("PUT")
	strPatch  = []byte("PATCH")

	strContentTypeJson           = "application/json"
	strContentTypeFormUrlEncoded = "application/x-www-form-urlencoded"
)

const TokenStr = "token"

var (
	httpInterceptor HttpInterceptor
)

type CookieKVPair struct {
	Key   string
	Value string
}

func RegisterInterceptor(interceptor HttpInterceptor) error {
	if httpInterceptor != nil {
		return fmt.Errorf("HttpInterceptor already exists")
	}
	httpInterceptor = interceptor
	return nil
}

type Response = fasthttp.Response

func Delete(ctx context.Context, url string, data []byte, timeout time.Duration, opts ...Option) (response *Response, err error) {
	if httpInterceptor != nil {
		return httpInterceptor(ctx, strDelete, url, nil, data, timeout, strContentTypeJson, request, opts...)
	}
	return request(ctx, strDelete, url, nil, data, timeout, strContentTypeJson, opts...)
}

func Post(ctx context.Context, url string, data []byte, timeout time.Duration, opts ...Option) (response *Response, err error) {
	if httpInterceptor != nil {
		return httpInterceptor(ctx, strPost, url, nil, data, timeout, strContentTypeJson, request, opts...)
	}
	return request(ctx, strPost, url, nil, data, timeout, strContentTypeJson, opts...)
}

func Put(ctx context.Context, url string, data []byte, timeout time.Duration, opts ...Option) (response *Response, err error) {
	if httpInterceptor != nil {
		return httpInterceptor(ctx, strPut, url, nil, data, timeout, strContentTypeJson, request, opts...)
	}
	return request(ctx, strPut, url, nil, data, timeout, strContentTypeJson, opts...)
}

func Patch(ctx context.Context, url string, data []byte, timeout time.Duration, opts ...Option) (response *Response, err error) {
	if httpInterceptor != nil {
		return httpInterceptor(ctx, strPatch, url, nil, data, timeout, strContentTypeJson, request, opts...)
	}
	return request(ctx, strPatch, url, nil, data, timeout, strContentTypeJson, opts...)
}

func PostMultiPart(ctx context.Context, url string, fields map[string]string, fileFieldName string, fileContent []byte, fileName string, timeout time.Duration, opts ...Option) (response *Response, err error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile(fileFieldName, fileName)
	if err != nil {
		fmt.Println("CreateFormFile failed", err)
		return
	}
	_, err = part.Write(fileContent)
	if err != nil {
		fmt.Println("part Write failed", err)
		return
	}

	for key, val := range fields {
		err = writer.WriteField(key, val)
		if err != nil {
			fmt.Println("WriteField failed", err)
			return
		}
	}

	err = writer.Close() // this method is not memory operation, so no need for defer
	if err != nil {
		fmt.Println("Close writer failed", err)
		return
	}

	contentType := writer.FormDataContentType()

	if httpInterceptor != nil {
		return httpInterceptor(ctx, strPost, url, nil, body.Bytes(), timeout, contentType, request, opts...)
	}
	return request(ctx, strPost, url, nil, body.Bytes(), timeout, contentType, opts...)
}

func Get(ctx context.Context, url string, params []byte, timeout time.Duration, opts ...Option) (response *Response, err error) {
	if httpInterceptor != nil {
		return httpInterceptor(ctx, strGet, url, params, nil, timeout, strContentTypeJson, request, opts...)
	}
	return request(ctx, strGet, url, params, nil, timeout, strContentTypeJson, opts...)
}

func PostForm(ctx context.Context, url string, data url.Values, timeout time.Duration, opts ...Option) (response *Response, err error) {
	if httpInterceptor != nil {
		return httpInterceptor(ctx, strPost, url, nil, []byte(data.Encode()), timeout, strContentTypeFormUrlEncoded, request, opts...)
	}
	return request(ctx, strPost, url, nil, []byte(data.Encode()), timeout, strContentTypeFormUrlEncoded, opts...)
}

var defaultClient = &fasthttp.Client{}
var readTimeoutClient = &fasthttp.Client{
	ReadTimeout: 120 * time.Second,
}

func request(ctx context.Context, method []byte, url string, params []byte, data []byte, timeout time.Duration, contentType string, opts ...Option) (rsp *fasthttp.Response, err error) {
	optStruct := new(option)
	for _, opt := range opts {
		opt(optStruct)
	}

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	header := &req.Header
	header.SetContentType(contentType)
	header.SetMethodBytes(method)

	for _, h := range optStruct.headers {
		header.Set(h.key, h.value)
	}

	for _, cookie := range optStruct.cookies {
		header.SetCookie(cookie.key, cookie.value)
	}

	req.SetRequestURI(url)

	if params != nil {
		req.URI().SetQueryStringBytes(params)
	}

	if data != nil {
		req.SetBody(data)
	}

	rsp = new(fasthttp.Response)
	reqCopy := fasthttp.AcquireRequest()
	req.CopyTo(reqCopy)
	defer fasthttp.ReleaseRequest(reqCopy)

	var client *fasthttp.Client
	if optStruct.useReadTimeOutClient {
		client = readTimeoutClient
	} else {
		client = defaultClient
	}

	err = client.DoTimeout(req, rsp, timeout)
	if err != nil || rsp.StatusCode() >= 400 {
		for i := 0; i < optStruct.retryTimes; i++ {
			rsp = new(fasthttp.Response)
			reqCopy.CopyTo(req)
			err = client.DoTimeout(req, rsp, timeout)
			if err == nil && rsp.StatusCode() < 400 {
				break
			}

			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}

	rspData := string(rsp.Body())

	printRspData := ""
	if err != nil {
		printRspData = err.Error()
	} else {
		if !optStruct.suppressRspLog {
			printRspData = rspData
		}
	}
	if len(printRspData) > 1024 {
		printRspData = printRspData[:1024]
	}

	rawData := string(data)
	if optStruct.isJwt {
		rawData = string(optStruct.jwtRawData)
	}

	printData := string(data)
	if !optStruct.enableJwtEncodedData {
		printData = ""
	}

	if string(method) == string(strPost) {
		rawData = removeToken(ctx, rawData, contentType)
		printData = removeToken(ctx, printData, contentType)
	}

	if !optStruct.suppressLog {
		fmt.Printf( "[send-http-request]url=%s,method=%s,params=%s,raw=%v,data=%s,timeout=%v,status_code=%d, response=%s",
			url, string(method), string(params), rawData, printData, timeout, rsp.StatusCode(), printRspData)
	}

	return rsp, err
}

func removeToken(ctx context.Context, data string, contentType string) string {
	if data == "" {
		return data
	}

	if contentType == strContentTypeJson {
		var res map[string]interface{}
		d := json.NewDecoder(strings.NewReader(data))
		d.UseNumber()
		err := d.Decode(&res)
		if err != nil {
			return data
		}

		if token, ok := res[TokenStr]; ok {
			if _, ok := token.(string); ok {
				res[TokenStr] = "***"
			}
		}

		rmTokenData, _ := json.Marshal(res)
		return string(rmTokenData)

	} else if contentType == strContentTypeFormUrlEncoded {
		vals, err := url.ParseQuery(data)
		if err != nil {
			return data
		}
		if _, ok := vals[TokenStr]; ok {
			vals.Set(TokenStr, "")
		}
		return vals.Encode()
	}

	return data
}
