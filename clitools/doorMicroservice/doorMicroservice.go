/*Package doorMicroservice is a small server intended to run on a
Raspberry Pi with a Piface attached & configured. It listens for
authenticated clients on a HTTP POST interface and switches a relay
for a given timespan for authenticated requests (Hmac+timestamp).
*/
package main

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/alecthomas/kingpin"
	"github.com/mailgun/iptools"
)

const (
	maxOpenTime = 10
)

var (
	localOnly  = kingpin.Flag("local-only", "Not working properly! Whether to only accept requests from local network IPs").Default("false").Bool()
	listenPort = kingpin.Flag("port", "What port to listen on").Short('p').Default("8080").Int()
	tokensFile = kingpin.Arg("tokens", "JSON list of authorised apps. Entries are objects with string keys: Key (b64 string), Name, DevName, DevEmail.").Required().ExistingFile()
	authKeys   []authApp
)

type authApp struct {
	Key      string
	Name     string
	DevName  string
	DevEmail string
}

func init() {
	kingpin.Parse()
	loadTokens(*tokensFile)
}

func loadTokens(fn string) {
	contents, err := ioutil.ReadFile(fn)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(contents, &authKeys)
	if err != nil {
		panic(err)
	}
}

func serve(unlockCallback func(int)) {
	log15.Info("Registering handler.")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Check for a HMAC header to authenticate body, and verify that
		// authenticated body has a recent timestamp.
		authedRequest, err := getAuthenticatedBody(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			return
		}
		log15.Info("Proceeding with authenticated request", log15.Ctx{"request": authedRequest})
		if authedRequest.Seconds > maxOpenTime || authedRequest.Seconds < 1 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Can only open door for between 1 and" + strconv.Itoa(authedRequest.Seconds) + "seconds."))
			log15.Error("Cannot open for requested timespan", log15.Ctx{"maxOpenTime": maxOpenTime, "requestedTime": authedRequest.Seconds})
			return
		}
		if *localOnly && (!iptools.IsPrivate(net.ParseIP(r.RemoteAddr))) {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte("Only local addresses may issue door commands."))
			log15.Error("Cannot serve authenticated request as origin is not local", log15.Ctx{"requestingIP": net.ParseIP(r.RemoteAddr)})
			return
		}
		unlockCallback(authedRequest.Seconds)
		w.Write([]byte("Success, door opening for " + strconv.Itoa(authedRequest.Seconds) + " seconds"))
		log15.Info("Opening door!", log15.Ctx{"Seconds": authedRequest.Seconds})
		return
	})
	http.ListenAndServe(":"+strconv.Itoa(*listenPort), nil)
}

func main() {
	log15.Info("Initialising PiFace")
	serve(unlockDoorForSeconds)
}
