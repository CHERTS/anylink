package handler

import (
	"crypto/sha1"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"time"

	"github.com/cherts/anylink/base"
	"github.com/cherts/anylink/dbdata"
	"github.com/cherts/anylink/pkg/utils"
	"github.com/gorilla/mux"
	"github.com/pires/go-proxyproto"
)

func startTls() {

	var (
		err error

		addr = base.Cfg.ServerAddr
		ln   net.Listener
	)

	// Determine certificate file
	// _, err = os.Stat(certFile)
	// if errors.Is(err, os.ErrNotExist) {
	//	// Automatically generate certificates
	//	certs[0], err = selfsign.GenerateSelfSignedWithDNS("vpn.anylink")
	// } else {
	//	// Use custom certificate
	//	certs[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	// }

	tlscert, _, err := dbdata.ParseCert()
	if err != nil {
		base.Fatal("Certificate loading failed", err)
	}
	dbdata.LoadCertificate(tlscert)

	// Calculate the certificate hash value
	s1 := sha1.New()
	s1.Write(tlscert.Certificate[0])
	h2s := hex.EncodeToString(s1.Sum(nil))
	certHash = strings.ToUpper(h2s)
	base.Info("certHash", certHash)

	// repair CVE-2016-2183
	// https://segmentfault.com/a/1190000038486901
	// nmap -sV --script ssl-enum-ciphers -p 443 www.example.com
	cipherSuites := tls.CipherSuites()
	selectedCipherSuites := make([]uint16, 0, len(cipherSuites))
	for _, s := range cipherSuites {
		selectedCipherSuites = append(selectedCipherSuites, s.ID)
	}

	// Set tls information
	tlsConfig := &tls.Config{
		NextProtos:   []string{"http/1.1"},
		MinVersion:   tls.VersionTLS12,
		CipherSuites: selectedCipherSuites,
		GetCertificate: func(chi *tls.ClientHelloInfo) (*tls.Certificate, error) {
			base.Trace("GetCertificate ServerName", chi.ServerName)
			return dbdata.GetCertificateBySNI(chi.ServerName)
		},
	}
	srv := &http.Server{
		Addr:         addr,
		Handler:      initRoute(),
		TLSConfig:    tlsConfig,
		ErrorLog:     base.GetServerLog(),
		ReadTimeout:  100 * time.Second,
		WriteTimeout: 100 * time.Second,
	}

	ln, err = net.Listen("tcp", addr)
	if err != nil {
		base.Fatal(err)
	}
	defer ln.Close()

	if base.Cfg.ProxyProtocol {
		ln = &proxyproto.Listener{
			Listener:          ln,
			ReadHeaderTimeout: 30 * time.Second,
		}
	}

	base.Info("Listen server", addr)
	err = srv.ServeTLS(ln, "", "")
	if err != nil {
		base.Fatal(err)
	}
}

func initRoute() http.Handler {
	r := mux.NewRouter()
	// Add security headers to all routes
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			utils.SetSecureHeader(w)
			next.ServeHTTP(w, req)
		})
	})

	r.HandleFunc("/", LinkHome).Methods(http.MethodGet)
	r.HandleFunc("/", LinkAuth).Methods(http.MethodPost)
	// r.Handle("/", antiBruteForce(http.HandlerFunc(LinkAuth))).Methods(http.MethodPost)
	r.HandleFunc("/CSCOSSLC/tunnel", LinkTunnel).Methods(http.MethodConnect)
	r.HandleFunc("/otp_qr", LinkOtpQr).Methods(http.MethodGet)
	r.HandleFunc("/otp-verification", LinkAuth_otp).Methods(http.MethodPost)
	// r.Handle("/otp-verification", antiBruteForce(http.HandlerFunc(LinkAuth_otp))).Methods(http.MethodPost)
	r.HandleFunc(fmt.Sprintf("/profile_%s.xml", base.Cfg.ProfileName), func(w http.ResponseWriter, r *http.Request) {
		b, _ := os.ReadFile(base.Cfg.Profile)
		w.Write(b)
	}).Methods(http.MethodGet)
	r.PathPrefix("/files/").Handler(
		http.StripPrefix("/files/",
			http.FileServer(http.Dir(base.Cfg.FilesPath)),
		),
	)
	// health check
	r.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "ok")
	}).Methods(http.MethodGet)
	r.NotFoundHandler = http.HandlerFunc(notFound)
	return r
}

func notFound(w http.ResponseWriter, r *http.Request) {
	if base.GetLogLevel() == base.LogLevelTrace {
		hd, _ := httputil.DumpRequest(r, true)
		base.Trace("NotFound: ", r.RemoteAddr, string(hd))
	}

	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintln(w, "404 page not found")
}
