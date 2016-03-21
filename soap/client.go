// Package soap provides a SOAP HTTP client.
package soap

import (
	"bytes"
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
)

// A RoundTripper executes a request passing the given req as the SOAP
// envelope body. The HTTP response is then de-serialized onto the resp
// object. Returns error in case an error occurs serializing req, making
// the HTTP request, or de-serializing the response.
type RoundTripper interface {
	RoundTrip(req, resp Message) error
}

// Message is an opaque type used by the RoundTripper to carry XML
// documents for SOAP.
type Message interface{}

// Header is an opaque type used as the SOAP Header element in requests.
type Header interface{}

// AuthHeader is a Header to be encoded as the SOAP Header element in
// requests, to convey credentials for authentication.
type AuthHeader struct {
	Namespace string `xml:"xmlns:ns,attr"`
	Username  string `xml:"ns:username"`
	Password  string `xml:"ns:password"`
}

// Client is a SOAP client.
type Client struct {
	URL         string       // URL of the server
	Namespace   string       // SOAP Namespace
	Envelope    string       // Optional SOAP Envelope
	Header      Header       // Optional SOAP Header
	ContentType string       // Optional Content-Type (default text/xml)
	Config      *http.Client // Optional HTTP client
}

// RoundTrip implements the RoundTripper interface.
func (c *Client) RoundTrip(in, out Message) error {
	req := &Envelope{
		EnvelopeAttr: c.Envelope,
		NSAttr:       c.Namespace,
		Header:       c.Header,
		Body:         Body{Message: in},
	}
	if req.EnvelopeAttr == "" {
		req.EnvelopeAttr = "http://schemas.xmlsoap.org/soap/envelope/"
	}
	if req.NSAttr == "" {
		req.NSAttr = c.URL
	}
	var b bytes.Buffer
	err := xml.NewEncoder(&b).Encode(req)
	if err != nil {
		return err
	}
	ct := c.ContentType
	if ct == "" {
		ct = "text/xml"
	}
	cli := c.Config
	if cli == nil {
		cli = http.DefaultClient
	}
	resp, err := cli.Post(c.URL, ct, &b)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if false {
		// to be removed
		z, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		log.Printf("%s", z)
	}
	return xml.NewDecoder(resp.Body).Decode(out)
}

// Envelope is a SOAP envelope.
type Envelope struct {
	XMLName      xml.Name `xml:"SOAP-ENV:Envelope"`
	EnvelopeAttr string   `xml:"xmlns:SOAP-ENV,attr"`
	NSAttr       string   `xml:"xmlns:ns,attr"`
	Header       Message  `xml:"SOAP-ENV:Header"`
	Body         Body
}

// Body is the body of a SOAP envelope.
type Body struct {
	XMLName xml.Name `xml:"SOAP-ENV:Body"`
	Message Message
}
