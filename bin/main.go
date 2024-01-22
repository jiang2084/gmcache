package main

import (
	"flag"
	"fmt"
	"gmcache"
	"log"
	"net/http"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func createGroup() *gmcache.Group {
	return gmcache.NewGroup("scores", 2<<10, gmcache.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		},
	))
}

func startCacheServer(addr string, addrs []string, gm *gmcache.Group) {
	peers := gmcache.NewHTTPPool(addr)
	peers.Set(addrs...)
	gm.RegisterPeers(peers)
	log.Println("gmcache is running at", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func startAPIServer(apiAddr string, gm *gmcache.Group) {
	http.Handle("/api", http.HandlerFunc(
		func(writer http.ResponseWriter, request *http.Request) {
			key := request.URL.Query().Get("key")
			view, err := gm.Get(key)
			if err != nil {
				http.Error(writer, err.Error(), http.StatusInternalServerError)
				return
			}

			writer.Header().Set("Content-Type", "application/octet-stream")
			writer.Write(view.ByteSlice())
		}))
	log.Println("fontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}

func main() {

	//addr := "localhost:9999"
	//peers := gmcache.NewHTTPPool(addr)
	//log.Println("gmcache is running at", addr)
	//log.Fatal(http.ListenAndServe(addr, peers))

	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "Gmcache server port")
	flag.BoolVar(&api, "api", false, "start a api server")
	flag.Parse()

	apiAddr := "http://localhost:9999"
	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap {
		addrs = append(addrs, v)
	}

	gm := createGroup()
	if api {
		go startAPIServer(apiAddr, gm)
	}
	startCacheServer(addrMap[port], []string(addrs), gm)
}
