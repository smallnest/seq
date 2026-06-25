# PRD: Optional[T] —— 轻量可选值类型

## Introduction

本项目（`seq`）的众多终端操作返回 Go 惯用的 `(T, bool)` 双值，表达"可能没有结果"——例如 `Find` / `First` / `Last` / `Nth` / `Reduce` / `Max` / `Min` / `MaxBy` 等共 15+ 处。这种约定地道、与标准库一致，但**无法链式后处理**：拿到 `(v, ok)` 后只能写 `if ok { ... }`，无法 `.Map(...).OrElse(...)`。

本特性新增一个**本包内、零依赖**的 `Optional[T]` 类型，作为 `(T, bool)` 的可选包装，提供链式后处理能力。它是**并行新增**而非替换：所有现有 `(T, bool)` 方法签名保持不变，用户通过包级函数 `ToOptional(v, ok)` 按需桥接。

设计上严格遵守项目两条铁律：（1）零第三方依赖（`go.mod` 当前仅一行）；（2）不破坏与 `iter.Seq` / `slices` / `maps` 标准库生态的零成本互操作——`Optional` 不出现在任何现有方法的签名里，仅作为用户侧的可选工具。

> **关于 `Result[T, error]`：** 评估后**不在本期实现**。本项目是纯函数式管道，代码库中**没有任何 `error` 语义**——库本身不产生错误，"缺失值"用 `(T, bool)` 表达，唯一失败模式（传入 nil func）按约定 panic（见 `seq.go:23-25`）。`Result` 解决的是"操作可能失败并带原因"，本项目无此场景。详见 Open Questions。

## Goals

- 提供 `Optional[T]` 类型，可链式表达"存在则变换，否则回退"的逻辑
- 保持**零第三方依赖**与 `go.mod` 现状
- **不改动**任何现有 `(T, bool)` 方法签名，不在现有 API 表面引入 `Optional`
- 提供 `ToOptional(v, ok) Optional[T]` 作为从现有方法桥接的唯一入口
- 提供包级 `MapOptional[T, U]` 以绕开 Go 无 HKT、方法不能改类型参数的限制
- 测试覆盖与文档风格与现有文件（如 `terminal.go`、`comparable.go`）保持一致

## User Stories

### US-001: 定义 Optional[T] 类型与构造函数
**Description:** As a library user, I want an `Optional[T]` type with clear constructors so that I can represent "a value that may be absent" explicitly.

**Acceptance Criteria:**
- [ ] 新增文件 `optional.go`，`package seq`
- [ ] 定义 `Optional[T any]` 类型（结构体，含未导出的 `value T` 与 `present bool` 字段）
- [ ] 提供构造函数 `Some[T any](v T) Optional[T]`（present=true）
- [ ] 提供构造函数 `None[T any]() Optional[T]`（present=false，value 为零值）
- [ ] 提供桥接函数 `ToOptional[T any](v T, ok bool) Optional[T]`：ok 为 true 时等价 `Some(v)`，否则 `None[T]()`
- [ ] 每个导出标识符有符合项目风格的 doc comment
- [ ] `go vet ./...` 与 `gofmt` 通过

### US-002: 实现取值与判定方法
**Description:** As a user, I want to inspect and extract the contained value safely so that I can interoperate with existing `(T, bool)` code.

**Acceptance Criteria:**
- [ ] `(o Optional[T]) Get() (T, bool)` —— 返回值与存在性，与现有约定对称（便于 `if v, ok := o.Get(); ok`）
- [ ] `(o Optional[T]) IsPresent() bool`
- [ ] `(o Optional[T]) IsEmpty() bool`（`!IsPresent()`，命名与 `Seq.IsEmpty` 呼应）
- [ ] `(o Optional[T]) OrElse(fallback T) T` —— present 返回值，否则返回 fallback
- [ ] `(o Optional[T]) OrZero() T` —— present 返回值，否则返回 T 的零值
- [ ] `Get` 在 None 时返回 `(zero, false)`，零值由 `var zero T` 产生
- [ ] 对应单元测试覆盖 Some / None 两条路径

### US-003: 实现同类型链式方法
**Description:** As a user, I want to transform and filter an Optional in-place so that I can chain post-processing without unwrapping.

**Acceptance Criteria:**
- [ ] `(o Optional[T]) Map(f func(T) T) Optional[T]` —— present 时返回 `Some(f(value))`，None 时原样返回 None；f 不得为 nil（present 且 f 为 nil 时 panic，与项目 nil-func 约定一致）
- [ ] `(o Optional[T]) Filter(pred func(T) bool) Optional[T]` —— present 且 `pred(value)` 为 true 时保留，否则变为 None
- [ ] `(o Optional[T]) FlatMap(f func(T) Optional[T]) Optional[T]` —— present 时返回 `f(value)`，None 时返回 None
- [ ] `(o Optional[T]) Or(other Optional[T]) Optional[T]` —— present 返回自身，否则返回 other
- [ ] `(o Optional[T]) IfPresent(f func(T))` —— present 时调用 `f(value)`，无返回值
- [ ] doc comment 明确说明 `Map` 因 Go 无 HKT 只能返回 `Optional[T]`（同类型），跨类型见 `MapOptional`
- [ ] 单元测试覆盖每个方法的 Some / None 分支

### US-004: 实现包级跨类型转换 MapOptional
**Description:** As a user, I want to transform `Optional[T]` into `Optional[U]` so that I can change the element type, which methods cannot do in Go.

**Acceptance Criteria:**
- [ ] 包级函数 `MapOptional[T, U any](o Optional[T], f func(T) U) Optional[U]`
- [ ] present 时返回 `Some(f(value))`，None 时返回 `None[U]()`
- [ ] f 为 nil 且 present 时 panic（与 nil-func 约定一致）
- [ ] doc comment 说明这是 `Map` 的跨类型补充，引用"无 HKT"原因
- [ ] 单元测试：含 `Optional[int]` → `Optional[string]` 的转换用例及 None 透传用例

### US-005: 实现 ToSlice 与 String
**Description:** As a user, I want to materialize and print an Optional so that it integrates with slices and debugging output.

**Acceptance Criteria:**
- [ ] `(o Optional[T]) ToSlice() []T` —— present 返回单元素切片 `[]T{value}`，None 返回长度为 0 的非 nil 切片 `[]T{}`（行为在 doc 中明确）
- [ ] `(o Optional[T]) String() string` —— present 返回 `Some(<value>)`，None 返回 `None`，实现 `fmt.Stringer`
- [ ] `String` 使用 `fmt.Sprintf("Some(%v)", value)` 形式
- [ ] 单元测试断言两种状态的 `ToSlice` 与 `String` 输出

### US-006: 文档与示例
**Description:** As a user reading the docs, I want a runnable example so that I understand the bridge pattern from existing `(T, bool)` methods.

**Acceptance Criteria:**
- [ ] `optional_test.go` 中新增 `ExampleToOptional` 或 `ExampleOptional_Map`，演示 `seq.ToOptional(s.Find(pred)).Map(f).OrElse(x)`，含 `// Output:` 注释且 `go test` 通过
- [ ] `README.md` 与 `README_CN.md` 新增 Optional 简介小节（命名、与 `(T, bool)` 的桥接、无 HKT 限制说明）
- [ ] `docs/` 下若有 API 清单文档，补充 Optional 条目
- [ ] `go test ./...` 全绿

## Functional Requirements

- FR-1: 系统必须提供 `Optional[T any]` 类型，内部以 `(value T, present bool)` 表示存在性，不导出字段。
- FR-2: 系统必须提供 `Some[T any](v T)`、`None[T any]()` 两个构造函数。
- FR-3: 系统必须提供 `ToOptional[T any](v T, ok bool) Optional[T]`，作为从现有 `(T, bool)` 方法桥接的入口。
- FR-4: 系统必须提供 `Get() (T, bool)`，输出与项目现有 `(T, bool)` 约定对称。
- FR-5: 系统必须提供 `IsPresent()`、`IsEmpty()`、`OrElse(T)`、`OrZero()`。
- FR-6: 系统必须提供同类型链式方法 `Map`、`Filter`、`FlatMap`、`Or`、`IfPresent`。
- FR-7: 系统必须提供包级 `MapOptional[T, U any]` 以支持跨类型转换。
- FR-8: 系统必须提供 `ToSlice() []T` 与 `String() string`（实现 `fmt.Stringer`）。
- FR-9: 当 present 为 true 且高阶函数参数为 nil 时，系统必须 panic，与现有 nil-func 约定（`seq.go:23-25`）保持一致。
- FR-10: 系统不得改动任何现有方法的签名，不得在现有方法签名中引入 `Optional`。
- FR-11: 系统不得引入任何第三方依赖；`go.mod` 仅允许保留现有 `module` 与 `go` 指令。

## Non-Goals (Out of Scope)

- **不实现 `Result[T, error]`**（本期）——本项目无 error 语义，详见 Introduction 与 Open Questions。
- 不为 `Seq` / `Seq2` / 子类型新增任何 `*Opt` 形式的并行方法（决策 2B：仅独立类型 + 桥接函数）。
- 不修改、不弃用任何现有 `(T, bool)` 方法。
- 不引入 `mo` / `samber` 或其他第三方 monad 库。
- 不实现 `Either` / `Future` / `IO` 等其他 monad 类型。
- 不提供 `Optional[T]` 的 JSON 序列化/反序列化（如有需要单独评估）。

## Technical Considerations

- **无 HKT 限制：** Go 方法不能引入新类型参数改变接收者的类型参数，故 `Map` 只能 `Optional[T] -> Optional[T]`；跨类型由包级 `MapOptional[T, U]` 承担。这是 PRD 的硬约束，实现与文档都须明示。
- **零值语义：** `None[T]()` 与 `Optional[T]{}`（零值结构体）应行为等价（present=false）——保证未初始化的 `Optional` 即 None。
- **nil-func 约定：** 与 `seq.go` 文档一致，present 时传入 nil 高阶函数 panic；None 时因不调用 func，不 panic（在 doc 中说明此短路行为）。
- **风格对齐：** 新文件命名、doc comment 密度、测试组织参照 `terminal.go` / `terminal_test.go`。
- **测试矩阵：** 每个方法至少覆盖 Some 与 None 两条路径；含值的方法额外覆盖零值类型（如 `Optional[int]` 的 0、`Optional[*X]` 的 nil 指针——指针为 nil 但 present=true 时仍视为存在）。

## Success Metrics

- 桥接一行可达：`seq.ToOptional(s.Find(p)).Map(f).OrElse(x)` 编译并按预期运行。
- `go.mod` 依赖数保持为 0。
- 新增代码测试覆盖率不低于现有终端方法文件水平；`go test ./...` 全绿。
- 现有测试无任何改动即继续通过（证明零侵入）。

## Open Questions

1. **`Result[T, error]` 是否值得未来引入？** 当前评估为"否"：库无 error 来源。唯一可能场景是用户希望 `Map` 支持可失败的变换函数（`func(T) (U, error)`），但这会污染整个管道 API 表面。建议：若未来出现真实需求（例如新增解析/IO 类操作），再以独立 PRD 评估，且大概率应作为 `Optional` 之外的独立类型而非改造现有管道。
2. `Map`/`FlatMap` 是否需要同时提供包级版本以统一风格，还是仅 `MapOptional` 一个跨类型出口即可？（倾向后者，避免 API 膨胀）
3. 是否需要 `Optional[T]` 与 `Seq[T]` 的互转（如 `Optional.ToSeq()` / 从单元素构造 Seq）？本期未列入，可按需补充。
