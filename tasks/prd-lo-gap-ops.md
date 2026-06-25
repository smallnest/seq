# PRD: lo 对照补全 —— PartitionBy / Times / Replace

## Introduction

在与 [`samber/lo`](https://github.com/samber/lo) 核心切片函数逐一对照后，`seq` 已覆盖绝大多数常用能力，仅余少量真实缺口。本特性补齐其中三项与本库惰性/函数式定位契合、且遵守"划分铁律"（T 受约束的操作不得是 `Seq[T]` 方法）的操作：

1. **`PartitionBy`** —— 按 key 函数把序列分成多个组，**保持 key 首次出现的顺序**。填补现有 `Partition`（仅按布尔谓词二分）与 `GroupBy`（返回无序 `map`）之间的空白。返回 `Seq2[K, []T]`（惰性类型，但见下方技术约束）。
2. **`Times`** —— 调用 `f(i)` 共 n 次并收集为序列的构造函数，补全 `Generate`（无限）/ `Repeat`（定值）/ `Iterate`（迭代）构造器家族。
3. **`Replace` / `ReplaceAll`** —— 按值替换：把序列中等于 `old` 的元素替换为 `new`（`Replace` 限前 n 个，`ReplaceAll` 全部）。因需 `==` 比较，按铁律作为 `SeqComparable[T]` 的惰性方法。

全部遵守项目两条铁律：零第三方依赖（`go.mod` 保持现状）、不破坏与 `iter.Seq` / 标准库的零成本互操作。

## Goals

- 新增 `PartitionBy`，按 key 保序分组，返回 `Seq2[K, []T]`
- 新增包级构造函数 `Times[T any](n int, f func(int) T) Seq[T]`
- 在 `SeqComparable[T]` 上新增惰性方法 `Replace` 与 `ReplaceAll`
- 保持零第三方依赖与零侵入（不改动任何现有签名）
- 测试、文档风格与现有文件（`intermediate.go` / `subtypes.go` / `constructors.go`）一致

## User Stories

### US-001: 实现 PartitionBy（按 key 保序分组）
**Description:** As a user, I want to split a sequence into groups by a key function while preserving the order in which keys first appear, so that I get deterministic grouping unlike the unordered `GroupBy` map.

**Acceptance Criteria:**
- [ ] 新增 `PartitionBy[T any, K comparable](s Seq[T], key func(T) K) Seq2[K, []T]`（包级函数；K 需 comparable 故不能是 Seq[T] 方法）
- [ ] 产出的 `Seq2` 按 key **首次出现顺序**逐组 yield `(key, 组内元素切片)`
- [ ] 同一 key 的元素按其在源序列中的相对顺序收集进切片
- [ ] 空序列产出空 `Seq2`（不 yield 任何对）
- [ ] `key` 为 nil 时，按项目 nil-func 约定在驱动迭代时 panic
- [ ] doc comment 说明：实现需在 yield 前缓冲全部元素（无法在源耗尽前确定某组已完整），因此是"类型惰性、内部物化"
- [ ] 单元测试覆盖：保序性、同 key 聚合、空序列、单元素
- [ ] `go build/vet/test` 与 `gofmt` 通过

### US-002: 实现 Times 构造函数
**Description:** As a user, I want a constructor that calls f(i) n times so that I can build index-derived sequences (e.g. squares 0,1,4,9...).

**Acceptance Criteria:**
- [ ] 新增 `Times[T any](n int, f func(int) T) Seq[T]`，置于 `constructors.go`
- [ ] 产出 `f(0), f(1), …, f(n-1)`，共 n 个元素
- [ ] `n <= 0` 产出空序列（不调用 f），与 `Repeat` 的"空输入空输出"约定一致
- [ ] 惰性：`f` 在驱动迭代时按需调用，而非构造时
- [ ] `f` 为 nil 时按项目 nil-func 约定在驱动迭代时 panic
- [ ] doc comment 引用 `Generate` / `Repeat` / `Iterate` 作为同族构造器
- [ ] 单元测试覆盖：n>0 正常、n=0、n<0、惰性（用计数器验证 f 调用次数）

### US-003: 实现 SeqComparable.Replace 与 ReplaceAll
**Description:** As a user, I want to replace elements equal to a given value with another value so that I can substitute specific values in a lazy pipeline.

**Acceptance Criteria:**
- [ ] `(s SeqComparable[T]) Replace(old, new T, n int) SeqComparable[T]`：把前 `n` 个等于 `old` 的元素替换为 `new`，其余原样；`n < 0` 视为不限（等同 ReplaceAll 行为，doc 说明）
- [ ] `(s SeqComparable[T]) ReplaceAll(old, new T) SeqComparable[T]`：替换全部等于 `old` 的元素
- [ ] 两者均为**惰性中间操作**，返回 `SeqComparable[T]`（保持子类型链可继续 `.Distinct()` 等）
- [ ] 不等于 `old` 的元素原样透传，顺序不变
- [ ] 空序列产出空序列
- [ ] 单元测试覆盖：替换前 n 个、ReplaceAll、n=0（不替换）、无匹配、惰性
- [ ] doc comment 说明 `Replace(old,new,n)` 中 n 的边界语义

### US-004: 文档与示例
**Description:** As a user reading the docs, I want the new operations documented so that I can discover and use them.

**Acceptance Criteria:**
- [ ] `README.md` 与 `README_CN.md` 在 API 一览表的相应分组补充 `PartitionBy`、`Times`、`Replace`/`ReplaceAll`
- [ ] 至少一个可运行 `Example`（如 `ExamplePartitionBy` 或 `ExampleTimes`），含 `// Output:` 且通过
- [ ] `docs/` 下若有 API 清单文档，补充三项条目
- [ ] `go test ./...` 全绿；`go.mod` 依赖数仍为 0

## Functional Requirements

- FR-1: 系统必须提供包级 `PartitionBy[T any, K comparable](s Seq[T], key func(T) K) Seq2[K, []T]`，按 key 首次出现顺序分组。
- FR-2: `PartitionBy` 同一 key 的元素必须按源序列相对顺序收集。
- FR-3: 系统必须提供包级 `Times[T any](n int, f func(int) T) Seq[T]`，产出 `f(0)..f(n-1)`。
- FR-4: `Times` 在 `n <= 0` 时必须产出空序列且不调用 f。
- FR-5: 系统必须在 `SeqComparable[T]` 提供惰性方法 `Replace(old, new T, n int) SeqComparable[T]`，替换前 n 个匹配。
- FR-6: 系统必须在 `SeqComparable[T]` 提供惰性方法 `ReplaceAll(old, new T) SeqComparable[T]`，替换全部匹配。
- FR-7: 所有新增高阶函数参数为 nil 时，必须按现有 nil-func 约定（惰性操作在驱动迭代时 panic）处理。
- FR-8: 系统不得改动任何现有方法签名，不得引入第三方依赖。

## Non-Goals (Out of Scope)

- 不实现 `Sample` / `Shuffle` / `Fill` / `Splice` / `DropByIndex`——随机性或原地/随机访问语义与惰性单遍迭代模型冲突（`lo/mutable` 已被 README Scope 明确排除）。
- 不实现 `Clamp`——标量工具，非序列操作。
- 不实现 `Every` / `Some`（lo 命名）——已有等价的 `All` / `Any`。
- 不实现 `DropRightWhile`——需反向扫描+物化，惰性收益为零，本批次不做。
- 不实现 `CountValuesBy`——与现有 `GroupCount` 语义重叠，避免冗余。
- `PartitionBy` 不提供 `[][]T` 形式（本批次按决策只做 `Seq2[K, []T]`）。

## Technical Considerations

- **PartitionBy 是"类型惰性、内部物化"**：返回 `Seq2[K, []T]` 看似惰性，但必须在 yield 任何组之前消费完整个源序列（在源耗尽前无法确定某组已完整）。实现应用 `map[K]int`（记录组在结果中的下标）+ `[]K`（保序）+ `[][]T`（各组缓冲），源遍历完后再按 `[]K` 顺序 yield。doc 必须明示此特性，避免用户误以为可对无限序列使用。
- **划分铁律**：`PartitionBy` 的 K、`Replace` 的 `==` 均需约束，故 `PartitionBy` 是带方法级 K 约束的包级函数、`Replace`/`ReplaceAll` 是 `SeqComparable[T]` 方法——与 `GroupBy`（Seq 方法因 K 在方法级）、`Distinct`（Comparable 子类型）的既有模式一致。
- **Replace 的惰性**：替换是逐元素 1:1 变换，天然惰性，无需物化；用一个计数器跟踪已替换数即可实现 `Replace(old,new,n)`。
- **风格对齐**：`PartitionBy` 参照 `terminal.go` 的分组实现与 `multiseq.go` 的 Seq2 产出；`Times` 参照 `constructors.go` 的 `Generate`/`Repeat`；`Replace` 参照 `subtypes.go` 的 `SeqComparable` 惰性方法。

## Success Metrics

- 三项操作均编译、测试通过；`go.mod` 依赖保持为 0。
- `PartitionBy` 保序性、`Times` 边界（n<=0）、`Replace` 的 n 边界均有测试钉住。
- 现有测试零改动仍通过（零侵入）。

## Open Questions

1. `Replace(old, new, n int)` 中 `n < 0` 的语义：本 PRD 定为"不限（等同 ReplaceAll）"。是否更倾向 `n < 0` 视为 0（不替换）？倾向前者（与 lo 及 strings.Replace 的 n<0 语义一致）。
2. 是否未来补 `PartitionBy` 的 `[][]T` 终端形式以贴近 lo 原签名？本批次不做，按需再议。
3. `Times` 是否需要并行/缓存变体？不需要——保持最小惰性构造器。
