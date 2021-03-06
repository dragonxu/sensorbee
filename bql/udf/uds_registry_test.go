package udf

import (
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/sensorbee/sensorbee.v0/core"
	"gopkg.in/sensorbee/sensorbee.v0/data"
	"testing"
)

type testSharedState struct {
}

func (s *testSharedState) Terminate(ctx *core.Context) error {
	return nil
}

func TestEmptyDefaultUDSCreatorRegistry(t *testing.T) {
	Convey("Given an empty default UDS registry", t, func() {
		r := NewDefaultUDSCreatorRegistry()

		Convey("When adding a creator function", func() {
			err := r.Register("test_state_func", UDSCreatorFunc(func(ctx *core.Context, params data.Map) (core.SharedState, error) {
				return &testSharedState{}, nil
			}))

			Convey("Then it should succeed", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("When looking up a nonexistent creator", func() {
			_, err := r.Lookup("test_state_func")

			Convey("Then it should fail", func() {
				So(core.IsNotExist(err), ShouldBeTrue)
			})
		})

		Convey("When retrieving a list of creators", func() {
			m, err := r.List()

			Convey("Then it should succeed", func() {
				So(err, ShouldBeNil)

				Convey("And the list should be empty", func() {
					So(m, ShouldBeEmpty)
				})
			})
		})

		Convey("When unregistering a nonexistent creator", func() {
			err := r.Unregister("test_state_func")

			Convey("Then it should fail", func() {
				So(core.IsNotExist(err), ShouldBeTrue)
			})
		})
	})
}

func TestDefaultUDSCreatorRegistry(t *testing.T) {
	ctx := core.NewContext(nil)

	Convey("Given an default UDS registry having two types", t, func() {
		r := NewDefaultUDSCreatorRegistry()
		So(r.Register("TEST_state_func", UDSCreatorFunc(func(ctx *core.Context, params data.Map) (core.SharedState, error) {
			return &testSharedState{}, nil
		})), ShouldBeNil)
		So(r.Register("TEST_state_func2", UDSCreatorFunc(func(ctx *core.Context, params data.Map) (core.SharedState, error) {
			return &testSharedState{}, nil
		})), ShouldBeNil)

		Convey("When adding a new type having the registered type name", func() {
			err := r.Register("test_STATE_FUNC", UDSCreatorFunc(func(ctx *core.Context, params data.Map) (core.SharedState, error) {
				return &testSharedState{}, nil
			}))

			Convey("Then it should fail", func() {
				So(err, ShouldNotBeNil)
			})
		})

		Convey("When looking up a creator", func() {
			c, err := r.Lookup("test_state_FUNC2")

			Convey("Then it should succeed", func() {
				So(err, ShouldBeNil)

				Convey("And it should have the expected type", func() {
					s, err := c.CreateState(ctx, nil)
					So(err, ShouldBeNil)
					So(s, ShouldHaveSameTypeAs, &testSharedState{})
				})
			})
		})

		Convey("When retrieving a list of creators", func() {
			m, err := r.List()

			Convey("Then it should succeed", func() {
				So(err, ShouldBeNil)

				Convey("And the list should have all creators", func() {
					So(len(m), ShouldEqual, 2)
					So(m["test_state_func"], ShouldNotBeNil)
					So(m["test_state_func2"], ShouldNotBeNil)
				})
			})
		})

		Convey("When unregistering a creator", func() {
			err := r.Unregister("TEST_STATE_FUNC")

			Convey("Then it should succeed", func() {
				So(err, ShouldBeNil)

				Convey("And the unregistered creator shouldn't be found", func() {
					_, err := r.Lookup("test_state_func")
					So(core.IsNotExist(err), ShouldBeTrue)
				})

				Convey("And the other creator should be found", func() {
					_, err := r.Lookup("test_state_func2")
					So(err, ShouldBeNil)
				})
			})
		})
	})
}
