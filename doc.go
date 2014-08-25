package json

import (
    "encoding/gob"
    "errors"
    "strconv"
    "io"
)

// error codes

// ErrorInvalidDocumentText is returned for misconstructed JSON text.
var ErrorInvalidDocumentText = errors.New("json.invalidDocumentText")

// ErrorEmptyDocument is returned when Document does not contain a valid value.
var ErrorEmptyDocument = errors.New("json.emptyDocument")

// Document object for JSON.
type Document struct {
    Docid    []byte
    M        map[string]interface{}
    // local variables
    pointers []string
    // config
    nk       NumberKind // how to save number in `M`
    ws       SpaceKind  // how to parse whitespace
    jsonp    bool       // whether to gather json-pointers while parsing
}

// NewDocument constructs a new Document object for {id, value} pair. `value`
// can be JSON text or go-native representation, map[string]interface{}, of
// parsed JSON object.
func NewDocument(id []byte, value interface{}) (doc *Document, err error) {
    if value == nil {
        return
    }

    var m interface{}
    var ok bool

    doc = &Document{Docid: id, nk: FloatNumber, ws: AnsiSpace}
    switch v := value.(type) {
    case []byte:
        p := NewParser(doc.nk, doc.ws, false /*jsonp*/)
        m, doc.pointers, err = p.Parse(v)
        if doc.M, ok = m.(map[string]interface{}); !ok {
            err = ErrorInvalidDocumentText
        }

    case map[string]interface{}:
        doc.M = v
    }
    return
}

// NewDocuments constructs a list of Document object for {_, value} pair.
// Document-id shall be populated by the caller.
func NewDocuments(txt []byte) (docs []*Document, err error) {
    p := NewParser(FloatNumber, AnsiSpace, false /*jsonp*/)
    ms, _, err := p.ParseMany(txt)
    if err != nil {
        return nil, err
    }

    var doc *Document

    docs = make([]*Document, 0, len(ms))
    for _, m := range ms {
        if M, ok := m.(map[string]interface{}); m != nil && ok {
            if doc, err = NewDocument(nil, M); err != nil {
                return nil, err
            }
            docs = append(docs, doc)
        }
    }
    return
}

func (doc *Document) ListPointers() ([]string, error) {
    if doc == nil || doc.M == nil {
        return nil, ErrorEmptyDocument
    }
    return traverseObject(doc.M), nil
}

func traverseObject(obj interface{}) []string {
    var pointers []string

    switch v := obj.(type) {
    case []interface{}:
        if len(v) > 0 {
            pointers = make([]string, 0, 4)
            for i, value := range v {
                prefix := "/" + strconv.Itoa(i)
                for _, pointer := range traverseObject(value) {
                    pointers = append(pointers, prefix + pointer)
                }
            }
        }

    case map[string]interface{}:
        pointers = make([]string, 0, 4)
        pointers = append(pointers, "")
        if len(v) > 0 {
            for key, value := range v {
                prefix := "/" + key
                for _, pointer := range traverseObject(value) {
                    pointers = append(pointers, prefix + pointer)
                }
            }
        }
    }
    return pointers
}

// GobEncode shall marshal a Document stucture (only exported fields) to byte
// array, suitable for transmission over wire.
func (doc *Document) GobEncode(buf io.Writer) error {
    e := gob.NewEncoder(buf)
    return e.Encode(doc.M)
}

// GobDecode shall unmarshal byte array returned by GobEncode() back to
// Document structure.
func (doc *Document) GobDecode(buf io.Reader) error {
    d := gob.NewDecoder(buf)
    return d.Decode(&doc.M)
}

// Factory object to construct `Document` structure.
type Factory struct {
    p     *Parser
    // config
    nk    NumberKind // how to save number in `M`
    ws    SpaceKind  // how to parse whitespace
    jsonp bool       // gather json-pointers
}

// NewFactory create a factory with specified parameters.
func NewFactory(nk NumberKind, ws SpaceKind, jsonp bool) *Factory {
    factory := &Factory{nk: nk, ws:ws, jsonp: jsonp}
    factory.p = NewParser(nk, ws, jsonp)
    return factory
}

// NewDocument will use the factory to create a Document object for {id, value}
// pair.
func (f *Factory) NewDocument(id []byte, value interface{}) (doc *Document, err error) {
    if value == nil {
        return
    }

    var m interface{}
    var ok bool

    doc = &Document{Docid: id, nk: f.nk, ws: f.ws}
    switch v := value.(type) {
    case []byte:
        m, doc.pointers, err = f.p.Parse(v)
        if doc.M, ok = m.(map[string]interface{}); !ok {
            err = ErrorInvalidDocumentText
        }
    case map[string]interface{}:
        doc.M = v
    }
    return
}
