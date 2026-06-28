# API — `seq`

> 权威 API 清单。本文件按「方法 / 自由函数」分区列出 `seq` 库全部 API 的签名与一句语义，并标注每个方法是**中间(intermediate)**还是**终结(terminal)**，每个自由函数注明为何不能是方法。文档与代码签名一致（可人工核对）。
>
> 目标 Go 版本：**1.27**（泛型方法，golang/go#77273）。`go.mod` 声明 `go 1.27`。

## 核心类型（FR-1）

| 类型 | 定义 | 说明 |
|---|---|---|
| `Seq[T any]` | `iter.Seq[T]` 的定义类型 | 与 `iter.Seq[T]` 零成本互转；`T` 声明为 `any` |
| `Seq2[K, V any]` | `iter.Seq2[K, V]` 的定义类型 | 键值迭代器，与 `iter.Seq2` 零成本互转 |
| `Pair[A, B any]` | `struct { Left A; Right B }` | 二元组，`Zip` 产出、`Unzip`/`Entries` 消费 |
| `Tuple3[A, B, C any]` | `struct { First A; Second B; Third C }` | 三元组，`Zip3` 产出 |
| `Tuple4[A, B, C, D any]` | `struct { First A; Second B; Third C; Fourth D }` | 四元组，`Zip4` 产出（封顶四元） |
| `Numeric` | `interface { ~int \| ~int8 \| … \| ~float64 }` | 整型 + 浮点约束，供数值自由函数与 `SeqNumeric` 使用 |

---

## 一、`Seq[T]` 方法

### 1.1 中间方法（intermediate，返回新 `Seq`，惰性）

| 方法 | 语义 |
|---|---|
| `Map[U any](f func(T) U) Seq[U]` | 元素 `T→U` 变换（**1.27 泛型方法核心**，可改变元素类型） |
| `FlatMap[U any](f func(T) Seq[U]) Seq[U]` | 变换为子序列再展平一层 |
| `FilterMap[U any](f func(T) (U, bool)) Seq[U]` | 过滤 + 变换合一 |
| `Filter(pred func(T) bool) Seq[T]` | 保留满足 `pred` 的元素 |
| `Reject(pred func(T) bool) Seq[T]` | 剔除满足 `pred` 的元素 |
| `Take(n int) Seq[T]` | 取前 `n` 个（`n<=0` 为空） |
| `Drop(n int) Seq[T]` | 跳过前 `n` 个 |
| `TakeWhile(pred func(T) bool) Seq[T]` | 取到第一个不满足为止 |
| `DropWhile(pred func(T) bool) Seq[T]` | 跳过开头满足的元素 |
| `Scan[U any](init U, f func(U, T) U) Seq[U]` | 发射每步累积值（含 `init`，左扫描） |
| `Peek(f func(T)) Seq[T]` | 旁路副作用（调试/日志），透传元素 |
| `Intersperse(sep T) Seq[T]` | 元素间插入分隔元素 |
| `Concat(other Seq[T]) Seq[T]` | 在尾部追加另一序列（二源方法版） |
| `DistinctBy[K comparable](key func(T) K) Seq[T]` | 按 key 去重（`K` 为方法自带约束参数） |

### 1.2 中间方法（需内部物化，结果可重复遍历）

| 方法 | 语义 |
|---|---|
| `TakeRight(n int) Seq[T]` | 取末 `n` 个（lodash takeRight） |
| `DropRight(n int) Seq[T]` | 去掉末 `n` 个（lodash dropRight） |
| `Slice(start, end int) Seq[T]` | 子区间 `[start, end)`，越界自动钳位（lodash slice） |
| `Init() Seq[T]` | 去掉最后一个元素（Scala init） |
| `Tail() Seq[T]` | 去掉第一个元素（Scala tail，等价 `Drop(1)`） |
| `SortBy(less func(a, b T) bool) Seq[T]` | 按比较器排序（稳定） |
| `Reverse() Seq[T]` | 逆序 |
| `Enumerate() Seq2[int, T]` | 配索引，转为 `Seq2[int, T]`（Scala zipWithIndex） |

### 1.3 ⚠ 偏差：`Chunk` / `Window`（返回 `iter.Seq[Seq[T]]`）

> **签名偏差**：PRD 原定 `Chunk(size int) Seq[Seq[T]]` 与 `Window(size, step int) Seq[Seq[T]]`。Go 1.27rc1 禁止 `Seq[T]` 上的泛型方法实例化 `Seq[Seq[T]]`（instantiation cycle: `T` instantiated as `Seq[T]`）。这是**自嵌套**形实例化循环（golang/go#80109，Griesemer 判 working as intended：单态化生成全部方法，`A[int]`→`A[A[int]]`→… 无穷展开，论证成立），故预期**不会**随稳定版恢复为 `Seq[Seq[T]]`。二者返回底层 `iter.Seq[Seq[T]]`，**内层为完整 `Seq[T]`（可继续方法链）**；如需对外层继续链式，用零成本包装 `Seq[Seq[T]](s.Chunk(n))`。详见 issue #6 笔记。

| 方法 | 语义 |
|---|---|
| `Chunk(size int) iter.Seq[Seq[T]]` | 定长不重叠分块（Scala grouped） |
| `Window(size, step int) iter.Seq[Seq[T]]` | 滑动窗口（Scala sliding） |

### 1.4 终结方法（terminal，驱动求值，返回非 `Seq`）

| 方法 | 语义 |
|---|---|
| `Collect() []T` | 物化为 slice |
| `ForEach(f func(T))` | 对每个元素执行副作用 |
| `ForEachIndexed(f func(int, T))` | 带索引遍历 |
| `Fold[U any](init U, f func(U, T) U) U` | 左折叠（Scala foldLeft，`U` 为方法自带类型参数） |
| `Reduce(f func(T, T) T) (T, bool)` | 无初值归约；空序列返回 `(zero, false)` |
| `Count() int` | 元素个数 |
| `CountBy(pred func(T) bool) int` | 满足 `pred` 的个数 |
| `GroupCount[K comparable](key func(T) K) map[K]int` | 按 key 分组计数（`K` 为方法自带约束参数） |
| `Find(pred func(T) bool) (T, bool)` | 首个满足者；未命中 `(zero, false)` |
| `FindIndex(pred func(T) bool) (int, bool)` | 首个满足者的索引；未命中 `(0, false)` |
| `FindLast(pred func(T) bool) (T, bool)` | 末个满足者；未命中 `(zero, false)` |
| `FindLastIndex(pred func(T) bool) (int, bool)` | 末个满足者的索引；未命中 `(0, false)` |
| `Any(pred func(T) bool) bool` | 是否存在满足者（短路，Scala exists） |
| `All(pred func(T) bool) bool` | 是否全部满足（短路，Scala forall） |
| `None(pred func(T) bool) bool` | 是否无人满足（短路） |
| `First() (T, bool)` | 首元素；空 `(zero, false)` |
| `Last() (T, bool)` | 末元素；空 `(zero, false)` |
| `Nth(n int) (T, bool)` | 第 `n` 个（`n<0` 或越界返回 `(zero, false)`） |
| `IsEmpty() bool` | 是否为空（短路） |
| `GroupBy[K comparable](key func(T) K) map[K][]T` | 按 key 分组 |
| `KeyBy[K comparable](key func(T) K) map[K]T` | 按 key 建索引（同 key 后者覆盖） |
| `Partition(pred func(T) bool) ([]T, []T)` | 按谓词二分 |
| `Span(pred func(T) bool) ([]T, []T)` | 在首个不满足处切分（Scala span） |
| `MaxBy(less func(a, b T) bool) (T, bool)` | 按比较器取最大；空 `(zero, false)` |
| `MinBy(less func(a, b T) bool) (T, bool)` | 按比较器取最小；空 `(zero, false)` |
| `MaxByKey[K cmp.Ordered](key func(T) K) (T, bool)` | 按投影 key 取最大（`K` 为方法自带约束参数） |
| `MinByKey[K cmp.Ordered](key func(T) K) (T, bool)` | 按投影 key 取最小（`K` 为方法自带约束参数） |
| `SumBy[U Numeric](f func(T) U) U` | 投影后求和（`U` 为方法自带约束参数） |
| `MeanBy[U Numeric](f func(T) U) float64` | 投影后求平均；空 `0` |
| `Join(sep string, str func(T) string) string` | 元素转字符串后用 `sep` 拼接（Scala mkString） |

### 命名约定（`By` / `ByKey`）

- `By` = 按投影**值**聚合（方法）：`SumBy`/`MeanBy`/`MaxBy`/`MinBy`
- `ByKey` = 按投影 **key** 取极值（方法）：`MaxByKey`/`MinByKey`
- 裸 `Max`/`Min`/`Sum` = 约束 `T` 本身的**自由函数**；想链式调用则转 `SeqNumeric`：`From(xs).Numbers().Sum()`

---

## 二、`Seq2[K, V]` 方法（FR-7a）

### 2.1 中间方法

| 方法 | 语义 |
|---|---|
| `MapValues[U any](f func(V) U) Seq2[K, U]` | 仅变换 value |
| `MapKeys[J any](f func(K) J) Seq2[J, V]` | 仅变换 key（映射后 key 可能碰撞，`ToMap` 时后者覆盖） |
| `Map[J, U any](f func(K, V) (J, U)) Seq2[J, U]` | 同时变换 key/value |
| `Filter(pred func(K, V) bool) Seq2[K, V]` | 按 `(k,v)` 过滤 |
| `Keys() Seq[K]` | 投影出 key 序列（惰性，返回完整 `Seq`，可链） |
| `Values() Seq[V]` | 投影出 value 序列（惰性，可链） |

### 2.2 终结方法

| 方法 | 语义 |
|---|---|
| `ForEach(f func(K, V))` | 遍历副作用 |
| `Fold[U any](init U, f func(U, K, V) U) U` | 左折叠 |
| `Count() int` | 对数 |
| `Find(pred func(K, V) bool) (K, V, bool)` | 首个满足的 `(k,v)`；未命中 `(zero, zero, false)` |
| `Any(pred func(K, V) bool) bool` | 存在性（短路） |
| `All(pred func(K, V) bool) bool` | 全称（短路） |

---

## 三、自由函数

> **划分铁律**：凡约束元素类型 `T` 本身（`comparable` / `cmp.Ordered` / `Numeric`）、或接收者须为特定实例化（如 `Seq[Seq[T]]`，接收者类型实参须为标识符）的操作，语言层面无法成为 `Seq[T]` 方法，故只能是自由函数。多源组合器（Zip 族、`Concat`、`Interleave`）在 1.27 下可带类型参数写成方法，取自由函数形式是为多源对称；其中 `Zip`/`Zip3`/`Zip4` 若返回同一定义类型 `Seq[Pair[…]]`/`Seq[Tuple…]` 会触发实例化循环（go1.27rc1 实测，`T instantiated as Pair[T,R]`），须返回 `Seq2`/底层 `iter.Seq`。该循环分两类（详见末节「工具链说明」）：**自嵌套**形（`Seq[Seq[T]]`，golang/go#80109 已判 working as intended）与**派生非自嵌套**形（`Seq[Pair[T,R]]`，golang/go#80172 仍 OPEN、无 Go 团队裁决，等价自由函数可编译）。下表每条注明理由。

### 3.1 构造入口（FR-2）—— 无接收者，故均为函数

| 函数 | 语义 |
|---|---|
| `From[T any](s []T) Seq[T]` | 从 slice 创建（按引用，可重复遍历） |
| `Of[T any](items ...T) Seq[T]` | 从可变参数创建 |
| `Empty[T any]() Seq[T]` | 空序列 |
| `Range(start, end int) Seq[int]` | `[start, end)` 整数序列（`start>end` 时降序） |
| `RangeStep(start, end, step int) Seq[int]` | 带步长；`step<=0` 为空 |
| `Repeat[T any](n int, v T) Seq[T]` | 重复 `v` 共 `n` 次（`n<=0` 为空） |
| `RepeatInf[T any](v T) Seq[T]` | 无限重复 `v`（须由 `Take` 等约束消费） |
| `Generate[T any](f func() T) Seq[T]` | 无限序列，每次调用 `f` 产元素 |
| `Iterate[T any](init T, f func(T) T) Seq[T]` | `init, f(init), f(f(init))...` |
| `FromChannel[T any](ch <-chan T) Seq[T]` | 从 channel 创建（一次性源） |
| `FromMap[K comparable, V any](m map[K]V) Seq2[K, V]` | 从 map 创建 `Seq2` |

### 3.2 约束型自由函数（FR-5）—— 约束 `T` 本身，故不能是方法

| 函数 | 语义 | 空序列约定 |
|---|---|---|
| `Distinct[T comparable](s Seq[T]) Seq[T]` | 去重，保持首次出现序 | — |
| `Contains[T comparable](s Seq[T], v T) bool` | 是否包含 `v`（短路） | `false` |
| `IndexOf[T comparable](s Seq[T], v T) (int, bool)` | `v` 的首个索引 | `(0, false)` |
| `LastIndexOf[T comparable](s Seq[T], v T) (int, bool)` | `v` 的末个索引 | `(0, false)` |
| `CountValues[T comparable](s Seq[T]) map[T]int` | 各值出现次数 | 空 map |
| `Max[T cmp.Ordered](s Seq[T]) (T, bool)` | 最大值 | `(zero, false)` |
| `Min[T cmp.Ordered](s Seq[T]) (T, bool)` | 最小值 | `(zero, false)` |
| `Mean[T Numeric](s Seq[T]) float64` | 平均值 | `0` |
| `Sum[T Numeric](s Seq[T]) T` | 求和 | `zero` |
| `Product[T Numeric](s Seq[T]) T` | 求积 | `1` |
| `Sort[T cmp.Ordered](s Seq[T]) Seq[T]` | 升序排序（内部物化） | 空序列 |
| `Equal[T comparable](a, b Seq[T]) bool` | 逐元素相等（流式短路） | 两空为 `true` |
| `Compact[T comparable](s Seq[T]) Seq[T]` | 去掉零值元素 | 空序列 |
| `Without[T comparable](s Seq[T], vals ...T) Seq[T]` | 排除指定值 | 空序列 |
| `Union[T comparable](seqs ...Seq[T]) Seq[T]` | 并集（去重，保首次出现序） | 空序列 |
| `Intersect[T comparable](a, b Seq[T]) Seq[T]` | 交集（去重） | 空序列 |
| `Difference[T comparable](a, b Seq[T]) Seq[T]` | 差集 `a−b`（去重） | — |
| `SymmetricDifference[T comparable](a, b Seq[T]) Seq[T]` | 对称差（lodash xor） | — |
| `ToSet[T comparable](s Seq[T]) map[T]struct{}` | 转集合 | 空 map |
| `JoinStrings(s Seq[string], sep string) string` | 字符串序列拼接 | `""` |

### 3.3 多序列 / 嵌套自由函数（FR-6）—— 嵌套实例化不可为方法；多源组合取函数形式

| 函数 | 语义 | 为何取自由函数形式 |
|---|---|---|
| `Zip[A, B any](a Seq[A], b Seq[B]) Seq2[A, B]` | 配对两序列，短者止 | 多源对称；若为方法返 `Seq[Pair[T,R]]` 触发实例化循环（#80172，未定论；返 `Seq2` 可） |
| `ZipWith[A, B, C any](a Seq[A], b Seq[B], f func(A, B) C) Seq[C]` | 配对并用 `f` 合并 | 多源对称（返 `Seq[C]` 可为方法） |
| `ZipMap[K comparable, V any](keys Seq[K], vals Seq[V]) map[K]V` | 两序列配成 map（同 key 后者覆盖） | 终止物化 + 约束 K |
| `Zip3[A, B, C any](a, b, c) Seq[Tuple3[A,B,C]]` | 三路配对 | 多源对称；自引用返回实例化循环（#80172 类） |
| `Zip4[A, B, C, D any](...) Seq[Tuple4[A,B,C,D]]` | 四路配对 | 多源对称；自引用返回实例化循环（#80172 类） |
| `Unzip[A, B any](s Seq2[A, B]) (Seq[A], Seq[B])` | 拆分 `Seq2` 为两序列（物化一次，两侧均可重复遍历） | 多源对称（可为方法） |
| `Flatten[T any](s Seq[Seq[T]]) Seq[T]` | 展平一层嵌套 | 接收者类型实参须为标识符（语言不可） |
| `Concat[T any](seqs ...Seq[T]) Seq[T]` | 顺序拼接多序列（变参版） | 变参聚合；另有 2 源方法 `Concat` |
| `Interleave[T any](seqs ...Seq[T]) Seq[T]` | 轮流交错取元素（跳过已耗尽者，直至全部耗尽） | 变参聚合 |

### 3.4 `Seq2` 配套自由函数（FR-7b）—— 约束 K，故为函数

| 函数 | 语义 |
|---|---|
| `ToMap[K comparable, V any](s Seq2[K, V]) map[K]V` | 物化为 map（同 key 后者覆盖；空 → 非空 map） |
| `CollectPairs[K, V any](s Seq2[K, V]) []Pair[K, V]` | 物化为 `Pair` slice（保序） |
| `Entries[K, V any](pairs []Pair[K, V]) Seq2[K, V]` | 从 `Pair` slice 创建（`CollectPairs` 的逆，可重复遍历） |
| `Associate[T any, K comparable, V any](s Seq[T], f func(T) (K, V)) Seq2[K, V]` | 由 `Seq[T]` 经 `f` 投影为 `Seq2[K,V]`（惰性） |

---

## 四、约束型子类型（FR-8）—— 把约束钉在类型上以恢复链式

> 铁律的副作用：约束 `T` 的操作沦为自由函数后，管道又退回从内往外读（`Sum(Distinct(From(xs)))`）。解法是引入三个约束型子类型，把对 `T` 的约束提前钉在类型上，使其上的 `Distinct`/`Max`/`Sum` 重新成为方法。

### 4.1 子类型定义

| 类型 | 约束 | 定义 |
|---|---|---|
| `SeqComparable[T comparable]` | `comparable` | `iter.Seq[T]` 的定义类型 |
| `SeqOrdered[T cmp.Ordered]` | `cmp.Ordered` | `iter.Seq[T]` 的定义类型 |
| `SeqNumeric[T Numeric]` | `Numeric` | `iter.Seq[T]` 的定义类型 |

### 4.2 入口自由函数（约束 `T`，故为函数）

> 跨越约束边界（从 `any` 进入约束世界）必须由自由函数完成：`Seq[T any]` 的 `T` 是 `any`，无法在方法里要求它满足 `comparable`。

| 函数 | 语义 |
|---|---|
| `Comparable[T comparable](s Seq[T]) SeqComparable[T]` | 转为可比较序列 |
| `Ordered[T cmp.Ordered](s Seq[T]) SeqOrdered[T]` | 转为可排序序列 |
| `Numbers[T Numeric](s Seq[T]) SeqNumeric[T]` | 转为数值序列（名避 `Numeric` 约束本身） |

### 4.3 降级方法（强约束满足弱约束，故可为方法）

> 约束边界之内用方法：`Numeric ⊂ Ordered ⊂ comparable`，强约束天然满足弱约束，降级无需对 `T` 追加新约束。

| 方法 | 语义 |
|---|---|
| `(SeqNumeric[T]) Ordered() SeqOrdered[T]` | 降级为可排序序列 |
| `(SeqOrdered[T]) Comparable() SeqComparable[T]` | 降级为可比较序列 |
| `(SeqComparable[T]) Seq() Seq[T]` | 退回裸序列（用于 `Map` 等改变 `T` 的操作前） |

### 4.4 子类型上的约束型方法

> 「继承」在 Go 中由在更强子类型上**重暴露**同名方法实现，返回**同子类型**（经零成本降级链路由到对应自由函数），使链不断。

**`SeqComparable[T]`**：`Distinct()`、`Contains(v)`、`IndexOf(v)`、`CountValues()`、`ToSet()`、`Union(others...)`、`Intersect(o)`、`Difference(o)`、`Equal(o)`、`Collect()`

**`SeqOrdered[T]`**（继承 comparable 全部 +）：`Max()`、`Min()`、`Sort()`

**`SeqNumeric[T]`**（继承 ordered + comparable 全部 +）：`Sum()`、`Product()`、`Mean()`

### 4.5 子类型上重暴露的 T 保持型中间方法

> 每个 `SeqComparable`/`SeqOrdered`/`SeqNumeric` 均重暴露：`Filter`、`Reject`、`Take`、`Drop`、`TakeWhile`、`DropWhile`、`Peek`，返回**同一子类型**。改变 `T` 的 `Map`/`FlatMap` **不在子类型上提供**，需先 `Seq()` 退回裸序列。

### 4.6 净效果

```go
// 旧：从内往外，两层包裹
sum := seq.Sum(seq.Distinct(seq.From([]int{1, 2, 2, 3})))

// 新：一次入口，之后全程方法链
sum := seq.Numbers(seq.From([]int{1, 2, 2, 3})).Distinct().Sum()
```

### 4.7 与裸 `Seq` 上自由函数的关系（FR-8e）

裸 `Seq[T]` 上的等价自由函数（FR-5 的 `Distinct`/`Max`/`Sum` 等）保留，作为不转类型时的兜底，**与子类型方法语义一致**。

---

## 五、能力边界（相对 Scala）

- 无 for-comprehension 语法糖：一切为裸方法链
- 无 HKT / type class：无法抽象 Functor/Monad
- 无普适相等：`==` 是约束不是普适能力（故 `Distinct`/`Max` 必须约束或转子类型）
- 元组封顶 `Tuple4`；更深嵌套用 `Flatten` 多次调用或自定义 struct
- `Flatten` 仅展平一层（Go 类型系统无法表达任意深度）
- 源可重复遍历性：slice 源可重复遍历；channel 源（`FromChannel`）为一次性

## 六、工具链说明（go1.27rc1）

- `go build ./...` / `go vet ./...` / `go test ./...` 全部通过
- `gofmt -l` 对**泛型方法**签名行报 `method must have no type parameters` —— 这是 go1.27rc1 的 **formatter/parser 误报**（编译器接受、vet/test 通过），非代码缺陷。稳定版 1.27 发布后应消失；当前对这些文件以肉眼核对格式
- `Chunk`/`Window` 的 `iter.Seq[Seq[T]]` 返回类型偏差：属自嵌套实例化循环（golang/go#80109，已判 working as intended），预期不会恢复为 `Seq[Seq[T]]`（见 issue #6 笔记）
