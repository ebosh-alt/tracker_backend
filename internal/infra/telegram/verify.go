package telegram

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

var ErrInvalidSignature = errors.New("invalid signature")
var ErrExpired = errors.New("auth data expired")

func ParseInt64(v string) (int64, bool) {
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, false
	}
	return n, true
}

// VerifyInitData validates Telegram Web App initData.
// https://core.telegram.org/bots/webapps#validating-data-received-via-the-web-app
func (cl *Client) VerifyInitData(initData string) (map[string]string, error) {
	vals, err := url.ParseQuery(initData)
	if err != nil {
		return nil, err
	}

	data := make(map[string]string)
	for k, v := range vals {
		if len(v) > 0 {
			data[k] = v[0]
		}
	}

	hash := data["hash"]
	if hash == "" {
		return nil, ErrInvalidSignature
	}
	delete(data, "hash")

	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	pairs := make([]string, 0, len(keys))
	for _, k := range keys {
		pairs = append(pairs, k+"="+data[k])
	}
	checkString := strings.Join(pairs, "\n")

	// Per Telegram docs:
	// secret_key = HMAC_SHA256(key="WebAppData", data=bot_token)
	secret := hmac.New(sha256.New, []byte("WebAppData"))
	secret.Write([]byte(cl.token))
	secretKey := secret.Sum(nil)

	sig := hmac.New(sha256.New, secretKey)
	sig.Write([]byte(checkString))
	expected := hex.EncodeToString(sig.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(hash)) {
		return nil, ErrInvalidSignature
	}

	if tsStr := data["auth_date"]; tsStr != "" {
		ts, _ := strconv.ParseInt(tsStr, 10, 64)
		if ts > 0 {
			if time.Since(time.Unix(ts, 0)) > 24*time.Hour {
				return nil, ErrExpired
			}
		}
	}

	return data, nil
}

// ExtractUser parses `user` field from verified initData payload.
func ExtractUser(payload map[string]string) (map[string]any, bool) {
	userStr := payload["user"]
	if userStr == "" {
		return nil, false
	}
	var user map[string]any
	if err := json.Unmarshal([]byte(userStr), &user); err != nil {
		return nil, false
	}
	return user, true
}
