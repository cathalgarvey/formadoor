/*Package totpClient is the default client for doorMicroservice. It requires a
valid API key for the door, and a list of JSON objects representing permitted
members. For each member, a string desribing a time-policy is provided along
with the member's TOTP secret and contact details.

The tool opens a small code-entry shell accepting numeric input and expecting
a return to enter a valid code. To prevent malicious input use a USB keypad,
preferably disabling or remapping invalid keys and, if possible, numlock.
*/
package main

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"os"

	"gopkg.in/inconshreveable/log15.v2"

	"github.com/alecthomas/kingpin"
	"github.com/cathalgarvey/formadoor/doorapi"
	"github.com/cathalgarvey/formadoor/totpset"
)

var (
	accounts       []FormiteAccount
	totps          *totpset.Set
	accountsFile   = kingpin.Arg("accounts", "Accounts JSON File").Required().ExistingFile()
	apiKey         = kingpin.Arg("apiKey", "API key for the door service (base64)").Required().String()
	secondsGranted = kingpin.Flag("seconds-granted", "Seconds to unlock door for to permit entry on successful authentication").Default("5").Short('s').Int()
	doorPort       = kingpin.Flag("door-port", "Port the door microservice API listens on").Default("8080").Short('p').Int()
	door           doorapi.Door
)

func init() {
	kingpin.Parse()
	accountsFileContents, err := ioutil.ReadFile(*accountsFile)
	if err != nil {
		panic(err)
	}
	err = json.Unmarshal(accountsFileContents, &accounts)
	if err != nil {
		panic(err)
	}
	totps = totpset.NewSet(3)
	totps.ValidityCallback = passcodeToTimePolicy
	for _, account := range accounts {
		accKey := totpset.NewKey(account.Name, account.Secret)
		accKey.Metadata["account"] = account
		totps.Keys = append(totps.Keys, accKey)
	}
	door = doorapi.Door{Port: *doorPort, Secret: *apiKey}
}

func main() {
	for {
		print("Please enter 6-digit code: ")
		reader := bufio.NewReader(os.Stdin)
		codeAttempt, err := reader.ReadString('\n')
		if err != nil {
			log15.Error("Error getting input", log15.Ctx{"err": err, "attempt": codeAttempt})
			continue
		}
		ok, who, err := totps.Validate(codeAttempt, func(s string) {
			log15.Info("While validating: " + s)
		})
		if err != nil {
			log15.Error("Error validating code", log15.Ctx{"err": err, "who": who, "ok": ok, "attempt": codeAttempt})
			continue
		}
		whoPolicy := who.Metadata["account"].(FormiteAccount).TimePolicy
		if ok {
			log15.Info("Code validated and access granted", log15.Ctx{"who": who, "code": codeAttempt, "policy": whoPolicy})
			err = door.InstructDoorToOpenForSeconds(*secondsGranted)
			if err != nil {
				log15.Error("Error instructing door to open", log15.Ctx{"who": who, "code": codeAttempt, "policy": whoPolicy, "err": err})
			}
		} else {
			log15.Info("Code validated but access denied", log15.Ctx{"who": who, "code": codeAttempt, "policy": whoPolicy})
			continue
		}
	}
}
