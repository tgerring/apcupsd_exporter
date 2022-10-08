// Command apcupsd_exporter provides a Prometheus exporter for apcupsd.
package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"strings"

	"github.com/mdlayher/apcupsd"
	apcupsdexporter "github.com/mdlayher/apcupsd_exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const DELIMIT = ","

var (
	telemetryAddr = flag.String("telemetry.addr", ":9162", "address for apcupsd exporter")
	metricsPath   = flag.String("telemetry.path", "/metrics", "URL path for surfacing collected metrics")

	apcupsdAddr    = flag.String("apcupsd.addr", ":3551", "address of apcupsd Network Information Server (NIS): separate multiple entries by comma")
	apcupsdNetwork = flag.String("apcupsd.network", "tcp", `network of apcupsd Network Information Server (NIS): typically "tcp", "tcp4", or "tcp6"`)
)

func main() {
	flag.Parse()

	if *apcupsdAddr == "" {
		log.Fatal("address of apcupsd Network Information Server (NIS) must be specified with '-apcupsd.addr' flag")
	}

	// split & loop through list of apcupsd addressess
	var addrList = strings.Split(*apcupsdAddr, DELIMIT)
	for i := 0; i < len(addrList); i++ {
		log.Printf("starting apcupsd exporter for server %s://%s", *apcupsdNetwork, addrList[i])
		fn := newClient(*apcupsdNetwork, addrList[i])
		prometheus.MustRegister(apcupsdexporter.New(fn))
	}

	// start exporter
	log.Printf("starting prometheus exporter on %q",
		*telemetryAddr)

	http.Handle(*metricsPath, promhttp.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, *metricsPath, http.StatusMovedPermanently)
	})

	if err := http.ListenAndServe(*telemetryAddr, nil); err != nil {
		log.Fatalf("cannot start apcupsd exporter: %s", err)
	}
}

func newClient(network, addr string) apcupsdexporter.ClientFunc {
	return func(ctx context.Context) (*apcupsd.Client, error) {
		return apcupsd.DialContext(ctx, network, addr)
	}
}
