package mbox

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"strings"
	"time"

	"takeout_services/internal/domain/model"
	"takeout_services/internal/ports/outbound"
)

// Parser implements outbound.MboxParser.
type Parser struct{}

// NewParser creates a new instance of Parser.
func NewParser() outbound.MboxParser {
	return &Parser{}
}

var wordDecoder = mime.WordDecoder{}

func decodeHeader(s string) string {
	decoded, err := wordDecoder.DecodeHeader(s)
	if err != nil {
		return s
	}
	return decoded
}

func parseAddresses(headerVal string) []string {
	if headerVal == "" {
		return nil
	}
	list, err := mail.ParseAddressList(headerVal)
	if err != nil {
		// Fallback to simple split if mail.ParseAddressList fails
		parts := strings.Split(headerVal, ",")
		var res []string
		for _, p := range parts {
			trimmed := strings.TrimSpace(p)
			if trimmed != "" {
				res = append(res, trimmed)
			}
		}
		return res
	}
	var res []string
	for _, addr := range list {
		res = append(res, addr.String())
	}
	return res
}

func decodeBody(r io.Reader, encoding string) io.Reader {
	encoding = strings.ToLower(strings.TrimSpace(encoding))
	switch encoding {
	case "quoted-printable":
		return quotedprintable.NewReader(r)
	case "base64":
		// Strip newlines/whitespace which is common in base64 encoded mail bodies
		// standard base64 decoder in go is strict about characters but StdEncoding handles it if wrapped properly.
		// base64.NewDecoder expects base64 stream. Standard base64 lines are wrapped in CRLF.
		return base64.NewDecoder(base64.StdEncoding, newLaxBase64Reader(r))
	default:
		return r
	}
}

// laxBase64Reader strips whitespaces/newlines from standard reader stream
// as Go's base64 decoder is strict and might error on newlines.
type laxBase64Reader struct {
	r io.Reader
}

func newLaxBase64Reader(r io.Reader) io.Reader {
	return &laxBase64Reader{r: r}
}

func (l *laxBase64Reader) Read(p []byte) (int, error) {
	buf := make([]byte, len(p))
	n, err := l.r.Read(buf)
	if n == 0 {
		return 0, err
	}

	w := 0
	for i := 0; i < n; i++ {
		c := buf[i]
		if c != '\r' && c != '\n' && c != ' ' && c != '\t' {
			p[w] = c
			w++
		}
	}
	return w, err
}

func parseMessageBody(msg *mail.Message) (string, string, []model.Attachment, error) {
	contentType := msg.Header.Get("Content-Type")
	encoding := msg.Header.Get("Content-Transfer-Encoding")

	if contentType == "" {
		bodyReader := decodeBody(msg.Body, encoding)
		bodyBytes, err := io.ReadAll(bodyReader)
		if err != nil {
			return "", "", nil, err
		}
		return string(bodyBytes), "", nil, nil
	}

	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		bodyReader := decodeBody(msg.Body, encoding)
		bodyBytes, err := io.ReadAll(bodyReader)
		if err != nil {
			return "", "", nil, err
		}
		return string(bodyBytes), "", nil, nil
	}

	if strings.HasPrefix(mediaType, "multipart/") {
		boundary := params["boundary"]
		if boundary == "" {
			return "", "", nil, fmt.Errorf("multipart message missing boundary")
		}
		return parseMultipart(msg.Body, boundary)
	}

	bodyReader := decodeBody(msg.Body, encoding)
	bodyBytes, err := io.ReadAll(bodyReader)
	if err != nil {
		return "", "", nil, err
	}

	if mediaType == "text/html" {
		return "", string(bodyBytes), nil, nil
	}
	return string(bodyBytes), "", nil, nil
}

func parseMultipart(r io.Reader, boundary string) (string, string, []model.Attachment, error) {
	var bodyText, bodyHTML string
	var attachments []model.Attachment

	mr := multipart.NewReader(r, boundary)
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return bodyText, bodyHTML, attachments, err
		}

		partType := part.Header.Get("Content-Type")
		partMediaType, partParams, _ := mime.ParseMediaType(partType)

		disposition := part.Header.Get("Content-Disposition")
		dispType, dispParams, _ := mime.ParseMediaType(disposition)

		filename := dispParams["filename"]
		if filename == "" {
			filename = partParams["name"]
		}

		// Check if it's an attachment
		if dispType == "attachment" || filename != "" {
			partEncoding := part.Header.Get("Content-Transfer-Encoding")
			partReader := decodeBody(part, partEncoding)
			data, err := io.ReadAll(partReader)
			if err != nil {
				return bodyText, bodyHTML, attachments, err
			}
			attachments = append(attachments, model.Attachment{
				Filename:    decodeHeader(filename),
				ContentType: partType,
				Size:        int64(len(data)),
				Data:        data,
			})
			continue
		}

		// Recursively parse sub-multiparts
		if strings.HasPrefix(partMediaType, "multipart/") {
			subBoundary := partParams["boundary"]
			if subBoundary != "" {
				subText, subHTML, subAtt, err := parseMultipart(part, subBoundary)
				if err == nil {
					if subText != "" {
						bodyText += subText
					}
					if subHTML != "" {
						bodyHTML += subHTML
					}
					attachments = append(attachments, subAtt...)
				}
			}
			continue
		}

		// Regular text / html parts
		partEncoding := part.Header.Get("Content-Transfer-Encoding")
		partReader := decodeBody(part, partEncoding)
		partBody, err := io.ReadAll(partReader)
		if err != nil {
			return bodyText, bodyHTML, attachments, err
		}

		if partMediaType == "text/html" {
			bodyHTML += string(partBody)
		} else {
			bodyText += string(partBody)
		}
	}

	return bodyText, bodyHTML, attachments, nil
}

func parseRawMessage(raw []byte) (*model.Email, error) {
	msg, err := mail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		return nil, err
	}

	from := decodeHeader(msg.Header.Get("From"))
	subject := decodeHeader(msg.Header.Get("Subject"))
	rawDate := msg.Header.Get("Date")
	messageID := msg.Header.Get("Message-ID")

	var parsedDate time.Time
	if rawDate != "" {
		parsedDate, _ = mail.ParseDate(rawDate)
	}

	to := parseAddresses(decodeHeader(msg.Header.Get("To")))
	cc := parseAddresses(decodeHeader(msg.Header.Get("Cc")))
	bcc := parseAddresses(decodeHeader(msg.Header.Get("Bcc")))

	bodyText, bodyHTML, attachments, err := parseMessageBody(msg)
	if err != nil {
		// Continue even if body parsing yields errors
	}

	headers := make(map[string]string)
	for key, values := range msg.Header {
		if len(values) > 0 {
			headers[key] = decodeHeader(values[0])
		}
	}

	id := messageID
	if id == "" {
		hash := sha256.New()
		hash.Write([]byte(from + rawDate + subject))
		if len(bodyText) > 100 {
			hash.Write([]byte(bodyText[:100]))
		} else {
			hash.Write([]byte(bodyText))
		}
		id = hex.EncodeToString(hash.Sum(nil))
	} else {
		id = strings.Trim(id, "<>")
	}

	return model.NewEmail(
		id,
		strings.Trim(messageID, "<>"),
		from,
		to,
		cc,
		bcc,
		subject,
		parsedDate,
		rawDate,
		bodyText,
		bodyHTML,
		headers,
		attachments,
	), nil
}

// Parse splits the mbox input stream and processes each message via the callback function.
func (p *Parser) Parse(r io.Reader, callback func(*model.Email) error) error {
	reader := bufio.NewReader(r)
	var currentMsg bytes.Buffer
	var hasMsg bool

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return err
		}

		isEOF := (err == io.EOF)
		isFromLine := false

		if bytes.HasPrefix(line, []byte("From ")) {
			// To avoid splitting on a body line starting with "From ", mboxo/mboxrd standard states
			// the From line should start with "From " and usually does not have a colon.
			isFromLine = true
		}

		if isFromLine || isEOF {
			if hasMsg {
				rawBytes := currentMsg.Bytes()
				// Strip the trailing blank line that separates mbox messages.
				if bytes.HasSuffix(rawBytes, []byte("\n\n")) {
					rawBytes = rawBytes[:len(rawBytes)-1]
				} else if bytes.HasSuffix(rawBytes, []byte("\r\n\r\n")) {
					rawBytes = rawBytes[:len(rawBytes)-2]
				}

				email, parseErr := parseRawMessage(rawBytes)
				if parseErr == nil && email != nil {
					if err := callback(email); err != nil {
						return err
					}
				}
				currentMsg.Reset()
			}
			hasMsg = true
			if isEOF {
				break
			}
			// The mbox envelope separator line (From ...) is not part of the standard raw RFC 822 email,
			// so we don't feed it to net/mail.
			continue
		}

		if hasMsg {
			currentMsg.Write(line)
		}
	}

	return nil
}
