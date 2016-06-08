package wxpay

import (
	"testing"

	"github.com/philchia/wxpay"
)

var wp *WxPay

func init() {
	wp = &WxPay{
		AppId:    "wxc90c08d45f3da985",
		MchId:    "1336275601",
		SubMchId: "1352041601",
		ApiKey:   "",
		NonceStr: wxpay.NewNonceString(),
	}

}

func TestNewApiKey(t *testing.T) {
	key, err := NewApiKey()
	if err != nil {
		t.Error(err)
	}
	t.Logf("%s", key)
}

func TestToShortUrl(t *testing.T) {
	_, err := wp.ToShort("1")
	if err != nil {
		t.Error(err)
	}
}
func TestToUnifiedOrder(t *testing.T) {

	wp.UnifiedOrder("o7iaMs91t51kHWJHwmlJzJyPGQL8", "测试", wp.NewTradeNo(), "8.8.8.8",
		"http://pay.weforpay.com/wxpaygw", "NATIVE", 1)
}
