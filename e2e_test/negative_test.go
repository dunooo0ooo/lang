package e2e_test

//
//import (
//	"github.com/dunooo0ooo/lang/internal/runtime/compilation"
//	"testing"a
//)
//
//func TestNegative_UndeclaredAssign(t *testing.T) {
//	src := `
//fn main() -> int {
//    x = 1;
//    return 0;
//}
//`
//	expectCompileFail(t, src)
//}
//
//func TestNegative_UnknownFunction(t *testing.T) {
//	src := `
//fn main() -> int {
//    foo();
//    return 0;
//}
//`
//	expectCompileFail(t, src)
//}
//
//func TestNegative_PrintArity(t *testing.T) {
//	src := `
//fn main() -> int {
//    print(1, 2);
//    return 0;
//}
//`
//	expectCompileFail(t, src)
//}
//
//func TestNegative_TooManyLocals(t *testing.T) {
//	// VM использует slot как byte => максимум 256 local slots.
//	// В компиляторе сейчас нет явной проверки, но тест полезен: хотя бы не должен silently succeed.
//	const n = 300
//	src := "fn main() -> int {\n"
//	for i := 0; i < n; i++ {
//		src += "    let x" + itoa(i) + ": int = 0;\n"
//	}
//	src += "    return 0;\n}\n"
//
//	expectCompileFail(t, src)
//}
//
//// ---- helpers ----
//
//func expectCompileFail(t *testing.T, src string) {
//	t.Helper()
//
//	prog := mustParse(t, src)
//	mustSema(t, prog)
//
//	comp := compilation.NewCompiler()
//
//	defer func() {
//		if r := recover(); r != nil {
//			// panic считается валидным "compile fail" на текущей стадии
//			return
//		}
//	}()
//
//	_, err := comp.CompileProgram(prog)
//	if err == nil {
//		t.Fatalf("expected compilation failure, got nil error")
//	}
//}
//
//func itoa(i int) string {
//	// без strconv, чтобы не добавлять импорт в файл тестов
//	if i == 0 {
//		return "0"
//	}
//	neg := false
//	if i < 0 {
//		neg = true
//		i = -i
//	}
//	var buf [32]byte
//	pos := len(buf)
//	for i > 0 {
//		pos--
//		buf[pos] = byte('0' + (i % 10))
//		i /= 10
//	}
//	if neg {
//		pos--
//		buf[pos] = '-'
//	}
//	return string(buf[pos:])
//}
