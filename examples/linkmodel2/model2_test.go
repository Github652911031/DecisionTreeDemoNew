package linkmodel2

import "testing"

func TestDemo01JumpFlow(t *testing.T) {
	factory := NewFactory()
	result, err := factory.Demo01().Apply("1", NewDynamicContext())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Info != "hi 小傅哥！" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestDemo01StopInApplyBefore(t *testing.T) {
	factory := NewFactory()
	result, err := factory.Demo01().Apply("2", NewDynamicContext())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Info != "applyBefore 拦截结果" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestDemo01NormalFlow(t *testing.T) {
	factory := NewFactory()
	result, err := factory.Demo01().Apply("3", NewDynamicContext())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Info != "hi 小傅哥！" {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestDemo02Flow(t *testing.T) {
	factory := NewFactory()
	result, err := factory.Demo02().Apply("123", NewDynamicContext())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.Info != "hi 小傅哥！" {
		t.Fatalf("unexpected result: %+v", result)
	}
}
