package subtest_test

import (
	"encoding/json"
	"regexp"
	"testing"

	"github.com/searis/subtest"
)

func TestDeepEqual(t *testing.T) {
	t.Run("given check DeepEqual(true)", func(t *testing.T) {
		cf := subtest.DeepEqual(true)
		t.Run("when cheking against true", func(t *testing.T) {
			vf := subtest.Value(cf(true))
			t.Run("then there should be no failure", vf.NoError())
		})
		t.Run("when cheking against false", func(t *testing.T) {
			vf := subtest.Value(cf(false))
			t.Run("then an appropriate error should be returned", vf.ErrorIs(subtest.Failure{
				Prefix: "values are not deep equal",
				Got:    "bool\n\tfalse",
				Expect: "bool\n\ttrue",
			}))
		})
		t.Run("when testing against true with different syntax", func(t *testing.T) {
			v := true
			vf := subtest.Value(v)
			t.Run("then <ValueFunc>.<CheckFuncName> should pass", vf.DeepEqual(true))
			t.Run("the <ValueFunc>.Test(<CheckFunc>) should pass", vf.Test(subtest.DeepEqual(true))) // equivalent
		})
	})

	t.Run("given a nested struct type T", func(t *testing.T) {
		type T struct {
			A string
			B map[string]string
		}

		t.Run("when checking a non-nil *T value", func(t *testing.T) {
			v := &T{A: "a", B: map[string]string{"C": "D"}}

			t.Run("then it should match an equivalent *T value", // equal value.
				subtest.Value(v).DeepEqual(&T{A: "a", B: map[string]string{"C": "D"}}),
			)
			t.Run("then it should not match a different *T value", // different value.
				subtest.Value(subtest.DeepEqual(&T{A: "a", B: map[string]string{"C": "E"}})(v)).Error(),
			)
			t.Run("then it should not match an equivalent T value", // equal value, different type
				subtest.Value(subtest.DeepEqual(T{A: "a", B: map[string]string{"C": "D"}})(v)).Error(),
			)

		})
	})

}

func TestNotDeepEqual(t *testing.T) {
	t.Run("given check NotDeepEqual(false)", func(t *testing.T) {
		cf := subtest.NotDeepEqual(false)
		t.Run("when cheking against true", func(t *testing.T) {
			vf := subtest.Value(cf(true))
			t.Run("then there should be no failure", vf.NoError())
		})
		t.Run("when cheking against false", func(t *testing.T) {
			vf := subtest.Value(cf(false))
			t.Run("then an appropriate error should be returned", vf.ErrorIs(subtest.Failure{
				Prefix: "values are deep equal",
				Got:    "bool\n\tfalse",
				Reject: "bool\n\tfalse",
			}))
		})
		t.Run("when testing against false with different syntax", func(t *testing.T) {
			v := false
			vf := subtest.Value(v)
			t.Run("then <ValueFunc>.<CheckFuncName> should pass", vf.NotDeepEqual(true))
			t.Run("the <ValueFunc>.Test(<CheckFunc>) should pass", vf.Test(subtest.NotDeepEqual(true))) // equivalent

		})
	})

	t.Run("given a nested struct type T", func(t *testing.T) {
		type T struct {
			A string
			B map[string]string
		}

		t.Run("when checking a non-nil *T value", func(t *testing.T) {
			v := &T{A: "a", B: map[string]string{"C": "D"}}

			t.Run("then it should not accept an equivalent *T value",
				subtest.Value(subtest.NotDeepEqual(&T{A: "a", B: map[string]string{"C": "D"}})(v)).Error(),
			)
			t.Run("then it should accept a different *T value",
				subtest.Value(v).NotDeepEqual(&T{A: "a", B: map[string]string{"C": "E"}}),
			)
			t.Run("then it should not match an equivalent T value",
				subtest.Value(v).NotDeepEqual(T{A: "a", B: map[string]string{"C": "D"}}),
			)

		})
	})
}

func TestCheckReflectNil(t *testing.T) {
	type T struct{ Foo string }

	cf := subtest.ReflectNil()

	t.Run("when cheking against untyped nil", func(t *testing.T) {
		vf := subtest.Value(cf(nil))
		t.Run("then it should fail", vf.NoError())
	})
	t.Run("when cheking against a nil struct pointer", func(t *testing.T) {
		vf := subtest.Value(cf((*T)(nil)))
		t.Run("then it should fail", vf.NoError())
	})

	t.Run("when cheking against a non-nil struct pointer", func(t *testing.T) {
		vf := subtest.Value(cf(&T{}))
		t.Run("then an appropriate error should be returned", vf.ErrorIs(subtest.Failure{
			Prefix: "value is neither typed nor untyped nil",
			Got:    "*subtest_test.T\n\t{Foo:}",
		}))
	})

}

func TestCheckNotReflectNil(t *testing.T) {
	type T struct{ Foo string }

	cf := subtest.NotReflectNil()

	t.Run("when cheking against untyped nil", func(t *testing.T) {
		vf := subtest.Value(cf(nil))
		t.Run("then an appropriate error should be returned", vf.ErrorIs(subtest.Failure{
			Prefix: "value is typed or untyped nil",
			Got:    "untyped nil",
		}))
	})
	t.Run("when cheking against a nil struct pointer", func(t *testing.T) {
		vf := subtest.Value(cf((*T)(nil)))
		t.Run("then an appropriate error should be returned", vf.ErrorIs(subtest.Failure{
			Prefix: "value is typed or untyped nil",
			Got:    "*subtest_test.T\n\tnil",
		}))
	})

	t.Run("when cheking against a non-nil struct pointer", func(t *testing.T) {
		vf := subtest.Value(cf(&T{}))
		t.Run("then it should fail", vf.NoError())
	})

}

func TestLessThan(t *testing.T) {
	t.Run("given a value float64(42)", func(t *testing.T) {
		v := float64(42)
		t.Run("when cheking against 43", func(t *testing.T) {
			cf := subtest.LessThan(43)
			vf := subtest.Value(cf(v))
			t.Run("then it should not fail", vf.NoError())
		})
		t.Run("when cheking against 42", func(t *testing.T) {
			cf := subtest.LessThan(42)
			vf := subtest.Value(cf(v))
			expect := subtest.Failure{
				Prefix: "not < 42.000000",
				Got:    "float64\n\t42",
			}
			t.Run("then it should fail", vf.ErrorIs(expect))
		})
	})
	t.Run("given a value int16(42)", func(t *testing.T) {
		v := int16(42)
		t.Run("when cheking against 43", func(t *testing.T) {
			cf := subtest.LessThan(43)
			vf := subtest.Value(cf(v))
			t.Run("then it should not fail", vf.NoError())
		})
		t.Run("when cheking against 42", func(t *testing.T) {
			cf := subtest.LessThan(42)
			vf := subtest.Value(cf(v))
			expect := subtest.Failure{
				Prefix: "not < 42.000000",
				Got:    "int16\n\t42",
			}
			t.Run("then it should fail", vf.ErrorIs(expect))
		})
	})
}

func TestLessThanOrEqual(t *testing.T) {
	t.Run("given a value float64(42)", func(t *testing.T) {
		v := float64(42)
		t.Run("when cheking against 43", func(t *testing.T) {
			cf := subtest.LessThanOrEqual(43)
			vf := subtest.Value(cf(v))
			t.Run("then it should not fail", vf.NoError())
		})
		t.Run("when cheking against 42", func(t *testing.T) {
			cf := subtest.LessThanOrEqual(42)
			vf := subtest.Value(cf(v))
			t.Run("then it should not fail", vf.NoError())
		})
		t.Run("when cheking against 41", func(t *testing.T) {
			cf := subtest.LessThanOrEqual(41)
			vf := subtest.Value(cf(v))
			expect := subtest.Failure{
				Prefix: "not <= 41.000000",
				Got:    "float64\n\t42",
			}
			t.Run("then it should fail", vf.ErrorIs(expect))
		})
	})
	t.Run("given a value int16(42)", func(t *testing.T) {
		v := int16(42)
		t.Run("when cheking against 42", func(t *testing.T) {
			cf := subtest.LessThanOrEqual(42)
			vf := subtest.Value(cf(v))
			t.Run("then it should not fail", vf.NoError())
		})
		t.Run("when cheking against 41", func(t *testing.T) {
			cf := subtest.LessThanOrEqual(41)
			vf := subtest.Value(cf(v))
			expect := subtest.Failure{
				Prefix: "not <= 41.000000",
				Got:    "int16\n\t42",
			}
			t.Run("then it should fail", vf.ErrorIs(expect))
		})
	})
}

func TestGreaterThan(t *testing.T) {
	t.Run("given a value float64(42)", func(t *testing.T) {
		v := float64(42)
		t.Run("when cheking against 41", func(t *testing.T) {
			cf := subtest.GreaterThan(41)
			vf := subtest.Value(cf(v))
			t.Run("then it should not fail", vf.NoError())
		})
		t.Run("when cheking against 42", func(t *testing.T) {
			cf := subtest.GreaterThan(42)
			vf := subtest.Value(cf(v))
			expect := subtest.Failure{
				Prefix: "not > 42.000000",
				Got:    "float64\n\t42",
			}
			t.Run("then it should fail", vf.ErrorIs(expect))
		})
	})
	t.Run("given a value int16(42)", func(t *testing.T) {
		v := int16(42)
		t.Run("when cheking against 41", func(t *testing.T) {
			cf := subtest.GreaterThan(41)
			vf := subtest.Value(cf(v))
			t.Run("then it should not fail", vf.NoError())
		})
		t.Run("when cheking against 42", func(t *testing.T) {
			cf := subtest.GreaterThan(42)
			vf := subtest.Value(cf(v))
			expect := subtest.Failure{
				Prefix: "not > 42.000000",
				Got:    "int16\n\t42",
			}
			t.Run("then it should fail", vf.ErrorIs(expect))
		})
	})
}

func TestGreaterThanOrEqual(t *testing.T) {
	t.Run("given a value float64(42)", func(t *testing.T) {
		v := float64(42)
		t.Run("when cheking against 41", func(t *testing.T) {
			cf := subtest.GreaterThanOrEqual(41)
			vf := subtest.Value(cf(v))
			t.Run("then it should not fail", vf.NoError())
		})
		t.Run("when cheking against 42", func(t *testing.T) {
			cf := subtest.GreaterThanOrEqual(42)
			vf := subtest.Value(cf(v))
			t.Run("then it should not fail", vf.NoError())
		})
		t.Run("when cheking against 43", func(t *testing.T) {
			cf := subtest.GreaterThanOrEqual(43)
			vf := subtest.Value(cf(v))
			expect := subtest.Failure{
				Prefix: "not >= 43.000000",
				Got:    "float64\n\t42",
			}
			t.Run("then it should fail", vf.ErrorIs(expect))
		})
	})
	t.Run("given a value int16(42)", func(t *testing.T) {
		v := int16(42)
		t.Run("when cheking against 42", func(t *testing.T) {
			cf := subtest.GreaterThanOrEqual(42)
			vf := subtest.Value(cf(v))
			t.Run("then it should not fail", vf.NoError())
		})
		t.Run("when cheking against 43", func(t *testing.T) {
			cf := subtest.GreaterThanOrEqual(43)
			vf := subtest.Value(cf(v))
			expect := subtest.Failure{
				Prefix: "not >= 43.000000",
				Got:    "int16\n\t42",
			}
			t.Run("then it should fail", vf.ErrorIs(expect))
		})
	})
}

func TestNumericEqual(t *testing.T) {
	t.Run("given a value float64(42)", func(t *testing.T) {
		v := float64(42)
		t.Run("when cheking against 41", func(t *testing.T) {
			cf := subtest.NumericEqual(41)
			vf := subtest.Value(cf(v))
			expect := subtest.Failure{
				Prefix: "not == 41.000000",
				Got:    "float64\n\t42",
			}
			t.Run("then it should fail", vf.ErrorIs(expect))
		})
		t.Run("when cheking against 42", func(t *testing.T) {
			cf := subtest.NumericEqual(42)
			vf := subtest.Value(cf(v))
			t.Run("then it should not fail", vf.NoError())
		})
		t.Run("when cheking against 43", func(t *testing.T) {
			cf := subtest.NumericEqual(43)
			vf := subtest.Value(cf(v))
			expect := subtest.Failure{
				Prefix: "not == 43.000000",
				Got:    "float64\n\t42",
			}
			t.Run("then it should fail", vf.ErrorIs(expect))
		})
	})
}

func TestNumericNotEqual(t *testing.T) {
	t.Run("given a value float64(42)", func(t *testing.T) {
		v := float64(42)
		t.Run("when cheking against 41", func(t *testing.T) {
			cf := subtest.NumericEqual(41)
			vf := subtest.Value(cf(v))
			expect := subtest.Failure{
				Prefix: "not == 41.000000",
				Got:    "float64\n\t42",
			}
			t.Run("then it should fail", vf.ErrorIs(expect))
		})
		t.Run("when cheking against 42", func(t *testing.T) {
			cf := subtest.NumericEqual(42)
			vf := subtest.Value(cf(v))
			t.Run("then it should not fail", vf.NoError())
		})
		t.Run("when cheking against 43", func(t *testing.T) {
			cf := subtest.NumericEqual(43)
			vf := subtest.Value(cf(v))
			expect := subtest.Failure{
				Prefix: "not == 43.000000",
				Got:    "float64\n\t42",
			}
			t.Run("then it should fail", vf.ErrorIs(expect))
		})
	})
	t.Run("given a value int16(42)", func(t *testing.T) {
		v := int16(42)
		t.Run("when cheking against 42", func(t *testing.T) {
			cf := subtest.GreaterThanOrEqual(42)
			vf := subtest.Value(cf(v))
			t.Run("then it should not fail", vf.NoError())
		})
		t.Run("when cheking against 43", func(t *testing.T) {
			cf := subtest.GreaterThanOrEqual(43)
			vf := subtest.Value(cf(v))
			expect := subtest.Failure{
				Prefix: "not >= 43.000000",
				Got:    "int16\n\t42",
			}
			t.Run("then it should fail", vf.ErrorIs(expect))
		})
	})
}

func TestRegexp(t *testing.T) {
	t.Run("given a regular expression check function", func(t *testing.T) {
		cf := subtest.MatchRegexp(regexp.MustCompile(`^"f.*a.?r"$`))
		t.Run("when cheking against a non matching string", func(t *testing.T) {
			vf := subtest.Value(cf(`"foo"`))
			t.Run("then it should fail", vf.Error())
		})
		t.Run("when cheking against a matching string", func(t *testing.T) {
			vf := subtest.Value(cf(`"foobar"`))
			t.Run("then it should not fail", vf.NoError())
		})
		t.Run("when cheking against a matching []byte", func(t *testing.T) {
			vf := subtest.Value(cf([]byte(`"foobar"`)))
			t.Run("then it should not fail", vf.NoError())
		})
		t.Run("when cheking against a matching json.RawMessage", func(t *testing.T) {
			vf := subtest.Value(cf(json.RawMessage(`"foobar"`)))
			t.Run("then it should not fail", vf.NoError())
		})
	})
}
