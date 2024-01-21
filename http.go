package gmcache

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

const defaultBasePath = "/gmcache"

type HTTPPool struct {
	self     string // 用来记录自己的地址，包括主机名和端口
	bashPath string // 节点间的访问
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		bashPath: defaultBasePath,
	}
}

func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.bashPath) {
		panic("HTTPPool seving unexpected path:" + r.URL.Path)
	}

	p.Log("%s %s", r.Method, r.URL.Path)
	// 格式：/<basepath>/<groupname>/<key>
	parts := strings.Split(r.URL.Path[len(p.bashPath):], "/")
	if len(parts) != 3 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[1]
	key := parts[2]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "no such group:"+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}
