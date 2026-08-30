---
layout: post
title: "The Config That Finally Worked — Part 4 of 5"
date: 2026-06-09
categories: [llm, infrastructure, self-hosting]
tags: [vllm, gpuhub, qwen3, self-hosted-llm, turboquant, rtx-5090, claude-code]
description: "The full install script, every vLLM flag explained, the chat template patch that makes Claude Code tool calls work, and the settings.json that ties it together. Part 4 of 5."
series: self-hosting-llm
series_part: 4
series_title: "Why I Stopped Paying Per Token"
---

Part 3 ended with a working GPU instance. cuInit returned 0. PyTorch saw the GPU. Nothing was broken.

Now the actual configuration work begins.

---

## What This Part Covers

Getting vLLM running cleanly on a fresh gpuhub container requires seven steps in the right order. Each one has a gotcha. Miss any of them and Claude Code either fails silently, crashes mid-session, or produces garbage tool calls.

This is the exact install sequence that works.

---

## The Install Script

Everything lives in `/root/autodl-tmp/install_vllm.sh`. Run it once on a fresh container and it handles the full setup — PyTorch, vLLM, patches, chat template fix, and the serve script.

{% raw %}
```bash
#!/bin/bash
# ============================================================
# vLLM install script for RTX 5090 (Blackwell / SM 12.x)
# Tested: CUDA 13.0 | PyTorch 2.11.0 | vLLM 0.22.1 | Python 3.12
# ============================================================
set -e

MODEL_PATH="${MODEL_PATH:-/root/autodl-tmp/models/Qwen3.6-35B-A3B-AWQ}"
INSTALL_DIR="${INSTALL_DIR:-/root/miniconda3}"
PYTHON="$INSTALL_DIR/bin/python3"
PIP="$INSTALL_DIR/bin/pip3"

echo "==> [1/7] Installing PyTorch 2.11.0 with CUDA 13 support"
$PIP install torch==2.11.0 torchvision torchaudio \
    --index-url https://download.pytorch.org/whl/cu130

echo "==> [2/7] Installing vLLM 0.22.1"
$PIP install vllm==0.22.1

echo "==> [2b] Pinning transformers 5.10.2"
$PIP install transformers==5.10.2 tokenizers==0.22.2 --force-reinstall

echo "==> [3/7] Applying flash_attn.ops stub (SM 12.x / Blackwell workaround)"
FLASH_ATTN_DIR=$($PYTHON -c "
import importlib.util, pathlib
spec = importlib.util.find_spec('flash_attn')
if spec and spec.submodule_search_locations:
    print(list(spec.submodule_search_locations)[0])
else:
    import site; print(site.getsitepackages()[0] + '/flash_attn')
")

mkdir -p "$FLASH_ATTN_DIR/ops/triton"
touch "$FLASH_ATTN_DIR/__init__.py"
touch "$FLASH_ATTN_DIR/ops/__init__.py"
touch "$FLASH_ATTN_DIR/ops/triton/__init__.py"

cat > "$FLASH_ATTN_DIR/ops/triton/rotary.py" << 'ROTARY_EOF'
import torch

def apply_rotary(x, cos, sin, seqlen_offsets=0, interleaved=False, inplace=False):
    batch, seqlen, nheads, headdim = x.shape
    rotary_dim = cos.shape[-1] * 2
    x_rot = x[..., :rotary_dim].float()
    x_pass = x[..., rotary_dim:]
    c = cos[:seqlen].unsqueeze(0).unsqueeze(2)
    s = sin[:seqlen].unsqueeze(0).unsqueeze(2)
    half = rotary_dim // 2
    x1, x2 = x_rot[..., :half], x_rot[..., half:]
    x_rot_out = torch.cat([x1 * c - x2 * s, x1 * s + x2 * c], dim=-1)
    out = torch.cat([x_rot_out.to(x.dtype), x_pass], dim=-1)
    if inplace:
        x.copy_(out); return x
    return out
ROTARY_EOF

echo "==> [4/7] Patching TurboQuant workspace lock bug"
TQATTN=$($PYTHON -c "import vllm, os; print(os.path.join(os.path.dirname(vllm.__file__), 'v1/attention/backends/turboquant_attn.py'))")

python3 - "$TQATTN" << 'PATCH_EOF'
import sys
path = sys.argv[1]
with open(path) as f:
    src = f.read()

old = '''        k_buf, v_buf = current_workspace_manager().get_simultaneous(
            (buf_shape, torch.float16),
            (buf_shape, torch.float16),
        )'''
new = '''        try:
            k_buf, v_buf = current_workspace_manager().get_simultaneous(
                (buf_shape, torch.float16),
                (buf_shape, torch.float16),
            )
        except AssertionError:
            k_buf = torch.empty(buf_shape, dtype=torch.float16, device=device)
            v_buf = torch.empty(buf_shape, dtype=torch.float16, device=device)'''

applied = 1 if old in src or new in src else 0
if old in src:
    src = src.replace(old, new, 1)
with open(path, 'w') as f:
    f.write(src)
print(f'  {applied}/1 patches applied: {path}')
PATCH_EOF

echo "==> [5/7] Patching Qwen3.6 chat template (tool_result content type)"
python3 << 'CHAT_PATCH_EOF'
from transformers import AutoTokenizer
import re, os

model_path = os.environ.get('MODEL_PATH', '/root/autodl-tmp/models/Qwen3.6-35B-A3B-AWQ')
tok = AutoTokenizer.from_pretrained(model_path, trust_remote_code=True)
template = tok.chat_template

patched = re.sub(
    r"\{%-?\s*else\s*%\}\s*\{\{-?\s*raise_exception\('Unexpected item type in content\.'\)\s*-?\}\}",
    "{%- else %}\n                {{- '' }}",
    template
)

if patched != template:
    tok.chat_template = patched
    tok.save_pretrained(model_path)
    remaining = re.findall(r"raise_exception\('Unexpected item type", patched)
    print(f"  Patched. Remaining raise_exception: {len(remaining)}")
else:
    print("  Already patched or pattern not found — skipping")
CHAT_PATCH_EOF

echo "==> [6/7] Verifying installation"
$PYTHON -c "
import torch, vllm
print(f'  torch    : {torch.__version__}')
print(f'  vllm     : {vllm.__version__}')
print(f'  CUDA     : {torch.version.cuda}')
print(f'  GPU      : {torch.cuda.get_device_name(0) if torch.cuda.is_available() else \"N/A\"}')
"

echo "==> [7/7] Writing start_vllm_35b.sh"
cat > /root/autodl-tmp/start_vllm_35b.sh << 'SCRIPT_EOF'
#!/bin/bash
export OMP_NUM_THREADS=1
export PYTORCH_CUDA_ALLOC_CONF=expandable_segments:True
export VLLM_ALLOW_LONG_MAX_MODEL_LEN=1
export VLLM_USE_DEEP_GEMM=0
export VLLM_USE_FLASHINFER_SAMPLER=0
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
    --served-model-name "claude-sonnet-4-6" "claude-sonnet-4-7" "claude-opus-4-6" \
        "claude-opus-4-7" "claude-haiku-4-5" "claude-haiku-4-5-20251001"
SCRIPT_EOF

chmod +x /root/autodl-tmp/start_vllm_35b.sh

echo "============================================================"
echo " Installation complete."
echo " Start: bash /root/autodl-tmp/start_vllm_35b.sh"
echo " Verify: curl http://localhost:6006/v1/models"
echo "============================================================"
EOF

chmod +x /root/autodl-tmp/install_vllm.sh
```
{% endraw %}

---

## What Each Step Does

**Step 1 — PyTorch 2.11.0 with CUDA 13**

gpuhub RTX 5090 containers ship CUDA 13.0. Standard PyTorch wheels target CUDA 12.x. Installing from `https://download.pytorch.org/whl/cu130` gets the right build. Wrong CUDA version here causes silent failures later.

**Step 2 — vLLM 0.22.1**

Not the latest. Pinned deliberately. vLLM moves fast and breaking changes between minor versions are common. 0.22.1 is confirmed stable with TurboQuant, Qwen3.6 hybrid architecture, and the Anthropic Messages API integration.

**Step 2b — transformers 5.10.2**

vLLM installs its own transformers version. Qwen3.6-35B-A3B-AWQ needs transformers 5.10.2 for correct tokenizer behaviour. Installing it with `--force-reinstall` after vLLM ensures it is not silently downgraded.

**Step 3 — flash_attn stub**

RTX 5090 is Blackwell architecture, SM 12.x. The flash_attn library is not yet compiled for SM 12.x in most containers. vLLM calls `flash_attn.ops.triton.rotary` at startup and crashes if it is not there. The stub is a pure PyTorch implementation that satisfies the import without requiring the compiled CUDA extension. Without this, vLLM fails before loading the model.

**Step 4 — TurboQuant workspace lock patch**

TurboQuant allocates large contiguous workspace buffers for KV cache computation. Under certain memory layouts it hits an `AssertionError` in `get_simultaneous` when the buffer cannot be allocated contiguously. The patch wraps the allocation in a try/except and falls back to direct `torch.empty` allocation. Without this, the server crashes on first inference at high memory utilization.

**Step 5 — Chat template patch**

This is the one that took the longest to find.

Claude Code sends `tool_result` content blocks back to the model after tool execution. Qwen3.6's Jinja chat template does not handle the `tool_result` content type. When it encounters one it raises:

```
raise_exception('Unexpected item type in content.')
```

The model receives a 500 error. Claude Code receives a failed tool call. The terminal shows nothing useful. Requests go through, responses come back, but tool actions do not execute. It looks like a network issue or a timeout. It is a template bug.

The regex patch:

{% raw %}
```python
import re
from transformers import AutoTokenizer

model_path = '/root/autodl-tmp/models/Qwen3.6-35B-A3B-AWQ'
tok = AutoTokenizer.from_pretrained(model_path, trust_remote_code=True)
template = tok.chat_template

patched = re.sub(
    r"\{%-?\s*else\s*%\}\s*\{\{-?\s*raise_exception\('Unexpected item type in content\.'\)\s*-?\}\}",
    "{%- else %}\n                {{- '' }}",
    template
)

if patched != template:
    tok.chat_template = patched
    tok.save_pretrained(model_path)
    remaining = re.findall(r"raise_exception\('Unexpected item type", patched)
    print(f"Patched. Remaining raise_exception: {len(remaining)}")
```
{% endraw %}

It replaces every instance of the `raise_exception` call for unexpected item types with an empty string output. The template silently ignores content types it does not recognise instead of crashing. After this patch `tool_result` blocks pass through cleanly and Claude Code tool calls work correctly.

The patch is applied directly to the saved model tokenizer. It persists across server restarts and does not need to be reapplied.

**Step 6 — Verification**

Confirms torch, vLLM, CUDA, and GPU visibility before starting the server.

**Step 7 — Write start_vllm_35b.sh**

Writes the serve script with all confirmed working flags.

---

## The Serve Command — Flag by Flag

```bash
export OMP_NUM_THREADS=1
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
    --served-model-name "claude-sonnet-4-6" "claude-sonnet-4-7" "claude-opus-4-6" \
        "claude-opus-4-7" "claude-haiku-4-5" "claude-haiku-4-5-20251001"
```

Without this, vLLM's OpenMP thread pool initialises across all available CPU threads, causing a startup crash on gpuhub containers with certain CPU topologies. This is the most unintuitive flag in the whole setup. The error without it looks completely unrelated to thread count.

**`--gpu-memory-utilization 0.96`**

Leaves 4% of VRAM as headroom for CUDA context, driver allocations, and TurboQuant workspace buffers. Higher values cause OOM crashes during the memory profiling phase before the model even starts serving.

**`--max-model-len 262144`**

262,144 tokens. Confirmed stable on RTX 5090 32GB with TurboQuant KV compression.

**`--max-num-seqs 6`**

Six parallel sequences. With TurboQuant compressing the KV cache to roughly 487K tokens of pool, six sequences is tight but stable. Eight caused intermittent OOM under sustained load.

**`--kv-cache-dtype turboquant_4bit_nc`**

TurboQuant 4-bit non-contiguous KV cache compression. This is the flag that makes six parallel sequences viable on 32GB. Without it, the KV cache alone would consume more VRAM than is available after weights are loaded.

**`--moe-backend flashinfer_trtllm`**

Replaces the deprecated environment variable `VLLM_USE_FLASHINFER_MOE_FP16=1`. Required for correct MoE routing on Qwen3.6-35B-A3B.

**`--language-model-only`**

Skips loading the vision encoder. Saves roughly 2GB of VRAM. Without this flag that VRAM is unavailable for KV cache.

**`--chat-template-content-format openai`**

Required for Claude Code compatibility. Without it, content block formatting mismatches what Claude Code expects and tool results come back malformed.

**`--tool-call-parser qwen3_coder`**

Uses the Qwen3 Coder tool call parser which handles the specific JSON format Qwen3.6 produces for tool calls.

**`--reasoning-parser qwen3`**

Strips `<think>` reasoning blocks from responses before they reach Claude Code.

**`--served-model-name "claude-sonnet-4-6" "claude-opus-4-6" ...`**

Claude Code sends requests with specific model name strings. vLLM must recognise those strings or it returns a 404. Listing all current Claude model name variants means the server responds correctly regardless of which model Claude Code requests.

---

## Claude Code settings.json

With vLLM running, Claude Code needs one config file to point at it. Location: `~/.claude/settings.json`

```json
{
  "env": {
    "ANTHROPIC_BASE_URL": "https://u425-883c-8cc6fbc3.singapore-a.gpuhub.com:8443/v1",
    "ANTHROPIC_API_KEY": "not-needed",
    "CLAUDE_CODE_ATTRIBUTION_HEADER": "0"
  },
  "contextWindowSize": 200000,
  "maxTokens": 32768,
  "autoCompact": true
}
```

**`ANTHROPIC_BASE_URL`** — the gpuhub endpoint. Port 8443 is the HTTPS forwarding port that maps to container port 6006. This changes every time a new container is provisioned. Update it each time.

**`ANTHROPIC_API_KEY`** — vLLM does not validate API keys. Any non-empty string works.

**`CLAUDE_CODE_ATTRIBUTION_HEADER: "0"`** — must be in `env` in `settings.json`, not exported in the terminal. Without it, Claude Code injects a per-request attribution hash into the system prompt on every call. That hash changes every request, invalidating the KV cache prefix on the vLLM side. Inference becomes roughly 90% slower. This one line is the difference between 103 tok/s and 10 tok/s.

**`contextWindowSize: 200000`** — set deliberately 62,000 tokens below the vLLM hard limit of 262,144. The buffer gives auto-compaction room to run cleanly without hitting the hard wall.

**`maxTokens: 32768`** — maximum output tokens per response. Enough for any single code generation or edit.

**`autoCompact: true`** — enables automatic context compaction when the window approaches contextWindowSize. Without this, long sessions hit the context wall hard and the session dies.

---

## First Run

After the install script completes:

```bash
bash /root/autodl-tmp/start_vllm_35b.sh
```

The first run takes 3–4 minutes before the first token is produced. `torch.compile` runs at startup and compiles the model's computational graph for the specific GPU. This is cached after the first run. Subsequent starts take under 30 seconds.

Verify the server is ready:

```bash
curl http://localhost:6006/v1/models
```

If the response lists `claude-sonnet-4-6`, `claude-opus-4-6`, and `claude-haiku-4-5`, the server is ready. Open Claude Code and start a session. The first successful tool call confirms the chat template patch worked.

---

## What is next

Seven steps. One working server. Claude Code pointing at a local model on rented hardware.

The last question is whether it is actually worth doing. The cost, the quality, the maintenance overhead, and an honest answer to whether self-hosting beats just paying for the API.

That is Part 5.

---

*Next: [Is It Actually Worth It? — cost, quality, and an honest verdict →](#)*

*This is Part 4 of a 5-part series on self-hosting LLM inference on rented GPUs.*

*I write about Java, PKI, LLM inference infrastructure, and building things on the side. Find me on [GitHub](https://github.com/kayis-rahman) and [LinkedIn](https://linkedin.com/in/kayisrahman).*
