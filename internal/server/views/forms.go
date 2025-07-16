package views

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/schema"

	"github.com/reddec/dreaming-bard/internal/dbo"
)

var getDecoder = sync.OnceValue(func() *schema.Decoder {
	dec := schema.NewDecoder()
	dec.IgnoreUnknownKeys(true)
	return dec
})

func BindForm[T any](r *http.Request) (*T, error) {
	if err := r.ParseForm(); err != nil {
		return nil, err
	}
	var doc T
	if err := getDecoder().Decode(&doc, r.PostForm); err != nil {
		return nil, err
	}
	return &doc, nil
}

func IsHTMX(r *http.Request) bool {
	v, _ := strconv.ParseBool(r.Header.Get("Hx-Request"))
	return v
}

func PrefHandler[T any](pref *dbo.Pref[T], handler func(value string) (T, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		v, err := handler(r.FormValue(pref.Name()))
		if err != nil {
			RenderError(w, err)
			return
		}
		err = pref.Set(r.Context(), v)
		if err != nil {
			RenderError(w, err)
			return
		}
		if IsHTMX(r) {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.Header().Set("Location", ".?"+r.URL.RawQuery)
		w.WriteHeader(http.StatusSeeOther)
	}
}

func BoolHandler(pref *dbo.Pref[bool]) http.HandlerFunc {
	return PrefHandler(pref, func(value string) (bool, error) {
		return strconv.ParseBool(value)
	})
}
