package squeeze

import (
	"encoding/json"
	"github.com/Offout/go-link-shortener/src/auth"
	"github.com/google/uuid"
	"net/http"
	"sort"
	"strconv"
	"strings"
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

type sorting struct {
	column string
	order  string
}

const defaultLimit = 10

func getDefaultSort() []sorting {
	return []sorting{{"counter", "desc"}}
}

// short => SqueezedLink
var squeezedStorage = make(map[string]squeezedLink)

func Squeeze(w http.ResponseWriter, r *http.Request) {
	var userName = auth.CheckSession(r)
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
		short = id.String()[0:5]
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

	var sortingQuery = r.URL.Query()["sort"]

	var squeezed []squeezedLinkResponse

	for short, element := range squeezedStorage {
		if userName == element.userName {
			squeezed = append(squeezed, squeezedLinkResponse{short, element.target, element.counter})
		}
	}

	var sortingParsed []sorting
	if 0 == len(sortingQuery) {
		sortingParsed = getDefaultSort()
	} else {
		for _, element := range sortingQuery {
			parts := strings.Split(element, "_")
			sortingParsed = append(sortingParsed, sorting{parts[0], parts[1]})
		}
	}

	sort.SliceStable(squeezed, func(i, j int) bool {
		switch sortingParsed[0].column {
		case "short":
			return sortByShortSmall(squeezed, i, j, sortingParsed[0].order, sortingParsed[1:])
		case "target":
			return sortByTargetSmall(squeezed, i, j, sortingParsed[0].order, sortingParsed[1:])
		case "counter":
			return sortByCounterSmall(squeezed, i, j, sortingParsed[0].order, sortingParsed[1:])
		}
		return true
	})

	if offsetInt > len(squeezed) {
		offsetInt = len(squeezed)
	}

	if limitInt+offsetInt >= len(squeezed) {
		limitInt = len(squeezed) - offsetInt
	}

	squeezed = squeezed[offsetInt : limitInt+offsetInt]

	if 0 == len(squeezed) {
		squeezed = []squeezedLinkResponse{}
	}

	err = json.NewEncoder(w).Encode(squeezed)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func sortByShortSmall(arr []squeezedLinkResponse, i int, j int, order string, sortParams []sorting) bool {
	if arr[i].Short == arr[j].Short && 0 != len(sortParams) {
		switch sortParams[0].column {
		case "counter":
			return sortByCounterSmall(arr, i, j, sortParams[0].order, sortParams[1:])
		case "target":
			return sortByTargetSmall(arr, i, j, sortParams[0].order, sortParams[1:])
		}
	}
	if "asc" == order {
		return arr[i].Short < arr[j].Short
	} else {
		return arr[i].Short > arr[j].Short
	}
}

func sortByTargetSmall(arr []squeezedLinkResponse, i int, j int, order string, sortParams []sorting) bool {
	if arr[i].Target == arr[j].Target && 0 != len(sortParams) {
		switch sortParams[0].column {
		case "counter":
			return sortByCounterSmall(arr, i, j, sortParams[0].order, sortParams[1:])
		case "short":
			return sortByShortSmall(arr, i, j, sortParams[0].order, sortParams[1:])
		}
	}
	if "asc" == order {
		return arr[i].Target < arr[j].Target
	} else {
		return arr[i].Target > arr[j].Target
	}
}

func sortByCounterSmall(arr []squeezedLinkResponse, i int, j int, order string, sortParams []sorting) bool {
	if arr[i].Counter == arr[j].Counter && 0 != len(sortParams) {
		switch sortParams[0].column {
		case "target":
			return sortByTargetSmall(arr, i, j, sortParams[0].order, sortParams[1:])
		case "short":
			return sortByShortSmall(arr, i, j, sortParams[0].order, sortParams[1:])
		}
	}
	if "asc" == order {
		return arr[i].Counter < arr[j].Counter
	} else {
		return arr[i].Counter > arr[j].Counter
	}
}
