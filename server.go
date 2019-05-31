package main

import (
  "net/http"
  "fmt"
  "time"
  "encoding/json"
  "golang.org/x/net/context/ctxhttp"
  "go.elastic.co/apm" // missing from docs
  "go.elastic.co/apm/module/apmhttp"
)

type user struct {
  Name string
}

var (
  port = ":9020"
  client = &http.Client{Timeout: 5 * time.Second}
  tracingClient = apmhttp.WrapClient(client)
)

func getUser() http.HandlerFunc {
  return func(w http.ResponseWriter, r *http.Request){
    url := "https://jsonplaceholder.typicode.com/users/1"
    res, err := ctxhttp.Get(r.Context(), tracingClient, url)
    if err != nil {
      fmt.Printf("%s", err)
      apm.CaptureError(r.Context(), err).Send() // err arg missing from docs
      http.Error(w, "failed to query json api", 500)
      return
    }
    defer res.Body.Close()
    var u user
    // unsure if the entire body is read
    // this is important to reuse the http client
    if err := json.NewDecoder(res.Body).Decode(&u); err != nil {
      fmt.Printf("%s", err)
      http.Error(w, "failed to parse json", 500)
      return
    }
    fmt.Fprintf(w, "hello %s\n", u.Name)
  }
}

func main() {
  mux := http.NewServeMux()
  mux.Handle("/user", getUser())
  fmt.Printf("server running on port %s\n", port)
  http.ListenAndServe(port, apmhttp.Wrap(mux))
}
