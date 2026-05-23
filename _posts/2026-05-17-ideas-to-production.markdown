---
layout: post
title: "Ideas to Production — Ten Years, Five Companies, One Direction"
subtitle: "A decade of building things, across five companies and two continents."
date: 2026-05-17 09:00:00 +0100
background: '/assets/img/posts/ideas-to-production.jpg'
tags: [career, engineering, security, backend]
word_count: 3800
reading_time: 11
---

I finished college in April 2015 and spent two months doing absolutely nothing useful.

June came. I packed a bag, left my hometown in Malappuram, and moved to Kochi. Nineteen years old, hostel room, first job. The office had a ceiling fan with strong opinions about when it wanted to work, monitors that had seen better decades, and six or seven people who were all figuring things out at the same time. No mechanical keyboards. No ping pong table. No cold brew on tap.

Just desks, a deadline, and the particular energy of people building something real for the first time.

The main project was MediSoft — a hospital management system running across multiple hospitals in the region. Java EE, Hibernate, JSP. The deployment process was exactly what you'd imagine: log into the server, push the WAR file, refresh and hope. That was it. Ship and pray a little.

What kept it interesting was the variety. Alongside MediSoft I got pulled into a couple of .NET projects — Windows desktop applications in C#. Coming from Java, that felt like switching languages mid-conversation. Stateful UIs, event-driven logic, a completely different mental model. Nobody hired me as a .NET developer. Nobody cared. There was work, I was there, I figured it out.

Those two years weren't glamorous. But the feedback loop was tight and honest. Build it, deploy it, watch a nurse at a hospital counter use it. That's a different kind of education than any classroom gives you.

Ideas to production. That's the job. Everything else is commentary.

## TCS — 2018

TCS in 2018 was a different world entirely.

I didn't start writing code immediately. My first posting was application support — monitoring live systems, responding to incidents, keeping things running. No greenfield development, no clean abstractions. Just live systems, real users, and problems that needed solving now.

Most people treat support as a lesser thing. I didn't see it that way. Reading a system you didn't build, under pressure, and understanding it well enough to fix it — that's a skill. It gave me a perspective on production that developers who've only ever written new code don't always have.

From there came the integration work. Vendor application upgrades, interfaces stitching together systems that were never designed to talk to each other. Some of that code was fifteen, seventeen years old. The kind of codebase where you move carefully and test everything twice. There's a quiet respect you develop for software that's still running long after the people who wrote it have moved on.

Then the gift management system — Spring Boot, LDAP, SSO authentication, batch jobs, AWS. This is where Spring Security stopped being something I'd read about and became something I actually understood. Authentication flows, trust boundaries, the mechanics of who gets to access what and why.

Then a greenfield project from scratch. Microservices, Dockerized containers, Jenkins pipelines, Kubernetes orchestration, auto-scaling. Everything I'd studied but never built under real constraints. That's where CI/CD stopped being a concept and became something I knew how to design and defend.

The client was US-based. Calls at six-thirty or seven in the evening, Indian time. You adjust. You eat dinner late. After a while you stop noticing.

## Ribalta — 2021

Ribalta in 2021 was a choice I made with open eyes.

I wanted real startup experience — not the kind where a company calls itself a startup but operates like a cautious mid-sized enterprise. Ribalta was early stage, small team, three engineers covering everything in and out. I knew exactly what I was walking into. That was the point.

The stack was whatever the problem demanded. Golang for backend services. Java for parts of the Android work. A little PHP where it already lived. And Android TV — building a lean-back streaming experience for television is an entirely different design problem from a phone app, and I got to figure that out from scratch with no playbook and no one to ask.

We started from an office in Trivandrum. New city, good people, the specific electricity that only exists in small teams where everyone's outcome is tied to everyone else's decisions. We used to go on road trips together — colleagues who quietly became people you actually spent time with, because when there are only three of you the line between work friends and real friends stops being meaningful. Then the work shifted remote. The camaraderie didn't.

When something broke, there was no escalation path. There was the three of us, the problem, and whatever we could work out between us. We ran sprint planning and backlog grooming not because a process document told us to but because without structure, three people can pull in three directions almost instantly.

That's where I learned what ownership actually means. Not ownership as a slide in a company all-hands. Ownership as in — if this breaks at midnight, it's on you, and you'd better already know the system well enough to fix it.

## Ernst & Young — 2023

EY, January 2023 to September 2024 — cloud at scale.

AWS, GCP, Azure, GKE, Terraform, microservices, GitLab CI pipelines — not one or two of these, all of them, on a platform that actually mattered to the people using it. The team was larger, the domain was enterprise, and the problems were the kind that punish you for thinking out loud before thinking properly.

We automated Kubernetes namespace provisioning through Terraform — the kind of infrastructure work that looks invisible when it's done right and catastrophic when it isn't. Redesigned a legacy email archival system using S3, SQS and Pub/Sub, turning something expensive and fragile into something quiet and cheap. Introduced a branching strategy that removed enormous amounts of redundant deployment work the team had quietly accepted as normal.

But the part nobody warns you about when you become a Technical Lead is the people half of the job.

Seven engineers. Different time zones, different working styles, different tolerances for ambiguity. Keeping them aligned on a platform that kept acquiring new requirements meant learning when to absorb scope creep and when to push back hard — and getting that wrong in both directions before I got it approximately right. I approved architectural decisions I hadn't fully thought through. I had to walk them back. That's uncomfortable in ways that shipping a bug never quite is.

What I'll remember as much as any technical win is the people. Every company I've worked at, the colleagues were good. I'm still in touch with people from all five. That doesn't happen by accident. It means something about how you choose where to work and who you choose to be while you're there.

Then September 2024 came. I handed in my notice and enrolled in a master's programme the same month.

## Dublin — 2024

No plan B on the table. A couple of weeks later I was sitting in a Dublin Business School lecture hall, student again at twenty-nine, wondering if I'd made a spectacular mistake.

No salary. New country. Part-time work to cover rent. Ireland doesn't ease you in gently — the weather alone makes sure of that. But there's something clarifying about stripping everything back. No title, no team, no accumulated reputation. Just you and what you actually know when nobody's watching.

The MSc was Information Systems with Computing. The coursework was solid. What I didn't expect was learning to write about systems instead of just building them — to step back far enough to describe what's actually happening and why. It turns out that's a different skill entirely. And a more useful one than I gave it credit for before I had it.

Then Qualcomm called. January 2026. I packed up Dublin and headed to Cork. New city, new domain, new everything. Again.

Some people need stability to do their best work. I've started to think I need the opposite.

## Qualcomm — 2026

Qualcomm since January 2026 is where everything converges.

Staff Security Software Engineer. PKI, HSM, secure device provisioning. The work is cryptographic infrastructure — establishing verifiable trust between hardware devices and backend systems in ways that cannot be forged, intercepted, or replayed.

It's slower work than application engineering. More deliberate. The failure modes are different in a way that takes time to internalise. A bug in an app throws an exception. A flaw in cryptographic infrastructure sits silently and creates exposure you might not discover for years. That changes how you read a specification. It changes what *done* actually means. It changes the weight of a code review.

Some days it feels like I'm nineteen again — new office, new problem space, figuring out the shape of what I don't yet know. That feeling used to make me anxious. Now I recognise it for what it is.

It just means I'm in the right place.

---

Ten years. Five companies. A hostel room in Kochi, a lecture hall in Dublin, a desk in Cork.

The stack changed every few years. The questions underneath it never did.

*Does it work. Is it correct. Will it hold.*

---

I write about Java, PKI, LLM inference, and building things on the side. Find me on [GitHub](https://github.com/kayisrahman) and [LinkedIn](https://linkedin.com/in/kayisrahman).
