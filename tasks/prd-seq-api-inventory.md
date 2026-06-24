# PRD: seq —— 基于 iter.Seq 的链式泛型库 API 清单

> 语言版本：简体中文 | [English](prd-seq-api-inventory-en.md)


## Introduction

`seq` 是一个 Go 泛型库，围绕标准库的惰性迭代器 `iter.Seq` / `iter.Seq2`（Go 1.23+）构建，借助 **Go 1.27 的泛型方法特性**（[golang/go#77273](https://github.com/golang/go/issues/77273)）提供 Scala 风格的、可链式调用的集合操作。

它要解决的核心问题：今天 `samber/lo` 这类库只能用顶层函数（`lo.Map(lo.Filter(...))`，从内往外读），因为在 Go 1.27 之前，**方法不能声明自己的类型参数**，无法表达"输入 `[]T`、输出 `[]U`"的链式 `.Map()`。Go 1.27 解除了这条限制，使 `From(xs).Filter(...).Map(...).Collect()` 这种可发现、可链、惰性的管道第一次成为可能。

本文档**只产出 API 清单**：完整列出 `Seq[T]` 与 `Seq2[K,V]` 的方法、配套的自由函数（package-level functions），并明确每个 API 归属"方法"还是"自由函数"及其判定理由。不含实现。读者可能是初级开发者或 AI agent，因此每个 API 附一句语义说明。

### 关键设计约束（决定方法 vs 自由函数的划分）

这是整个库设计的中心规则，所有划分都源于此：

1. **方法的类型参数是全新、独立的，不能给接收者的 `T` 追加约束。** `Seq[T any]` 在类型层把 `T` 声明为 `any`，因此任何要求**元素本身**满足 `comparable` / `cmp.Ordered` / 数值约束的操作**不能做成 `Seq[T]` 的方法**，只能是自由函数。例：`Distinct`（需 `comparable`）、`Max`（需 `Ordered`）、`Sum`（需 `Numeric`）。
2. **逃生舱一：方法自己的类型参数可以带约束。** 所以"按 key"的变体能回到方法形态。例：`DistinctBy[K comparable](key func(T) K)`、`GroupBy[K comparable]`，因为 `K` 是方法自己的参数，不是接收者的 `T`。
3. **逃生舱二：约束型子类型把约束钉在类型上，恢复链式。** 引入 `SeqComparable[T comparable]` / `SeqOrdered[T cmp.Ordered]` / `SeqNumeric[T Numeric]`，其上的 `Distinct`/`Max`/`Sum` 是方法。进入这些类型的入口（`Comparable`/`Ordered`/`Numbers`）必须是自由函数（约束 `T`），但约束间降级（`Numeric→Ordered→comparable`）可做成方法，因强约束满足弱约束。详见 FR-8。
4. **涉及多个发生器类型或特定实例化接收者的操作只能是自由函数。** 例：`Zip[A, B](a Seq[A], b Seq[B])`（两个独立类型）、`Flatten[T](s Seq[Seq[T]])`（接收者必须是 `Seq[Seq[T]]` 这一具体实例化，方法无法泛型地挂上去）。
5. **泛型方法不满足接口**（提案的硬线）——本库不尝试提供集合的多态接口抽象。

### 已知的能力边界（相对 Scala）

- 无 for-comprehension 语法糖：一切都是裸方法链。
- 无 HKT / type class：无法抽象 Functor/Monad，每个方法都是具体类型上的具体方法。
- 无普适相等性：`==` 是约束而非普适能力，故 `Distinct`/`Max` 等必须是自由函数。
- 无元组：`Zip` 等返回自定义 `Pair[A,B]` 结构体。

## Goals

- 产出一份**完整、无歧义**的 API 清单，每条标注归属（方法 / 自由函数）+ 一句语义。
- 明确 `Seq[T]` 与 `Seq2[K,V]` 两个核心类型的职责边界与互转入口。
- 让划分规则可被机械验证：凡约束 `T` 本身者必为自由函数，凡仅用方法自带约束参数者可为方法。
- 覆盖 `lo` 常用操作 + Scala 集合可移植部分 + 合理的长尾操作。
- 为后续 `/prd-to-spec`（技术设计）与 `/to-issues`（拆 Issue）提供可直接落地的 API 契约。

## User Stories

### US-001: 定义核心类型与 Pair 辅助类型
**Description:** 作为库作者，我需要定义 `Seq[T]`、`Seq2[K,V]` 定义类型及 `Pair[A,B]`，作为所有 API 的载体。

**Acceptance Criteria:**
- [ ] `type Seq[T any] iter.Seq[T]` 定义，可与 `iter.Seq[T]` 零成本互转
- [ ] `type Seq2[K, V any] iter.Seq2[K, V]` 定义，可与 `iter.Seq2[K,V]` 零成本互转
- [ ] `type Pair[A, B any] struct { Left A; Right B }` 定义
- [ ] `type Numeric` 约束（整型 + 浮点）定义，供数值自由函数使用
- [ ] `go build ./...` 通过，`go vet` 无告警

### US-002: 构造入口自由函数（Seq[T]）
**Description:** 作为使用者，我需要从 slice、可变参数、范围、channel 等来源创建 `Seq[T]` 以开始链式调用。

**Acceptance Criteria:**
- [ ] 实现 FR-2 列出的全部构造函数，签名与文档一致
- [ ] `From([]int{1,2,3}).Collect()` 返回 `[1 2 3]`
- [ ] `Range(0, 3).Collect()` 返回 `[0 1 2]`
- [ ] 无限源（`Generate`/`Iterate`/`Repeat` 无限版）配合 `Take` 不死循环
- [ ] 单元测试覆盖每个构造函数，`go test ./...` 通过

### US-003: 中间转换方法（惰性，返回 Seq）
**Description:** 作为使用者，我需要 Map/Filter/FlatMap 等惰性中间操作，可链式拼接而不立即物化。

**Acceptance Criteria:**
- [ ] 实现 FR-3 列出的全部中间方法
- [ ] `Map[U]` 能改变元素类型：`From([]int{1,2}).Map(strconv.Itoa).Collect()` 返回 `["1","2"]`
- [ ] 中间操作惰性：在终结操作前不遍历源（用计数器副作用验证）
- [ ] `TakeRight`/`DropRight`/`Slice`/`Init` 等需物化的中间操作语义正确
- [ ] 链式可拼接：`Filter(...).Map(...).Take(n)` 编译且语义正确
- [ ] 单元测试覆盖，`go test ./...` 通过

### US-004: 终结方法（驱动求值，返回非 Seq）
**Description:** 作为使用者，我需要 Collect/Fold/Reduce/Find 等终结操作来驱动管道并取出结果。

**Acceptance Criteria:**
- [ ] 实现 FR-4 列出的全部终结方法
- [ ] `Reduce` 在空序列时返回 `(zero, false)`
- [ ] `Find`/`First`/`Last`/`Nth` 命中返回 `(v, true)`，未命中返回 `(zero, false)`
- [ ] `FindLast`/`FindLastIndex` 返回末个满足者及其索引
- [ ] `CountBy` 返回满足谓词的个数；`GroupCount` 返回按 key 分组的 `map[K]int`
- [ ] `MaxByKey`/`MinByKey`/`SumBy`/`MeanBy` 按投影函数聚合
- [ ] `Any`/`All`/`None` 短路求值（命中即停，用计数器验证）
- [ ] 单元测试覆盖含空序列边界，`go test ./...` 通过

### US-005: 约束型自由函数（comparable / Ordered / Numeric）
**Description:** 作为使用者，我需要 Distinct/Max/Sum/Sort 等依赖元素约束的操作，它们因语言约束规则只能是自由函数。

**Acceptance Criteria:**
- [ ] 实现 FR-5 列出的全部约束型自由函数
- [ ] `Distinct[T comparable]` 去重保持首次出现顺序
- [ ] `Compact` 去掉零值；`Without` 排除指定值
- [ ] `SymmetricDifference` 返回对称差（lodash xor 语义）
- [ ] `Max`/`Min`/`Sum`/`Product`/`Mean` 在空序列返回约定值（见 FR-5 说明）
- [ ] 每个函数签名带正确约束（`comparable`/`cmp.Ordered`/`Numeric`），编译强制
- [ ] 单元测试覆盖，`go test ./...` 通过

### US-006: 多序列 / 嵌套自由函数（Zip / Flatten / Concat）
**Description:** 作为使用者，我需要组合多个序列或展平嵌套序列，这些因多类型参数或特定实例化接收者只能是自由函数。

**Acceptance Criteria:**
- [ ] 实现 FR-6 列出的全部多序列函数
- [ ] `Zip[A,B]` 在较短序列耗尽时停止
- [ ] `ZipWith` 用合并函数产出 `Seq[C]`；`ZipMap` 配成 `map[K]V`
- [ ] `Flatten[T](Seq[Seq[T]])` 正确展平一层
- [ ] `Concat` 按顺序拼接多个序列
- [ ] 单元测试覆盖，`go test ./...` 通过

### US-007: Seq2[K,V] 方法与自由函数
**Description:** 作为使用者，我需要在 key-value / 带索引 / map 场景使用 `Seq2[K,V]` 的转换与终结操作。

**Acceptance Criteria:**
- [ ] 实现 FR-7 列出的全部 `Seq2` 方法与配套自由函数
- [ ] `Keys()`/`Values()` 返回 `Seq[K]`/`Seq[V]`
- [ ] `ToMap[K comparable,V]` 是自由函数（因 map key 需 comparable）
- [ ] `Associate[T,K comparable,V]` 由 `Seq[T]` 投影出 `Seq2[K,V]`
- [ ] `FromMap`/`Enumerate` 入口可用
- [ ] 单元测试覆盖，`go test ./...` 通过

### US-008: 约束型子类型与链式恢复（SeqComparable / SeqOrdered / SeqNumeric）
**Description:** 作为使用者，我需要在数值/可比较/可排序元素上链式调用 `Sum`/`Max`/`Distinct`，而不必退回从内往外的自由函数嵌套。

**Acceptance Criteria:**
- [ ] 实现 FR-8 列出的三个约束型子类型及其方法、入口函数、降级方法
- [ ] 入口为自由函数：`Comparable`/`Ordered`/`Numbers` 把 `Seq[T]` 转成对应子类型
- [ ] 降级为方法：`SeqNumeric.Ordered()`、`SeqOrdered.Comparable()` 编译通过
- [ ] `Numbers(From([]int{1,2,3})).Distinct().Sum()` 全程链式且结果正确
- [ ] 保持 `T` 的中间方法（`Filter`/`Take` 等）在子类型上返回同类型，链不断
- [ ] 单元测试覆盖（含 string 走 `Comparable().Distinct()`、float 走 `Numbers().Mean()`），`go test ./...` 通过

### US-009: API 清单文档与划分理由表
**Description:** 作为维护者与下游，我需要一份权威的 API 清单文档，逐条标注归属与理由，作为契约。

**Acceptance Criteria:**
- [ ] 项目内生成 `API.md`，按"方法 / 自由函数"分区列出全部 API 签名 + 一句语义
- [ ] 每个"自由函数"条目注明为何不能是方法（约束 T / 多类型 / 嵌套实例化）
- [ ] 约束型子类型一节说明"入口为函数、降级为方法"的理由
- [ ] 文档与代码签名一致（可人工核对）
- [ ] Markdown 渲染无格式错误

## Functional Requirements

### FR-1: 核心类型
- FR-1: 系统须定义 `Seq[T any]`（`iter.Seq[T]` 的定义类型）、`Seq2[K,V any]`、`Pair[A,B any]` 结构体、`Tuple3[A,B,C any]`、`Tuple4[A,B,C,D any]` 结构体、`Numeric` 约束。

### FR-2: 构造入口（自由函数 —— 无接收者，故均为函数）
- FR-2: 系统须提供以下构造自由函数：

| 函数 | 语义 |
|---|---|
| `From[T any](s []T) Seq[T]` | 从 slice 创建序列 |
| `Of[T any](items ...T) Seq[T]` | 从可变参数创建序列 |
| `Empty[T any]() Seq[T]` | 空序列 |
| `Range(start, end int) Seq[int]` | `[start, end)` 整数序列 |
| `RangeStep(start, end, step int) Seq[int]` | 带步长的整数序列 |
| `Repeat[T any](n int, v T) Seq[T]` | 重复 `v` 共 `n` 次 |
| `RepeatInf[T any](v T) Seq[T]` | 无限重复 `v` |
| `Generate[T any](f func() T) Seq[T]` | 无限序列，每次调用 `f` 产元素 |
| `Iterate[T any](init T, f func(T) T) Seq[T]` | 无限序列 `init, f(init), f(f(init))...` |
| `FromChannel[T any](ch <-chan T) Seq[T]` | 从 channel 创建（一次性源） |
| `FromMap[K comparable, V any](m map[K]V) Seq2[K,V]` | 从 map 创建 Seq2 |

### FR-3: 中间转换方法（`Seq[T]` 上的惰性方法，返回新 Seq）
- FR-3: 系统须在 `Seq[T]` 上提供以下方法：

| 方法 | 语义 | 备注 |
|---|---|---|
| `Map[U any](f func(T) U) Seq[U]` | 元素 `T→U` 变换 | 1.27 泛型方法核心 |
| `FlatMap[U any](f func(T) Seq[U]) Seq[U]` | 变换为子序列再展平 | |
| `FilterMap[U any](f func(T) (U, bool)) Seq[U]` | 过滤 + 变换合一 | |
| `Filter(pred func(T) bool) Seq[T]` | 保留满足 `pred` 的元素 | |
| `Reject(pred func(T) bool) Seq[T]` | 剔除满足 `pred` 的元素 | |
| `Take(n int) Seq[T]` | 取前 `n` 个 | |
| `Drop(n int) Seq[T]` | 跳过前 `n` 个 | |
| `TakeRight(n int) Seq[T]` | 取末 `n` 个（内部物化，lodash takeRight） | |
| `DropRight(n int) Seq[T]` | 去掉末 `n` 个（内部物化，lodash dropRight） | |
| `Slice(start, end int) Seq[T]` | 子区间 `[start, end)`（lodash slice） | |
| `Init() Seq[T]` | 去掉最后一个元素（Scala init，内部物化） | |
| `Tail() Seq[T]` | 去掉第一个元素（Scala tail，等价 `Drop(1)`） | |
| `TakeWhile(pred func(T) bool) Seq[T]` | 取到第一个不满足为止 | |
| `DropWhile(pred func(T) bool) Seq[T]` | 跳过开头满足的元素 | |
| `Scan[U any](init U, f func(U, T) U) Seq[U]` | 发射每步累积值（含 init） | |
| `Peek(f func(T)) Seq[T]` | 旁路副作用（调试/日志），透传元素 | |
| `Chunk(size int) Seq[Seq[T]]` | 定长不重叠分块（Scala grouped） | |
| `Window(size, step int) Seq[Seq[T]]` | 滑动窗口（Scala sliding） | |
| `Intersperse(sep T) Seq[T]` | 元素间插入分隔元素 | |
| `DistinctBy[K comparable](key func(T) K) Seq[T]` | 按 key 去重 | `K` 为方法自带约束参数 |
| `SortBy(less func(T, T) bool) Seq[T]` | 按比较器排序（内部物化） | |
| `Reverse() Seq[T]` | 逆序（内部物化） | |
| `Concat(other Seq[T]) Seq[T]` | 在尾部追加另一序列 | |
| `Enumerate() Seq2[int, T]` | 配索引，转为 `Seq2[int,T]`（Scala zipWithIndex） | |

### FR-4: 终结方法（`Seq[T]` 上，驱动求值）
- FR-4: 系统须在 `Seq[T]` 上提供以下终结方法：

| 方法 | 语义 |
|---|---|
| `Collect() []T` | 物化为 slice |
| `ForEach(f func(T))` | 对每个元素执行副作用 |
| `ForEachIndexed(f func(int, T))` | 带索引遍历 |
| `Fold[U any](init U, f func(U, T) U) U` | 左折叠到累积值（Scala foldLeft） |
| `Reduce(f func(T, T) T) (T, bool)` | 无初值归约，空序列返回 `(zero,false)` |
| `Count() int` | 元素个数 |
| `CountBy(pred func(T) bool) int` | 满足 `pred` 的个数 |
| `GroupCount[K comparable](key func(T) K) map[K]int` | 按 key 分组计数（lodash countBy 语义，`K` 为方法自带约束参数） |
| `Find(pred func(T) bool) (T, bool)` | 首个满足者 |
| `FindIndex(pred func(T) bool) (int, bool)` | 首个满足者的索引 |
| `FindLast(pred func(T) bool) (T, bool)` | 末个满足者（lodash findLast，内部物化） |
| `FindLastIndex(pred func(T) bool) (int, bool)` | 末个满足者的索引（内部物化） |
| `Any(pred func(T) bool) bool` | 是否存在满足者（短路，Scala exists） |
| `All(pred func(T) bool) bool` | 是否全部满足（短路，Scala forall） |
| `None(pred func(T) bool) bool` | 是否无人满足 |
| `First() (T, bool)` | 首元素 |
| `Last() (T, bool)` | 末元素 |
| `Nth(n int) (T, bool)` | 第 `n` 个元素 |
| `IsEmpty() bool` | 是否为空 |
| `GroupBy[K comparable](key func(T) K) map[K][]T` | 按 key 分组 |
| `KeyBy[K comparable](key func(T) K) map[K]T` | 按 key 建索引（同 key 后者覆盖） |
| `Partition(pred func(T) bool) ([]T, []T)` | 按谓词二分（满足 / 不满足） |
| `Span(pred func(T) bool) ([]T, []T)` | 在首个不满足处切分（Scala span） |
| `MaxBy(less func(T, T) bool) (T, bool)` | 按比较器取最大 |
| `MinBy(less func(T, T) bool) (T, bool)` | 按比较器取最小 |
| `MaxByKey[K cmp.Ordered](key func(T) K) (T, bool)` | 按投影 key 取最大（lodash maxBy，`K` 为方法自带约束参数） |
| `MinByKey[K cmp.Ordered](key func(T) K) (T, bool)` | 按投影 key 取最小（lodash minBy，`K` 为方法自带约束参数） |
| `SumBy[U Numeric](f func(T) U) U` | 投影后求和（lodash sumBy，`U` 为方法自带约束参数） |
| `MeanBy[U Numeric](f func(T) U) float64` | 投影后求平均（lodash meanBy） |
| `Join(sep string, str func(T) string) string` | 元素转字符串后用 `sep` 拼接（Scala mkString，原 `JoinFunc`） |

### FR-5: 约束型自由函数（约束 `T` 本身 —— 故不能是方法）
- FR-5: 系统须提供以下自由函数，每个因约束 `T` 本身而无法成为方法：

| 函数 | 语义 | 空序列约定 |
|---|---|---|
| `Distinct[T comparable](s Seq[T]) Seq[T]` | 去重，保持首次出现序 | — |
| `Contains[T comparable](s Seq[T], v T) bool` | 是否包含 `v`（短路） | `false` |
| `IndexOf[T comparable](s Seq[T], v T) (int, bool)` | `v` 的首个索引 | `(0,false)` |
| `LastIndexOf[T comparable](s Seq[T], v T) (int, bool)` | `v` 的末个索引（lodash lastIndexOf，内部物化） | `(0,false)` |
| `Count Values` → `CountValues[T comparable](s Seq[T]) map[T]int` | 各值出现次数 | 空 map |
| `Max[T cmp.Ordered](s Seq[T]) (T, bool)` | 最大值 | `(zero,false)` |
| `Min[T cmp.Ordered](s Seq[T]) (T, bool)` | 最小值 | `(zero,false)` |
| `Mean[T Numeric](s Seq[T]) float64` | 平均值（lodash mean） | `0` |
| `Sum[T Numeric](s Seq[T]) T` | 求和 | `zero` |
| `Product[T Numeric](s Seq[T]) T` | 求积 | `1` |
| `Sort[T cmp.Ordered](s Seq[T]) Seq[T]` | 升序排序（内部物化） | 空序列 |
| `Equal[T comparable](a, b Seq[T]) bool` | 逐元素相等 | 两空为 `true` |
| `Compact[T comparable](s Seq[T]) Seq[T]` | 去掉零值元素（lodash compact） | 空序列 |
| `Without[T comparable](s Seq[T], vals ...T) Seq[T]` | 排除指定值（lodash without） | 空序列 |
| `Union[T comparable](seqs ...Seq[T]) Seq[T]` | 并集（去重） | 空序列 |
| `Intersect[T comparable](a, b Seq[T]) Seq[T]` | 交集 | 空序列 |
| `Difference[T comparable](a, b Seq[T]) Seq[T]` | 差集 `a−b` | — |
| `SymmetricDifference[T comparable](a, b Seq[T]) Seq[T]` | 对称差（lodash xor） | — |
| `ToSet[T comparable](s Seq[T]) map[T]struct{}` | 转集合 | 空 map |
| `JoinStrings(s Seq[string], sep string) string` | 字符串序列拼接（自由函数版 Join） | `""` |

### FR-6: 多序列 / 嵌套自由函数（多类型参数或嵌套实例化 —— 故不能是方法）
- FR-6: 系统须提供以下自由函数：

| 函数 | 语义 | 为何不是方法 |
|---|---|---|
| `Zip[A, B any](a Seq[A], b Seq[B]) Seq2[A, B]` | 配对两序列，短者止 | 两个独立类型参数 |
| `ZipWith[A, B, C any](a Seq[A], b Seq[B], f func(A, B) C) Seq[C]` | 配对并用 `f` 合并（lodash zipWith） | 多个独立类型参数 |
| `ZipMap[K comparable, V any](keys Seq[K], vals Seq[V]) map[K]V` | 两序列配成 map（lodash zipObject） | 多类型参数 + 约束 K |
| `Zip3[A, B, C any](a Seq[A], b Seq[B], c Seq[C]) Seq[Tuple3[A,B,C]]` | 三路配对 | 多类型参数 |
| `Zip4[A, B, C, D any](a Seq[A], b Seq[B], c Seq[C], d Seq[D]) Seq[Tuple4[A,B,C,D]]` | 四路配对 | 多类型参数 |
| `Unzip[A, B any](s Seq2[A, B]) (Seq[A], Seq[B])` | 拆分 Seq2 为两序列 | 多返回类型 |
| `Flatten[T any](s Seq[Seq[T]]) Seq[T]` | 展平一层嵌套 | 接收者须为 `Seq[Seq[T]]` 实例化 |
| `Concat[T any](seqs ...Seq[T]) Seq[T]` | 顺序拼接多序列（变参版） | 变参聚合更适合函数 |
| `Interleave[T any](seqs ...Seq[T]) Seq[T]` | 轮流交错取元素 | 变参聚合 |

> 注：`Tuple3[A,B,C]`、`Tuple4[A,B,C,D]` 在 FR-1 同处定义（`Pair` 的多元扩展，封顶四元，更多元用自定义 struct）。

### FR-7: Seq2[K,V] 方法与自由函数
- FR-7a: 系统须在 `Seq2[K,V]` 上提供以下方法：

| 方法 | 语义 |
|---|---|
| `MapValues[U any](f func(V) U) Seq2[K, U]` | 仅变换 value |
| `MapKeys[J any](f func(K) J) Seq2[J, V]` | 仅变换 key |
| `Map[J, U any](f func(K, V) (J, U)) Seq2[J, U]` | 同时变换 key/value |
| `Filter(pred func(K, V) bool) Seq2[K, V]` | 按 (k,v) 过滤 |
| `Keys() Seq[K]` | 投影出 key 序列 |
| `Values() Seq[V]` | 投影出 value 序列 |
| `ForEach(f func(K, V))` | 遍历副作用 |
| `Fold[U any](init U, f func(U, K, V) U) U` | 左折叠 |
| `Count() int` | 对数 |
| `Find(pred func(K, V) bool) (K, V, bool)` | 首个满足的 (k,v) |
| `Any(pred func(K, V) bool) bool` | 存在性（短路） |
| `All(pred func(K, V) bool) bool` | 全称（短路） |

- FR-7b: 系统须提供以下 `Seq2` 配套自由函数（约束 K 本身，故为函数）：

| 函数 | 语义 |
|---|---|
| `ToMap[K comparable, V any](s Seq2[K, V]) map[K]V` | 物化为 map |
| `CollectPairs[K, V any](s Seq2[K, V]) []Pair[K, V]` | 物化为 Pair slice |
| `Entries[K, V any](pairs []Pair[K, V]) Seq2[K, V]` | 从 Pair slice 创建 |
| `Associate[T any, K comparable, V any](s Seq[T], f func(T) (K, V)) Seq2[K, V]` | 由 `Seq[T]` 经 `f` 投影为 `Seq2[K,V]`（收口 Open Question；约束 K，故为自由函数） |

### FR-8: 约束型子类型（把约束钉在类型上以恢复链式）
- FR-8a: 系统须定义三个约束型子类型（均为 `iter.Seq[T]` 的定义类型）：

| 类型 | 约束 | 其上的约束型方法 |
|---|---|---|
| `SeqComparable[T comparable]` | `comparable` | `Distinct()`、`Contains(v)`、`IndexOf(v)`、`CountValues()`、`ToSet()`、`Union(others...)`、`Intersect(o)`、`Difference(o)`、`Equal(o)` |
| `SeqOrdered[T cmp.Ordered]` | `cmp.Ordered` | `Max()`、`Min()`、`Sort()`（外加继承自 comparable 的全部） |
| `SeqNumeric[T Numeric]` | `Numeric` | `Sum()`、`Product()`、`Mean()`（外加继承自 ordered 的全部） |

- FR-8b: 系统须提供进入这些类型的**入口自由函数**（约束 `T`，故为函数）：

| 函数 | 语义 |
|---|---|
| `Comparable[T comparable](s Seq[T]) SeqComparable[T]` | 转为可比较序列 |
| `Ordered[T cmp.Ordered](s Seq[T]) SeqOrdered[T]` | 转为可排序序列 |
| `Numbers[T Numeric](s Seq[T]) SeqNumeric[T]` | 转为数值序列（名字避开 `Numeric` 约束本身） |

- FR-8c: 系统须提供约束间的**降级方法**（强约束满足弱约束，故可为方法）：

| 方法 | 语义 |
|---|---|
| `(SeqNumeric[T]) Ordered() SeqOrdered[T]` | 降级为可排序序列 |
| `(SeqOrdered[T]) Comparable() SeqComparable[T]` | 降级为可比较序列 |
| `(SeqComparable[T]) Seq() Seq[T]` | 退回裸序列（用于 `Map` 等改变 T 的操作前） |

- FR-8d: 系统须在每个子类型上**重新暴露保持 `T` 的中间方法**（`Filter`/`Reject`/`Take`/`Drop`/`TakeWhile`/`DropWhile`/`Peek` 等），返回同一子类型，使链不断。改变 `T` 的 `Map`/`FlatMap` 不在子类型上提供，需先 `Seq()` 退回裸序列。
- FR-8e: 裸 `Seq[T]` 上的等价自由函数（FR-5 的 `Distinct`/`Max`/`Sum` 等）保留，作为不转类型时的兜底，与子类型方法语义一致。

### FR-9: 文档
- FR-9: 系统须在项目根生成 `API.md`，按"方法 / 自由函数"分区列出全部 API 的签名与一句语义，自由函数须注明无法成为方法的原因类别，并设专节说明约束型子类型"入口为函数、降级为方法"的规则。

## Non-Goals (Out of Scope)

- **错误处理 / `Seq2[T, error]` 短路链**：本版明确不做。Scala 的 `Try`/`Either`/for-comprehension 错误流转在 Go 惰性序列上无优雅对应，留作后续独立 PRD。
- **HKT / type class 抽象**（Functor/Monad/可抽象的 `Collection` 接口）：语言不支持，不尝试。
- **for-comprehension 语法糖**：Go 无对应语法，不模拟。
- **并行执行**（对标 `lo/parallel`）：本版只做顺序惰性管道；并行留后续。
- **原地可变操作**（对标 `lo/mutable`）：与惰性不可变定位冲突，不做。
- **具体实现与性能调优**：本 PRD 只定义 API 契约，不含实现代码与基准。
- **泛型方法满足接口**：提案硬线，不提供集合多态接口。
- **任意深度展平**（lodash `flattenDeep`/`flattenDepth`）：Go 类型系统无法表达「任意深度嵌套序列」这一类型，故只提供固定一层的 `Flatten`；更深嵌套须显式多次调用或自定义结构。
- **Scala 式高元数 Tuple（Tuple5–Tuple22）**：本库元组封顶 `Tuple4`。Go 无元组字面量与模式解构语法，高元数 `TupleN` 字段只能是 `Field1..FieldN` 这类无语义名，可读性差且 `ZipN`/`Unzip`/测试需逐元维护；更多字段的聚合应使用具名 `struct`（字段名表意），故不提供五元及以上元组与对应 `Zip5+`。
- **随机性操作**（lodash `sample`/`sampleSize`/`shuffle`）：引入随机性与「惰性 + 可重复遍历」语义冲突，本版不做，留作后续独立 PRD。
- **函数式工具**（lodash `debounce`/`throttle`/`memoize`/`once`/`curry` 等）：属于函数控制流而非集合操作，明确不在本库范围。

## Technical Considerations

- **目标 Go 版本：1.27**（泛型方法）。`go.mod` 声明 `go 1.27`，本地需安装 1.27 工具链（当前为 `go1.27rc1`，正式版发布后切换）。编译/测试在 1.27 工具链下进行。
- **定义类型而非 struct 包装**：`type Seq[T any] iter.Seq[T]` 与 `iter.Seq` 零成本互转，可无缝喂给 `slices.Collect`、`maps.Keys` 等标准库。
- **惰性语义**：中间方法返回未求值的 `Seq`，仅终结方法驱动遍历；需在 `API.md` 标注每个方法是"中间(intermediate)"还是"终结(terminal)"。
- **re-iteration 语义**：源自 slice 的 `Seq` 可重复遍历；源自 channel 的为一次性。须在文档明确。
- **`cmp.Ordered`** 来自标准库 `cmp` 包；`Numeric` 需自定义约束。
- **性能提示**：深链 = 嵌套 `yield` 闭包调用，热点内循环可能不及手写 loop；文档应给出适用边界提示。

## Success Metrics

- API 清单 100% 覆盖 FR-2 至 FR-9 列出的条目，无遗漏。
- 每个裸 `Seq[T]` 上的自由函数都能对应到三类判定理由之一（约束 T / 多类型参数 / 嵌套实例化）。
- 划分规则可机械复核：裸 `Seq[T]` 的清单中不存在"约束 T 本身却被列为方法"的条目；约束型子类型上的方法均满足"约束已钉在类型上"或"降级到弱约束"。
- 下游可直接据此清单运行 `/to-issues` 拆分为可实现 Issue，无需二次澄清 API 形态。

## Open Questions

> 以下决策已在评审中拍板（保留备查）：
> - **MapIndexed**：不单独提供，索引一律走 `Enumerate().Map(...)`，保持 `Map(func(T) U)` 单一形态。
> - **Tuple 元数**：提供到 `Tuple4` 封顶，更多元用自定义 struct。
> - **Join 命名**：统一为 `Join` 系列 —— 方法 `Join(sep, fn)`、自由函数 `JoinStrings(s, sep)`。
> - **计数命名**：`CountBy(pred)` = 谓词计数（保持原义）；按 key 分组计数命名为 `GroupCount[K]`。
> - **By vs ByKey 约定**：`SumBy`/`MeanBy` = 按投影值聚合；`MaxByKey`/`MinByKey` = 按投影 key 取极值；`MaxBy`/`MinBy` = 比较器版。文档须明确该约定。

剩余待验证项（非设计决策，落实现/测试阶段确认）：

- `Chunk`/`Window` 返回 `Seq[Seq[T]]` 的内层是否也支持完整方法链（应当——它就是 `Seq`），需在测试中验证嵌套链可用。
