# PRD: 终端可选结果统一为 Optional[T]

## Introduction

`seq` 库当前用 Go 惯用的 `(T, bool)` 双值表达"可能没有结果"的终端操作（`Find`/`First`/`Last`/`Nth`/`Reduce`/`Max`/`Min` 等约 26 处）。这是地道的 Go 写法，但无法链式后处理——拿到 `(v, ok)` 只能 `if ok {…}`，不能 `.Map().OrElse()`。

库已有零依赖的 `Optional[T]` 类型（含 `Map`/`Filter`/`FlatMap`/`OrElse`/`Unwrap` 等）。本特性**趁库刚发布、尚无下游引用**，一步到位地将所有返回 `(…, bool)` 的终端方法与函数**统一改为返回 `Optional`**，使库转向函数式优先、链式至上的风格。

这是一次**有意的破坏性变更**（breaking change）——之所以现在做，正因为没有引用、成本最低；拖到有引用后再改代价高得多。`Optional.Get() (T, bool)` 保留，作为回到 Go 惯用 `(T, bool)` 的反向桥接。

## Goals

- 所有"可能无结果"的终端操作统一返回 `Optional[T]`（或 `Optional[Pair[K,V]]`），消除 `(…, bool)` 与 `Optional` 并存的不一致
- 保留 `Optional.Get() (T, bool)` 作为反向桥接，不破坏与 Go `(T,bool)` 惯用法的互通
- 保持零第三方依赖
- 同步更新全部测试、README 双语、`docs/index.html` 中受影响的签名与示例
- 库整体编译、测试通过

## User Stories

### US-001: Seq[T] 单值终端方法改为 Optional[T]
**Description:** As a user, I want Seq's partial-result methods to return Optional so that I can chain post-processing.

**Acceptance Criteria:**
- [ ] `Find` / `FindLast` / `First` / `Last` / `Nth` 由 `(T, bool)` 改为 `Optional[T]`
- [ ] `Reduce` 由 `(T, bool)` 改为 `Optional[T]`（空序列返回 `None`）
- [ ] `MaxBy` / `MinBy` / `MaxByKey` / `MinByKey` 由 `(T, bool)` 改为 `Optional[T]`
- [ ] 空序列 / 无匹配一律返回 `None[T]()`
- [ ] 所有相关单元测试更新为断言 `Optional`（用 `.Get()` 或 `.IsPresent()`/`.OrElse()`）
- [ ] `go build/vet/test` 与 `gofmt` 通过

### US-002: Seq[T] 下标类终端方法改为 Optional[int]
**Description:** As a user, I want index-returning methods to return Optional[int] so the "not found" case is consistent with the rest.

**Acceptance Criteria:**
- [ ] `FindIndex` / `FindLastIndex` 由 `(int, bool)` 改为 `Optional[int]`
- [ ] 无匹配返回 `None[int]()`
- [ ] 相关测试更新
- [ ] `go build/vet/test` 通过

### US-003: 包级约束函数改为 Optional
**Description:** As a user, I want package-level Max/Min/IndexOf/LastIndexOf to return Optional so they match the methods.

**Acceptance Criteria:**
- [ ] `Max[T]` / `Min[T]`（numeric.go）由 `(T, bool)` 改为 `Optional[T]`
- [ ] `IndexOf[T]` / `LastIndexOf[T]`（comparable.go）由 `(int, bool)` 改为 `Optional[int]`
- [ ] 空序列 / 无匹配返回 `None`
- [ ] 相关测试更新
- [ ] `go build/vet/test` 通过

### US-004: 约束子类型方法改为 Optional
**Description:** As a user, I want subtype Max/Min/IndexOf to return Optional consistently.

**Acceptance Criteria:**
- [ ] `SeqOrdered.Max` / `SeqOrdered.Min` / `SeqOrdered.IndexOf` 改为 `Optional`
- [ ] `SeqNumeric.Max` / `SeqNumeric.Min` / `SeqNumeric.IndexOf` 改为 `Optional`
- [ ] `SeqComparable.IndexOf` 改为 `Optional[int]`
- [ ] 这些方法多为转调包级函数（US-003），保持委托实现
- [ ] 相关测试更新
- [ ] `go build/vet/test` 通过

### US-005: Seq2.Find 改为 Optional[Pair[K,V]]
**Description:** As a user, I want Seq2.Find to return an Optional too, wrapping the key/value in a Pair.

**Acceptance Criteria:**
- [ ] `(s Seq2[K,V]) Find(pred) ` 由 `(K, V, bool)` 改为 `Optional[Pair[K, V]]`
- [ ] 命中时返回 `Some(Pair[K,V]{Left:k, Right:v})`；无匹配返回 `None`
- [ ] doc 说明用 `Pair` 承载键值（复用现有类型，不新增 Optional2）
- [ ] 相关测试更新
- [ ] `go build/vet/test` 通过

### US-006: 文档与示例同步
**Description:** As a user reading docs, I want all examples to reflect the new Optional return types.

**Acceptance Criteria:**
- [ ] `README.md` / `README_CN.md` 中所有受影响方法的签名与示例更新为 `Optional` 形式
- [ ] `docs/index.html` 中受影响条目的签名与示例更新（终结、聚合、约束自由函数、子类型、Seq2 各节）
- [ ] `Optional` 一节补充：从终端结果直接链式（`s.Find(p).Map(f).OrElse(x)`）及用 `.Get()` 回到 `(T,bool)` 的桥接示例
- [ ] 受影响的 `Example`（如有 `// Output:`）更新且 `go test` 通过
- [ ] `go.mod` 依赖数仍为 0

## Functional Requirements

- FR-1: 系统必须将 `Seq.Find`/`FindLast`/`First`/`Last`/`Nth`/`Reduce` 的返回类型从 `(T, bool)` 改为 `Optional[T]`。
- FR-2: 系统必须将 `Seq.MaxBy`/`MinBy`/`MaxByKey`/`MinByKey` 的返回类型改为 `Optional[T]`。
- FR-3: 系统必须将 `Seq.FindIndex`/`FindLastIndex` 的返回类型从 `(int, bool)` 改为 `Optional[int]`。
- FR-4: 系统必须将包级 `Max`/`Min` 改为 `Optional[T]`，`IndexOf`/`LastIndexOf` 改为 `Optional[int]`。
- FR-5: 系统必须将 `SeqOrdered`/`SeqNumeric` 的 `Max`/`Min`/`IndexOf` 与 `SeqComparable.IndexOf` 改为对应 `Optional`。
- FR-6: 系统必须将 `Seq2.Find` 改为 `Optional[Pair[K, V]]`。
- FR-7: 所有"空序列 / 无匹配"情形必须返回 `None`。
- FR-8: 系统必须保留 `Optional.Get() (T, bool)` 作为反向桥接，不得移除。
- FR-9: 系统不得引入第三方依赖。
- FR-10: `FilterMap` 的 `f func(T) (U, bool)` 参数**不在**改造范围（它是用户提供的回调签名，非"可能无结果"的返回值）。

## Non-Goals (Out of Scope)

- 不改 `FilterMap[U](f func(T) (U, bool))`——`(U, bool)` 是入参回调约定，不是返回的可选结果。
- 不改 `Any`/`All`/`None`/`Contains`/`Equal`/`IsEmpty` 等返回纯 `bool` 的判定方法（`bool` 是结果本身，不是"可能无值"）。
- 不移除或改名 `Optional.Get()`——它是回到 `(T,bool)` 的桥接，必须保留。
- 不新增 `Optional2[K,V]` 类型——`Seq2.Find` 用 `Optional[Pair[K,V]]`。
- 不新增任何 `*Opt` 并行方法——这是**替换**而非并行新增。

## Technical Considerations

- **破坏性变更**：这是 breaking change，仅因库无引用才此时进行。建议在 commit/PR 中明确标注 BREAKING，必要时未来发版以 minor 版本号体现（当前为 Draft，尚无 tag）。
- **委托实现复用**：子类型的 `Max`/`Min`/`IndexOf` 多为转调包级函数，改造包级函数后子类型方法自然受益，保持薄委托。
- **测试改造模式**：原 `v, ok := s.Find(p); if ok {…}` → `opt := s.Find(p); if opt.IsPresent() {…}` 或断言 `opt.Get()`。批量但机械。
- **Pair 已存在**：`seq.go` 已定义 `Pair[A,B]{Left, Right}`，`Seq2.Find` 直接复用。
- **风格对齐**：返回 `None[T]()` / `Some(v)`，与 `optional.go` 现有实现一致。

## Success Metrics

- 全库无任何 `(…, bool)` 形式的"可能无结果"终端 API（判定类 `bool` 除外）；`Optional` 成为统一表达。
- `go build/vet/test ./...` 全绿；`go.mod` 依赖数为 0。
- `Optional.Get()` 保留，可一行桥接回 `(T,bool)`。
- 文档（README ×2 + docs/index.html）签名与示例与代码一致。

## Open Questions

1. 是否借此机会给 `Optional` 增加 `Get()` 之外更顺手的解构方式（如 `func (o Optional[T]) Unpack() (T, bool)` 别名）？倾向不加，`Get()` 已够。
2. `Seq2.Find` 用 `Optional[Pair[K,V]]` 后，是否需要在 `Pair` 上加 `Unpack() (A,B)` 便于解构？可选，本批不做，按需再议。
3. 未来若要发布稳定版，本次 breaking change 应与其它待定 API 调整一起在同一个 pre-1.0 窗口完成，避免多次破坏。
