package gson

import "encoding/gob"
import "strconv"
import "io"

// Document object for JSON.
//   meta fields,
//      _id     document id
type Document struct {
	// marshalled fields
	M map[string]interface{}
	// local fields
	pointers []string
	config   Config
}

// NewDocument constructs a new Document object for {id, value} pair.
// `value` can be JSON text or go-native representation,
// map[string]interface{}, of parsed JSON object.
func NewDocument(
	id []byte, value interface{}, config Config) (doc *Document, err error) {

	doc = &Document{config: config, M: make(map[string]interface{})}
	doc.M["_id"] = id

	if value == nil {
		return doc, nil
	}

	switch v := value.(type) {
	case []byte:
		txt := bytes2str(v)
		if m, _, err := config.Parse(txt); err != nil {
			return nil, err
		} else if val, ok := m.(map[string]interface{}); !ok {
			return nil, ErrorInvalidDocumentType
		} else {
			doc.Mixin(val)
		}

	case map[string]interface{}:
		doc.Mixin(v)

	default:
		return nil, ErrorInvalidValueType
	}
	return
}

// ID returns the document's id.
func (doc *Document) ID() []byte {
	return doc.M["_id"].([]byte)
}

// SetID will update the document's id.
func (doc *Document) SetID(id []byte) *Document {
	doc.M["_id"] = id
	return doc
}

// SetValue will update the document's value object.
func (doc *Document) SetValue(value map[string]interface{}) *Document {
	for _, metaf := range MetaFields {
		value[metaf] = doc.M[metaf]
	}
	doc.M = value
	return doc
}

// Mixin will update the document's value map into one or more
// supplied maps.
func (doc *Document) Mixin(vals ...map[string]interface{}) *Document {
	for _, val := range vals {
		for key, value := range val {
			doc.M[key] = value
		}
	}
	return doc
}

// ListPointers will compose json-pointers into document object.
// composed pointers will also include meta-fields, like, "/_id"
func (doc *Document) ListPointers() []string {
	if doc == nil || doc.M == nil {
		return nil
	}
	return traverseObject(doc.M)
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
					pointers = append(pointers, prefix+pointer)
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
					pointers = append(pointers, prefix+pointer)
				}
			}
		}
	}
	return pointers
}

// GobEncode shall marshal a Document stucture (only exported fields)
// to byte array, suitable for transmission over wire.
func (doc *Document) GobEncode(buf io.Writer) error {
	e := gob.NewEncoder(buf)
	return e.Encode(doc.M)
}

// GobDecode shall unmarshal byte array returned by GobEncode()
// back to Document structure.
func (doc *Document) GobDecode(buf io.Reader) error {
	d := gob.NewDecoder(buf)
	return d.Decode(&doc.M)
}
