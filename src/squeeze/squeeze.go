package squeeze

import (
	"encoding/json"
	"github.com/Offout/go-link-shortener/src/auth"
	"github.com/google/uuid"
	"net/http"
	"sort"
	"strconv"
)

type squeezeForm struct {
	Link string `json:"link"`
}

type squeezedLink struct {
	target   string
	counter  int
	userName string
}

type squeezedLinkResponse struct {
	Short   string `json:"short"`
	Target  string `json:"target"`
	Counter int    `json:"counter"`
}

type squeezeResponse struct {
	Short string `json:"short"`
}

const defaultLimit = 10
const defaultSort = "asc"

// short => SqueezedLink
var squeezedStorage = make(map[string]squeezedLink)

func Squeeze(w http.ResponseWriter, r *http.Request) {
	var userName = auth.CheckSession(r)
	if "" == userName {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var form squeezeForm
	err := json.NewDecoder(r.Body).Decode(&form)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var shortLink = generateUniqueShort()
	squeezedStorage[shortLink] = squeezedLink{form.Link, 0, userName}
	err = json.NewEncoder(w).Encode(squeezeResponse{shortLink})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func generateUniqueShort() string {
	var short = ""
	for {
		id := uuid.New()
		short = id.String()[0:10]
		var _, ok = squeezedStorage[short]
		if !ok {
			break
		}
	}
	return short
}

func Redirect(w http.ResponseWriter, r *http.Request) {
	var short = r.URL.Path[3:]
	var squeezed, ok = squeezedStorage[short]
	if !ok {
		http.Error(w, "No such short link", http.StatusBadRequest)
		return
	}
	squeezed.counter++
	squeezedStorage[short] = squeezed
	http.Redirect(w, r, squeezed.target, http.StatusTemporaryRedirect)
}

func Statistics(w http.ResponseWriter, r *http.Request) {
	var userName = auth.CheckSession(r)
	if "" == userName {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	var offset, limit = r.URL.Query().Get("offset"), r.URL.Query().Get("limit")
	var offsetInt int
	var err error
	if "" == offset {
		offsetInt = 0
	} else {
		offsetInt, err = strconv.Atoi(offset)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	var limitInt int
	limitInt, err = strconv.Atoi(limit)
	if "" == limit {
		limitInt = defaultLimit
	} else {
		limitInt, err = strconv.Atoi(limit)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	var sorting = r.URL.Query().Get("sort")
	if "" == sorting {
		sorting = defaultSort
	}

	var squeezed []squeezedLinkResponse

	for short, element := range squeezedStorage {
		if userName == element.userName {
			squeezed = append(squeezed, squeezedLinkResponse{short, element.target, element.counter})
		}
	}
	sort.SliceStable(squeezed, func(i, j int) bool {
		return squeezed[i].Short > squeezed[j].Short
	})

	if "asc" == sorting {
		sort.SliceStable(squeezed, func(i, j int) bool {
			return squeezed[i].Counter < squeezed[j].Counter
		})
	} else if "desc" == sorting {
		sort.SliceStable(squeezed, func(i, j int) bool {
			return squeezed[i].Counter > squeezed[j].Counter
		})
	}
	if offsetInt > len(squeezed) {
		offsetInt = len(squeezed)
	}

	if limitInt+offsetInt >= len(squeezed) {
		limitInt = len(squeezed) - offsetInt
	}

	squeezed = squeezed[offsetInt : limitInt+offsetInt]

	err = json.NewEncoder(w).Encode(squeezed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
