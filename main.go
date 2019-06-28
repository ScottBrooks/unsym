package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strings"
)

type SymFile struct {
	Input io.ReadSeeker

	Records []Record
}

type Record struct {
	Address uint64
	Line    uint32
	File    string
	Symbol  string
}

func (r Record) String() string {
	return fmt.Sprintf("0x%x %s:%d %s", r.Address, r.File, r.Line, r.Symbol)
}

func (s SymFile) Close() error {
	return nil
}

const MAXINT = math.MaxUint32

func ReadLine(in []byte) (string, error) {
	b := bytes.NewBuffer(in)
	str, err := b.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(str), nil
}

func (s *SymFile) Parse() error {
	var header RawRecordsHeader
	err := binary.Read(s.Input, binary.LittleEndian, &header)
	if err != nil {
		return err
	}
	records := make([]RawRecord, header.RecordCount)
	err = binary.Read(s.Input, binary.LittleEndian, &records)
	if err != nil {
		return err
	}

	s.Records = make([]Record, 0, header.RecordCount)

	stringTable, err := ioutil.ReadAll(s.Input)

	for _, r := range records {
		rec := Record{Address: r.Address, Line: r.LineNumber}
		if r.FileRelativeOffset != MAXINT {
			rec.File, err = ReadLine(stringTable[r.FileRelativeOffset:])
			if err != nil {
				return err
			}
		}
		if r.SymbolRelativeOffset != MAXINT {

			rec.Symbol, err = ReadLine(stringTable[r.SymbolRelativeOffset:])
			if err != nil {
				return err
			}

		}
		s.Records = append(s.Records, rec)
	}

	return nil

}

func (s SymFile) Dump(n int) {

	log.Printf("Records: %d", len(s.Records))
	for i := 0; i < n; i++ {
		log.Printf("Rec: %+v", s.Records[i])
	}

}

func Abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (s SymFile) DumpAddr(addr uint64) {
	closestDelta := MAXINT
	closestIdx := MAXINT
	found := false
	for i, r := range s.Records {
		delta := Abs(int(r.Address) - int(addr))
		if delta < closestDelta {
			closestDelta = delta
			closestIdx = i
		}
		if r.Address == addr {
			fmt.Printf("%s\n", r)
			found = true
		}
	}

	if !found {
		fmt.Printf("Best Guess: %d %s\n", closestDelta, s.Records[closestIdx])

	}

}

type RawRecordsHeader struct {
	RecordCount uint32
}

type RawRecord struct {
	Address              uint64
	LineNumber           uint32
	FileRelativeOffset   uint32
	SymbolRelativeOffset uint32
}

func main() {
	fmt.Println("Unsym")

	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	sf := SymFile{Input: f}

	err = sf.Parse()
	if err != nil {
		log.Fatal(err)
	}

	/*
	 */

	baseAddresses := []uint64{0x200000}

	for _, ba := range baseAddresses {
		fmt.Printf("Using base address: %x\n", ba)
		sf.DumpAddr(0x5ecaa05 - ba)
		sf.DumpAddr(0x5f5fafd - ba)
		sf.DumpAddr(0x5f69350 - ba)
		sf.DumpAddr(0x31d8b50 - ba)
		sf.DumpAddr(0x5f2b8b6 - ba)
		sf.DumpAddr(0x60109f8 - ba)
		sf.DumpAddr(0x600fb67 - ba)
		sf.DumpAddr(0x60525b6 - ba)
		sf.DumpAddr(0x60520a4 - ba)
		sf.DumpAddr(0x6021c60 - ba)
		sf.DumpAddr(0x6020936 - ba)

		/*
			sf.DumpAddr(0x3284e7b - ba)
			sf.DumpAddr(0x3261074 - ba)
			sf.DumpAddr(0x3260098 - ba)
			sf.DumpAddr(0x325ff41 - ba)
			sf.DumpAddr(0x32ce837 - ba)
			sf.DumpAddr(0x32b2059 - ba)
			log.Printf("------")
		*/
	}

}
