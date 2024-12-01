package proxy

import (
    "fmt"
    "io"
    "log"
    "net/http"
    "net/url"
    "time"
)

func StartProxy(port string, target string) {
    proxyURL, err := url.Parse(target)
    if err != nil {
        log.Fatalf("URL inválida: %v", err)
    }

    requests := make(chan *http.Request)

    go func() {
        ticker := time.NewTicker(time.Second / 2)
        defer ticker.Stop()
        for req := range requests {
            <-ticker.C
            processRequest(req, proxyURL)
        }
    }()

    mux := http.NewServeMux()
    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        requests <- r
        fmt.Printf("Recebido request: %s %s\n", r.Method, r.URL)

        proxyReq, err := http.NewRequest(r.Method, proxyURL.ResolveReference(r.URL).String(), r.Body)
        if err != nil {
            http.Error(w, "Erro ao criar request", http.StatusInternalServerError)
            return
        }

        proxyReq.Header = r.Header

        client := &http.Client{}
        resp, err := client.Do(proxyReq)
        if err != nil {
            http.Error(w, "Erro ao fazer request para o destino", http.StatusInternalServerError)
            return
        }
        defer resp.Body.Close()

        for key, values := range resp.Header {
            for _, value := range values {
                w.Header().Add(key, value)
            }
        }
        w.WriteHeader(resp.StatusCode)
        io.Copy(w, resp.Body)
    })

    log.Printf("Proxy escutando na porta %s", port)
    log.Fatal(http.ListenAndServe(port, mux))
}

func processRequest(r *http.Request, proxyURL *url.URL) {
    proxyReq, err := http.NewRequest(r.Method, proxyURL.ResolveReference(r.URL).String(), r.Body)
    if err != nil {
        log.Printf("Erro ao criar request: %v", err)
        return
    }

    proxyReq.Header = r.Header

    client := &http.Client{}
    resp, err := client.Do(proxyReq)
    if err != nil {
        log.Printf("Erro ao fazer request para o destino: %v", err)
        return
    }
    defer resp.Body.Close()

    for key, values := range resp.Header {
        for _, value := range values {
            r.Header.Add(key, value)
        }
    }
    r.Body = resp.Body
}