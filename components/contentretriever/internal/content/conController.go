package content

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/De-cROMPOS/pastebin/contentretriever/internal/connections"
	er "github.com/De-cROMPOS/pastebin/contentretriever/internal/errorhandler"
)

type ContentController struct {
	connections.PGClient
	connections.RClient
}

func (cc *ContentController) Initialize() error {
	err := cc.PGClient.PGInit()
	if err != nil {
		return fmt.Errorf("smth went wrong while initializing content controller: %s", err)
	}

	err = cc.RedisInit()
	if err != nil {
		return fmt.Errorf("smth went wrong while initializing content controller: %s", err)
	}

	return nil
}

func (cc *ContentController) Close() {
	cc.RClient.Close()
	cc.PGClient.Close()
	log.Printf("connections closed")
}
func (cc *ContentController) GetHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte(`{"error": "method not allowed"}`))
		return
	}

	hash := r.URL.Query().Get("hash")

	link, err := cc.RClient.GetLink(hash)
	if err != nil && !errors.Is(err, er.ErrRecordNotFound) {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"error": "redis client died"}`))
		log.Printf("error while getting link from redis: %v", err)
		return
	}

	if err == nil {
		http.Redirect(w, r, link, http.StatusFound)
		log.Printf("redirected hash %s to %s (from cache)", hash, link)
		return
	}

	meta, err := cc.PGClient.GetMeta(hash)
	if err != nil {
		log.Printf("error while getting link from pg: %s", err)
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"error": "unable to find url by hash"}`))
		return
	}

	http.Redirect(w, r, meta.S3URL, http.StatusFound)

	go func() {
		if err := cc.RClient.AddMeta(meta); err != nil {
			log.Printf("failed to cache meta %s: %v", hash, err)
		}
	}()

	log.Printf("redirected hash %s to %s", hash, link)
}
