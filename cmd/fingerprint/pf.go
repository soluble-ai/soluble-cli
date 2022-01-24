package fingerprint

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"

	"github.com/soluble-ai/soluble-cli/pkg/assessments/fingerprint"
	"github.com/spf13/cobra"
)

func Command() *cobra.Command {
	c := &cobra.Command{
		Use:    "pf",
		Short:  "Display partial fingerprints and diffs",
		Hidden: true,
	}
	c.AddCommand(showCommand(), diffCommand())
	return c
}

type fileFingerprints struct {
	lines             []string
	fingeprints       []string
	fingerprintToLine map[string]int
}

func newFileFingerprints(r io.Reader) (*fileFingerprints, error) {
	d, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	ff := &fileFingerprints{
		fingerprintToLine: map[string]int{},
	}
	sc := bufio.NewScanner(bytes.NewReader(d))
	for sc.Scan() {
		line := sc.Text()
		ff.lines = append(ff.lines, line)
	}
	ff.fingeprints = make([]string, len(ff.lines))
	b := bufio.NewReader(bytes.NewReader(d))
	err = fingerprint.Partial(b, func(i int, s string) {
		ff.fingeprints[i-1] = s
		ff.fingerprintToLine[s] = i
	})
	return ff, err
}

func newFileFingerprintsFromPath(p string) (*fileFingerprints, error) {
	f, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return newFileFingerprints(f)
}

func pf(r io.Reader) error {
	ff, err := newFileFingerprints(r)
	if err != nil {
		return err
	}
	for i := range ff.lines {
		fmt.Printf("%s %s\n", ff.fingeprints[i], ff.lines[i])
	}
	return nil
}

func maxInt(i, j int) int {
	if i > j {
		return i
	}
	return j
}

func pfSame(p1, p2 string) error {
	ff1, err := newFileFingerprintsFromPath(p1)
	if err != nil {
		return err
	}
	ff2, err := newFileFingerprintsFromPath(p2)
	if err != nil {
		return err
	}
	n := maxInt(len(ff1.lines), len(ff2.lines))
	h := int(math.Log10(float64(n+1))) + 1
	fingerprints := ff1.fingeprints
	if len(ff2.fingeprints) > len(fingerprints) {
		fingerprints = ff2.fingeprints
	}
	fmt.Printf("A %s\n", p1)
	fmt.Printf("B %s\n", p2)
	for _, f := range fingerprints {
		ln1 := ff1.fingerprintToLine[f] - 1
		ln2 := ff2.fingerprintToLine[f] - 1
		if ln1 >= 0 && ln2 >= 0 {
			fmt.Printf("%s A:%-*d B:%-*d %s\n", f, h, ln1, h, ln2, ff1.lines[ln1])
		}
	}
	return nil
}

func diffCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "diff",
		Short: "Display a diff and where partial fingerprint are the same in 2 files",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			// #nosec:G204
			diff := exec.Command("diff", "-u", args[0], args[1])
			diff.Stdout = os.Stdout
			diff.Stderr = os.Stderr
			if err := diff.Run(); err != nil {
				_, ok := err.(*exec.ExitError)
				if !ok {
					return err
				}
			}
			fmt.Println()
			return pfSame(args[0], args[1])
		},
	}
}

func showCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Display partial fingerprints",
		Args:  cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return pf(os.Stdin)
			}
			for _, path := range args {
				f, err := os.Open(path)
				if err != nil {
					return err
				}
				if err := pf(f); err != nil {
					return err
				}
				_ = f.Close()
			}
			return nil
		},
	}
}
