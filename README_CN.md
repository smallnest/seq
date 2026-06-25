# seq

[![Go Reference](https://pkg.go.dev/badge/github.com/smallnest/seq.svg)](https://pkg.go.dev/github.com/smallnest/seq)
[![CI](https://github.com/smallnest/seq/actions/workflows/ci.yml/badge.svg)](https://github.com/smallnest/seq/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-%3E%3D1.27-blue)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

[English](README.md) | 简体中文

**基于 `iter.Seq` 构建的、可链式调用的惰性集合管道库。**

`seq` 是一个 Go 泛型库，封装标准库的惰性迭代器 `iter.Seq` / `iter.Seq2`（Go 1.23+），为它们提供 Scala 风格、从左到右、可链式调用的集合操作：

```go
sum := seq.From([]int{1, 2, 3, 4, 5, 6}).
    Filter(func(x int) bool { return x%2 == 0 }).
    SumBy(func(x int) int { return x * x })
```

> ⚠️ **本库需要 Go 1.27。** 链式方法依赖泛型方法提案（[golang/go#77273](https://github.com/golang/go/issues/77273)），该提案**已被接受**并在 Go 1.27 中实现。构建需安装 1.27 工具链（当前为 `go1.27rc1`）。详见下方 [状态](#状态)。

## 为什么需要它

今天，像 [`samber/lo`](https://github.com/samber/lo) 这样的库只能提供顶层函数，于是管道读起来是从内往外的：

```go
// 阅读顺序与数据流向相反：你得先看最内层的 Filter。
sum := lo.Sum(
    lo.Map(
        lo.Filter([]int{1, 2, 3, 4, 5, 6}, func(x int, _ int) bool { return x%2 == 0 }),
        func(x int, _ int) int { return x * x },
    ),
)
```

这不是风格偏好，而是语言限制。在 Go 1.27 之前，**方法不能声明自己的类型参数**，所以一个把 `Seq[int]` 变成 `Seq[string]` 的 `.Map()` 根本无法表达：

```go
// 1.27 之前：这行编译不过。方法不允许有自己的 [U any]。
func (s Seq[T]) Map[U any](f func(T) U) Seq[U]
```

[golang/go#77273](https://github.com/golang/go/issues/77273) 解除了这条限制。可链式、惰性、可发现的管道才成为可能——这是本库存在的全部理由。

## 设计

### 核心类型是「定义类型」，不是 struct 包装

```go
type Seq[T any]      iter.Seq[T]       // = func(yield func(T) bool)
type Seq2[K, V any]  iter.Seq2[K, V]
```

这带来零成本互转：任何 `iter.Seq[T]` 都能直接当作 `Seq[T]` 使用，反之亦然，结果也能直接喂给 `slices.Collect`、`maps.Keys` 等。代价是 `Seq[T]` 在类型层把 `T` 声明成了 `any`——而这一点决定了一半 API 的归属。

### 划分规则：方法 vs. 自由函数

因为 `Seq[T any]` 把 `T` 钉死成 `any`，方法的类型参数是全新、独立的——**它们无法反过来给接收者的 `T` 追加约束**。于是：

- 约束 `T` 本身的操作**必须是自由函数**：

  ```go
  func Distinct[T comparable](s Seq[T]) Seq[T]
  func Max[T cmp.Ordered](s Seq[T]) (T, bool)
  func Sum[T Numeric](s Seq[T]) T
  ```

- 只用方法自带约束参数的操作**可以是方法**（「逃生舱」）：

  ```go
  func (s Seq[T]) DistinctBy[K comparable](key func(T) K) Seq[T]
  func (s Seq[T]) GroupBy[K comparable](key func(T) K) map[K][]T
  ```

- 涉及多个发生器类型或特定实例化接收者的操作**必须是自由函数**：

  ```go
  func Zip[A, B any](a Seq[A], b Seq[B]) Seq2[A, B]   // 两个独立类型
  func Flatten[T any](s Seq[Seq[T]]) Seq[T]           // 接收者必须是 Seq[Seq[T]]
  ```

这条规则可机械验证：问一句「它约束了 `T` 本身吗？」——约束了就是函数，没约束就可以是方法。

## API 一览

| 分组 | 形态 | 示例 |
|------|------|------|
| 构造入口 | 自由函数 | `From`、`Of`、`Range`、`RangeStep`、`Repeat`、`Generate`、`Iterate`、`FromChannel`、`FromMap` |
| 中间转换（惰性） | 方法 | `Map`、`Filter`、`FlatMap`、`FilterMap`、`Reject`、`Take`/`Drop`、`TakeWhile`/`DropWhile`、`Scan`、`Chunk`、`Window`、`DistinctBy`、`Enumerate` |
| 终结 | 方法 | `Collect`、`Fold`、`Reduce`、`Find`、`Any`/`All`/`None`、`GroupBy`、`KeyBy`、`Partition`、`SumBy`、`MaxByKey`、`Join` |
| 约束型 | 自由函数 | `Distinct`、`Contains`、`Max`/`Min`、`Sum`/`Product`/`Mean`、`Sort`、`Union`/`Intersect`/`Difference`、`Compact`、`Without` |
| 约束型子类型 | 函数（入口）+ 方法 | `Comparable`/`Ordered`/`Numbers` 进入；之后链式 `.Distinct()`、`.Max()`、`.Sum()`；用 `.Ordered()`/`.Comparable()` 降级 |
| 多序列 | 自由函数 | `Zip`/`Zip3`/`Zip4`、`ZipWith`、`ZipMap`、`Unzip`、`Flatten`、`Concat`、`Interleave` |
| `Seq2[K,V]` | 方法 + 函数 | `MapValues`、`MapKeys`、`Keys`、`Values`；`ToMap`、`CollectPairs`、`Associate` |
| `Optional[T]` | 类型 + 方法 + 函数 | `Some`/`None`/`ToOptional`；`Get`、`Unwrap`/`UnwrapOr`/`OrElse`、`Map`、`Filter`、`FlatMap`；`MapOptional`（可改类型） |

完整、权威的清单（每条附一句语义）将放在 `API.md`。设计取舍见 [`tasks/design-seq.md`](tasks/design-seq.md)，完整 API 清单见 [`tasks/prd-seq-api-inventory.md`](tasks/prd-seq-api-inventory.md)。

### Optional[T]

`Optional[T]` 是对 `Find`、`First`、`Last`、`Nth`、`Reduce` 等方法所返回 `(T, bool)` 的零依赖、可选包装。它只是调用方侧的语法糖——**没有任何 `Seq`/`Seq2` 方法接受或返回它**，因此包仍保持与 `iter.Seq`、标准库的零成本互操作。用 `ToOptional` 从任意 `(T, bool)` 方法桥接：

```go
out := seq.ToOptional(s.Find(func(x int) bool { return x > 2 })).
    Map(func(x int) int { return x * 10 }).
    OrElse(-1)
```

由于 Go 方法不能引入新类型参数，`Optional.Map` 只能同类型转换；要改变元素类型（如 `Optional[int]` → `Optional[string]`）用包级 `MapOptional[T, U]`。

## 范围边界

**不包含**（均为刻意决策，理由见设计文档）：

- **错误处理链**（`Seq2[T, error]` 短路）——在 Go 惰性迭代器上没有优雅对应，留作独立提案。
- **并行执行**（`lo/parallel`）——本版只做顺序、惰性管道。
- **原地可变操作**（`lo/mutable`）——与惰性不可变模型冲突。
- **任意深度展平**（`flattenDeep`）——Go 类型系统无法表达，只提供一层 `Flatten`。
- **高元数元组**（`Tuple5`–`Tuple22`）——封顶 `Tuple4`，更多字段请用具名 struct。
- **HKT / type class** 与 **泛型方法满足接口**——语言不支持。

## 状态

**草案。** 本库需要 **Go 1.27**（`go.mod` 声明 `go 1.27`），构建需安装 1.27 工具链，当前为 `go1.27rc1`。工作分为两批：

- **第一批 —— 核心类型与自由函数**（`From`、`Distinct`、`Max`、`Zip`…）。本身独立可用——相当于「一个带正确约束划分的 `lo`」——若单独抽出可在 Go 1.23+ 编译。
- **第二批 —— 链式方法**（`Map`、`Filter`、`Fold`…），依赖 Go 1.27 的泛型方法。

泛型方法提案（[golang/go#77273](https://github.com/golang/go/issues/77273)）已被接受并在 Go 1.27 中实现。剩下的唯一差距是：1.27 目前是 RC、尚非稳定版。把最低版本钉在 1.27 是我们为方法链接受的代价。

## 许可证

见 [LICENSE](LICENSE)。
