package brpc

import (
	"testing"
)

func TestMethodType(t *testing.T) {
	var mt = &methodType{}
	var got = mt.NumCalls()
	var expect int64 = 0
	if got != expect {
		t.Errorf("MethodType.NumCalls() = %d, expect = %d",
			got, expect)
	}
}

func TestNewService(t *testing.T) {
	var s = NewService()
	t.Run("GetName", func(t *testing.T) {
		var got = s.GetName()
		var expect = ""
		if got != expect {
			t.Errorf("service.GetName() = %s, expect = %s",
				got, expect)
		}
	})
}

type TestClassNoMethod struct {
}

type TestClassOneMethod struct {
}

func (x *TestClassOneMethod) add() bool {
	return true
}

type TestClassOneMethodExport struct {
}

func (x *TestClassOneMethodExport) Add(a *int, b *int) error {
	return nil
}

func TestServiceInstall(t *testing.T) {
	t.Run("Bad Service Name", func(t *testing.T) {
	})
	t.Run("TypeHasNotExportMethod_NoMethod", func(t *testing.T) {
		var srv = NewService()
		var s = &TestClassNoMethod{}

		var err = srv.Install(s)

		var expect = "Type Has Not Export Method"
		if err == nil || err.Error() != expect {
			t.Errorf("service.Install(%T) = err(%v), expect = %s",
				s, err, expect)
		}
	})
	t.Run("TypeHasNotExportMethod_OneMethod", func(t *testing.T) {

		var srv = NewService()
		var s = &TestClassOneMethod{}
		var err = srv.Install(s)

		var expect = "Type Has Not Export Method"
		if err == nil || err.Error() != expect {
			t.Errorf("service.Install(%T) = err(%v), expect = %s",
				s, err, expect)
		}
	})
	t.Run("TypeHasNotExportMethod_OneMethodExport", func(t *testing.T) {

		var srv = NewService()
		var s = &TestClassOneMethodExport{}
		var err = srv.Install(s)

		if err != nil {
			t.Errorf("service.Install(%T) = %v, expect = nil",
				s, err)
		}
	})
}
