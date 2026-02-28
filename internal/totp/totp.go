package totp

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"math"
	"strings"
	"time"
)

func Code(secret string, t time.Time, digits int, period int) (string, error) {
	key, err := base32.StdEncoding.
		WithPadding(base32.NoPadding).
		DecodeString(strings.ToUpper(strings.ReplaceAll(secret, " ", "")))
	if err != nil {
		return "", fmt.Errorf("decode secret: %w", err)
	}

	counter := uint64(t.Unix()) / uint64(period)

	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, counter)

	mac := hmac.New(sha1.New, key)
	mac.Write(buf)
	h := mac.Sum(nil)

	offset := h[len(h)-1] & 0x0f
	code := int(binary.BigEndian.Uint32(h[offset:offset+4]) & 0x7fffffff)
	code = code % int(math.Pow10(digits))

	return fmt.Sprintf("%0*d", digits, code), nil
}

func Remaining(period int, t time.Time) int {
	return period - int(t.Unix())%period
}
