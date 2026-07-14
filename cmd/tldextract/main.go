package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/jenic/go-tldextract"
)

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	flags := flag.NewFlagSet("tldextract", flag.ContinueOnError)
	flags.SetOutput(stderr)
	punycode := flags.Bool("punycode", false, "output registrable domains in ASCII (IDNA/punycode)")
	if err := flags.Parse(args); err != nil {
		return err
	}

	in := stdin
	var file *os.File
	if flags.NArg() > 0 && flags.Arg(0) != "-" {
		opened, err := os.Open(flags.Arg(0))
		if err != nil {
			return fmt.Errorf("open: %w", err)
		}
		file = opened
		in = file
		defer file.Close()
	}

	var inputs []string
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		inputs = append(inputs, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read: %w", err)
	}

	domains, err := tldextract.ExtractAll(inputs, tldextract.Options{Punycode: *punycode})
	if err != nil {
		return err
	}
	for _, domain := range domains {
		if _, err := fmt.Fprintln(stdout, domain); err != nil {
			return fmt.Errorf("write: %w", err)
		}
	}
	return nil
}
