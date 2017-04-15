package auth

import (
	"bytes"
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"io"
	"io/ioutil"
	"math/big"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

type keyCache interface {
	// Get gets or updates a cache entry.
	Get(kid string) *rsa.PublicKey
}

// pubkeyCache caches *rsa.PublicKeys according to keyid/kid parameter
// any time a cache miss occurs, we fetch from remote
type pubkeyCache struct {
	mu          sync.RWMutex
	cache       map[string]*rsa.PublicKey
	remoteUrl   string
	client      *http.Client
	updateToken chan struct{}
	ll          *zap.Logger
}

var _ keyCache = &pubkeyCache{}

type pubKey struct {
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type pubKeys []*pubKey
type pubKeyResp struct {
	Keys pubKeys `json:"keys"`
}

func newPubkeyCache(remoteUrl string, ll *zap.Logger) *pubkeyCache {
	updateToken := make(chan struct{}, 1)

	// token bucket to only allow one update every 5s
	updateToken <- struct{}{}
	go func() {
		t := time.NewTicker(5 * time.Second)
		for range t.C {
			updateToken <- struct{}{}
		}
	}()
	return &pubkeyCache{
		cache:     make(map[string]*rsa.PublicKey),
		remoteUrl: remoteUrl,
		client:    http.DefaultClient,

		updateToken: updateToken,
		ll:          ll,
	}
}

func (p *pubkeyCache) Get(kid string) *rsa.PublicKey {
	p.mu.RLock()

	var key *rsa.PublicKey
	var ok bool
	if key, ok = p.cache[kid]; !ok {
		p.mu.RUnlock()
		p.fetchKeys()

		// try again!
		p.mu.RLock()
		key = p.cache[kid]
		p.mu.RUnlock()
	}

	return key
}

func (p *pubkeyCache) fetchKeys() {
	p.mu.Lock()
	defer p.mu.Unlock()

	// prevent many updates on a cache miss with the token bucket
	select {
	case <-p.updateToken:
		p.ll.Info("updating pubkey cache")
		m, err := doFetch(p.client, p.remoteUrl)
		if err != nil {
			p.ll.Error("error updating cached keys", zap.Error(err))
			return
		}

		p.cache = m
	}

	return
}

func doFetch(client *http.Client, url string) (map[string]*rsa.PublicKey, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		_, _ = io.Copy(ioutil.Discard, resp.Body)
		resp.Body.Close()
	}()

	kresp := &pubKeyResp{}
	err = json.NewDecoder(resp.Body).Decode(kresp)
	if err != nil {
		return nil, err
	}

	m := make(map[string]*rsa.PublicKey)
	for _, key := range kresp.Keys {
		rsakey, err := googleKeyToRSAPublicKey(key.N, key.E)
		if err != nil {
			return nil, err
		}
		m[key.Kid] = rsakey
	}

	return m, nil
}

// googleKeyToRSAPublicKey converts a base64 publickey from Google's oauth cert list into an rsa.PublicKey
func googleKeyToRSAPublicKey(nstr, estr string) (*rsa.PublicKey, error) {
	decN, err := base64.URLEncoding.DecodeString(nstr)
	if err != nil {
		return nil, err
	}
	n := big.NewInt(0)
	n.SetBytes(decN)

	decE, err := base64.URLEncoding.DecodeString(estr)
	if err != nil {
		return nil, err
	}
	var eBytes []byte
	if len(decE) < 8 {
		eBytes = make([]byte, 8-len(decE), 8)
		eBytes = append(eBytes, decE...)
	} else {
		eBytes = decE
	}
	eReader := bytes.NewReader(eBytes)
	var e uint64
	err = binary.Read(eReader, binary.BigEndian, &e)
	if err != nil {
		return nil, err
	}
	return &rsa.PublicKey{N: n, E: int(e)}, nil
}
