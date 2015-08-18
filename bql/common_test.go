package bql

import (
	"errors"
	"pfi/sensorbee/sensorbee/bql/parser"
	"pfi/sensorbee/sensorbee/bql/udf"
	"pfi/sensorbee/sensorbee/core"
	"pfi/sensorbee/sensorbee/data"
	"sync"
)

func newTestTopology() core.Topology {
	return core.NewDefaultTopology(core.NewContext(nil), "testTopology")
}

func addBQLToTopology(tb *TopologyBuilder, bql string) error {
	p := parser.New()
	// execute all parsed statements
	stmts, err := p.ParseStmts(bql)
	if err != nil {
		return err
	}
	for _, stmt := range stmts {
		_, err := tb.AddStmt(stmt)
		if err != nil {
			return err
		}
	}
	return nil
}

type dummyUDS struct {
	num int64
}

func newDummyUDS(ctx *core.Context, params data.Map) (core.SharedState, error) {
	s := &dummyUDS{}
	if v, ok := params["num"]; ok {
		if n, err := data.ToInt(v); err != nil {
			return nil, err
		} else {
			s.num = n
		}
	}
	return s, nil
}

func (s *dummyUDS) Terminate(ctx *core.Context) error {
	return nil
}

type dummyUpdatableUDS struct {
	*dummyUDS
}

func (s *dummyUpdatableUDS) Update(params data.Map) error {
	return nil
}

func newDummyUpdatableUDS(ctx *core.Context, params data.Map) (core.SharedState, error) {
	state, _ := newDummyUDS(ctx, params)
	uds, _ := state.(*dummyUDS)
	s := &dummyUpdatableUDS{
		dummyUDS: uds,
	}
	return s, nil
}

func init() {
	udf.MustRegisterGlobalUDSCreator("dummy_uds", udf.UDSCreatorFunc(newDummyUDS))
	udf.MustRegisterGlobalUDSCreator("dummy_updatable_uds", udf.UDSCreatorFunc(newDummyUpdatableUDS))
}

type duplicateUDSF struct {
	dup int
}

func (d *duplicateUDSF) Process(ctx *core.Context, t *core.Tuple, w core.Writer) error {
	for i := 0; i < d.dup; i++ {
		w.Write(ctx, t.Copy())
	}
	return nil
}

func (d *duplicateUDSF) Terminate(ctx *core.Context) error {
	return nil
}

func createDuplicateUDSF(decl udf.UDSFDeclarer, stream string, dup int) (udf.UDSF, error) {
	if err := decl.Input(stream, &udf.UDSFInputConfig{
		InputName: "test",
	}); err != nil {
		return nil, err
	}

	return &duplicateUDSF{
		dup: dup,
	}, nil
}

func failingUDSFCreator(decl udf.UDSFDeclarer, stream string, dup int) (udf.UDSF, error) {
	return nil, errors.New("test UDSF creation failed")
}

func init() {
	udf.MustRegisterGlobalUDSFCreator("duplicate", udf.MustConvertToUDSFCreator(createDuplicateUDSF))
	udf.MustRegisterGlobalUDSFCreator("failing_duplicate", udf.MustConvertToUDSFCreator(failingUDSFCreator))
}

type sequenceUDSF struct {
	cnt int
}

// TODO: remove this WaitGroup after supporting pause/resume of streams
var (
	wgSequenceUDSF sync.WaitGroup
)

func (s *sequenceUDSF) Process(ctx *core.Context, t *core.Tuple, w core.Writer) error {
	wgSequenceUDSF.Wait()
	for i := 1; i <= s.cnt; i++ {
		if err := w.Write(ctx, core.NewTuple(data.Map{"int": data.Int(i)})); err != nil {
			if err == core.ErrSourceStopped {
				return err
			}
		}
	}
	return nil
}

func (s *sequenceUDSF) Terminate(ctx *core.Context) error {
	return nil
}

func createSequenceUDSF(decl udf.UDSFDeclarer, cnt int) (udf.UDSF, error) {
	return &sequenceUDSF{
		cnt: cnt,
	}, nil
}

func init() {
	udf.MustRegisterGlobalUDSFCreator("test_sequence", udf.MustConvertToUDSFCreator(createSequenceUDSF))
}
