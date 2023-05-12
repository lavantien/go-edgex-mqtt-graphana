package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"thirdparty_go/common"
)

type Data struct {
	Temperature float64 `json:"temperature,omitempty"`
	Humidity    float64 `json:"humidity,omitempty"`
}

type DataRepository []Data

type dataHandler struct {
	sync.Mutex
	repository DataRepository
}

func (ph *dataHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Authorization") != "Bearer "+common.ThirdPartyJwtToken {
		log.Println("Authorization:", r.Header.Get("Authorization"))
		log.Println("Bearer " + common.ThirdPartyJwtToken)
		respondError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	log.Println("endpoint hit from", r.Header.Get("X-Forwarded-For"))
	switch r.Method {
	case "GET":
		ph.get(w, r)
	case "POST":
		ph.post(w, r)
	case "PUT":
		ph.put(w, r)
	case "PATCH":
		ph.patch(w, r)
	case "DELETE":
		ph.delete(w, r)
	default:
		respondError(w, http.StatusMethodNotAllowed, "invalid method")
	}
}

func (ph *dataHandler) get(w http.ResponseWriter, r *http.Request) {
	defer ph.Unlock()
	ph.Lock()
	id, err := idFromUrl(r)
	if err != nil {
		respondJson(w, http.StatusOK, ph.repository)
		return
	}
	if id > len(ph.repository) || id < 0 {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	respondJson(w, http.StatusOK, ph.repository[id])
}

func (ph *dataHandler) post(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		log.Println("client", r.Header.Get("X-Forwarded-For"), "error:", err.Error())
		return
	}
	ct := r.Header.Get("content-type")
	if ct != common.ContentType {
		respondError(w, http.StatusUnsupportedMediaType, "content type "+"'"+common.ContentType+"'"+" required")
		return
	}
	var data Data
	err = json.Unmarshal(body, &data)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		log.Println("client", r.Header.Get("X-Forwarded-For"), "error:", err.Error())
		return
	}
	defer ph.Unlock()
	ph.Lock()
	ph.repository = append(ph.repository, data)
	respondJson(w, http.StatusCreated, data)
	log.Println("POST hit from", r.Header.Get("X-Forwarded-For"), "with data:\n\t", data)
}

func (ph *dataHandler) put(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	id, err := idFromUrl(r)
	if err != nil {
		respondError(w, http.StatusNotFound, err.Error())
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	ct := r.Header.Get("content-type")
	if ct != common.ContentType {
		respondError(w, http.StatusUnsupportedMediaType, "content type "+"'"+common.ContentType+"'"+" required")
		return
	}
	var data Data
	err = json.Unmarshal(body, &data)
	if err != nil {
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}
	defer ph.Unlock()
	ph.Lock()
	if id > len(ph.repository) || id < 0 {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	if data.Temperature != 0.0 {
		ph.repository[id].Temperature = data.Temperature
	}
	if data.Humidity != 0.0 {
		ph.repository[id].Humidity = data.Humidity
	}
	respondJson(w, http.StatusOK, ph.repository[id])
}

func (ph *dataHandler) patch(w http.ResponseWriter, r *http.Request) {
	ph.put(w, r)
}

func (ph *dataHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := idFromUrl(r)
	if err != nil {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	defer ph.Unlock()
	ph.Lock()
	if id > len(ph.repository) || id < 0 {
		respondError(w, http.StatusNotFound, "not found")
		return
	}
	if id < len(ph.repository)-1 {
		ph.repository[len(ph.repository)-1], ph.repository[id] = ph.repository[id], ph.repository[len(ph.repository)-1]
	}
	ph.repository = ph.repository[:len(ph.repository)-1]
	respondJson(w, http.StatusNoContent, "")
}

func respondError(w http.ResponseWriter, code int, msg string) {
	respondJson(w, code, map[string]string{"error": msg})
}

func respondJson(w http.ResponseWriter, code int, data interface{}) {
	response, err := json.Marshal(data)
	handleError(err)
	w.Header().Add("content-type", common.ContentType)
	w.WriteHeader(code)
	w.Write(response)
}

func handleError(err error) {
	if err != nil {
		log.Panicln(err)
	}
}

func idFromUrl(r *http.Request) (int, error) {
	parts := strings.Split(r.URL.String(), "/")
	if len(parts) != 3 {
		return -1, errors.New("not found")
	}
	id, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		return -1, errors.New("not found")
	}
	return id, nil
}

func newProductsHandler() *dataHandler {
	return &dataHandler{
		repository: DataRepository{
			Data{25.00, 40.00},
			Data{25.15, 41.00},
			Data{25.50, 41.20},
		},
	}
}

func ListenHttpRest(ctx context.Context) {
	ph := newProductsHandler()
	http.Handle("/data", ph)
	http.Handle("/data/", ph)
	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(rw, "permission denied")
	})
	log.Println("listenning on", common.ThirdPartyEndpoint, "...")
	log.Panicln(http.ListenAndServe(common.ThirdPartyEndpoint, nil))
}
