package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	t "github.com/jjg-akers/socketses/internal/types"
	"github.com/patrickmn/go-cache"
)

type SetHandler struct {
	PermCh chan t.Permission
	OkChan chan string
	Cache  *cache.Cache
}

func (c *SetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()

	request := t.SetRequest{}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&request)
	if err != nil {
		log.Println("Unable to decode request")
		w.WriteHeader(http.StatusBadRequest)
		// body, _ := ioutil.ReadAll(req.Body)
		// clog.Debug(ctx, "Error decoding json reqeust", zap.Error(err), zap.String("request", string(body)))
		return
	}
	// ask for permission
	//fmt.Println("start set handler")
	//time.Sleep(time.Second * 1)
	cacheKey := fmt.Sprintf("localCache_%d_%s", request.App, request.Device)

	//check cache
	identMap := make(map[string]string)

	localVal, found := c.Cache.Get(cacheKey)
	if found {
		switch idents := localVal.(type) {
		case map[string]string:
			identMap = idents
		default:
			log.Println("could not convert ident type")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// check for current value
	if v, ok := identMap[request.Name]; ok {
		if v == request.Value {
			w.Header().Set("Action", "Checked")
			return
		}
	}

	// not found locally
	// Get permission to continue

	permissionKey := fmt.Sprintf("%s-%s", request.Device, request.Name)

	c.PermCh <- t.Permission{
		Key: t.Key{
			Type: "p",
			Key:  permissionKey,
		},
	}

	// fmt.Println("waiting for ok")
	ok := <-c.OkChan
	fmt.Println("got ok: ", ok)

	fmt.Println("getting from DB")
	identMap = getFromDB()
	// check for current value
	if v, ok := identMap[request.Name]; ok {
		if v == request.Value {
			fmt.Println("found in results from db")
			w.Header().Set("Action", "Checked")
			return
		}
	}

	// not in db need to insert or updated
	identMap[request.Name] = request.Value

	//send update and return results
	c.Cache.Set(cacheKey, identMap, 0)

	// fmt.Println("doing some work")
	//time.Sleep(time.Second * 5)

	fmt.Println("done with work")
	c.PermCh <- t.Permission{
		Key: t.Key{
			Type: "d",
			Key:  permissionKey,
		},
		CacheUpdate: t.CacheUpdate{
			Key:   cacheKey,
			Value: identMap,
		},
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(identMap)
	fmt.Println("done time: ", time.Since(start))

	return

	// fmt.Println("done time: ", time.Since(start))
	// fmt.Fprintln(w, "GOT IT")
}

//mock getting idents from db
func getFromDB() map[string]string {

	identMap := make(map[string]string)

	identMap["key1"] = "value1"
	identMap["key2"] = "value2"

	return identMap

}
