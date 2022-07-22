package libalp

import (
	"fmt"
	"io"
	"os"

	"github.com/tkuchiki/alp"
	"github.com/tkuchiki/alp/errors"
	"github.com/tkuchiki/alp/options"
	"github.com/tkuchiki/alp/parsers"
	"github.com/tkuchiki/alp/stats"
)

type AlpOption struct {
	File                    string
	SortType                string
	Reverse                 bool
	QueryString             bool
	QueryStringIgnoreValues bool
	DecodeUri               bool
	Format                  string
	Limit                   int
	Location                string
	Output                  string
	NoHeaders               bool
	ShowFooters             bool
	MatchingGroups          string
	Filters                 string
	Percentiles             []int
}

type AlpProfiler struct {
	profiler *alp.Profiler
	output   io.Writer
}

func NewAlpProfiler(outfile string) (*AlpProfiler, error) {
	f, err := os.Create(outfile)
	if err != nil {
		return nil, err
	}
	prof := alp.NewProfiler(f, os.Stderr)
	return &AlpProfiler{
		profiler: prof,
		output:   f,
	}, nil
}

func (ap *AlpProfiler) Run(opt AlpOption) error {
	p := ap.profiler

	sortOptions := stats.NewSortOptions()
	err := sortOptions.SetAndValidate(opt.SortType)
	if err != nil {
		return err
	}

	opts := options.NewOptions()
	opts = options.SetOptions(opts,
		options.File(opt.File),
		options.Sort(opt.SortType),
		options.Reverse(opt.Reverse),
		options.QueryString(opt.QueryString),
		options.QueryStringIgnoreValues(opt.QueryStringIgnoreValues),
		options.DecodeUri(opt.DecodeUri),
		options.Format(opt.Format),
		options.Limit(opt.Limit),
		options.Location(opt.Location),
		options.Output(opt.Output),
		options.NoHeaders(opt.NoHeaders),
		options.ShowFooters(opt.ShowFooters),
		options.CSVGroups(opt.MatchingGroups),
		options.Filters(opt.Filters),
		options.Percentiles(opt.Percentiles),
	)

	sts := stats.NewHTTPStats(true, false, false)

	err = sts.InitFilter(opts)
	if err != nil {
		return err
	}

	sts.SetOptions(opts)
	sts.SetSortOptions(sortOptions)

	printOptions := stats.NewPrintOptions(opts.NoHeaders, opts.ShowFooters, opts.DecodeUri, opts.PaginationLimit)
	printer := stats.NewPrinter(ap.output, opts.Output, opts.Format, opt.Percentiles, printOptions)
	if err = printer.Validate(); err != nil {
		return err
	}

	f, err := p.Open(opts.File)
	if err != nil {
		return err
	}
	defer f.Close()

	if len(opts.MatchingGroups) > 0 {
		err = sts.SetURIMatchingGroups(opts.MatchingGroups)
		if err != nil {
			return err
		}
	}

	var parser parsers.Parser
	label := parsers.NewLTSVLabel(opts.LTSV.UriLabel, opts.LTSV.MethodLabel, opts.LTSV.TimeLabel,
		opts.LTSV.ApptimeLabel, opts.LTSV.ReqtimeLabel, opts.LTSV.SizeLabel, opts.LTSV.StatusLabel,
	)
	parser = parsers.NewLTSVParser(f, label, opts.QueryString, opts.QueryStringIgnoreValues)

Loop:
	for {
		s, err := parser.Parse()
		if err != nil {
			if err == io.EOF {
				break
			} else if err == errors.SkipReadLineErr {
				continue Loop
			}

			return err
		}

		var b bool
		b, err = sts.DoFilter(s)
		if err != nil {
			return err
		}

		if !b {
			continue Loop
		}

		sts.Set(s.Uri, s.Method, s.Status, s.ResponseTime, s.BodyBytes, 0)

		if sts.CountUris() > opts.Limit {
			return fmt.Errorf("too many URI's (%d or less)", opts.Limit)
		}
	}

	sts.SortWithOptions()
	printer.Print(sts, nil)

	return nil
}
