---
layout: post
title: "The Model Hunt: Five Attempts, One Winner — Part 2 of 5"
date: 2026-05-26
categories: [llm, infrastructure, self-hosting]
tags: [vllm, gpuhub, qwen3, self-hosted-llm, turboquant, rtx-5090]
description: "Five model attempts, three frameworks, and one GPU graveyard later — the setup that finally worked. Qwen3.6-35B-A3B-AWQ on vLLM 0.22.1 with TurboQuant. Part 2 of 5."
background: /assets/img/posts/model-hunt.jpg
image: /assets/img/posts/model-hunt.jpg
word_count: 4500
reading_time: 12
series: "Self-Hosting LLM Inference on Rented GPUs"
series_part: 2
---

Part 1 ended with one change in `~/.claude/settings.json`:

```json
{
  "env": {
    "ANTHROPIC_BASE_URL": "http://localhost:6006"
  }
}
```

That is the entire switch. Claude Code does not negotiate, does not check certificates, does not care what is on the other end. It formats requests in Anthropic's Messages API format and sends them. Something needs to answer in the same format.

That something is a vLLM server. vLLM implements the Anthropic Messages API natively — the same format Claude Code uses to communicate with Anthropic's servers. Any model served by vLLM with proper tool calling support acts as a drop-in replacement.

Simple in theory. The question is which model, on which hardware, with which framework. Get any of those wrong and Claude Code either fails silently, produces garbage tool calls, or hits a context wall before the session is useful.

Three things decide whether a local Claude Code session actually works: model quality decides whether the answer is smart, tool-call formatting decides whether Claude Code can act on the answer, and context length decides whether the session survives past the first few edits. For local coding agents 32K is the floor. 64K is the sweet spot. Anything below 32K is a chat demo, not Claude Code.

Getting all three right took five attempts, several weeks, and more broken GPU instances than expected.

Here is exactly what happened.

-----

## Attempt 1: Qwen3.5-35B-A3B-AWQ on single RTX 4080S 32GB

The first question was which model family to use. Qwen3 was the obvious starting point. Open weights, strong coding benchmarks, long context window, and an active community already running it on consumer hardware. The AWQ 4-bit quantized variant specifically because vLLM has mature AWQ support and the compression is well understood.

The first mistake was underestimating what "fits in VRAM" actually means.

Qwen3.5-35B-A3B is a Mixture of Experts model. The "3B active" in the name refers to how many parameters are activated per forward pass during inference. It sounds like you only need 3B worth of VRAM. You do not. The full 35B parameter weights still need to reside in memory. AWQ 4-bit brings the footprint down to roughly 21.4GB. On a 32GB card that sounds comfortable. It is not.

Once the weights are loaded, whatever VRAM remains is split between the KV cache and compute buffers. With 21.4GB of weights on a 32GB card, you have roughly 10GB left. At 32K context that is adequate. At 86K it runs out. The vLLM profiler calculated exactly --max-model-len 86592 as the maximum that fit. Not 128K. Not 262K. Eighty-six thousand tokens.

For Claude Code that is borderline usable. A medium complexity session hits 53K tokens mid-way through. A longer refactor hits 131K. The context ceiling arrived before the session was finished more often than not.

Tensor parallelism across two GPUs would have solved both problems. But there was only one GPU. Dead end.

Quality was not the problem. Qwen3.5-35B-A3B scores 73.4% on SWE-bench Verified compared to Claude Sonnet 4.6 at 79.6%. Close enough for most coding tasks. The hardware was the bottleneck, not the model.

Lesson: VRAM headroom matters as much as VRAM capacity. A model that fits is not the same as a model that runs well.

-----

## Attempt 2: Split setup across two providers

The obvious response to a single GPU bottleneck was to use two providers simultaneously. Run the large reasoning model on gpuhub and a smaller fast model on Vast.ai, then route between them with LiteLLM on the Mac. Heavy tasks go to the 35B. Quick tasks go to the 9B. Best of both worlds.

It sounded sensible. It was not.

The setup was: Qwen3.5-35B-A3B-AWQ on gpuhub handling main reasoning, Qwen3.5-9B-AWQ on a Vast.ai RTX 3090 acting as a fast subagent for lighter tasks, and a LiteLLM proxy running locally on the Mac routing between both. Claude Code pointed at LiteLLM. LiteLLM decided which provider answered.

The first problem appeared almost immediately. Requests were returning 404 errors. The path in the error was revealing:

```
POST /v1/v1/messages 404
```

The double /v1 prefix. LiteLLM was appending its own /v1 to the api_base that already had a trailing /v1 in the config. The fix was removing the trailing /v1 from the api_base value in the LiteLLM configuration. Simple once you see it. Not obvious when two providers are running simultaneously and the error could have come from either one.

That is the core problem with multi-provider setups. When something breaks you have two SSH sessions open, two sets of logs, two billing accounts running, and no clean way to isolate which side the problem is on. The LiteLLM bug looked like a provider issue. It was a config issue. But diagnosing that required eliminating both providers as the cause first.

One provider, one instance, one clear failure point.

-----

## Attempt 3: GGUF with llama.cpp and SGLang

GGUF seemed attractive for a specific reason. The files are smaller, the format is widely supported, and llama.cpp runs on virtually anything including CPU fallback.

It was not.

The first path was vLLM with a GGUF file. vLLM does support GGUF in recent versions but the support is architecture-specific. The error was immediate:

```
GGUF model with architecture qwen3next is not supported yet.
```

SGLang was tried next. SGLang has some GGUF support but it is inconsistent across architectures. Llama 3.1 8B loaded cleanly. Qwen3 failed. For a production agentic setup you need certainty that the server starts reliably every time. Inconsistent GGUF support is not that.

llama.cpp built from source with CUDA support actually worked. The model loaded, the server started, requests came back correctly formatted. It seemed like a solution until parallel sessions revealed the fundamental architecture problem.

llama.cpp uses fixed pre-allocated KV cache slots rather than continuous batching. If you allocate 524K total context and run 5 parallel sessions, each session gets roughly 104K tokens regardless of whether it needs that much. Sessions cannot borrow unused context from each other. The total context pool is divided statically at startup.

For 3 to 5 simultaneous Claude Code sessions that architecture creates constant OOM risk. One long session starves the others. llama.cpp is not architected for high-concurrency parallel inference. Continuous batching in vLLM solves this by allocating KV cache dynamically. llama.cpp does not have that.

Thirty gigabytes downloaded before the framework incompatibility was discovered. Check framework compatibility before downloading.

-----

## Attempt 4: Qwen3-Coder-Next-AWQ on 2x RTX 4090 96GB combined

The previous three attempts had one thing in common. Not enough VRAM headroom. The fix was straightforward: get enough memory that the model breathes properly.

gpuhub offers RTX 4090 instances at 48GB each. Two instances combined via tensor parallelism gives 96GB total. That changed everything.

Qwen3-Coder-Next is an 80B MoE model with 3B active parameters per forward pass. The AWQ 4-bit quantized variant sits at approximately 45GB. Across 96GB combined VRAM that leaves roughly 51GB for KV cache after weights are loaded. That headroom unlocked 524,288 tokens of context with 8 parallel sequences running simultaneously.

The first container on a clean host, cuInit returned 0 on both GPUs. After three failed attempts that felt significant.

The serve command that worked:

```bash
export CUDA_VISIBLE_DEVICES=0,1

vllm serve /root/autodl-tmp/models/Qwen3-Coder-Next-AWQ-4bit \
    --host 0.0.0.0 --port 6006 \
    --tensor-parallel-size 2 \
    --gpu-memory-utilization 0.95 \
    --max-model-len 524288 \
    --max-num-seqs 8 \
    --kv-cache-dtype fp8 \
    --enable-prefix-caching \
    --enable-chunked-prefill \
    --tool-call-parser qwen3_coder \
    --reasoning-parser qwen3 \
    --enable-auto-tool-choice \
    --served-model-name "claude-sonnet-4-6" "claude-opus-4-6" "claude-haiku-4-5"
```

Saved immediately to /root/start_vllm.sh. Every fresh instance starts with one command.

The LiteLLM double /v1 bug resurfaced here too. Diagnosed cleanly this time with a single provider and single log stream. One config line fixed it.

Both GPUs were fully committed to TP=2. No room for a second model on the same instance. MLX on the Mac handled PDF extraction and lighter vision tasks as a workaround.

TurboQuant was attempted here to extend KV cache headroom further. It was blocked immediately:

```
NotImplementedError: The page size of the layer is not divisible
by the maximum page size.
```

GatedDeltaNet hybrid architecture layers have a page size incompatibility with vLLM's TurboQuant integration in versions prior to vLLM 0.22.1. FP8 KV cache dtype was used instead, which halves KV cache memory usage and was sufficient.

Quality on Qwen3-Coder-Next is strong. 44.3% on SWE-bench Pro, comparable to Claude Sonnet 4.5 on coding benchmarks. For most Claude Code sessions the output quality was indistinguishable from the paid API in practice.

This was the first setup that actually worked for daily use. But Qwen3.6-27B had just been released. And it looked better on paper.

-----

## Attempt 5: Qwen3.6-27B-AWQ — OOM in two configurations

Qwen3.6-27B was released by Alibaba Cloud on April 22, 2026. Dense architecture, 27B parameters, AWQ 4-bit at roughly 21.9GB. On 96GB combined VRAM that should have been straightforward. It was not.

The single instance attempt came first.

One RTX 4090 48GB, model loaded at 21.9GB, leaving 26GB for KV cache. That sounds workable. The vLLM profiler disagreed:

```
Available KV cache memory: 0.87 GiB
Estimated maximum model length is 72,896
```

The --language-model-only flag was tried next. Qwen3.6-27B includes a vision encoder that loads by default. Skipping it saves roughly 2GB. The available KV cache improved slightly but not enough.

Moved to 2x RTX 4090 96GB combined with TP=2. OOM again, this time at gpu-memory-utilization 0.96 during the profiling phase. The model loaded cleanly. The crash came during KV cache allocation.

The root cause took time to find. Qwen3.6-27B uses a hybrid GatedDeltaNet architecture. 75% of its layers use GDN as the primary attention mechanism. Those layers require approximately 2GB of extra activation memory per forward pass that vLLM's memory profiler does not pre-reserve. At 0.96 utilisation there was no room left when inference actually started.

SGLang was tried as an alternative. It failed with a different error entirely:

```
KeyError: qkqkv_proj.weight
```

AWQ quantization renames certain weight tensors. SGLang's loader expected the original naming convention. The mismatch blocked loading completely.

Back to vLLM. Two fixes applied together: drop gpu-memory-utilization from 0.96 to 0.88, reduce max-num-seqs from 8 to 2. Both changes together left enough headroom for the GDN activation memory spike. The model ran. But two parallel sequences instead of eight was a significant regression from the previous setup.

TurboQuant would have solved the activation memory problem. The same page size incompatibility that blocked it in Attempt 4 blocked it here too. The fix was not in vLLM yet.

-----

## Current setup: Qwen3.6-35B-A3B-AWQ on 1x RTX 5090

The 27B ran well. It was not the destination.

Qwen3.6-35B-A3B-AWQ landed cleanly on the same RTX 5090 32GB. 35B total parameters, 3B active per token via MoE routing. At AWQ 4-bit the weights sit at roughly 21GB — a comfortable fit with meaningful headroom left for KV cache. And unlike earlier attempts that needed patches to get TurboQuant working, the 35B on vLLM 0.22.1 just worked. `turboquant_4bit_nc` loaded without patches, without workarounds, without the page size errors that blocked earlier attempts.

The NVFP4 variant was tried first. It failed immediately with `No supported CUDA architectures found for major versions [12]`. NVFP4 requires SM 12.x CUTLASS support not compiled in the container's flashinfer build. The AWQ variant ran cleanly without that constraint.

The final serve command:

```bash
export OMP_NUM_THREADS=1
export PYTORCH_ALLOC_CONF=expandable_segments:True
export VLLM_ALLOW_LONG_MAX_MODEL_LEN=1
export CUDA_VISIBLE_DEVICES=0

vllm serve /root/autodl-tmp/models/Qwen3.6-35B-A3B-AWQ \
    --host 0.0.0.0 --port 6006 \
    --gpu-memory-utilization 0.96 \
    --max-model-len 262144 \
    --max-num-seqs 6 \
    --kv-cache-dtype turboquant_4bit_nc \
    --max-num-batched-tokens 8192 \
    --enable-prefix-caching \
    --language-model-only \
    --moe-backend flashinfer_trtllm \
    --tool-call-parser qwen3_coder \
    --reasoning-parser qwen3 \
    --trust-remote-code \
    --enable-auto-tool-choice \
    --enable-chunked-prefill \
    --chat-template-content-format openai \
    --served-model-name "claude-sonnet-4-6" "claude-opus-4-6" "claude-haiku-4-5"
```

One non-obvious fix that will bite anyone following this path: a regex patch applied to Qwen3.6's Jinja chat template is required to silence `raise_exception('Unexpected item type in content.')` errors triggered by Claude Code's `tool_result` content blocks. Without it, Claude Code tool calls fail silently. Requests go through, responses come back, but tool actions do not execute. The patch and the full install script are covered in Part 4.

| Setup | VRAM | Sequences | Context | KV cache | Cost/hr |
|---|---|---|---|---|---|
| 2x RTX 4090 Attempt 5 | 96GB | 2 | 262K | fp8 | $0.92 |
| 1x RTX 5090 vLLM 0.22.1 | 32GB | 6 | 262K | turboquant_4bit_nc | $0.36 |

This setup runs Claude Code across multiple concurrent projects on a single GPU. One model. One instance. 262K context. Six parallel sessions. $0.36 an hour.

-----

## Framework verdict: vLLM won

Three frameworks were tried across five attempts. The verdict is clear but the reasons are worth understanding because the wrong choice costs days not hours.

llama.cpp is the most accessible option. Runs on anything, wide model support, active community. For a single user chatting with a model it is perfectly fine. For agentic workloads with multiple simultaneous Claude Code sessions it is the wrong architecture. Fixed KV cache slot allocation means parallel sessions divide the context pool statically at startup rather than sharing it dynamically.

SGLang is legitimately fast. For hybrid architectures like Qwen3.6-27B, the official Qwen model card recommends SGLang first before vLLM. RadixAttention prefix caching is more efficient than vLLM's implementation for multi-turn agentic workloads. In practice two bugs blocked it completely. The qkqkv_proj.weight naming bug in AWQ quantized Qwen3 models caused a KeyError on load. And SGLang has no TurboQuant integration as of vLLM 0.22.1. Without TurboQuant the 8-sequence target on a single RTX 5090 was not achievable. SGLang is worth watching. It is not there yet.

vLLM won by being the most complete production-ready option for this specific combination of requirements. vLLM implements the Anthropic Messages API natively. vLLM 0.22.1 added TurboQuant 2-bit KV cache compression with 4x capacity, multiple tool parsers including dedicated Qwen3 Coder support, and Blackwell SM120 optimizations. The --language-model-only flag saves 2GB by skipping the vision encoder. Multiple --served-model-name aliases mean one server responds to claude-sonnet-4-6, claude-opus-4-6, and claude-haiku-4-5 simultaneously.

| Framework | Tool calling | Continuous batching | TurboQuant | Anthropic API native | Verdict |
|---|---|---|---|---|---|
| vLLM 0.22.1 | Reliable | Yes | Yes | Yes | Winner |
| SGLang | Unreliable on AWQ | Yes | No | Via proxy | Not yet |
| llama.cpp | Intermittent | No | Via fork | No | Wrong architecture |

-----

## What is next

Five attempts. One winner.

Qwen3.6-35B-A3B-AWQ on vLLM 0.22.1 with `turboquant_4bit_nc` on a single RTX 5090 — 262K context, 6 parallel Claude Code sessions, $0.36 an hour. That is the setup that runs today.

But getting there required provisioning a lot of GPU instances first. And roughly half of them did not work.

Not did not work well. Did not work at all.

nvidia-smi showed the GPUs. VRAM looked correct. Everything appeared normal. The moment any Python code tried to actually use CUDA it fell over with a single cryptic error:

```
cuInit: 999
```

Unknown error. No stack trace. No helpful message. Just 999.

This had nothing to do with vLLM, the model, or the configuration. The instances themselves were broken at the host level — a CUDA driver mismatch between the container runtime and the host kernel. The model never even loaded.

Diagnosing that, understanding why entire gpuhub host nodes were affected, and learning how to verify an instance before trusting it with a three-hour Claude Code session — that is Part 3.

---

*Next: [The GPU Graveyard — cuInit 999, broken hosts, and how to verify an instance before you trust it →](/llm/infrastructure/self-hosting/2026/06/02/the-gpu-graveyard-cuinit-999.html)*

*This is Part 2 of a 5-part series on self-hosting LLM inference on rented GPUs.*

*I write about Java, PKI, LLM inference infrastructure, and building things on the side. Find me on [GitHub](https://github.com/kayisrahman) and [LinkedIn](https://linkedin.com/in/kayisrahman).*
