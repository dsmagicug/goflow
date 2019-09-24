package webhooks

import (
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/http/httputil"
	"time"

	"github.com/nyaruka/goflow/flows"
	"github.com/nyaruka/goflow/flows/engine"
	"github.com/nyaruka/goflow/utils/dates"

	"github.com/pkg/errors"
)

const httpHeaderUserAgent = "User-Agent"

// response content-types that we'll fetch
var fetchResponseContentTypes = map[string]bool{
	"application/json":       true,
	"application/javascript": true,
	"application/xml":        true,
	"text/html":              true,
	"text/plain":             true,
	"text/xml":               true,
	"text/javascript":        true,
}

type provider struct {
	defaultUserAgent string
	maxBodyBytes     int
}

// NewService creates a new webhook service
func NewService(defaultUserAgent string, maxBodyBytes int) engine.WebhookService {
	return func(flows.Session) flows.WebhookProvider { return NewProvider(defaultUserAgent, maxBodyBytes) }
}

// NewProvider creates a new webhook provider
func NewProvider(defaultUserAgent string, maxBodyBytes int) flows.WebhookProvider {
	return &provider{defaultUserAgent: defaultUserAgent, maxBodyBytes: maxBodyBytes}
}

func (p *provider) Call(session flows.Session, request *http.Request, resthook string) (*flows.WebhookCall, error) {
	// if user-agent isn't set, use our default
	if request.Header.Get(httpHeaderUserAgent) == "" {
		request.Header.Set(httpHeaderUserAgent, p.defaultUserAgent)
	}

	dump, err := httputil.DumpRequestOut(request, true)
	if err != nil {
		return nil, err
	}

	start := dates.Now()
	response, err := session.Engine().HTTPClient().Do(request)
	timeTaken := dates.Now().Sub(start)

	if err != nil {
		return &flows.WebhookCall{
			URL:        request.URL.String(),
			Method:     request.Method,
			StatusCode: 0,
			Status:     flows.WebhookStatusConnectionError,
			Request:    dump,
			Response:   []byte(err.Error()),
		}, nil
	}

	return p.newCallFromResponse(dump, response, p.maxBodyBytes, timeTaken, resthook)
}

// creates a new call based on the passed in http response
func (p *provider) newCallFromResponse(requestTrace []byte, response *http.Response, maxBodyBytes int, timeTaken time.Duration, resthook string) (*flows.WebhookCall, error) {
	defer response.Body.Close()

	// save response trace without body which will be parsed separately
	responseTrace, err := httputil.DumpResponse(response, false)
	if err != nil {
		return nil, err
	}

	w := &flows.WebhookCall{
		URL:        response.Request.URL.String(),
		Method:     response.Request.Method,
		StatusCode: response.StatusCode,
		Status:     statusFromCode(response.StatusCode, resthook != ""),
		Request:    requestTrace,
		Response:   responseTrace,
		TimeTaken:  timeTaken,
		Resthook:   resthook,
	}

	// we will only read up to our max body bytes limit
	bodyReader := io.LimitReader(response.Body, int64(maxBodyBytes)+1)
	var bodySniffed []byte

	// hopefully we got a content-type header
	contentTypeHeader := response.Header.Get("Content-Type")
	contentType, _, _ := mime.ParseMediaType(contentTypeHeader)

	// but if not, read first 512 bytes to sniff the content-type
	if contentType == "" {
		bodySniffed = make([]byte, 512)
		bodyBytesRead, err := bodyReader.Read(bodySniffed)
		if err != nil && err != io.EOF {
			return nil, err
		}
		bodySniffed = bodySniffed[0:bodyBytesRead]

		contentType, _, _ = mime.ParseMediaType(http.DetectContentType(bodySniffed))
	}

	// only save response body's if we have a supported content-type
	saveBody := fetchResponseContentTypes[contentType]

	if saveBody {
		bodyBytes, err := ioutil.ReadAll(bodyReader)
		if err != nil {
			return nil, err
		}

		// if we have no remaining bytes, error because the body was too big
		if bodyReader.(*io.LimitedReader).N <= 0 {
			return nil, errors.Errorf("webhook response body exceeds %d bytes limit", maxBodyBytes)
		}

		if len(bodySniffed) > 0 {
			bodyBytes = append(bodySniffed, bodyBytes...)
		}

		w.Response = append(w.Response, bodyBytes...)
	} else {
		w.BodyIgnored = true
	}

	return w, nil
}

// determines the webhook status from the HTTP status code
func statusFromCode(code int, isResthook bool) flows.WebhookStatus {
	// https://zapier.com/developer/documentation/v2/rest-hooks/
	if isResthook && code == 410 {
		return flows.WebhookStatusSubscriberGone
	}
	if code/100 == 2 {
		return flows.WebhookStatusSuccess
	}
	return flows.WebhookStatusResponseError
}

var _ flows.WebhookProvider = (*provider)(nil)
