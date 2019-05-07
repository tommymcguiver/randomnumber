http://localhost:8080/random?start=10&end=633

```
curl -v 'localhost:8080/random?start=10&end=633'
```

# Go Test

go test -bench .
go test -bench . -benchmem
go test -bench Benchmark_ServeHttp -benchmem
go test -bench Benchmark_ServeHttp -benchmem -cpuprofile prof.cpu | tee bench.0
go test -bench Benchmark_ServeHttp -benchmem -memprofile prof.mem | tee bench.0
benchcmp bench.0 bench.1

go tool pprof randomnumber.test prof.cpu

# Race Detection

```
	atomic.AddInt64(&requestCount, 1)


func TestServer_HandleRandom_Parallel(t *testing.T) {

	var wg sync.WaitGroup
	var testServer = httptest.NewServer(http.HandlerFunc(HandleRandom))

	for i := 0; i < 2; i++ {

		wg.Add(1)
		go func() {
			defer wg.Done()
			res, err := http.Get(testServer.URL + "/random?start=1&end=2000")
			if err != nil {
				t.Error(err)
				return
			}

			if res.StatusCode != 200 {
				t.Error("Status code invalid")
				return
			}
		}()
	}

	//Wait until done
	wg.Wait()
}
```


# Make JSON Encoding Faster

Builder

```
var randomBuilder strings.Builder

func (rn *RandomNumber) MarshalJSON() ([]byte, error) {
	randomBuilder.Reset()
	fmt.Fprintf(&randomBuilder, `{"Number":%d}`, rn.Number)
	return []byte(randomBuilder.String()), nil
}
```


bytes.Buffer
```
func (rn *RandomNumber) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`{"Number":`)
	buffer.WriteString("1}")
	return buffer.Bytes(), nil
}
```

# Why is JSON encoding not faster when i prebuild a string?

# Fixes

Extra Alloc
```
	fmt.Fprintf(w, "%s", string(b))
```

Global RNG

```
	var rng = rand.New(mt19937.New())
```

Query Mapping

```
for k, v := range map[string][]string(r.URL.Query()) {
		switch k {
		case "start":
			if len(v) == 1 {
				start, err = strconv.Atoi(v[0])
			}
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		case "end":
			if len(v) == 1 {
				end, err = strconv.Atoi(v[0])
			}

			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
		}
	}

```

Sync Pool

Is ugly, but for maximum performance ( i.e. 0 allocations at runtime) you can do this.

```
var randomPool = sync.Pool{
	New: func() interface{} {
		return new(RandomNumber)
	},
}

var rng = rand.New(mt19937.New())

func NewRandomNumber() *RandomNumber {
	rn := randomPool.Get().(*RandomNumber)
	defer randomPool.Put(rn)
	rn.Number = 0
	rng.Seed(time.Now().UnixNano())
	rn.Number = rng.Int63()
	return rn
}
```

Contention

```
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
```
# HTTP Profiling

http://localhost:8080/debug/pprof/

```
	import _ "net/http/pprof"
	
	....

	server := http.DefaultServeMux
	server.HandleFunc("/health", HandleHealth)
	server.HandleFunc("/random", HandleRandom)
	log.Println("Starting server...")
	log.Fatal(http.ListenAndServe(":8080", server))
```