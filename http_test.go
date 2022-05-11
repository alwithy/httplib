package httplib

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/url"
	"testing"
	"time"
)

type LoginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRsp struct {
	Data interface{} `json:"data"`

	Message string `json:"message"`
	Retcode int    `json:"retcode"`
}

func TestDelete(t *testing.T) {
	url := "http://127.0.0.1:8000/test_delete"
	resp, err := Delete(context.Background(), url, []byte(""), 5*time.Second)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode())
}

func TestPost(t *testing.T) {
	purl := "https://e-procurement.test.com/api/srm/user/login/"

	b := &LoginReq{Email: "xian.li@gmail.com", Password: "123456"}
	data, _ := json.Marshal(b)
	rsp, err := Post(context.Background(), purl, data, 5*time.Second)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, rsp.StatusCode())

	lr := &LoginRsp{}
	err = json.Unmarshal(rsp.Body(), lr)
	assert.Nil(t, err)
	assert.Equal(t, 0, lr.Retcode)
	assert.Equal(t, []byte("application/json; charset=utf-8"), rsp.Header.Peek("Content-Type"))
}

func TestPostForm(t *testing.T) {
	rUrl := "https://admin.test.vn/api/v2/logistics/pickup/days/"

	type PickupDaysRsp struct {
		Retdesc string      `json:"retdesc"`
		Rettype string      `json:"rettype"`
		Retcode string      `json:"retcode"`
		Data    interface{} `json:"data"`
	}

	formData := url.Values{}
	formData.Set(TokenStr, "JpywakNgvqVesTzBSeQ3oYc8ZewOj0oP")
	formData.Set("orderid", "1009372")
	formData.Set("shopid", "7439150")

	rsp, err := PostForm(context.Background(), rUrl, formData, 5*time.Second)
	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, rsp.StatusCode())

	ret := new(PickupDaysRsp)
	err = json.Unmarshal(rsp.Body(), rsp)
	assert.Nil(t, err)
	assert.Equal(t, "0", ret.Retcode)
}

func Test_removeToken(t *testing.T) {
	d1 := `{"token":"JpywakNgvqVesTzBSeQ3oYc8ZewOj0oP","orderid":537773739677088512,"log_type":0,"country":"ID"}`
	d2 := `country=ID&token=JpywakNgvqVesTzBSeQ3oYc8ZewOj0oP`
	d3 := ""
	d4 := `{"token":"JpywakNgvqVesTzBSeQ3oYc8ZewOj0oP","orderid":537773739677088512,"log_type":0,"country":"ID"`

	ctx := context.Background()
	fmt.Println(removeToken(ctx, d1, strContentTypeJson))
	fmt.Println(removeToken(ctx, d2, strContentTypeFormUrlEncoded))
	fmt.Println(removeToken(ctx, d3, strContentTypeFormUrlEncoded))
	fmt.Println(removeToken(ctx, d4, strContentTypeJson))
}

func test_rm_token() {
	d1 := `{"token":"JpywakNgvqVesTzBSeQ3oYc8ZewOj0oP","orderid":537773739677088512,"log_type":0,"country":"ID"}`
	ctx := context.Background()
	removeToken(ctx, d1, strContentTypeJson)

}

func BenchmarkRemoveToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		test_rm_token()
	}
}

func TestPostRetry(t *testing.T) {
	purl := "http://127.0.0.1:7999/hello"

	for i := 0; i < 10; i++ {
		fmt.Println("START", i)
		opts := []Option{
			ReadTimeOutClient(true),
		}
		rsp, err := Get(context.Background(), purl, nil, 5*time.Second, opts...)
		fmt.Println(err)
		if rsp != nil {
			fmt.Println(rsp.StatusCode(), rsp.String())
		}
		fmt.Println(rsp.String())
		time.Sleep(10 * time.Second)
	}
}
