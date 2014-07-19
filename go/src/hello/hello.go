package main

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"sort"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

type Message struct {
	Message string `json:"message"`
}

type FlavorDetails struct {
	Id         string  `json:"id"`
	Name       string  `json:"name"`
	Ram        uint64  `json:"ram"`
	VCPUs      uint64  `json:"vcpus"`
	Swap       uint64  `json:"swap"`
	RxTxFactor float64 `json:"rxtx_factor"`
	Ephemeral  uint64  `json:"OS-FLV-EXT-DATA:ephemeral"`
	Disk       uint64  `json:"disk"`
	Links      []struct {
		Href string `json:"href"`
		Rel  string `json:"rel"`
	} `json:"links"`
	ExtraSpecs map[string]interface{} `json:"OS-FLV-WITH-EXT-SPECS:extra_specs"`
}

type Flavors struct {
	Flavors []FlavorDetails `json:"flavors"`
}

type World struct {
	Id           uint16 `json:"id"`
	RandomNumber uint16 `json:"randomNumber"`
}

type Fortune struct {
	Id      uint16 `json:"id"`
	Message string `json:"message"`
}

const (
	// Database
	connectionString   = "benchmarkdbuser:benchmarkdbpass@tcp(localhost:3306)/hello_world"
	worldSelect        = "SELECT id, randomNumber FROM World WHERE id = ?"
	worldUpdate        = "UPDATE World SET randomNumber = ? WHERE id = ?"
	fortuneSelect      = "SELECT id, message FROM Fortune;"
	worldRowCount      = 10000
	maxConnectionCount = 256

	helloWorldString = "Hello, World!"
	json1kString     = "{\"flavors\":[{\"id\":\"1\",\"links\":[{\"href\":\"http://openstack.example.com/v2/openstack/flavors/1\",\"rel\":\"self\"},{\"href\":\"http://openstack.example.com/openstack/flavors/1\",\"rel\":\"bookmark\"}],\"name\":\"m1.tiny\"},{\"id\":\"2\",\"links\":[{\"href\":\"http://openstack.example.com/v2/openstack/flavors/2\",\"rel\":\"self\"},{\"href\":\"http://openstack.example.com/openstack/flavors/2\",\"rel\":\"bookmark\"}],\"name\":\"m1.small\"},{\"id\":\"3\",\"links\":[{\"href\":\"http://openstack.example.com/v2/openstack/flavors/3\",\"rel\":\"self\"},{\"href\":\"http://openstack.example.com/openstack/flavors/3\",\"rel\":\"bookmark\"}],\"name\":\"m1.medium\"},{\"id\":\"4\",\"links\":[{\"href\":\"http://openstack.example.com/v2/openstack/flavors/4\",\"rel\":\"self\"},{\"href\":\"http://openstack.example.com/openstack/flavors/4\",\"rel\":\"bookmark\"}],\"name\":\"m1.large\"},{\"id\":\"5\",\"links\":[{\"href\":\"http://openstack.example.com/v2/openstack/flavors/5\",\"rel\":\"self\"},{\"href\":\"http://openstack.example.com/openstack/flavors/5\",\"rel\":\"bookmark\"}],\"name\":\"m1.xlarge\"}]}"
	json10kString    = "{\"flavors\":[{\"OS-FLV-WITH-EXT-SPECS:extra_specs\":{\"class\":\"standard1\",\"disk_io_index\":\"2\",\"number_of_data_disks\":\"0\"},\"name\":\"512MB Standard Instance\",\"links\":[{\"href\":\"https://iad.servers.api.rackspacecloud.com/v2/728975/flavors/2\",\"rel\":\"self\"},{\"href\":\"https://iad.servers.api.rackspacecloud.com/728975/flavors/2\",\"rel\":\"bookmark\"}],\"ram\":512,\"vcpus\":1,\"swap\":512,\"rxtx_factor\":80.0,\"OS-FLV-EXT-DATA:ephemeral\":0,\"disk\":20,\"id\":\"2\"},{\"OS-FLV-WITH-EXT-SPECS:extra_specs\":{\"class\":\"standard1\",\"disk_io_index\":\"2\",\"number_of_data_disks\":\"0\"},\"name\":\"1GB Standard Instance\",\"links\":[{\"href\":\"https://iad.servers.api.rackspacecloud.com/v2/728975/flavors/3\",\"rel\":\"self\"},{\"href\":\"https://iad.servers.api.rackspacecloud.com/728975/flavors/3\",\"rel\":\"bookmark\"}],\"ram\":1024,\"vcpus\":1,\"swap\":1024,\"rxtx_factor\":120.0,\"OS-FLV-EXT-DATA:ephemeral\":0,\"disk\":40,\"id\":\"3\"},{\"OS-FLV-WITH-EXT-SPECS:extra_specs\":{\"class\":\"standard1\",\"disk_io_index\":\"2\",\"number_of_data_disks\":\"0\"},\"name\":\"2GB Standard Instance\",\"links\":[{\"href\":\"https://iad.servers.api.rackspacecloud.com/v2/728975/flavors/4\",\"rel\":\"self\"},{\"href\":\"https://iad.servers.api.rackspacecloud.com/728975/flavors/4\",\"rel\":\"bookmark\"}],\"ram\":2048,\"vcpus\":2,\"swap\":2048,\"rxtx_factor\":240.0,\"OS-FLV-EXT-DATA:ephemeral\":0,\"disk\":80,\"id\":\"4\"},{\"OS-FLV-WITH-EXT-SPECS:extra_specs\":{\"class\":\"standard1\",\"disk_io_index\":\"2\",\"number_of_data_disks\":\"0\"},\"name\":\"4GB Standard Instance\",\"links\":[{\"href\":\"https://iad.servers.api.rackspacecloud.com/v2/728975/flavors/5\",\"rel\":\"self\"},{\"href\":\"https://iad.servers.api.rackspacecloud.com/728975/flavors/5\",\"rel\":\"bookmark\"}],\"ram\":4096,\"vcpus\":2,\"swap\":2048,\"rxtx_factor\":400.0,\"OS-FLV-EXT-DATA:ephemeral\":0,\"disk\":160,\"id\":\"5\"},{\"OS-FLV-WITH-EXT-SPECS:extra_specs\":{\"class\":\"standard1\",\"disk_io_index\":\"2\",\"number_of_data_disks\":\"0\"},\"name\":\"8GB Standard Instance\",\"links\":[{\"href\":\"https://iad.servers.api.rackspacecloud.com/v2/728975/flavors/6\",\"rel\":\"self\"},{\"href\":\"https://iad.servers.api.rackspacecloud.com/728975/flavors/6\",\"rel\":\"bookmark\"}],\"ram\":8192,\"vcpus\":4,\"swap\":2048,\"rxtx_factor\":600.0,\"OS-FLV-EXT-DATA:ephemeral\":0,\"disk\":320,\"id\":\"6\"},{\"OS-FLV-WITH-EXT-SPECS:extra_specs\":{\"class\":\"standard1\",\"disk_io_index\":\"2\",\"number_of_data_disks\":\"0\"},\"name\":\"15GB Standard Instance\",\"links\":[{\"href\":\"https://iad.servers.api.rackspacecloud.com/v2/728975/flavors/7\",\"rel\":\"self\"},{\"href\":\"https://iad.servers.api.rackspacecloud.com/728975/flavors/7\",\"rel\":\"bookmark\"}],\"ram\":15360,\"vcpus\":6,\"swap\":2048,\"rxtx_factor\":800.0,\"OS-FLV-EXT-DATA:ephemeral\":0,\"disk\":620,\"id\":\"7\"},{\"OS-FLV-WITH-EXT-SPECS:extra_specs\":{\"class\":\"standard1\",\"disk_io_index\":\"2\",\"number_of_data_disks\":\"0\"},\"name\":\"30GB Standard Instance\",\"links\":[{\"href\":\"https://iad.servers.api.rackspacecloud.com/v2/728975/flavors/8\",\"rel\":\"self\"},{\"href\":\"https://iad.servers.api.rackspacecloud.com/728975/flavors/8\",\"rel\":\"bookmark\"}],\"ram\":30720,\"vcpus\":8,\"swap\":2048,\"rxtx_factor\":1200.0,\"OS-FLV-EXT-DATA:ephemeral\":0,\"disk\":1200,\"id\":\"8\"},{\"OS-FLV-WITH-EXT-SPECS:extra_specs\":{\"quota_resources\":\"instances=onmetal-compute-v1-instances,ram=onmetal-compute-v1-ram\",\"class\":\"onmetal\",\"policy_class\":\"onmetal_flavor\"},\"name\":\"OnMetal Compute v1\",\"links\":[{\"href\":\"https://iad.servers.api.rackspacecloud.com/v2/728975/flavors/onmetal-compute1\",\"rel\":\"self\"},{\"href\":\"https://iad.servers.api.rackspacecloud.com/728975/flavors/onmetal-compute1\",\"rel\":\"bookmark\"}],\"ram\":32768,\"vcpus\":20,\"swap\":\"\",\"rxtx_factor\":20000.0,\"OS-FLV-EXT-DATA:ephemeral\":0,\"disk\":32,\"id\":\"onmetal-compute1\"},{\"OS-FLV-WITH-EXT-SPECS:extra_specs\":{\"quota_resources\":\"instances=onmetal-io-v1-instances,ram=onmetal-io-v1-ram\",\"class\":\"onmetal\",\"policy_class\":\"onmetal_flavor\"},\"name\":\"OnMetal I/O v1\",\"links\":[{\"href\":\"https://iad.servers.api.rackspacecloud.com/v2/728975/flavors/onmetal-io1\",\"rel\":\"self\"},{\"href\":\"https://iad.servers.api.rackspacecloud.com/728975/flavors/onmetal-io1\",\"rel\":\"bookmark\"}],\"ram\":131072,\"vcpus\":40,\"swap\":\"\",\"rxtx_factor\":20000.0,\"OS-FLV-EXT-DATA:ephemeral\":3200,\"disk\":32,\"id\":\"onmetal-io1\"},{\"OS-FLV-WITH-EXT-SPECS:extra_specs\":{\"quota_resources\":\"instances=onmetal-memory-v1-instances,ram=onmetal-memory-v1-ram\",\"class\":\"onmetal\",\"policy_class\":\"onmetal_flavor\"},\"name\":\"OnMetal Memory v1\",\"links\":[{\"href\":\"https://iad.servers.api.rackspacecloud.com/v2/728975/flavors/onmetal-memory1\",\"rel\":\"self\"},{\"href\":\"https://iad.servers.api.rackspacecloud.com/728975/flavors/onmetal-memory1\",\"rel\":\"bookmark\"}],\"ram\":524288,\"vcpus\":24,\"swap\":\"\",\"rxtx_factor\":20000.0,\"OS-FLV-EXT-DATA:ephemeral\":0,\"disk\":32,\"id\":\"onmetal-memory1\"},{\"OS-FLV-WITH-EXT-SPECS:extra_specs\":{\"resize_policy_class\":\"performance_flavor\",\"class\":\"performance1\",\"disk_io_index\":\"40\",\"number_of_data_disks\":\"0\"},\"name\":\"1 GB Performance\",\"links\":[{\"href\":\"https://iad.servers.api.rackspacecloud.com/v2/728975/flavors/performance1-1\",\"rel\":\"self\"},{\"href\":\"https://iad.servers.api.rackspacecloud.com/728975/flavors/performance1-1\",\"rel\":\"bookmark\"}],\"ram\":1024,\"vcpus\":1,\"swap\":\"\",\"rxtx_factor\":200.0,\"OS-FLV-EXT-DATA:ephemeral\":0,\"disk\":20,\"id\":\"performance1-1\"},{\"OS-FLV-WITH-EXT-SPECS:extra_specs\":{\"resize_policy_class\":\"performance_flavor\",\"class\":\"performance1\",\"disk_io_index\":\"40\",\"number_of_data_disks\":\"1\"},\"name\":\"2 GB Performance\",\"links\":[{\"href\":\"https://iad.servers.api.rackspacecloud.com/v2/728975/flavors/performance1-2\",\"rel\":\"self\"},{\"href\":\"https://iad.servers.api.rackspacecloud.com/728975/flavors/performance1-2\",\"rel\":\"bookmark\"}],\"ram\":2048,\"vcpus\":2,\"swap\":\"\",\"rxtx_factor\":400.0,\"OS-FLV-EXT-DATA:ephemeral\":20,\"disk\":40,\"id\":\"performance1-2\"},{\"OS-FLV-WITH-EXT-SPECS:extra_specs\":{\"resize_policy_class\":\"performance_flavor\",\"class\":\"performance1\",\"disk_io_index\":\"40\",\"number_of_data_disks\":\"1\"},\"name\":\"4 GB Performance\",\"links\":[{\"href\":\"https://iad.servers.api.rackspacecloud.com/v2/728975/flavors/performance1-4\",\"rel\":\"self\"},{\"href\":\"https://iad.servers.api.rackspacecloud.com/728975/flavors/performance1-4\",\"rel\":\"bookmark\"}],\"ram\":4096,\"vcpus\":4,\"swap\":\"\",\"rxtx_factor\":800.0,\"OS-FLV-EXT-DATA:ephemeral\":40,\"disk\":40,\"id\":\"performance1-4\"},{\"OS-FLV-WITH-EXT-SPECS:extra_specs\":{\"resize_policy_class\":\"performance_flavor\",\"class\":\"performance1\",\"disk_io_index\":\"40\",\"number_of_data_disks\":\"1\"},\"name\":\"8 GB Performance\",\"links\":[{\"href\":\"https://iad.servers.api.rackspacecloud.com/v2/728975/flavors/performance1-8\",\"rel\":\"self\"},{\"href\":\"https://iad.servers.api.rackspacecloud.com/728975/flavors/performance1-8\",\"rel\":\"bookmark\"}],\"ram\":8192,\"vcpus\":8,\"swap\":\"\",\"rxtx_factor\":1600.0,\"OS-FLV-EXT-DATA:ephemeral\":80,\"disk\":40,\"id\":\"performance1-8\"},{\"OS-FLV-WITH-EXT-SPECS:extra_specs\":{\"resize_policy_class\":\"performance_flavor\",\"class\":\"performance2\",\"disk_io_index\":\"80\",\"number_of_data_disks\":\"4\"},\"name\":\"120 GB Performance\",\"links\":[{\"href\":\"https://iad.servers.api.rackspacecloud.com/v2/728975/flavors/performance2-120\",\"rel\":\"self\"},{\"href\":\"https://iad.servers.api.rackspacecloud.com/728975/flavors/performance2-120\",\"rel\":\"bookmark\"}],\"ram\":122880,\"vcpus\":32,\"swap\":\"\",\"rxtx_factor\":10000.0,\"OS-FLV-EXT-DATA:ephemeral\":1200,\"disk\":40,\"id\":\"performance2-120\"},{\"OS-FLV-WITH-EXT-SPECS:extra_specs\":{\"resize_policy_class\":\"performance_flavor\",\"class\":\"performance2\",\"disk_io_index\":\"40\",\"number_of_data_disks\":\"1\"},\"name\":\"15 GB Performance\",\"links\":[{\"href\":\"https://iad.servers.api.rackspacecloud.com/v2/728975/flavors/performance2-15\",\"rel\":\"self\"},{\"href\":\"https://iad.servers.api.rackspacecloud.com/728975/flavors/performance2-15\",\"rel\":\"bookmark\"}],\"ram\":15360,\"vcpus\":4,\"swap\":\"\",\"rxtx_factor\":1250.0,\"OS-FLV-EXT-DATA:ephemeral\":150,\"disk\":40,\"id\":\"performance2-15\"},{\"OS-FLV-WITH-EXT-SPECS:extra_specs\":{\"resize_policy_class\":\"performance_flavor\",\"class\":\"performance2\",\"disk_io_index\":\"40\",\"number_of_data_disks\":\"1\"},\"name\":\"30 GB Performance\",\"links\":[{\"href\":\"https://iad.servers.api.rackspacecloud.com/v2/728975/flavors/performance2-30\",\"rel\":\"self\"},{\"href\":\"https://iad.servers.api.rackspacecloud.com/728975/flavors/performance2-30\",\"rel\":\"bookmark\"}],\"ram\":30720,\"vcpus\":8,\"swap\":\"\",\"rxtx_factor\":2500.0,\"OS-FLV-EXT-DATA:ephemeral\":300,\"disk\":40,\"id\":\"performance2-30\"},{\"OS-FLV-WITH-EXT-SPECS:extra_specs\":{\"resize_policy_class\":\"performance_flavor\",\"class\":\"performance2\",\"disk_io_index\":\"60\",\"number_of_data_disks\":\"2\"},\"name\":\"60 GB Performance\",\"links\":[{\"href\":\"https://iad.servers.api.rackspacecloud.com/v2/728975/flavors/performance2-60\",\"rel\":\"self\"},{\"href\":\"https://iad.servers.api.rackspacecloud.com/728975/flavors/performance2-60\",\"rel\":\"bookmark\"}],\"ram\":61440,\"vcpus\":16,\"swap\":\"\",\"rxtx_factor\":5000.0,\"OS-FLV-EXT-DATA:ephemeral\":600,\"disk\":40,\"id\":\"performance2-60\"},{\"OS-FLV-WITH-EXT-SPECS:extra_specs\":{\"resize_policy_class\":\"performance_flavor\",\"class\":\"performance2\",\"disk_io_index\":\"70\",\"number_of_data_disks\":\"3\"},\"name\":\"90 GB Performance\",\"links\":[{\"href\":\"https://iad.servers.api.rackspacecloud.com/v2/728975/flavors/performance2-90\",\"rel\":\"self\"},{\"href\":\"https://iad.servers.api.rackspacecloud.com/728975/flavors/performance2-90\",\"rel\":\"bookmark\"}],\"ram\":92160,\"vcpus\":24,\"swap\":\"\",\"rxtx_factor\":7500.0,\"OS-FLV-EXT-DATA:ephemeral\":900,\"disk\":40,\"id\":\"performance2-90\"}]}"
)

var (
	// Templates
	tmpl = template.Must(template.ParseFiles("templates/layout.html", "templates/fortune.html"))

	// Database
	worldStatement   *sql.Stmt
	fortuneStatement *sql.Stmt
	updateStatement  *sql.Stmt

	helloWorldBytes = []byte(helloWorldString)
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}
	db.SetMaxIdleConns(maxConnectionCount)
	worldStatement, err = db.Prepare(worldSelect)
	if err != nil {
		log.Fatal(err)
	}
	fortuneStatement, err = db.Prepare(fortuneSelect)
	if err != nil {
		log.Fatal(err)
	}
	updateStatement, err = db.Prepare(worldUpdate)
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/db", dbHandler)
	http.HandleFunc("/queries", queriesHandler)
	http.HandleFunc("/json", jsonHandler)
	http.HandleFunc("/fortune", fortuneHandler)
	http.HandleFunc("/update", updateHandler)
	http.HandleFunc("/plaintext", plaintextHandler)
	http.HandleFunc("/json1k", json1kHandler)
	http.HandleFunc("/json10k", json10kHandler)
	http.ListenAndServe(":8080", nil)
}

// Test 1: JSON serialization
func jsonHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript")
	json.NewEncoder(w).Encode(&Message{helloWorldString})
}

// Test 2: Single database query
func dbHandler(w http.ResponseWriter, r *http.Request) {
	var world World
	err := worldStatement.QueryRow(rand.Intn(worldRowCount)+1).Scan(&world.Id, &world.RandomNumber)
	if err != nil {
		log.Fatalf("Error scanning world row: %s", err.Error())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(&world)
}

// Test 3: Multiple database queries
func queriesHandler(w http.ResponseWriter, r *http.Request) {
	n := 1
	if nStr := r.URL.Query().Get("queries"); len(nStr) > 0 {
		n, _ = strconv.Atoi(nStr)
	}

	if n <= 1 {
		dbHandler(w, r)
		return
	}

	world := make([]World, n)
	for i := 0; i < n; i++ {
		err := worldStatement.QueryRow(rand.Intn(worldRowCount)+1).Scan(&world[i].Id, &world[i].RandomNumber)
		if err != nil {
			log.Fatalf("Error scanning world row: %s", err.Error())
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(world)
}

// Test 4: Fortunes
func fortuneHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := fortuneStatement.Query()
	if err != nil {
		log.Fatalf("Error preparing statement: %v", err)
	}

	fortunes := make(Fortunes, 0, 16)
	for rows.Next() { //Fetch rows
		fortune := Fortune{}
		if err := rows.Scan(&fortune.Id, &fortune.Message); err != nil {
			log.Fatalf("Error scanning fortune row: %s", err.Error())
		}
		fortunes = append(fortunes, &fortune)
	}
	fortunes = append(fortunes, &Fortune{Message: "Additional fortune added at request time."})

	sort.Sort(ByMessage{fortunes})
	w.Header().Set("Content-Type", "text/html")
	if err := tmpl.Execute(w, fortunes); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Test 5: Database updates
func updateHandler(w http.ResponseWriter, r *http.Request) {
	n := 1
	if nStr := r.URL.Query().Get("queries"); len(nStr) > 0 {
		n, _ = strconv.Atoi(nStr)
	}

	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	if n <= 1 {
		var world World
		worldStatement.QueryRow(rand.Intn(worldRowCount)+1).Scan(&world.Id, &world.RandomNumber)
		world.RandomNumber = uint16(rand.Intn(worldRowCount) + 1)
		updateStatement.Exec(world.RandomNumber, world.Id)
		encoder.Encode(&world)
	} else {
		world := make([]World, n)
		for i := 0; i < n; i++ {
			if err := worldStatement.QueryRow(rand.Intn(worldRowCount)+1).Scan(&world[i].Id, &world[i].RandomNumber); err != nil {
				log.Fatalf("Error scanning world row: %s", err.Error())
			}
			world[i].RandomNumber = uint16(rand.Intn(worldRowCount) + 1)
			if _, err := updateStatement.Exec(world[i].RandomNumber, world[i].Id); err != nil {
				log.Fatalf("Error updating world row: %s", err.Error())
			}
		}
		encoder.Encode(world)
	}
}

// Test 6: Plaintext
func plaintextHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.Write(helloWorldBytes)
}

// Test 7: JSON 1K size response
func json1kHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	resp := new(Flavors)
	json.Unmarshal([]byte(json1kString), resp)
	json.NewEncoder(w).Encode(resp)
}

// Test 8: JSON 10K size
func json10kHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	resp := new(Flavors)
	json.Unmarshal([]byte(json10kString), resp)
	json.NewEncoder(w).Encode(resp)
}

type Fortunes []*Fortune

func (s Fortunes) Len() int      { return len(s) }
func (s Fortunes) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

type ByMessage struct{ Fortunes }

func (s ByMessage) Less(i, j int) bool { return s.Fortunes[i].Message < s.Fortunes[j].Message }
