package gmcache

import (
	"strconv"
	"testing"
)

func TestHashing(t *testing.T) {
	hash := New(3, func(data []byte) uint32 {
		// 使用了自定义的哈希算法，传入字符串表示的数字，返回对应的数字即可
		i, _ := strconv.Atoi(string(data))
		return uint32(i)
	})

	/*
		原值 m.keys [6 16 26 4 14 24 2 12 22]
		排序后 m.keys [2 4 6 12 14 16 22 24 26]
	*/
	hash.Add("6", "4", "2")

	testCases := map[string]string{
		"2":  "2",
		"11": "2",
		"23": "4", // 下一个位置应该是在24这里，所以对应的是4
		"27": "2", // 超过了长度，环形 取模后从头开始算起
	}

	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("Asking for %s, should have yielded %s", k, v)
		}
	}

	// add 8, 18, 28
	hash.Add("8")

	// 27 对应的是 8
	testCases["27"] = "8"

	for k, v := range testCases {
		if hash.Get(k) != v {
			t.Errorf("Asking for %s, should have yielded %s", k, v)
		}
	}
}
