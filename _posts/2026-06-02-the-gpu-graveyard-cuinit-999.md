---
layout: post
title: "The GPU Graveyard: cuInit 999, Broken Hosts, and How to Verify an Instance — Part 3 of 5"
date: 2026-06-02
categories: [llm, infrastructure, self-hosting]
tags: [vllm, gpuhub, cuda, self-hosted-llm, debugging, rtx-5090]
description: "Half the GPU instances I provisioned wouldn't start. cuInit 999, broken host drivers, an account lockout, and the ten-second check that saves hours. Part 3 of 5."
background: /assets/img/posts/gpu.jpg
image: /assets/img/posts/gpu.jpg
word_count: 3500
reading_time: 10
series: "Self-Hosting LLM Inference on Rented GPUs"
series_part: 3
---

Part 2 ended with a working model and a working framework. What it did not have was a working GPU instance.

That turned out to be its own problem entirely.

---

## The Error That Tells You Nothing

The first sign something was wrong was this:

```bash
python3 -c "
import ctypes
libcuda = ctypes.cdll.LoadLibrary('libcuda.so.1')
ret = libcuda.cuInit(0)
print('cuInit:', ret)
"
cuInit: 999
```

[`cuInit: 999` means CUDA unknown error](https://github.com/NVIDIA/open-gpu-kernel-modules/issues/689). That is the entire error message. No stack trace. No hint about what is wrong. Just 999.

The confusing part was that `nvidia-smi` worked perfectly:

```bash
nvidia-smi -L
GPU 0: NVIDIA GeForce RTX 4080 (UUID: GPU-68c44fa0...)
GPU 1: NVIDIA GeForce RTX 4080 SUPER (UUID: GPU-288b15b5...)
GPU 2: NVIDIA GeForce RTX 4080 (UUID: GPU-95ee52e5...)
```

The GPUs were there. VRAM looked correct. Names showed up. Everything appeared completely normal. The moment any Python code tried to actually use CUDA, it fell over.

PyTorch confirmed the same thing:

```
CUDA initialization: CUDA unknown error - this may be due to an
incorrectly set up environment, e.g. changing env variable
CUDA_VISIBLE_DEVICES after program start.
cuda: False
gpus: 3
RuntimeError: CUDA unknown error
```

Sees three GPUs. Cannot use any of them.

[This issue reproduces across multiple fresh pods and templates on GPU rental providers running driver 580.xx with CUDA 13.0 — nvidia-smi works, `/dev/nvidia-caps/*` nodes are missing, cuInit returns 999, PyTorch reports CUDA unavailable.](https://www.answeroverflow.com/m/1486185962403795006)

This had nothing to do with vLLM, the model, or the configuration. The instance was broken before vLLM started.

---

## Diagnosis

The commands that revealed the root cause:

```bash
nvidia-smi | head -3
cat /proc/driver/nvidia/version
```

Output:

```
| NVIDIA-SMI 580.105.08   Driver Version: 580.105.08   CUDA Version: 13.0 |

NVRM version: NVIDIA UNIX Open Kernel Module for x86_64
580.105.08 Release Build Wed Oct 29 22:29:53 UTC 2025
```

Host driver: `580.105.08`, reporting CUDA 13.0. Container CUDA runtime: 12.8.

That mismatch causes `cuInit: 999`. [After host upgrades to driver 580, CUDA fails to initialize in containers with a driver/library version mismatch — the fix has to come from the host side, not the container.](https://forums.linuxmint.com/viewtopic.php?t=455662)

The libcuda.so inside the container cannot communicate with the kernel module on the host because they are speaking different versions of the same protocol. This is not a configuration error you can fix from inside the container.

The workarounds that were tried and failed:

```bash
# Attempt 1 — LD_PRELOAD to use host libcuda
export LD_PRELOAD=/usr/lib/x86_64-linux-gnu/libcuda.so.1
python3 -c "import torch; print(torch.cuda.is_available())"
# Result: still False

# Attempt 2 — Install CUDA 13.0 runtime pip package
pip install nvidia-cuda-runtime-cu12==12.9.* --timeout 300
python3 -c "import torch; print(torch.cuda.is_available())"
# Result: still False

# Attempt 3 — Install cuda-python to match host
pip install cuda-python==13.0.0 --timeout 300
python3 -c "import torch; print(torch.cuda.is_available())"
# Result: still False
```

None of it worked. The mismatch happens at the kernel module level. No amount of Python package installation fixes that.

---

## The Host Problem

The instinct after hitting this was to spin up a fresh instance. Same error. Spin up another. Same error.

All broken instances shared the same physical host machine. Every new container provisioned from that host inherited the same broken driver environment. Spinning up fresh did not help because there was nothing wrong with the container — the problem was the physical machine underneath it.

Three separate containers. Same `cuInit: 999` on all three. Same host on all three.

This is the thing that costs the most time. The error looks like it could be anything — a bad container image, a misconfigured environment variable, a PyTorch version mismatch. You spend an hour trying fixes before realising the instance itself is unfixable. The host is the problem and you cannot fix the host.

The RTX 4080S instances on that broken host were abandoned entirely. Total time lost: several hours across multiple provisioning attempts.

---

## The Account Lockout

While debugging the broken instances, a different problem arrived.

The gpuhub account was locked. Access suspended. No warning.

The reason: automated systems had flagged the account for behaviour resembling cryptocurrency mining. High GPU utilisation, sustained compute load, Singapore region. The pattern matched.

Running LLM inference at full utilisation on a rented GPU looks identical to running a mining rig from the outside. Same sustained load, same power draw, same compute profile. The automated flag was wrong but not unreasonable.

Support was contacted with the exact use case — self-hosted LLM inference for Claude Code, vLLM serving Qwen3.6 models, personal development infrastructure. The response was fast. The account was reinstated. gpuhub confirmed that LLM inference is legitimate usage and that support would contact before taking action on automated flags in future rather than locking first.

Worth knowing if you run sustained inference workloads on rented GPU infrastructure. The utilisation profile will trigger mining detectors. Have a clear explanation ready.

---

## Which Instances Actually Worked

Not every instance had the broken driver. The pattern became clear after enough attempts.

RTX 5090 instances — different host hardware, different driver stack, cuInit returned 0 cleanly on the first try. This is where the model eventually settled.

RTX PRO 6000 96GB instances — also clean. cuInit 0, stable, no driver issues.

RTX 4090 48GB instances on different hosts — clean. cuInit 0.

The broken instances were all RTX 4080S nodes, all on the same physical host. Different GPU types on different hosts were unaffected. It was not a GPU family problem. It was a specific host that had received a driver upgrade without the container images being updated to match.

---

## The Mamba Assertion Error

Once a clean instance was found, a second error appeared on first serve attempt:

```
AssertionError: In Mamba cache align mode, block_size (2096) must
be <= max_num_batched_tokens (2048).
```

This one is unrelated to the host driver problem but appears on almost every first attempt to serve Qwen3.5 or Qwen3.6 hybrid models. The fix is straightforward: [`--max-num-batched-tokens` must be at least 2096 when serving hybrid Mamba-attention architectures with FP8 or TurboQuant KV cache enabled.](https://patrickgawron.com/articles/qwen36-35b-nvfp4-vllm-on-rtx5090/) vLLM sets the cache block size to 2096 for these models and the batched token limit must not be smaller than the block size.

```bash
# Wrong — causes assertion error
--max-num-batched-tokens 2048

# Correct
--max-num-batched-tokens 8192
```

This error is in the vLLM documentation but easy to miss on first setup.

---

## The Verification Checklist

Ten seconds of verification before installing anything saves hours of debugging. Run this on every new instance before doing anything else:

```bash
# Step 1 — Check CUDA init
python3 -c "
import ctypes
libcuda = ctypes.cdll.LoadLibrary('libcuda.so.1')
ret = libcuda.cuInit(0)
print('cuInit:', ret)  # Must be 0
"

# Step 2 — Check driver version
nvidia-smi | head -3
cat /proc/driver/nvidia/version

# Step 3 — Check PyTorch sees the GPU
python3 -c "
import torch
print('cuda available:', torch.cuda.is_available())
print('gpu count:', torch.cuda.device_count())
for i in range(torch.cuda.device_count()):
    print(f'  GPU {i}:', torch.cuda.get_device_name(i))
"
```

If cuInit returns anything other than 0 — terminate the instance and reprovision. Do not attempt to fix it from inside the container. The problem is the host. A new instance on a different host will work. The same host will not.

If the driver version in `nvidia-smi` shows CUDA 13.0 but your container image ships CUDA 12.8, that is the source of the mismatch. Terminate and reprovision.

If PyTorch reports `cuda available: False` after cuInit returns 0, then the problem is a container-level PyTorch/CUDA version mismatch and can be fixed by reinstalling PyTorch with the correct CUDA wheel. [That is a different problem with a different fix.](https://markaicode.com/errors/vllm-gpu-not-detected-fix/)

The ten-second check disambiguates all three cases before you spend hours on the wrong diagnosis.

---

## What is next

Instance verified. Working host confirmed. The model loads. CUDA initialises.

Now the actual configuration work begins — the full install script, the vLLM flags, the MoE backend flag that replaced a deprecated environment variable, the chat template patch that makes Claude Code tool calls work, and the Claude Code settings that prevent context wall crashes.

That is Part 4.

---

*Next: [The Config That Finally Worked — vLLM flags, install script, and the chat template patch →](#)*

*This is Part 3 of a 5-part series on self-hosting LLM inference on rented GPUs.*

*I write about Java, PKI, LLM inference infrastructure, and building things on the side. Find me on [GitHub](https://github.com/kayisrahman) and [LinkedIn](https://linkedin.com/in/kayisrahman).*
