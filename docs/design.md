# 项目设计文档

## 项目目标

`llm-eval` 是一个可扩展的大模型评估与基准测试框架。它负责加载测试套件、调用真实模型 Provider、将模型原始输出与期望答案进行比对、缓存评估结果，并生成结构化报告。

## 模块划分

- `cmd/llm-eval`：CLI 入口，负责解析命令行参数并启动评估任务。
- `internal/suite`：测试套件加载模块，支持 YAML 和 JSON。
- `internal/provider`：模型 Provider 抽象层，当前包含 `deepseek` 和 `qwen`。
- `internal/evaluator`：评分器模块，包含通用评分器和 benchmark 专用评分器。
- `internal/cache`：本地文件缓存，缓存键由 provider、prompt、case id 和 input 共同生成。
- `internal/runner`：评估流程编排模块，使用 worker pool 并发执行测试用例。
- `internal/report`：报告生成模块，支持 JSON、Markdown 和 HTML。
- `benchmarks`：内置 GSM8K、MATH、MMLU 三类 benchmark 风格测试套件。

## 评估流程

1. CLI 根据 `--suite` 参数加载测试套件。
2. Runner 根据 `--model` 参数选择 `deepseek` 或 `qwen` Provider。
3. Runner 根据 `--concurrency` 启动多个 goroutine 并发处理测试用例，并将并发数限制在 `1-5` 之间。
4. 对每个测试用例，Runner 先检查本地缓存。
5. 如果缓存未命中，则调用 Provider 获取模型输出。
6. Evaluator 将模型输出与期望答案进行比对并计算得分。
7. Runner 按原始 case 顺序收集结果，并聚合通过数量、失败数量和整体分数。
8. Report 模块将结果写入 `reports/latest.json`、`reports/latest.md` 和 `reports/latest.html`。

## 测试套件

当前项目提供三个示例测试套件：

- `benchmarks/gsm8k.yaml`：GSM8K 风格的小学数学文字题，使用 `number_match` 抽取最终数字。
- `benchmarks/math.yaml`：MATH 风格的数学推理题，使用 `math_match` 处理 `\boxed{}` 和数学答案归一化。
- `benchmarks/mmlu.yaml`：MMLU 风格的多选题，使用 `choice_match` 抽取最终 `A/B/C/D` 选项。

这些套件采用统一 YAML 格式，后续可以替换为完整公开数据集或接入数据集加载器。

## 扩展接口

### Provider

新增模型服务商时，只需要实现 `provider.Provider` 接口：

```go
type Provider interface {
    Name() string
    Generate(ctx context.Context, req Request) (Response, error)
}
```

当前已经实现 DeepSeek 和 Qwen。通过该接口可以继续扩展 Anthropic、本地 vLLM 或其他 OpenAI-compatible 接口。

### Evaluator

新增评分方式时，只需要实现 `evaluator.Evaluator` 接口：

```go
type Evaluator interface {
    Name() string
    Evaluate(output string, expected string) Result
}
```

后续可以继续扩展 BLEU、ROUGE、JSON Schema 校验、自定义规则评分等指标。

当前内置评分器：

- `exact_match`：归一化空白后进行精确匹配。
- `contains`：检查模型输出是否包含期望答案。
- `regex`：使用正则表达式验证模型输出。
- `choice_match`：从输出中抽取选择题选项，适用于 MMLU。
- `number_match`：抽取最终数字并进行数值等价比较，适用于 GSM8K。
- `math_match`：抽取 `\boxed{}` 或最终数值，并进行数学答案归一化，适用于 MATH。

## 可视化报告

HTML 报告包含四部分：

- 总览：总题数、通过数、失败数和准确率。
- 评分器统计：按 evaluator 聚合通过率。
- 混淆矩阵：当套件使用 `choice_match` 时，展示 MMLU 风格选择题的 expected/predicted 分布。
- 明细表：展示每道题的期望答案、抽取后的预测答案、原始模型输出和判分结果。

## 模型 Provider 配置

DeepSeek 和 Qwen Provider 均使用 OpenAI-compatible Chat Completions 接口。

DeepSeek 配置：

- `DEEPSEEK_API_KEY`
- `DEEPSEEK_MODEL`，默认值为 `deepseek-v4-flash`
- `DEEPSEEK_BASE_URL`，默认值为 `https://api.deepseek.com`

Qwen 配置：

- `QWEN_API_KEY`
- `QWEN_MODEL`，默认值为 `qwen-plus`
- `QWEN_BASE_URL`，默认值为 `https://dashscope.aliyuncs.com/compatible-mode/v1`

程序启动时会读取项目根目录下的 `.env` 文件；如果系统环境变量中已经存在同名配置，则优先使用系统环境变量。

## 缓存策略

缓存键使用 provider 名称、suite prompt、case id 和 input 共同计算 SHA-256 摘要。缓存内容保留模型原始输出，重复运行同一评估任务时可以避免重复请求模型 API，降低成本并提升复现性。


