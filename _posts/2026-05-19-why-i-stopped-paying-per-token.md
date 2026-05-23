---
layout: post
title: "Why I Stopped Paying Per Token — Part 1 of 5"
date: 2026-05-19
categories: [llm, infrastructure, self-hosting]
tags: [vllm, gpuhub, claude-code, self-hosted-llm, qwen3]
description: "Claude Code mid-session, 500 errors, rate limits, and the decision to self-host. Part 1 of a 5-part series on running your own LLM on rented GPUs."
image: /img/home.jpg
word_count: 4500
reading_time: 12
series: "Self-Hosting LLM Inference on Rented GPUs"
series_part: 1
---

I was on the $20 Claude Pro plan and hitting rate limits every other session.

Mid-refactor. Mid-thought. "You've hit your session limit. Resets 3:45pm." Close the laptop, come back later, rebuild the context in your head all over again.

The fix seemed obvious — upgrade. But before spending $100 or $200 a month I wanted to understand what I was actually paying for, and whether there was a smarter way.

Here is what Claude Code is actually doing to your token budget.

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

Source: [32blog.com — Claude Code context window exceeded fix](https://32blog.com/en/claude-code/claude-code-context-window-exceeded-fix)

Real numbers from actual sessions:

- Mid-session, medium complexity: **53,152 tokens**
- Same session before hitting the wall: **131,841 tokens**
- Default output request per call: **32,000 tokens**
- Combined: exactly one token over the limit

The context didn't spike. It crept. File by file, tool call by tool call, edit by edit. By the time the error fired, the session had been growing for twenty minutes and nobody noticed until it hit the ceiling at full speed.

Also worth knowing: auto-compaction reserves roughly 33K tokens on a 200K window as a buffer so Claude can finish its current response when compaction triggers. Your effective working limit is always lower than the headline number. ([32blog.com](https://32blog.com/en/claude-code/claude-code-context-window-exceeded-fix))

This is why upgrading the plan alone doesn't fully solve it.

Rate limits cut you off at the plan level. You've exhausted your 5-hour window, wait for the reset. Context walls cut you off at the session level. A single task grew too large for the model to continue. Two different problems, same result: Claude Code stops mid-work.

On a paid plan, both happen. On a self-hosted instance, neither does. The context window is whatever your GPU's VRAM can hold. There's no rolling window, no shared quota, no plan ceiling.

But to get there, you first need to understand how sessions grow this fast — and what you can do about it while still on the API.

So you have two problems. Rate limits at the plan level and context walls at the session level.

For context walls, Anthropic's official fixes are /compact, /context, and disabling unused MCP servers. ([Claude Code error docs](https://code.claude.com/docs/en/errors)) Run /context first — you'll typically find MCP tool definitions quietly consuming thousands of tokens you never asked for.

/compact compresses conversation history to free space. If it fails because the window is already full, press Esc twice to step back a few turns, then compact from there. If that still fails, /clear and start fresh — your previous session is preserved and reopenable with /resume.

The community workaround of lowering contextWindowSize in settings exists too. It forces earlier compaction but trades depth for stability — longer sessions and multi-file refactors become harder the moment you cap the window. It treats the symptom, not the cause.

For rate limits there's no workaround. You wait for the reset or you upgrade. That's the plan ceiling — and it's why the maths on self-hosting started making sense.

-----

## The Cost Maths

I was on the $20 Pro plan. Hitting rate limits every other session.

The obvious next step was to upgrade. But before reaching for a credit card I did the maths — and the maths told a different story.

Here's what the options actually look like in 2026:

|Plan               |Monthly Cost |Token window       |Reality                         |
|-------------------|-------------|-------------------|--------------------------------|
|Claude Pro         |$20          |~44K tokens / 5hr* |Hits limit mid-session regularly|
|Claude Max 5x      |$100         |~88K tokens / 5hr* |Ok for light daily use          |
|Claude Max 20x     |$200         |~220K tokens / 5hr*|One heavy agentic week burns it |
|Sonnet 4.6 API     |Pay-per-token|Unlimited          |$3/M input, $15/M output        |
|2x RTX 4080S gpuhub|~$84/mo      |Unlimited          |Private, no throttle            |
|2x RTX 5090 gpuhub |~$130/mo     |Unlimited          |Private, no throttle            |

Token estimates are community-reported approximations — Anthropic does not publish exact per-window token counts. Sources: [Duet](https://duet.so/blog/claude-code-pricing), [CloudZero](https://www.cloudzero.com/blog/claude-code-pricing/)

Real gpuhub instance prices from actual sessions:

|GPU Instance   |VRAM         |Cost/hr   |Est. monthly (6hr/day)|
|---------------|-------------|----------|----------------------|
|2x RTX 4080S   |64GB combined|$0.46-0.48|~$84                  |
|RTX 5090 single|32GB         |$0.36     |~$65                  |
|2x RTX 5090    |64GB combined|$0.72     |~$130                 |
|RTX PRO 6000   |96GB         |$0.91     |~$164                 |

Prices from actual sessions — verify current rates at [gpuhub.com](https://gpuhub.com) before provisioning. gpuhub bills per second.

One important distinction from subscription plans: gpuhub is pure pay-as-you-go. The 6hr/day estimate in the table is just that — an estimate. If you use it for 2 hours one day and 8 the next, you pay exactly for those hours. No monthly floor, no wasted budget, no commitment. What you use is what you pay.

Token maths at scale:

|Monthly volume   |Sonnet 4.6 API cost|2x RTX 5090 gpuhub|
|-----------------|-------------------|------------------|
|10M in / 2M out  |~$60               |~$50-80           |
|50M in / 10M out |~$300              |~$130             |
|100M in / 20M out|~$600              |~$130             |
|200M in / 40M out|~$1,200            |~$130             |

Sonnet 4.6 API pricing source: [Anthropic](https://www.anthropic.com/pricing/api)

Upgrading to Max 20x would cost $200 a month — and still hit limits on a heavy agentic week. The API pay-per-token route spikes unpredictably on long sessions. One developer tracked 8 months of intensive Claude Code usage — API equivalent exceeded $15,000, Max plan for the same period was ~$800, a 93% reduction. ([CloudZero](https://www.cloudzero.com/blog/claude-code-pricing/), [Verdent](https://www.verdent.ai/guides/claude-code-pricing-2026))

One more thing worth knowing: your usage limit is shared across Claude.ai chat and Claude Code from the same bucket. Heavy chat usage eats directly into your Claude Code allocation. ([Duet](https://duet.so/blog/claude-code-pricing))

But even $84/month on gpuhub beats Max 5x at $100 — with no rolling window, no shared quota, and no data leaving your network.

The upgrade path wasn't worth it. Self-hosting was.

-----

## The Part Nobody Talks About — Privacy

Cost alone would have been enough. But there was a second constraint that made the API a non-starter regardless of price.

The apps being built process data that cannot leave the local network. Sending any of that to a third-party API isn't a preference — it's an architectural hard stop.

Here's what the actual data policy looks like in 2026, because most people haven't read it:

> **Free, Pro, and Max accounts are consumer accounts. By default, these train on your data unless you manually opt out. Only Commercial or Enterprise tiers prohibit data training by default. ([Anthropic Privacy Policy](https://www.anthropic.com/privacy))**

In August 2025, Anthropic introduced an opt-in toggle with a deadline of September 28. If you opted in, Anthropic could retain conversations in de-identified form for up to **5 years** — a 60x increase from the previous 30-day retention period. ([Anthropic](https://www.anthropic.com/privacy))

Even with training opted out:

- Consumer account data: retained 30 days
- Zero Data Retention (ZDR): **only available for commercial API keys or Claude Enterprise — not Pro or Max** ([Anthropic Privacy](https://www.anthropic.com/privacy))
- Claude Code still stores every read, bash command, search result, and file edit locally as a plaintext JSONL file — and the prompts containing all of it still transit Anthropic's servers

For financial data, proprietary logic, anything sensitive — self-hosting isn't paranoia. It's the only architecture where the data genuinely doesn't leave.

-----

The actual switch turned out to be a single line in one config file. Claude Code doesn't know or care what's on the other end of the URL. It just sends requests. But figuring out what to put on the other end took several weeks, five model attempts, and more broken GPU instances than expected.

Spoiler: the first model attempt ran out of VRAM in under ten seconds.

That's Part 2.

-----

*Next: Part 2 — The model hunt (five attempts, one winner) — coming soon.*

*This is Part 1 of a 5-part series on self-hosting LLM inference on rented GPUs.*

*I write about Java, PKI, LLM inference infrastructure, and building things on the side. Find me on [GitHub](https://github.com/kayisrahman) and [LinkedIn](https://linkedin.com/in/kayisrahman).*
