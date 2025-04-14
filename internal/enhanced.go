package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type EnhancedRequest struct {
	*http.Request
}
type EnhancedResponseWriter struct {
	http.ResponseWriter;
}

type EnhancedHandler func(w EnhancedResponseWriter, r *EnhancedRequest) *CustomError

func (e EnhancedHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	req := &EnhancedRequest{r}
	res := EnhancedResponseWriter{w}
    if err := e(res, req); err != nil {
		LogError(err, map[string]interface{}{"user_id": req.Context().Value("userID")})
		WriteError(res, err)
    }
}

func WriteError(w http.ResponseWriter, receivedError *CustomError) {
	js, err := json.MarshalIndent(receivedError.ErrorData(), "", "\t")
	if err != nil {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
	}

	js = append(js, '\n')
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(receivedError.status)
	w.Write(js)
}


func (r *EnhancedRequest) DecodeJSON(w *EnhancedResponseWriter, dst interface{}) error {
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(dst)
	if err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError
		var invalidUnmarshalError *json.InvalidUnmarshalError

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %q", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("body contains unknown key %s", fieldName)

		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	err = dec.Decode(&struct{}{})
	if err != io.EOF {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

func (w EnhancedResponseWriter) WriteResponse(status int, data any) {
	js, err := json.MarshalIndent(data, "", "\t")
	if err != nil {
		newErr := NewError(http.StatusInternalServerError, "response.marshal", "failed to create response body", err.Error())
		WriteError(w, newErr)
	}

	js = append(js, '\n')

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	w.Write(js)
}