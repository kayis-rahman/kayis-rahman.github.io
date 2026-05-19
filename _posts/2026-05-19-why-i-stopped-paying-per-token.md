---
layout: post
title: "Why I Stopped Paying Per Token — Part 1 of 5"
date: 2026-05-19
categories: [llm, infrastructure, self-hosting]
tags: [vllm, gpuhub, claude-code, self-hosted-llm, qwen3]
description: "Claude Code mid-session, 500 errors, rate limits, and the decision to self-host. Part 1 of a 5-part series on running your own LLM on rented GPUs."
---

It happened mid-session.

Claude Code was deep inside `TechnicalIndicators.java` — editing a smoothed value calculation, six files open, tool calls chained across the last twenty minutes of work. Then this:

```
500 {"type":"error","error":{"type":"internal_error","message":"This model's
maximum context length is 196176 tokens. However, you requested 32000 output
tokens and your prompt contains at least 164177 input tokens, for a total of
at least 196177 tokens."}}

Retrying in 10 seconds… (attempt 6/10)
```

Six retries. Ten seconds each. A minute of watching it fail repeatedly at the worst possible moment.

This wasn't a free tier problem. This was the Max plan. The plan that costs $200 a month.

---

## What Claude Code Actually Eats

Most people think of token usage in terms of messages — question in, answer out. Claude Code doesn't work like that.

Every tool call sends context. Every file read adds to the prompt. Every edit response comes back at up to 32,000 output tokens by default. And if you're running subagents — which Claude Code spins up automatically for parallel tasks — they hit the same instance, the same context window, at the same time.

Before you even type your first message, context is already pre-consumed:

| Category | Tokens used |
|---|---|
| System prompt | ~2,600 |
| System tools | ~17,600 |
| MCP tools | ~900 |
| Custom agents | ~935 |
| Memory files | ~300 |
| **Pre-consumed at session start** | **~22,000** |

Real numbers from actual sessions:

- Mid-session, medium complexity: **53,152 tokens**
- Same session before hitting the wall: **131,841 tokens**
- Default output request per call: **32,000 tokens**
- Combined: exactly one token over the limit

The error wasn't lying. Claude Code had grown the context so naturally, across so many incremental steps, that nobody noticed it approaching the wall until it hit it at full speed.

> **As of May 2026:** Claude Code operates on a 5-hour rolling window. Pro users get approximately 44,000 tokens per window, Max 5x around 88,000, and Max 20x roughly 220,000 tokens. Developers have reported burning through 4 hours of usage in 3 prompts when using plan mode to refactor a frontend architecture.

The official fix is `autoCompact` — a setting that compresses conversation history when context gets long. In theory it fires automatically. In practice, by the time it triggers, queued requests have already exceeded the model's limit. The workaround is to lower `contextWindowSize` so compaction fires earlier:

```json
{
  "autoCompact": true,
  "maxTokens": 16000,
  "contextWindowSize": 98304
}
```

Which means: deliberately cripple the context window to stop the tool from breaking itself. That's not a solution. That's a band-aid on a structural problem.

Run three or four parallel Claude Code sessions — which is exactly what building multiple apps simultaneously looks like — and a new problem appears. LiteLLM 429 errors. Internal rate limits throttling requests before they even reach the model. You're paying $200 a month and still getting throttled.

Also worth knowing: **agent teams use approximately 7x more tokens than standard sessions** when teammates run in plan mode, because each teammate maintains its own context window and runs as a separate Claude instance.

-----

## The Cost Maths

Here's what the options actually look like in 2026:

|Plan                   |Monthly Cost |Token window      |Reality                            |
|-----------------------|-------------|------------------|-----------------------------------|
|Claude Pro             |$20          |~44K tokens / 5hr |Hits limit in 3 Claude Code prompts|
|Claude Max 5x          |$100         |~88K tokens / 5hr |Ok for light daily use             |
|Claude Max 20x         |$200         |~220K tokens / 5hr|One heavy agentic week burns it    |
|Sonnet 4.6 API         |Pay-per-token|Unlimited         |$3/M input, $15/M output           |
|**2× RTX 4080S gpuhub**|**~$84/mo**  |**Unlimited**     |**Private, no throttle**           |
|**2× RTX 5090 gpuhub** |**~$130/mo** |**Unlimited**     |**Private, no throttle**           |

Real gpuhub instance prices from actual sessions:

|GPU Instance   |VRAM         |Cost/hr   |Est. monthly (6hr/day)|
|---------------|-------------|----------|----------------------|
|2× RTX 4080S   |64GB combined|$0.46–0.48|~$84                  |
|RTX 5090 single|32GB         |$0.36     |~$65                  |
|2× RTX 5090    |64GB combined|$0.72     |~$130                 |
|RTX PRO 6000   |96GB         |$0.91     |~$164                 |

*Prices from actual sessions — verify current rates at [gpuhub.com](https://gpuhub.com) before provisioning. gpuhub bills per second.*

Token maths at scale:

|Monthly volume   |Sonnet 4.6 API cost|2× RTX 5090 gpuhub|
|-----------------|-------------------|------------------|
|10M in / 2M out  |~$60               |~$50–80           |
|50M in / 10M out |~$300              |~$130             |
|100M in / 20M out|~$600              |~$130             |
|200M in / 40M out|~$1,200            |~$130             |

Break-even hits at medium usage. Above that, the gap widens every month. And this is worth knowing: **your usage limit is shared across Claude.ai chat and Claude Code from the same bucket.** Heavy chat usage eats into your Claude Code allocation.

One developer tracked 8 months of intensive Claude Code usage. API equivalent cost exceeded $15,000. Max plan for the same period: ~$800. That's a 93% reduction — but it still doesn't solve the privacy problem.

-----

## The Part Nobody Talks About — Privacy

Cost alone would have been enough. But there was a second constraint that made the API a non-starter regardless of price.

The apps being built process data that cannot leave the local network. Sending any of that to a third-party API isn't a preference — it's an architectural hard stop.

Here's what the actual data policy looks like in 2026, because most people haven't read it:

> **Free, Pro, and Max accounts are consumer accounts. By default, these train on your data unless you manually opt out. Only Commercial or Enterprise tiers prohibit data training by default.**

In August 2025, Anthropic introduced an opt-in toggle with a deadline of September 28. If you opted in, Anthropic could retain conversations in de-identified form for up to **5 years** — a 60x increase from the previous 30-day retention period.

Even with training opted out:

- Consumer account data: retained 30 days
- Zero Data Retention (ZDR): **only available for commercial API keys or Claude Enterprise — not Pro or Max**
- Claude Code still stores every read, bash command, search result, and file edit locally as a plaintext JSONL file — and the prompts containing all of it still transit Anthropic's servers

For financial data, proprietary logic, anything sensitive — self-hosting isn't paranoia. It's the only architecture where the data genuinely doesn't leave.

-----

## One Config Change

The actual switch to self-hosting is a few lines in `~/.claude/settings.json`:

```json
{
  "env": {
    "ANTHROPIC_MODEL": "claude-sonnet-4-6",
    "ANTHROPIC_BASE_URL": "http://localhost:6006",
    "ANTHROPIC_AUTH_TOKEN": "local-key",
    "CLAUDE_CODE_USE_NATIVE_TOOLS": "1",
    "CLAUDE_CODE_ATTRIBUTION_HEADER": "0"
  },
  "autoCompact": true,
  "maxTokens": 16000,
  "contextWindowSize": 180000
}
```

`ANTHROPIC_BASE_URL` pointing at localhost. Claude Code doesn't know or care what's running on the other end. It sends requests. Something answers. As far as the tool is concerned, nothing changed.

> **Critical gotcha:** `CLAUDE_CODE_ATTRIBUTION_HEADER: "0"` must be in `settings.json` — not exported in the terminal. Without it, Claude Code injects a per-request hash into the system prompt on every call, which invalidates the KV cache on local models and makes inference roughly **90% slower**. This one line matters.

For remote instances on gpuhub, one more step — `claude setup-token` generates a one-year OAuth token from your subscription that authenticates against the remote server without sending credentials over the wire:

```bash
# On your Mac
claude setup-token

# On gpuhub server
export CLAUDE_CODE_OAUTH_TOKEN=your-token-here
export ANTHROPIC_BASE_URL=http://localhost:6006/v1
claude
```

What followed was several weeks of figuring out what actually works on the other end of that URL. That's Part 2.

-----

*Next: [The model hunt — five attempts, one winner →](#)*

*This is Part 1 of a 5-part series on self-hosting LLM inference on rented GPUs.*

*I write about Java, PKI, LLM inference infrastructure, and building things on the side. Find me on [GitHub](https://github.com/kayisrahman) and [LinkedIn](https://linkedin.com/in/kayisrahman).*
