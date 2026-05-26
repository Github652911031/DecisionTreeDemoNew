package linkmodel1

import "testing"

func TestModel1Chain(t *testing.T) {
	factory := NewFactory()
	logicLink := factory.OpenLogicLink()

	result, err := logicLink.Apply("123", &DynamicContext{})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != "link model01 单实例链" {
		t.Fatalf("unexpected result: %s", result)
	}
}
