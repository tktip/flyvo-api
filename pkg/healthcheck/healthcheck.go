package healthcheck

import (
	"net/http"

	"github.com/sirupsen/logrus"
)

func health(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte("OK"))
}

// StartHealthService starts a health-check service
func StartHealthService() {
	http.HandleFunc("/health", health)
	logrus.Infof("Staring health check on http://localhost:8090/health")
	if err := http.ListenAndServe(":8090", nil); err != nil {
		panic(err)
	}
}
