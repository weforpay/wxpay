package wxpay

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/url"
	"time"

	"github.com/philchia/wxpay"
)

func NewApiKey() (key string, err error) {
	var bs [16]byte
	_, err = rand.Read(bs[:])
	key = hex.EncodeToString(bs[:])
	return
}

func BizPayUrl(appId, mchId, productId, nonce_str, apiKey string) (url string) {
	var params = make(map[string]string)
	params["appid"] = appId
	params["mch_id"] = mchId
	params["product_id"] = productId
	params["time_stamp"] = wxpay.NewTimestampString()
	params["nonce_str"] = nonce_str
	params["sign"] = wxpay.Sign(params, apiKey)
	urlParams := wxpay.SortAndConcat(params)
	url = fmt.Sprintf("weixin://wxpay/bizpayurl?%s\n", urlParams)
	fmt.Printf("BizPayUrl:%s", url)
	return
}

type WxPay struct {
	AppId    string
	AppSec   string
	MchId    string
	SubMchId string
	ApiKey   string
	NonceStr string
}

func (this *WxPay) PayUrl(productId string) (url string) {
	return BizPayUrl(this.AppId, this.SubMchId, productId, this.NonceStr, this.ApiKey)
}

func (this *WxPay) ToShort(productId string) (shortUrl string, err error) {
	var params = make(map[string]string)
	params["appid"] = this.AppId
	params["mch_id"] = this.MchId
	params["long_url"] = this.PayUrl(productId)
	params["nonce_str"] = this.NonceStr
	params["sub_mch_id"] = this.SubMchId
	params["sign"] = wxpay.Sign(params, this.ApiKey)

	var bs []byte
	type XmlResult struct {
		XMLName    xml.Name `xml:"xml"`
		ReturnCode string   `xml:"return_code"`
		ReturnMsg  string   `xml:"return_msg"`
		MchId      string   `xml:"mch_id"`
		AppId      string   `xml:"appid"`
		ResultCode string   `xml:"result_code"`
		ShortUrl   string   `xml:"short_url"`
		NonceStr   string   `xml:"nonce_str"`
		ErrCode    string   `xml:"err_code"`
		Sign       string   `xml:"sign"`
	}

	bs, err = doHttpPost("https://api.mch.weixin.qq.com/tools/shorturl", []byte(wxpay.ToXmlString(params)))
	fmt.Printf("bs:%s", string(bs))
	var out XmlResult
	err = xml.Unmarshal(bs, &out)

	if err != nil {
		return
	}
	fmt.Printf("out:%#v", out)
	params = make(map[string]string)
	params["appid"] = out.AppId
	params["mch_id"] = out.MchId
	params["short_url"] = out.ShortUrl
	params["nonce_str"] = out.NonceStr
	params["return_code"] = out.ReturnCode
	params["return_msg"] = out.ReturnMsg
	params["result_code"] = out.ResultCode
	params["err_code"] = out.ErrCode
	sign := wxpay.Sign(params, this.ApiKey)
	if sign != out.Sign {
		err = fmt.Errorf("api return sign error")
		return
	}
	return
}
func (this *WxPay) NewTradeNo() string {
	now := time.Now()
	return fmt.Sprintf("%04d%02d%02d%d",
		now.Year(), now.Month(), now.Day(),
		now.Nanosecond())
}

type UnifiedOrderResult struct {
	XMLName    xml.Name `xml:"xml"`
	ReturnCode string   `xml:"return_code"`
	ReturnMsg  string   `xml:"return_msg"`
	PrepayId   string   `xml:"prepay_id"`
	CodeUrl    string   `xml:"code_url"`
	MchId      string   `xml:"mch_id"`
	SubMchId   string   `xml:"sub_mch_id"`
	AppId      string   `xml:"appid"`
	OpenId     string   `xml:"openid"`
	ResultCode string   `xml:"result_code"`
	TradeType  string   `xml:"trade_type"`
	NonceStr   string   `xml:"nonce_str"`
	ErrCode    string   `xml:"err_code"`
	DeviceInfo string   `xml:"device_info"`
	ErrCodeDes string   `xml:"err_code_des"`
	Sign       string   `xml:"sign"`
	TimeStamp  string   `xml:"-"`
}

//统一下单接口
//body 商品描述
//out_trade_no 商户订单号
//total_fee 总金额 单位为分
//spbill_create_ip APP和网页支付提交用户端ip，Native支付填调用微信支付API的机器IP。
//notify_url 接收微信支付异步通知回调地址，通知url必须为直接可访问的url，不能携带参数。
//trade_type 交易类型 取值如下：JSAPI，NATIVE，APP
func (this *WxPay) UnifiedOrder(openId, body, out_trade_no, spbill_create_ip, notify_url, trade_type string, total_fee int) (result *UnifiedOrderResult, err error) {
	var params = make(map[string]string)
	params["appid"] = this.AppId
	params["mch_id"] = this.MchId
	params["sub_mch_id"] = this.SubMchId
	params["nonce_str"] = this.NonceStr
	params["body"] = body
	params["out_trade_no"] = out_trade_no
	params["total_fee"] = fmt.Sprintf("%.0d", total_fee)
	params["spbill_create_ip"] = spbill_create_ip
	params["notify_url"] = notify_url
	params["trade_type"] = trade_type

	params["openid"] = openId
	params["sign"] = wxpay.Sign(params, this.ApiKey)

	var bs []byte

	reqBody := wxpay.ToXmlString(params)

	bs, err = doHttpPost("https://api.mch.weixin.qq.com/pay/unifiedorder", []byte(reqBody))
	fmt.Printf("UnifiedOrder reqbody:%s,bs:%s", string(reqBody), string(bs))
	result = &UnifiedOrderResult{}
	err = xml.Unmarshal(bs, result)
	return
}

//微信默认授权获取openId
func (this *WxPay) GetWxH5Auth(state, uri string) string {
	uri = url.QueryEscape(uri)
	return fmt.Sprintf("https://open.weixin.qq.com/connect/oauth2/authorize?appid=%s&redirect_uri=%s&response_type=code&scope=snsapi_base&state=%s#wechat_redirect",
		this.AppId, uri, state,
	)
}

type WxH5AccessToken struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	OpenId       string `json:"openid"`
	Scope        string `json:"scope"`
}

func (this *WxPay) GetWxH5AccessToken(code string) (token *WxH5AccessToken, err error) {
	var url = fmt.Sprintf("https://api.weixin.qq.com/sns/oauth2/access_token?appid=%s&secret=%s&code=%s&grant_type=authorization_code",
		this.AppId, this.AppSec, code,
	)
	bs, err := doHttpGet(url, nil)
	if err != nil {
		return
	}
	fmt.Printf("bs:%s\n", bs)
	token = &WxH5AccessToken{}
	err = json.Unmarshal(bs, token)
	return
}
