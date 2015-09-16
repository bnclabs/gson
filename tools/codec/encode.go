package main

import "flag"
import "fmt"
import "log"
import "os"

//import "strings"
import "encoding/hex"

import "github.com/prataprc/collatejson"

var options struct {
	encode    bool
	decode    bool
	overheads bool
	inp       string
}

var usageHelp = `Usage : codec [OPTIONS]
to encode specify -encode switch
to decode specify -decode switch and the hexdump as -inp
`

func argParse() {
	flag.StringVar(&options.inp, "inp", "", "input to encode")
	flag.BoolVar(&options.encode, "encode", false, "encode input")
	flag.BoolVar(&options.decode, "decode", false, "decode input")
	flag.BoolVar(&options.overheads, "overheads", false, "collation overhead")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, usageHelp)
		flag.PrintDefaults()
	}
	flag.Parse()
}

func main() {
	var err error

	argParse()
	codec := collatejson.NewCodec(100)
	out := make([]byte, 0, len(options.inp)*3+collatejson.MinBufferSize)
	if options.overheads {
		computeOverheads(codec)

	} else if options.encode {
		out, err = codec.Encode([]byte(options.inp), out)
		if err != nil {
			log.Fatal(err)
		}
		hexout := make([]byte, len(out)*5)
		n := hex.Encode(hexout, out)
		fmt.Printf("in : %q\n", options.inp)
		fmt.Printf("out: %q\n", string(out))
		fmt.Printf("hex: %q\n", string(hexout[:n]))

	} else if options.decode {
		inpbs := make([]byte, len(options.inp)*5)
		n, err := hex.Decode(inpbs, []byte(options.inp))
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(n, inpbs[:n])
		out, err = codec.Decode([]byte(inpbs[:n]), out)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("in : %q\n", options.inp)
		fmt.Printf("out: %q\n", string(out))
	}
}

func computeOverheads(codec *collatejson.Codec) {
	var err error

	items := []string{
		"10",
		"10000",
		"1000000000",
		"100000000000000000.0",
		"123456789123565670.0",
		"10.2",
		"10.23456789012",
		"null",
		"true",
		"false",
		`"hello world"`,
		`[ 10, 10000, 1000000000, 10.2, 10.23456789012, null, true, false, "hello world" ]`,
		`{ "a": 10000, "b": 10.23456789012, "c": null, "d": true, "e": false, "f": "hello world" }`,
	}
	for _, item := range items {
		out := make([]byte, 0, 1024)
		out, err = codec.Encode([]byte(item), out)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("item: %q\n", item)
		fmt.Println("  i/p len:", len(item))
		fmt.Println("  o/p len:", len(out))
	}
}
