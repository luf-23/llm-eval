# llm-eval

`llm-eval` 是一个基于 Go 实现的可扩展 LLM 评估与基准测试框架，支持命令行运行、自定义评分器、多模型 Provider、结果缓存和报告生成。

## 功能特性

- 提供 CLI 评估入口。
- 支持加载 YAML 和 JSON 格式的测试套件。
- 内置 `deepseek` 和 `qwen` 两个真实模型 Provider。
- Provider 使用 OpenAI-compatible Chat Completions 接口，便于扩展其他模型服务商。
- 提供可扩展的评分器接口，当前支持 `exact_match`、`contains` 和 `regex`。
- 支持本地结果缓存，避免重复调用模型。
- 支持生成 JSON 和 Markdown 格式的评估报告。

## 快速开始

```bash
go mod tidy
go test ./...
go run ./cmd/evaluate --suite examples/gsm8k.yaml --model deepseek
```

运行完成后会生成报告：

```text
reports/latest.json
reports/latest.md
```

## 配置模型 Provider

复制配置模板并填写自己的 Key：

```powershell
Copy-Item .env.example .env
```

`.env` 示例：

```env
DEEPSEEK_API_KEY=your_deepseek_key
DEEPSEEK_MODEL=deepseek-v4-flash
DEEPSEEK_BASE_URL=https://api.deepseek.com

QWEN_API_KEY=your_qwen_key
QWEN_MODEL=qwen-plus
QWEN_BASE_URL=https://dashscope.aliyuncs.com/compatible-mode/v1
```

运行 DeepSeek：

```bash
go run ./cmd/evaluate --suite examples/gsm8k.yaml --model deepseek
```

运行 Qwen：

```bash
go run ./cmd/evaluate --suite examples/mmlu.yaml --model qwen
```

## 内置示例套件

项目提供三个 benchmark 风格的示例套件：

```text
examples/gsm8k.yaml  GSM8K 风格小学数学文字题
examples/math.yaml   MATH 风格数学推理题
examples/mmlu.yaml   MMLU 风格多选题
```

可以分别运行：

```bash
go run ./cmd/evaluate --suite examples/gsm8k.yaml --model deepseek
go run ./cmd/evaluate --suite examples/math.yaml --model deepseek
go run ./cmd/evaluate --suite examples/mmlu.yaml --model qwen
```

## 测试套件格式

```yaml
name: math
prompt: |
  Answer with only the final answer.
cases:
  - id: add
    input: "What is 2 + 3?"
    expected: "5"
    evaluator: exact_match
```

当前支持的评分器：

- `exact_match`
- `contains`
- `regex`

## 项目结构

```text
cmd/evaluate        CLI 入口
internal/suite      测试套件加载
internal/provider   模型 Provider
internal/evaluator  评分器
internal/cache      本地结果缓存
internal/runner     评估流程编排
internal/report     报告生成
docs/design.md      项目设计文档
examples/           GSM8K、MATH、MMLU 示例测试套件
```
