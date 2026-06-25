---
layout: page
title: Projects
description: "Things I'm building outside of work."
permalink: /projects/
background: '/assets/img/projects.jpg'
---

Things I build outside of work.

---

**Synapse**
*In Progress*

A self-hosted AI proxy running on a Raspberry Pi 5 that routes inference requests intelligently between a remote vLLM server on rented GPUs and lightweight local models depending on task complexity. Built to replace paid API subscriptions with private, fast, self-hosted inference. Claude Code, long-context reasoning and tool use go to the full model. Simple queries route elsewhere.

Stack: Java, Spring WebFlux, Raspberry Pi 5, vLLM, Qwen3.6, Redis

[GitHub →](https://github.com/kayisrahman)

---

**Swing Trading System**
*In Progress*

A fully automated swing trading engine targeting NSE/BSE Indian equities. One to four week holds. Spring Boot backend, TimescaleDB for OHLCV time-series, Redis for indicator caching, LangChain4j connecting to a locally hosted Qwen3 model on a single RTX 5090 as a qualitative sentiment filter. The LLM is not the signal generator. EMA crossovers, RSI momentum, and volume confirmation generate signals. Telegram alerts, Zerodha Kite Connect integration.

Stack: Java, Spring Boot, TimescaleDB, Redis, LangChain4j, Qwen3, Zerodha Kite Connect, Telegram

[GitHub →](https://github.com/kayisrahman)

---

**Pomodoro Timer iOS**
*In Progress*

A focus timer for iOS with an LLM intelligence layer underneath. Not just a countdown. The model adapts session lengths based on task type, surfaces focus patterns over time, and nudges you when behaviour suggests burnout rather than productivity.

Stack: Swift, iOS, LLM inference via Synapse

[GitHub →](https://github.com/kayisrahman)

---

**PKI Certificate Lifecycle Tool**
*Planned*

A certificate expiry and lifecycle management dashboard for engineers working with PKI infrastructure. Monitors certificate chains, alerts on expiry, tracks issuance and renewal across environments. Enterprises manage hundreds of certificates manually. They should not have to.

Stack: Java, Spring Boot, PostgreSQL, React

[GitHub →](https://github.com/kayisrahman)

---

**Nifty Screener**
*Planned*

A public-facing stock screener surfacing signals from the swing trading system without exposing the full engine. Technical indicators, momentum filters, volume confirmation. Built for retail investors who want signal without noise.

Stack: Java, Spring Boot, TimescaleDB, Angular

[GitHub →](https://github.com/kayisrahman)
