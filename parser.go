package json

import (
    "fmt"
    "errors"
)

// ErrorScan says error while sanning input JSON text.
var ErrorScan = errors.New("json.scanError")

// NumberKind to parse JSON numbers.
type NumberKind byte

const (
    // StringNumber will keep the number text as is.
    StringNumber NumberKind = iota + 1
    // IntNumber will use str.Atoi to parse JSON numbers.
    IntNumber
    // FloatNumber will use strconv.ParseFloat to parse JSON numbers.
    FloatNumber
)

// SpaceKind to skip white-spaces in JSON text.
type SpaceKind byte

const (
    // AnsiSpace will skip white space characters defined by ANSI spec.
    AnsiSpace SpaceKind = iota + 1
    // UnicodeSpace will skip white space characters defined by Unicode spec.
    UnicodeSpace
)

// Parser to parse input JSON text to go native representation.
type Parser struct {
    nk NumberKind
    ws SpaceKind
    jsonp bool
}

// NewParser with specified NumberKind and SpaceKind.
func NewParser(nk NumberKind, ws SpaceKind, jsonp bool) *Parser {
    return &Parser{nk: nk, ws: ws, jsonp: jsonp}
}

// Parse input JSON text to go native representation. `txt` is expected to
// contain a single JSON data.
func (p *Parser) Parse(txt []byte) (interface{}, []string, error) {
    i := len(txt)
    txt = skipWS(txt, p.ws)
    tok, txt, pointers, err := scanToken(txt, p.nk, p.ws, p.jsonp)
    if err != nil {
        j := len(txt)
        err = fmt.Errorf("error `%v` before %v", err, i-j)
        return nil, nil, err
    }
    return tok, pointers, nil
}

// ParseMany will parse input JSON text to one or more go native
// representation.
func (p *Parser) ParseMany(txt []byte) ([]interface{}, [][]string, error) {
    ms, ps := make([]interface{}, 0), make([][]string, 0)

    i := len(txt)
    for len(txt) > 0 {
        txt = skipWS(txt, p.ws)
        tok, txt, pointers, err := scanToken(txt, p.nk, p.ws, p.jsonp)
        if err != nil {
            j := len(txt)
            err = fmt.Errorf("error `%v` before %v", err, i-j)
            return nil, nil, err
        }
        ms = append(ms, tok)
        ps = append(ps, pointers)
    }
    return ms, ps, nil
}
