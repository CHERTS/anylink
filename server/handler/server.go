package handler

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
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
			base.Trace("GetCertificate", chi.ServerName)
			return dbdata.GetCertificateBySNI(chi.ServerName)
		},
	}
	srv := &http.Server{
		Addr:         addr,
		Handler:      initRoute(),
		TLSConfig:    tlsConfig,
		ErrorLog:     base.GetBaseLog(),
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	ln, err = net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
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
	r.HandleFunc("/CSCOSSLC/tunnel", LinkTunnel).Methods(http.MethodConnect)
	r.HandleFunc("/otp_qr", LinkOtpQr).Methods(http.MethodGet)
	r.HandleFunc("/profile.xml", func(w http.ResponseWriter, r *http.Request) {
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
	// fmt.Println(r.RemoteAddr)
	// hu, _ := httputil.DumpRequest(r, true)
	// fmt.Println("NotFound: ", string(hu))

	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintln(w, "404 page not found")
}
