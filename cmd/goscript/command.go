package main

import (
	"context"
	"flag"
	"fmt"
	"go/ast"
	"os"
	"strings"
	"tools/pkg/conv/goast"
	"tools/pkg/io/read"
	"tools/pkg/log"
	"tools/pkg/util"

	"github.com/google/subcommands"
)

func newLogger(verbose bool) *log.Logger {
	level := func() log.Level {
		if verbose {
			return log.Debug
		}
		return log.Info
	}()
	return log.NewLogger(log.WithLevel(level))
}

type parse struct {
	verbose bool
	quiet   bool
	logger  *log.Logger
}

func (*parse) Name() string {
	return "parse"
}

func (*parse) Synopsis() string {
	return "parse and print ast"
}

func (*parse) Usage() string {
	return `echo 'func(s string)(string, error){return strings.ToLower(s),nil}' | goscript parse`
}

func (s *parse) SetFlags(fs *flag.FlagSet) {
	fs.BoolVar(&s.verbose, "v", false, "verbose")
	fs.BoolVar(&s.quiet, "q", false, "quiet")
}

func (s *parse) execute() error {
	var err error
	src, err := read.Read(os.Stdin)
	if err != nil {
		return err
	}
	node, err := func() (ast.Node, error) {
		if x, err := goast.ParseExpr(src); err == nil {
			s.logger.Debug("parsed as expression")
			return x, nil
		}
		if x, err := goast.ParseExpr(fmt.Sprintf("func(){%s}", src)); err == nil {
			s.logger.Debug("parsed as statement list")
			return x.(*ast.FuncLit).Body, nil
		}
		if x, err := goast.ParseFile(src); err == nil {
			s.logger.Debug("parsed as file")
			return x, nil
		}
		x, err := goast.ParseFile(fmt.Sprintf("package main\n%s", src))
		if err != nil {
			return nil, err
		}
		s.logger.Debug("parsed as file added main package")
		x.Name = nil
		return x, nil
	}()
	if err != nil {
		return err
	}
	if s.quiet {
		return nil
	}
	return goast.Print(os.Stdout, node)
}

func (s *parse) Execute(_ context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	s.logger = newLogger(s.verbose)
	elapsed := util.Elapsed()
	err := s.execute()
	s.logger.Info("elapsed: %v err: %v", elapsed(), err)
	if err != nil {
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}

type funcPipe struct {
	importSpecs string
	verbose     bool
	logger      *log.Logger
}

func (*funcPipe) Name() string {
	return "func"
}

func (*funcPipe) Synopsis() string {
	return "generate main that consumes stdin and execute main proc"
}

func (*funcPipe) Usage() string {
	return `echo 'func(s string)(string,error){return strings.ToLower(s),nil}' | goscript func | goimports
output: go file with main package, func main()
`
}

func (s *funcPipe) SetFlags(fs *flag.FlagSet) {
	fs.BoolVar(&s.verbose, "v", false, "verbose")
	fs.StringVar(&s.importSpecs, "i", "", "packages separeted by space")
}

func (s *funcPipe) execute() error {
	src, err := read.Read(os.Stdin)
	if err != nil {
		return err
	}
	f, err := goast.ParseExpr(src)
	if err != nil {
		return err
	}
	var iSpecs []string
	if len(s.importSpecs) > 0 {
		iSpecs = strings.Split(s.importSpecs, " ")
	}
	r, err := translateFilePipeFromFuncLit(f, iSpecs)
	if err != nil {
		return err
	}
	return goast.Dump(os.Stdout, r)
}

func (s *funcPipe) Execute(_ context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	s.logger = newLogger(s.verbose)
	elapsed := util.Elapsed()
	err := s.execute()
	s.logger.Info("elapsed: %v err: %v", elapsed(), err)
	if err != nil {
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}

type mainPipe struct {
	mainProc string
	verbose  bool
	logger   *log.Logger
}

func (*mainPipe) Name() string {
	return "main"
}

func (*mainPipe) Synopsis() string {
	return "generate main that consumes stdin and execute main proc"
}

func (*mainPipe) Usage() string {
	return `cat tmp.go | goscript main | goimports
input: go file with main package, and func Main(string) (string, error), without func main()
output: go file with func main()
`
}

func (s *mainPipe) SetFlags(fs *flag.FlagSet) {
	fs.BoolVar(&s.verbose, "v", false, "verbose")
	fs.StringVar(&s.mainProc, "m", "Main", "function name of main procedure")
}

func (s *mainPipe) execute() error {
	var err error
	src, err := read.Read(os.Stdin)
	if err != nil {
		return err
	}
	f, err := goast.ParseFile(src)
	if err != nil {
		return err
	}
	r, err := translateMainPipe(f, s.mainProc)
	if err != nil {
		return err
	}
	return goast.Dump(os.Stdout, r)
}

func (s *mainPipe) Execute(_ context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	s.logger = newLogger(s.verbose)
	elapsed := util.Elapsed()
	err := s.execute()
	s.logger.Info("elapsed: %v err: %v", elapsed(), err)
	if err != nil {
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}
