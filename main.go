package main

import (
	"encoding/json"
	"fmt"
	aero "github.com/aerospike/aerospike-client-go"
	"github.com/gorilla/mux"
	"io/ioutil"
	"log"
	"net/http"
	"flag"
)

var client *aero.Client

type sadisObject struct {
	ID string `json:"id"`
}

var (
	aeroIp		string
	aeroPort	int
	aeroNamespace	string
	port		int
)

/*
type SadisDevice struct {
	ID         string `json:"id"`
	HardwareID string `json:"hardwareIdentifier"`
	Uplink     int    `json:"uplinkPort"`
	IPAddress  string `json:"ipAddress"`
	NasID      string `json:"nasId"`
}

type SadisSubscriber struct {
	ID        string `json:"id"`
	CTag      int16  `json:"cTag"`
	STag      int16  `json:"sTag"`
	NasPortID string `json:"nasPortId"`
	CircuitID string `json:"circuitId"`
	RemoteID  string `json:"remoteId"`
}
*/
/*"entries" : [ {
    "id" : "EC1721000218",
    "uplinkPort" : 65536,
    "hardwareIdentifier" : "3c:2c:99:f7:c6:82",
    "ipAddress" : "10.64.1.205",
    "nasId" : "ATLEDGEVOLT1"
  }, {
    "id" : "ALPHe3d1ce3f-1",
    "cTag" : 20,
    "sTag" : 10,
    "nasPortId" : "PON 1/1/3/1:1.1.1",
    "circuitId" : "PON 1/1/3/1:1.1.1-CID",
    "remoteId" : "ATLEDGEVOLT1-RID"
  }]
*/
func getSubscriberHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	sadisRequestID := vars["id"]

	defer r.Body.Close()

	key, err := aero.NewKey(aeroNamespace, "sadis", sadisRequestID)
	if err != nil {
		log.Fatal(err)
	}

	record, err := client.Get(nil, key, "sadis-entry")
	if err != nil {
		http.NotFound(w, r)
		log.Println("key-search", err, key)
		return
	}

	if json, ok := record.Bins["sadis-entry"]; ok {
		log.Println("Sending Response 200")
		log.Println(json)
		w.Write([]byte(json.(string)))
		return
	}
}

func addSubscriberHandler(w http.ResponseWriter, r *http.Request) {

	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	log.Println(string(body))

	var id sadisObject
	err = json.Unmarshal(body, &id)
	if err != nil {
		panic(err)
	}

	key, err := aero.NewKey(aeroNamespace, "sadis", id.ID)
	if err != nil {
		log.Fatal(err)
	}

	bins := aero.BinMap{
		"sadis-entry": string(body),
	}

	err = client.Put(nil, key, bins)

	if err == nil {
		log.Println("Sending Response 200 OK")
		w.WriteHeader(http.StatusOK)
		return
	} else {
		log.Println("Sending Response 500")
		w.WriteHeader(http.StatusInternalServerError)
		panicOnError(err)
	}
}
func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	flag.StringVar(&aeroIp, "aeroIp", "127.0.0.1", "Aerospike DB IP address")
	flag.IntVar(&aeroPort, "aeroPort", 3000, "Aerospike DB Port")
	flag.IntVar(&port, "port", 8888, "Which port to run the rest api on")
	flag.StringVar(&aeroNamespace, "aeroNamespace", "sadis-data", "Aerospike DB namespace")

	flag.Parse()

	log.Println("Aerospike Ip address: ", aeroIp)

	var err error
	client, err = aero.NewClient(aeroIp, aeroPort)
	if err != nil {
		log.Fatal(err)
	}

	// read it back!
	//_, err = client.Get(nil, key)
	//panicOnError(err)

	// delete the key, and check if key exists
	//existed, err := client.Delete(nil, key)
	//panicOnError(err)
	//fmt.Printf("Record existed before delete? %v\n", existed)

	/*
		var devices []SadisDevice
		devices = append(devices, SadisDevice{
			ID:         "EC1721000218",
			Uplink:     65536,
			HardwareID: "3c:2c:99:f7:c6:82",
			IPAddress:  "10.64.1.205",
			NasID:      "ATLEDGEVOLT1",
		})

		var subscribers []SadisSubscriber
		subscribers = append(subscribers, SadisSubscriber{
			ID:        "ALPHe3d1ce3f-1",
			CTag:      20,
			STag:      10,
			NasPortID: "PON 1/1/3/1:1.1.1",
			CircuitID: "PON 1/1/3/1:1.1.1-CID",
			RemoteID:  "ATLEDGEVOLT1-RID",
		})

		sub, err := json.Marshal(&subscribers[0])
		dev, err := json.Marshal(&devices[0])

		// define some bins with data
		bins := aero.BinMap{
			"EC1721000218": dev,
			"ALPHe3d1ce3f-1": sub,
		}

		// write the bins
		err = client.Put(nil, key, bins)
		panicOnError(err)
	*/
	router := mux.NewRouter()
	router.HandleFunc("/subscriber/{id}", getSubscriberHandler).Methods("GET")
	router.HandleFunc("/subscriber", addSubscriberHandler).Methods("POST")
	http.Handle("/", router)

	log.Println("serving-port", port)
	panic(http.ListenAndServe(fmt.Sprintf(":%d", port), router))

}
