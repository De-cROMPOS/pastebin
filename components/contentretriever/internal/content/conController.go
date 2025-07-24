package content

import (
	"fmt"
	"log"
	"net/http"

	"github.com/De-cROMPOS/pastebin/contentretriever/internal/connections"
)

type ContentController struct {
	connections.PGClient
}

func (cc *ContentController) Initialize() error {
	err := cc.PGClient.PGInit()
	if err != nil {
		return fmt.Errorf("smth went wrong while initializing content controller: %s", err)
	}

	return nil
}

//todo: добавить редис...
func (cc *ContentController) GetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error": "method not allowed"}`))
		return
	}

	hash := r.URL.Query().Get("hash")
	link, err := cc.PGClient.GetLink(hash)
	if err != nil {
		log.Printf("error while getting link from pg: %s", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"error": "unable to find url by hash"}`))
		return
	}

    http.Redirect(w, r, link, http.StatusFound)
    
    log.Printf("redirected hash %s to %s", hash, link)
}


