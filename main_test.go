package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http/httptest"
	"testing"
)

func Benchmark_ServeHttp(b *testing.B) {
	req := httptest.NewRequest("GET", "http://example.com/random?start=10&end=20000", nil)
	w := httptest.NewRecorder()

	for n := 0; n < b.N; n++ {
		s := &Server{}
		s.ServeHTTP(w, req)
		if w.Result().StatusCode != 200 {
			b.Error("Invalid status code")
		}
	}
}

func reset(rw *httptest.ResponseRecorder) {
	m := rw.HeaderMap
	for k := range m {
		delete(m, k)
	}
	body := rw.Body
	body.Reset()
	*rw = httptest.ResponseRecorder{
		Body:      body,
		HeaderMap: m,
	}
}

func BenchmarkHealthParallel(b *testing.B) {
	r := httptest.NewRequest("GET", "http://example.com/health", nil)

	b.RunParallel(func(pb *testing.PB) {
		rw := httptest.NewRecorder()
		for pb.Next() {
			HandleHealth(rw, r)
			reset(rw)
		}
	})
}

func Benchmark_Random(b *testing.B) {
	for n := 0; n < b.N; n++ {

		req := httptest.NewRequest("GET", "http://example.com/random?start=10&end=20", nil)
		w := httptest.NewRecorder()
		HandleRandom(w, req)

		resp := w.Result()
		body, _ := ioutil.ReadAll(resp.Body)
		random := &RandomNumber{}

		if json.Unmarshal(body, random) != nil {
			b.Error("Failed to unmarshal")
		}

		if random.Number < 10 {
			b.Error("Random to low")
		}

		if random.Number > 20 {
			b.Error("Random to high")
		}
	}
}
