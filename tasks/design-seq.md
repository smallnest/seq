Title: 用 Go 1.27 的泛型方法，给 iter.Seq 装上可链式、惰性的集合管道
Author(s): chaoyuepan
Last updated: 2026-06-24
Discussion at: （暂缺，内部评审中——待建项目 issue 后补）
Status: Draft

> 语言版本：简体中文 | [English](design-seq-en.md)


## Abstract / 摘要

我们提议构建 `seq`：一个围绕标准库惰性迭代器 `iter.Seq` / `iter.Seq2`（Go 1.23+）的泛型集合库，让 `From(xs).Filter(...).Map(...).Collect()` 这种从左到右、可发现、惰性的管道第一次在 Go 里成立。它的地基是 **Go 1.27 的泛型方法特性**（[golang/go#77273](https://github.com/golang/go/issues/77273)）——在此之前，方法不能声明自己的类型参数，`.Map()` 无法表达"输入 `Seq[T]`、输出 `Seq[U]`"，所以 `samber/lo` 这类库只能写成从内往外读的嵌套函数 `lo.Map(lo.Filter(...))`。

最重要的承诺，也是最大的约束：**这是一份面向 Go 1.27 的设计。** 泛型方法提案已被接受并在 Go 1.27 中实现；本地已安装 Go 1.27 RC（`go1.27rc1`），`go.mod` 声明 `go 1.27`，全部 API 可编译、可测试。本文档不提供"退回到旧 Go 也能用"的兼容实现——这是刻意的，理由见 Rationale。唯一尚待的是 1.27 的正式稳定版（当前为 RC）。

本库**只定义 API 契约**，不含实现。它有一条贯穿始终的划分铁律：凡是约束元素类型 `T` 本身的操作（`Distinct` 要 `comparable`、`Max` 要 `Ordered`、`Sum` 要 `Numeric`）只能是自由函数；凡是只用方法自带约束参数的操作（`DistinctBy[K comparable]`、`GroupBy[K comparable]`）才能回到方法形态。这条规则是语言逼出来的，不是风格偏好，下面会讲清原因。

## Background / 背景与动机

先看今天用 `lo` 写一段"取偶数、平方、求和"的真实样子：

```go
// 从内往外读：先看最里层 Filter，再 Map，最后 Sum。
// 阅读顺序和数据流动顺序相反。
sum := lo.Sum(
    lo.Map(
        lo.Filter([]int{1, 2, 3, 4, 5, 6}, func(x int, _ int) bool {
            return x%2 == 0
        }),
        func(x int, _ int) int { return x * x },
    ),
)
```

这段代码有三个具体的痛点：

1. **阅读方向是拧着的。** 数据流是 filter → map → sum，但眼睛得先跳到最内层的 `Filter`，再往外退着读。嵌套越深，越要在括号间反复横跳。

2. **每一步都物化一个新 slice。** `lo.Filter` 返回一个新 `[]int`，`lo.Map` 又返回一个，全程无法惰性短路。想"找到第一个满足条件的偶数平方"就得自己拆开重写，享受不到管道的提前退出。

3. **类型变换处尤其别扭。** `[]int` 经 `Map` 变 `[]string` 时，函数签名 `lo.Map[int, string]` 的类型参数得显式或半显式地标，链不起来。

`lo` 为什么不做成 `xs.Filter(...).Map(...)`？因为在 Go 1.27 之前，**方法不能声明自己的类型参数**。`Map` 要把 `Seq[int]` 变成 `Seq[string]`，需要一个方法级的类型参数 `U`：

```go
// Go 1.27 之前：这行编译不过。方法不允许有自己的 [U any]。
func (s Seq[T]) Map[U any](f func(T) U) Seq[U]
```

这是一条语言级的硬墙。`lo` 不是不想做链式，是做不了。[golang/go#77273](https://github.com/golang/go/issues/77273) 解除了这条限制，链式管道才成为可能——这是本库存在的全部理由。

## Design / 设计

### 核心类型是"定义类型"，不是 struct 包装

我们把 `Seq[T]` 定义为 `iter.Seq[T]` 的定义类型，而不是包一层 struct：

```go
type Seq[T any] iter.Seq[T]      // = func(yield func(T) bool)
type Seq2[K, V any] iter.Seq2[K, V]
```

这样做的收益是零成本互转：任何 `iter.Seq[T]` 能直接当 `Seq[T]` 用，反之亦然，也能喂给标准库的 `slices.Collect`、`maps.Keys`。代价是 `Seq[T]` 在类型层把 `T` 声明成了 `any`——这一点马上会决定一半 API 的归属。

配套两个辅助类型和一个约束：

```go
type Pair[A, B any] struct { Left A; Right B }
type Tuple3[A, B, C any] struct { ... }   // 封顶四元
type Tuple4[A, B, C, D any] struct { ... }
type Numeric interface { ~int | ~int8 | ... | ~float64 }  // 整型 + 浮点
```

### 从最小的链开始

最简单的管道：从 slice 来，过滤，物化回 slice。

```go
seq.From([]int{1, 2, 3, 4}).
    Filter(func(x int) bool { return x%2 == 0 }).
    Collect()                                    // [2 4]
```

读起来和数据流向一致：从左到右。中间的 `Filter` 是**惰性中间操作**——它返回一个新的 `Seq[int]`，此刻没有任何元素被遍历。只有末尾的 `Collect`（**终结操作**）才驱动整条链跑起来。

### 类型变换：1.27 泛型方法的主场

`Map` 能改变元素类型，这是整个库的招牌能力，也是 1.27 之前做不到的那个 `.Map()`：
```go
seq.From([]int{1, 2}).
    Map(strconv.Itoa).        // Seq[int] → Seq[string]
    Collect()                  // ["1" "2"]
```

`Map[U any](f func(T) U) Seq[U]` 的 `U` 就是方法自己声明的类型参数。把开头那段拧巴的 `lo` 代码改写过来：

```go
// 改造后：从左到右，一眼读完，且惰性。
sum := seq.SumBy(
    seq.From([]int{1, 2, 3, 4, 5, 6}).
        Filter(func(x int) bool { return x%2 == 0 }).
        Map(func(x int) int { return x * x }),
)
// 想要"第一个偶数的平方"？只需把末端换成 .First()，链自动短路，不会跑完整个序列。
```

### 划分铁律：约束落在谁身上，决定它是方法还是函数

这是全库设计的中心，所有 API 归属都从这一条推出来。因为 `Seq[T]` 把 `T` 钉死成 `any`，**方法的类型参数是全新独立的，不能反过来给接收者的 `T` 追加约束**。于是：

- 需要 `T` 本身满足 `comparable` 的 `Distinct`、需要 `Ordered` 的 `Max`、需要 `Numeric` 的 `Sum`——**都不能是方法**，只能是自由函数：

```go
func Distinct[T comparable](s Seq[T]) Seq[T]
func Max[T cmp.Ordered](s Seq[T]) (T, bool)
func Sum[T Numeric](s Seq[T]) T
```

- 但"按 key"的变体能回到方法，因为约束落在方法自带的参数 `K` 上，不碰接收者的 `T`：

```go
func (s Seq[T]) DistinctBy[K comparable](key func(T) K) Seq[T]
func (s Seq[T]) GroupBy[K comparable](key func(T) K) map[K][]T
```

- 涉及多个独立发生器类型、或接收者必须是特定实例化的操作，也只能是自由函数：

```go
func Zip[A, B any](a Seq[A], b Seq[B]) Seq2[A, B]   // 两个独立类型
func Flatten[T any](s Seq[Seq[T]]) Seq[T]            // 接收者须为 Seq[Seq[T]] 实例化
```

这条规则的好处是**可机械验证**：拿任何一个 API，问"它约束了 `T` 本身吗"——约束了就必须是函数，没约束就可以是方法。清单里不允许出现"约束了 `T` 却被列为方法"的条目。
### 边界：这个库不做什么形态的事

- **没有 for-comprehension 语法糖**，一切都是裸方法链。
- **没有 HKT / type class**，无法抽象 Functor/Monad，每个方法都是具体类型上的具体方法。
- **泛型方法不满足接口**（这是 1.27 提案的硬线），所以本库不提供集合的多态接口抽象。
- **没有普适相等**，`==` 是约束不是普适能力，这正是 `Distinct`/`Max` 必须是函数的根因。

完整 API 清单（构造入口、中间方法、终结方法、约束型函数、多序列函数、`Seq2` 一族）见 PRD `tasks/prd-seq-api-inventory.md`，本文档不重复抄录。

## Rationale / 理由与取舍

### 为什么把方法链押在最新的语言版本上

这是最该被质疑的决策，先正面回答。我们大可以做一个"兼容旧 Go"的版本——全自由函数，不用方法链。我们没选这条路，因为那样做出来的东西就是又一个 `lo`，而 `lo` 已经足够好了。本库唯一不可替代的价值就是链式方法；抽掉它，这个项目没有存在的理由。泛型方法提案（#77273）已被接受并在 Go 1.27 实现，本地装的 1.27 RC 已能编译全部 API，所以这不是赌特性会不会通过，而是接受一个明确的代价：把支持的最低 Go 版本钉在 1.27，放弃旧版本用户。我们认为这个取舍值得。

### 为什么是"定义类型"而不是 struct 包装

被放弃的方案：`type Seq[T any] struct { it iter.Seq[T] }`。struct 包装的诱惑是能挂任意方法、字段可扩展。我们没选它，因为它切断了与 `iter.Seq` 的零成本互转——用户每次想喂给 `slices.Collect` 都得先 `.it` 解包，标准库生态的无缝衔接全没了。定义类型让 `Seq[T]` 和 `iter.Seq[T]` 可以隐式当彼此用，这个收益压倒了 struct 的灵活性。

### 为什么 Distinct/Max/Sum 是函数而不是方法——这不是我们的选择

有人会问"为什么不把 `Distinct` 做成 `s.Distinct()`，多顺手"。答案是：做不到，不是不想。`Seq[T any]` 在类型层声明 `T` 为 `any`，方法无法给 `T` 追加 `comparable` 约束（方法的类型参数是独立的新参数，不能回头约束接收者）。这是语言规则，不是 API 品味。我们能做的只是把规则讲透，并提供"按 key"的方法变体（`DistinctBy`）作为逃生舱，让常见场景仍能链下去。

### 为什么元组封顶 Tuple4，不学 Scala 做到 Tuple22

被放弃的方案：Scala 式 `Tuple1`–`Tuple22`。我们没选，因为 Go 没有元组字面量和模式解构语法，高元数 `TupleN` 的字段只能叫 `Field1..FieldN` 这种无语义名，可读性差，且 `ZipN`/`Unzip`/测试要逐元维护。`Zip`/`Zip3`/`Zip4` 已覆盖几乎全部实际场景，更多字段的聚合应该用具名 struct（字段名表意）。封顶四元是收益和噪音的平衡点。

### 为什么不做错误处理 / 并行 / 原地可变

- **错误流转（`Seq2[T, error]` 短路链）**：Scala 的 `Try`/`Either`/for-comprehension 在 Go 惰性序列上没有优雅对应，强行做会污染所有签名。留作后续独立 PRD。
- **并行（对标 `lo/parallel`）**：本版只做顺序惰性管道。并行涉及完全不同的语义和正确性保证，不该塞进 v1。
- **原地可变（对标 `lo/mutable`）**：与"惰性 + 不可变管道"的定位直接冲突，不做。

## Compatibility / 兼容性

**这是纯增量的新库，对现有任何代码零破坏**——它没有"现有调用者"。真正的代价不在向后兼容，而在它把支持的最低 Go 版本钉在了 1.27。

代价如实列出：

- **最低 Go 版本是 1.27。** `go.mod` 声明 `go 1.27`，本地需安装 1.27 工具链（当前为 `go1.27rc1`）。任何停留在 1.26 及更早版本的项目都无法引入本库——这把潜在用户挡在门外，必须说在明处。
- **当前依赖的是 RC，尚非稳定版。** 1.27 正式版发布前，泛型方法的语法细节仍可能微调（比如方法类型参数的声明位置）；若有变化，相关方法签名需跟着对齐。我们对齐的是已接受的提案，不假设细节一字不改。

迁移路径：本库自身就是 opt-in 的——没人被迫依赖它。对使用方，建议在 1.27 正式版发布后再引入生产（当前 RC 适合尝鲜与开发），发布前在 1.27 工具链上跑通全部单测作为门禁。

## Implementation / 实现与过渡

落地分三步，每步都有可验证的产出（本地 1.27 RC 已就绪，三步均可立即编译测试）：

1. **类型与契约先行。** 先落地 `Seq`/`Seq2`/`Pair`/`Tuple3`/`Tuple4`/`Numeric` 类型定义和全部**自由函数**（`From`、`Distinct`、`Max`、`Zip`…），验证划分铁律站得住。这一批不依赖泛型方法，即便单独抽出也能在 Go 1.23+ 编译，可作为可独立发布的子集。

2. **方法链部分。** `Map`/`Filter`/`FlatMap` 等带方法级类型参数的方法依赖 #77273（Go 1.27），以"全部单测在 1.27 工具链通过"为完成门禁。

3. **生成权威 API.md。** 按"方法 / 自由函数"分区列出全部签名 + 一句语义，每个自由函数注明它无法成为方法的原因类别（约束 T / 多类型参数 / 嵌套实例化）。文档与代码签名人工核对一致。

第 1 步的自由函数集本身是一个可在 Go 1.23+ 独立编译的子集——相当于一个"带正确约束划分的 lo"，对暂时不能升级到 1.27 的用户仍有价值。

## Appendix / 附录

### FAQ

**Q：为什么不直接用 `samber/lo`？**
A：`lo` 解决的是"有没有这些操作"，本库解决的是"能不能从左到右链着写、且惰性"。两者定位不同；`lo` 的全部能力在 1.27 之前是链不起来的，这正是本库的切入点。

**Q：`SumBy`（方法）和 `Sum`（函数）为什么并存？**
A：`Sum[T Numeric]` 约束 `T` 本身，只能是函数；`SumBy[U Numeric](f func(T) U)` 的约束落在方法自带的 `U` 上，可以是方法。命名约定：`By` = 按投影值聚合（方法），`ByKey` = 按投影 key 取极值（方法），裸 `Max`/`Min`/`Sum` = 约束 T 的函数。

**Q：性能如何？**
A：深链 = 嵌套 `yield` 闭包调用，热点内循环可能不及手写 loop。本库的卖点是可读性和组合性，不是极致性能。API.md 会给出适用边界提示。

**Q：源序列能重复遍历吗？**
A：看来源。源自 slice 的 `Seq` 可重复遍历；源自 channel 的是一次性源。文档会逐个标注。
