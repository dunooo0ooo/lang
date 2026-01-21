# lang

### Поддерживаемые возможности

- Типы: `int`, `float`, `bool`, `string`, `char`, `void`
- Массивы: `[]int`
- Арифметика и сравнения
- Условные операторы `if / else`
- Циклы `while`, `for`
- Рекурсия
- Функции
- Built-in функции:
    - `array(len)`
    - `get(arr, i)`
    - `set(arr, i, v)`
    - `print(x)`
    - `println(x)`

Пример:

```lang
fn fact(n: int) -> int {
    if n <= 1 { return 1; }
    return n * fact(n - 1);
}

fn main() -> int {
    return fact(10);
}