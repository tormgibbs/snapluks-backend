package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/tormgibbs/snapluks-backend/internal/validator"

	"github.com/julienschmidt/httprouter"
)

type envelope map[string]any

var validImageExts = []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}

func (app *application) readIDParam(r *http.Request) (int64, error) {
	params := httprouter.ParamsFromContext(r.Context())

	id, err := strconv.ParseInt(params.ByName("id"), 10, 64)
	if err != nil || id < 1 {
		return 0, errors.New("invalid id parameter")
	}

	return id, nil
}

func (app *application) readString(qs url.Values, key string, defaultValue string) string {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	return s
}

func (app *application) readCSV(qs url.Values, key string, defaultValue []string) []string {
	csv := qs.Get(key)

	if csv == "" {
		return defaultValue
	}

	return strings.Split(csv, ",")
}

func (app *application) readInt(qs url.Values, key string, defaultValue int, v *validator.Validator) int {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		v.AddError(key, "must be an integer value")
		return defaultValue
	}

	return i
}

func (app *application) readFloat(qs url.Values, key string, defaultValue *float64, v *validator.Validator) *float64 {
	s := qs.Get(key)

	if s == "" {
		return defaultValue
	}

	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		v.AddError(key, "must be a float value")
		return defaultValue
	}

	return &f
}

func (app *application) writeJSON(w http.ResponseWriter, status int, data envelope, headers http.Header) error {
	for key, value := range headers {
		w.Header()[key] = value
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "\t")

	if err := encoder.Encode(data); err != nil {
		return err
	}

	return nil
}

func (app *application) readJSON(w http.ResponseWriter, r *http.Request, v interface{}) error {
	maxBytes := 1_048_576
	r.Body = http.MaxBytesReader(w, r.Body, int64(maxBytes))

	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()

	err := dec.Decode(v)
	if err != nil {
		var (
			syntaxError           *json.SyntaxError
			unmarshalTypeError    *json.UnmarshalTypeError
			invalidUnmarshalError *json.InvalidUnmarshalError
		)

		switch {
		case errors.As(err, &syntaxError):
			return fmt.Errorf("body contains badly-formed JSON (at character %d)", syntaxError.Offset)

		case errors.Is(err, io.ErrUnexpectedEOF):
			return errors.New("body contains badly-formed JSON")

		case errors.As(err, &unmarshalTypeError):
			if unmarshalTypeError.Field != "" {
				return fmt.Errorf("body contains incorrect JSON type for field %s", unmarshalTypeError.Field)
			}
			return fmt.Errorf("body contains incorrect JSON type (at character %d)", unmarshalTypeError.Offset)

		case errors.Is(err, io.EOF):
			return errors.New("body must not be empty")

		case strings.HasPrefix(err.Error(), "json: unknown field"):
			fieldName := strings.Trim(strings.TrimPrefix(err.Error(), "json: unknown field "), `"`)
			return fmt.Errorf("body contains unknown key %s", fieldName)

		case err.Error() == "http: request body too large":
			return fmt.Errorf("body must not be larger than %d bytes", maxBytes)

		case errors.As(err, &invalidUnmarshalError):
			panic(err)

		default:
			return err
		}
	}

	if dec.More() {
		return errors.New("body must only contain a single JSON value")
	}

	return nil
}

func (app *application) background(fn func()) {
	app.wg.Add(1)
	go func() {
		defer func() {
			defer app.wg.Done()
			if err := recover(); err != nil {
				app.logger.PrintError(fmt.Errorf("%s", err), nil)
			}
		}()
		fn()
	}()
}

func (app *application) uploadImageToS3(fileHeader *multipart.FileHeader) (string, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	if !isValidImageType(fileHeader.Filename) {
		return "", errors.New("invalid image type")
	}

	if fileHeader.Size > 5*1024*1024 {
		return "", errors.New("image size limit exceeded")
	}

	ext := filepath.Ext(fileHeader.Filename)
	s3Key := fmt.Sprintf("services/%d_%s%s", time.Now().UnixNano(),
		strings.ReplaceAll(fileHeader.Filename[:len(fileHeader.Filename)-len(ext)], " ", "_"), ext)

	_, err = app.s3Client.UploadFile(file, s3Key, fileHeader.Header.Get("Content-Type"))
	if err != nil {
		return "", fmt.Errorf("failed to upload to S3: %w", err)
	}

	return s3Key, nil
}

func isValidImageType(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return slices.Contains(validImageExts, ext)
}

func (app *application) cleanupFailedServiceCreation(serviceID, providerID int64, uploadedImages []string) {
	err := app.models.Services.Delete(serviceID, providerID)
	if err != nil {
		app.logger.PrintError(err, map[string]string{
			"service_id":  fmt.Sprintf("%d", serviceID),
			"provider_id": fmt.Sprintf("%d", providerID),
		})
	}

	if len(uploadedImages) > 0 {
		err = app.s3Client.DeleteFiles(uploadedImages)
		if err != nil {
			app.logger.PrintError(err, map[string]string{
				"images": fmt.Sprintf("%v", uploadedImages),
			})
		}
	}
}

func (app *application) getFormValue(form *multipart.Form, key string) *string {
	if val, ok := form.Value[key]; ok && len(val) > 0 {
		val := strings.TrimSpace(val[0])
		return &val
	}
	return nil
}

func (app *application) _readMultipartForm(w http.ResponseWriter, r *http.Request, maxMemory int64) error {
	err := r.ParseMultipartForm(maxMemory)
	if err != nil {
		var maxBytesError *http.MaxBytesError

		switch {
		case errors.As(err, &maxBytesError):
			return fmt.Errorf("multipart form too large (maximum %d bytes)", maxMemory)

		case err == http.ErrNotMultipart:
			return errors.New("request is not multipart/form-data")

		case err == http.ErrMissingBoundary:
			return errors.New("multipart form missing boundary")

		case strings.Contains(err.Error(), "no multipart boundary param"):
			return errors.New("invalid multipart boundary")

		case strings.Contains(err.Error(), "malformed MIME header"):
			return errors.New("malformed multipart form")

		default:
			return fmt.Errorf("failed to parse multipart form: %w", err)
		}
	}

	return nil
}

func (app *application) readMultipartForm(r *http.Request, maxMemory int64, v any) error {
	err := r.ParseMultipartForm(maxMemory)
	if err != nil {
		var maxBytesError *http.MaxBytesError

		switch {
		case errors.As(err, &maxBytesError):
			return fmt.Errorf("multipart form too large (maximum %d bytes)", maxMemory)

		case err == http.ErrNotMultipart:
			return errors.New("request is not multipart/form-data")

		case err == http.ErrMissingBoundary:
			return errors.New("multipart form missing boundary")

		case strings.Contains(err.Error(), "no multipart boundary param"):
			return errors.New("invalid multipart boundary")

		case strings.Contains(err.Error(), "malformed MIME header"):
			return errors.New("malformed multipart form")

		default:
			return fmt.Errorf("failed to parse multipart form: %w", err)
		}
	}

	return app.populateStructFromForm(r, v)
}

func (app *application) populateStructFromForm(r *http.Request, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("destination must be a pointer to a struct")
	}

	rv = rv.Elem()
	rt := rv.Type()

	for i := 0; i < rv.NumField(); i++ {
		field := rv.Field(i)
		structField := rt.Field(i)

		if !field.CanSet() {
			continue
		}

		// Get the form tag name, fallback to json tag, then field name
		formTag := structField.Tag.Get("form")
		if formTag == "" {
			formTag = strings.ToLower(structField.Name)
		}

		// Handle different field types
		switch field.Kind() {
		case reflect.String:
			if value := r.FormValue(formTag); value != "" {
				field.SetString(value)
			}

		case reflect.Ptr:
			if field.Type().Elem().Kind() == reflect.String {
				if value := r.FormValue(formTag); value != "" {
					field.Set(reflect.ValueOf(&value))
				}
			} else if field.Type() == reflect.TypeOf((*multipart.FileHeader)(nil)) {
				if file, header, err := r.FormFile(formTag); err == nil {
					file.Close()
					field.Set(reflect.ValueOf(header))
				}
			} else if value := r.FormValue(formTag); value != "" {
				elemKind := field.Type().Elem().Kind()
				switch elemKind {
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					if num, err := strconv.ParseInt(value, 10, 64); err == nil {
						newVal := reflect.New(field.Type().Elem())
						newVal.Elem().SetInt(num)
						field.Set(newVal)
					}
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					if num, err := strconv.ParseUint(value, 10, 64); err == nil {
						newVal := reflect.New(field.Type().Elem())
						newVal.Elem().SetUint(num)
						field.Set(newVal)
					}
				case reflect.Float32, reflect.Float64:
					if f, err := strconv.ParseFloat(value, 64); err == nil {
						newVal := reflect.New(field.Type().Elem())
						newVal.Elem().SetFloat(f)
						field.Set(newVal)
					}
				case reflect.Bool:
					if b, err := strconv.ParseBool(value); err == nil {
						newVal := reflect.New(field.Type().Elem())
						newVal.Elem().SetBool(b)
						field.Set(newVal)
					}
				}
			}

		case reflect.Slice:
			if field.Type() == reflect.TypeOf([]*multipart.FileHeader{}) {
				// Handle []*multipart.FileHeader fields
				if r.MultipartForm != nil && r.MultipartForm.File != nil {
					if fileHeaders, exists := r.MultipartForm.File[formTag]; exists {
						field.Set(reflect.ValueOf(fileHeaders))
					}
				}
			} else {
				// Handle all other slice types
				values := r.Form[formTag]
				if len(values) > 0 {
					elemType := field.Type().Elem()
					elemKind := elemType.Kind()

					// Create a new slice with the correct element type
					slice := reflect.MakeSlice(field.Type(), 0, len(values))

					for _, v := range values {
						if v == "" {
							continue
						}

						var elemValue reflect.Value

						switch elemKind {
						case reflect.String:
							elemValue = reflect.ValueOf(v)

						case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
							if num, err := strconv.ParseInt(v, 10, 64); err == nil {
								elemValue = reflect.ValueOf(num).Convert(elemType)
							}

						case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
							if num, err := strconv.ParseUint(v, 10, 64); err == nil {
								elemValue = reflect.ValueOf(num).Convert(elemType)
							}

						case reflect.Float32, reflect.Float64:
							if f, err := strconv.ParseFloat(v, 64); err == nil {
								elemValue = reflect.ValueOf(f).Convert(elemType)
							}

						case reflect.Bool:
							if b, err := strconv.ParseBool(v); err == nil {
								elemValue = reflect.ValueOf(b)
							}

						case reflect.Ptr:
							ptrElemType := elemType.Elem()
							ptrElemKind := ptrElemType.Kind()

							var ptrValue reflect.Value

							switch ptrElemKind {
							case reflect.String:
								ptrValue = reflect.ValueOf(&v)

							case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
								if num, err := strconv.ParseInt(v, 10, 64); err == nil {
									convertedNum := reflect.ValueOf(num).Convert(ptrElemType).Interface()
									ptrValue = reflect.ValueOf(convertedNum).Addr()
								}

							case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
								if num, err := strconv.ParseUint(v, 10, 64); err == nil {
									convertedNum := reflect.ValueOf(num).Convert(ptrElemType).Interface()
									ptrValue = reflect.ValueOf(convertedNum).Addr()
								}

							case reflect.Float32, reflect.Float64:
								if f, err := strconv.ParseFloat(v, 64); err == nil {
									convertedFloat := reflect.ValueOf(f).Convert(ptrElemType).Interface()
									ptrValue = reflect.ValueOf(convertedFloat).Addr()
								}

							case reflect.Bool:
								if b, err := strconv.ParseBool(v); err == nil {
									ptrValue = reflect.ValueOf(&b)
								}
							}

							if ptrValue.IsValid() {
								elemValue = ptrValue
							}
						}

						if elemValue.IsValid() {
							slice = reflect.Append(slice, elemValue)
						}
					}

					if slice.Len() > 0 {
						field.Set(slice)
					}
				}
			}

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if value := r.FormValue(formTag); value != "" {
				if num, err := strconv.ParseInt(value, 10, 64); err == nil {
					field.SetInt(num)
				}
			}

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if value := r.FormValue(formTag); value != "" {
				if num, err := strconv.ParseUint(value, 10, 64); err == nil {
					field.SetUint(num)
				}
			}

		case reflect.Float32, reflect.Float64:
			if value := r.FormValue(formTag); value != "" {
				if f, err := strconv.ParseFloat(value, 64); err == nil {
					field.SetFloat(f)
				}
			}

		case reflect.Bool:
			if value := r.FormValue(formTag); value != "" {
				if b, err := strconv.ParseBool(value); err == nil {
					field.SetBool(b)
				}
			}
		}
	}

	return nil
}

func (app *application) getFormValueOptional(form *multipart.Form, key string) *string {
	values, exists := form.Value[key]
	if !exists || len(values) == 0 || values[0] == "" {
		return nil
	}
	return &values[0]
}

func (app *application) getFormValueRequired(form *multipart.Form, key string) (*string, error) {
	values, exists := form.Value[key]
	if !exists {
		return nil, fmt.Errorf("required field '%s' is missing", key)
	}

	if len(values) == 0 || values[0] == "" {
		return nil, fmt.Errorf("field '%s' cannot be empty", key)
	}

	return &values[0], nil
}

func (app *application) getFormFileOptional(r *http.Request, key string) (multipart.File, *multipart.FileHeader, error) {
	file, header, err := r.FormFile(key)
	if err != nil {
		if err == http.ErrMissingFile {
			return nil, nil, nil // Not an error, just missing
		}
		return nil, nil, fmt.Errorf("failed to get file '%s': %w", key, err)
	}
	return file, header, nil
}

func (app *application) getFormFileRequired(r *http.Request, key string) (multipart.File, *multipart.FileHeader, error) {
	file, header, err := r.FormFile(key)
	if err != nil {
		if err == http.ErrMissingFile {
			return nil, nil, fmt.Errorf("required file field '%s' is missing", key)
		}
		return nil, nil, fmt.Errorf("failed to get file '%s': %w", key, err)
	}
	return file, header, nil
}
