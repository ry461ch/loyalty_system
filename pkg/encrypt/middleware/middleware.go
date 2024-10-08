package encryptmiddleware

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/ry461ch/loyalty_system/pkg/encrypt"
)

type ResponseEncrypter struct {
	http.ResponseWriter
	encrypter *encrypt.Encrypter
}

func (re *ResponseEncrypter) Write(b []byte) (int, error) {
	bodyHash := re.encrypter.EncryptMessage(b)
	re.Header().Set("HashSHA256", fmt.Sprintf("%x", bodyHash))
	return re.ResponseWriter.Write(b)
}

func CheckRequestAndEncryptResponse(encrypter *encrypt.Encrypter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqHeaderHash256 := r.Header.Get("HashSHA256")
			if reqHeaderHash256 == "" {
				next.ServeHTTP(w, r)
				return
			}

			var buf bytes.Buffer
			_, err := buf.ReadFrom(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			reqBody := buf.Bytes()
			reqHash := encrypter.EncryptMessage(reqBody)

			if fmt.Sprintf("%x", reqHash) != reqHeaderHash256 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewBuffer(reqBody))
			next.ServeHTTP(&ResponseEncrypter{ResponseWriter: w, encrypter: encrypter}, r)
		})
	}
}
