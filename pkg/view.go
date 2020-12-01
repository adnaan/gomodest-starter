package pkg

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"unicode"

	"github.com/Masterminds/sprig"
	"github.com/foolin/goview"

	"github.com/go-chi/chi"
)

func First(str string) string {
	if len(str) == 0 {
		return ""
	}
	tmp := []rune(str)
	tmp[0] = unicode.ToUpper(tmp[0])
	return string(tmp)
}

func viewEngine(baseTemplate string) (*goview.ViewEngine, error) {

	fileInfo, err := ioutil.ReadDir("web/html/partials")
	if err != nil {
		return nil, err
	}
	var partials []string
	for _, file := range fileInfo {
		if !strings.HasSuffix(file.Name(), ".html") {
			continue
		}
		partials = append(partials, fmt.Sprintf("partials/%s", strings.TrimSuffix(file.Name(), ".html")))
	}

	return goview.New(goview.Config{
		Root:         "web/html",
		Extension:    ".html",
		Master:       fmt.Sprintf("layouts/%s", baseTemplate),
		Partials:     partials,
		DisableCache: true,
		Funcs:        sprig.FuncMap(), // http://masterminds.github.io/sprig/
	}), nil
}

type PageHandlerFunc func(appCtx AppContext, w http.ResponseWriter, r *http.Request) (goview.M, error)

func simplePage(appCtx AppContext, w http.ResponseWriter, r *http.Request) (goview.M, error) {
	return goview.M{}, nil
}

func newRenderer(appCtx AppContext) func(page string, pageHandlerFuncs ...PageHandlerFunc) http.HandlerFunc {
	return func(page string, pageHandlerFuncs ...PageHandlerFunc) http.HandlerFunc {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// merge app context data set by isAuthenticated Middleware with the passed page data
			pageData := make(map[string]interface{})
			// default page data set by setPageData middleware
			appCtxData, ok := r.Context().Value(appCtxDataKey).(map[string]interface{})
			if ok {
				for k, v := range appCtxData {
					pageData[k] = v
				}
			}
			// set default page renderer
			if len(pageHandlerFuncs) == 0 {
				pageHandlerFuncs = append(pageHandlerFuncs, simplePage)
			}

			for _, pageHandlerFunc := range pageHandlerFuncs {
				appCtx.pageData = pageData
				pageHandlerData, err := pageHandlerFunc(appCtx, w, r)
				if err != nil {
					fmt.Println(err)
					userError := errors.Unwrap(err)
					if userError != nil {
						pageData["userError"] = First(strings.ToLower(userError.Error()))
					} else {
						pageData["userError"] = "Internal Error"
					}
				}
				// set returned page data from the handler to the main pageData map
				for k, v := range pageHandlerData {
					pageData[k] = v
				}
			}

			err := appCtx.viewEngine.Render(w, http.StatusOK, page, pageData)
			if err != nil {
				fmt.Println(err)
				fmt.Fprintf(w, "umm...awkward.")
				return
			}
		})
	}
}

// fileServer conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func fileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}
