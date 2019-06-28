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
	"bufio"
	"strings"
	"strconv"
)

type SymFile struct {
	Input io.ReadSeeker

	Records []Record
}

type StackFile struct {
	Input io.Reader
	Pid int 
	Stacks []CallStack 
}

type CallStack struct {
	Tid int
	Calls []Call
}

type Call struct {
	Addr uint64
}


func (s *StackFile) Parse() error{
	var err error
	var c *CallStack

	scan := bufio.NewScanner(s.Input)

	scan.Scan()
	line := scan.Text()
	fields := strings.Fields(line)
	if fields[0] == "PID" {

		pid, err := strconv.Atoi(fields[1])
		if err != nil {
			return fmt.Errorf("Could not convert PID to number: %s: %s", fields[1], err)
		}
		s.Pid = pid
	}

	if fields[0] == "TID" {
		c = &CallStack{}
		c.Tid, err = strconv.Atoi(strings.TrimRight(fields[1], ":"))
		if err != nil {
			return fmt.Errorf("Unable to parse TID: %s %s %s", fields[1], line, err)
		}

	}


	for scan.Scan() {
		line = scan.Text()

		fields = strings.Fields(line)
		if fields[0] == "TID" {
			if c != nil {
				s.Stacks = append(s.Stacks, *c)
			}
			c = &CallStack{}
			c.Tid, err = strconv.Atoi(strings.TrimRight(fields[1], ":"))
			if err != nil {
				return fmt.Errorf("Unable to parse TID: %s %s %s", fields[1], line, err)
			}
		}
		if fields[0][0] == '#' {
			addr, err := strconv.ParseUint(strings.Replace(fields[1], "0x", "", -1), 16, 64)
			if err != nil {
				log.Fatal(err)
			}

			c.Calls = append(c.Calls, Call{Addr: addr})
		}
	}
	s.Stacks = append(s.Stacks, *c)


	return nil

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

func (s SymFile) LookupAddr(addr uint64) string {
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
			return r.String()
			found = true
		}
	}

	if !found && closestIdx != MAXINT{
		return "G: "+ s.Records[closestIdx].String()
	}

	return ""
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

	if len(os.Args ) != 3 {
		log.Fatalf("Expected use: eu-stacktrace | ./unsym UnrealrealServer.sym 0x200000")
	}

	baseAddr, err := strconv.ParseUint(strings.Replace(os.Args[2], "0x", "", -1), 16, 64)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Base Address: 0x%x\n", baseAddr)

	stackFile := StackFile{Input: os.Stdin}

	err = stackFile.Parse()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Stacks for process: %d\n", stackFile.Pid)
	for _, thread := range stackFile.Stacks {
		fmt.Printf("\tThreadID: %d\n", thread.Tid)
		for _, call := range thread.Calls {
			delta := call.Addr - uint64(baseAddr)
			lookup := sf.LookupAddr(delta)
			fmt.Printf("\t\t0x%x %s\n", delta, lookup)
		}
	}
}
