---
{
  "title": "A Tale of Three Months with AI: Guardrails, and How We Are More Alike Than We Wish to Admit",
  "description": "Three months of hands-on agentic development: moving beyond greenfield vibecoding into the mud of real-world software engineering, and why deterministic guardrails matter more than ever.",
  "date": "2026-10-20T11:00:00+02:00",
  "draft": false,
  "tags": ["ai", "programming"],
}
---

As I promised in [A fresh leaf](<{{< relref "2026-10-08-a-fresh-leaf.md" >}}>), I want to share my experiences with using LLMs for agentic development in the last three months.
Summarised: I started as an initial sceptic, became quite enthusiastic, and am now moderately enthusiastic and guarded for the future.
Allow me to take you through my journey.

Coming from a background of rigorous engineering with up-front thinking about program structures, algorithms, and types, it seemed incomprehensible to me how a jumped-up neural net could generate working code, let alone sensible systems.
How to write understandable software has been hotly debated for decades, continuously shifting as new languages, type systems, and ecosystems introduce different trade-offs.

Early in my career, coming from academia, I was fixated on finding the "just-right" solution: elegant abstractions, strict correctness, and formal soundness.
As problems grew larger and messier, and time became scarce, holding out for perfection became an obstacle.
I found far more satisfaction in achieving _something_ functional, with elegance as a welcome second.
Having both overlap is fantastic; achieving the actual goal remains both a priority and a necessity.

This shift in perspective explains why LLMs, and agentic development, eventually proved so attractive to me.
Expecting a model to generate a "just-right" architecture, code, or text only causes frustration.
Conversely, it acts as an engine to reach concrete milestones in areas I was always curious about without having the time to do things "properly", or simply eliminating the _toil_ of writing boilerplate when I can describe my intent with less mechanical precision.

If your goal is simply to achieve something concrete, they are remarkably effective.
That spans from something as simple as integrating an unfamiliar library to breaking ground in an unfamiliar problem domain where you lack the tools and vocabulary.

## The greenfield honeymoon

My catalyst was a device sitting in my living room. 
After reports emerged that its successor was aggressively hostile to user privacy, I wanted to verify whether my own hardware behaved the same way. 
Since I had no background in firmware inspection or packet analysis on embedded devices, I was initially discouraged. 
Luckily, I noticed right then that I had a free trial month of Codex lying around, so I decided to see how far these so-called "amazing frontier" models could carry me.

I started out by talking in ChatGPT and asking questions on how to solve my problem.

{{< observation title="Search engine on steroids" >}}
I realised I was using ChatGPT more and more as a search engine on steroids: give it a simple question, let it figure out where to search, and synthesise the answer.
{{< / observation >}}

This quickly became using Codex to work on my laptop, letting it download images, and running local scripts.
The way of work was very interactive, prompt-based, and relatively unstructured and unguided.

{{< observation title="Over-caffeinated Junior Engineers" >}}
I soon began to realise that LLMs responded a lot like humans, and later like over-caffeinated junior engineers.
The more concrete and narrow the scope of an assignment or question the better.
{{< / observation >}}

I was amazed by how well the model worked when told to achieve a very specific goal and even more so when giving the context and constraints in a single prompt: as long as the prompt wasn't too big.

When I pointed out errors, the models complimented me lavishly on the discovery and apologised profusely for the defect.[^sycophancy]
Yet that contrition yielded no durable change in their subsequent behaviour.
It was just text, learned from reams of documents.

### Asserting control

As soon as the codebase and program to detect any nefarious behaviour was growing, I noticed the performance of the models degrade.[^context-degradation]
Responses lost their initial speed, precision, and correctness.
The model began overlooking interdependent updates that required lock-step coordination, shedding context established in earlier turns.
In other words, it required more hand-holding, for which I had neither time, nor patience, nor appetite.

At that point, I also started working on some other ideas that had been lying dormant in my head.
I used an agent to create a Terraform project to spin up a small timeseries-db backup in Google Cloud constrained such that it would stay within the free tier.
Creating the provisioning scripts was the work of mere moments.
It checked the exact sizing requirements for staying within the free tier including some caveats, whipped the Terraform code together, and added the Tailscale integration.
All from a simple prompt of what I wanted to achieve and the fact that I wanted to be able to reconstruct the VM from a single command.
That was awesome.

It was here that I first started suspecting a deeper link between models and humans: it all came down to how bounded the task was.
Over the years, I noticed that when you ask a person an open-ended question, their answer is mostly influenced by their recent history, their emotional state, and whatever problem they were just wrestling with. 
Models work almost identically. 
Ask them a broad, open-ended question and you get a generic response shaped by their pre-trained weights and immediate context window. 
Getting either a human or a model to think outside that default box requires tight, explicit constraints, and careful prodding (or prompting).

One of the often touted best-practices I encountered in the beginning was to have a three-phase approach.
Use one agent to plan the work interactively, producing a handover document.
Then task a second agent with implementing that handover document unattended.
Finally pass the control to a review agent, which gets the diff and the plan, and reviews it.
The obvious improvement is to make the last two communicate and work in a loop until no more findings are established.

I first tried automating this division of work (planning, implementation, and review) by writing conventions as prose in `AGENTS.md`.
As code grew and threads grew longer, however, prose instructions became less effective and easily ignored by the model. 
To take this workflow further, I wanted to combine the specialised strengths of different models: Codex as the fast, local workhorse for writing code, and Gemini as the thoughtful reviewer for architecture. 
To let them talk directly to each other, I built `gemini-mcp`. 

Here I hit another quirk of model psychology: autonomous models possess such supreme self-certainty that review triggers require explicit programmatic conditions. 
A loose prose directive like _"Ask Gemini for help when uncertain"_ generated precisely zero invocations; the model was never uncertain. 
Conversely, a deterministic circuit breaker ("Attempt to solve each problem a maximum of three times; if still failing, consult Gemini via MCP, implement its recommendations, and repeat") reliably prompted Codex to escalate for second opinions. 

This put me with an immediate security challenge: because I was using this setup to assess the privacy of the TV, I needed a classifier to prevent prompts between models from looking like offensive security probes, which could get my Google account suspended.

Scaffolding the classifier was quick work. 
Making it function reliably, however, involved dozens of subtle architectural choices that flew past too fast for me to properly evaluate. 
Steering the model through prose in `AGENTS.md` simply wasn't giving me the control I needed. 
That was the moment I turned to systematic, deterministic tooling, starting with a simple directive: run `make check` after every modification to verify that the project was still coherent.

This worked so well that ever since I have been extending what is behind this check continuously.
Relying on what I already knew worked well in larger projects with multiple contributors, I decided to use strict tools to check as much as possible.
This meant automatic linting, formatting, and upgrades of the different types of files in the project:
Markdown, code, YAML files, ADRs, formatting, modernisation of Go, and dependency upgrades.
Moreover, I decided to take more out of the hands from the agents and not even bother them with formatting and easy to fix linting errors by automatically running formatting, automatic imports, and adding `--fix` to golangci-lint already as part of the *check* step to ensure all easy fixes were implemented before the agent even saw the result.[^static-analysis]

{{< observation title="What helps human engineers, helps agents even more" >}}
Battle-tested tooling such as compilers, type systems, linters, tests, libraries, proper documentation, and proper (actionable) error messages help humans a lot.
They help agents continue forwards even faster, multiplying the leverage.
{{< / observation >}}

I introduced this rigorous tracking in my first big vibe-coded project, [`semantic-editor`](https://github.com/spockz/semedit). 
The whole point of the project was to let the agent worry about intent while delegating the nitty-gritty editing work to deterministic AST tools. 
I applied that same mindset to the repository itself, using strict checks for the markdown, the ADRs, and the schemas.
This approach worked pretty well, and the project was progressing smoothly.

Around that time, another personal project came into my head.
I had always felt sad that in this age of search engines and lately omniscient (and often wrong) LLMs, my kids didn't have the experience to just browse an encyclopedia, guided by interest.
For me the concrete example of this was Microsoft Encarta '97.
It gave me endless browsing, learning of all kinds of things, and important to my parents (and now me), relatively safe because it was curated content.

As you may have guessed, I decided to vibe code the answer to this problem as well and make a viewer of Encarta data. 
Because in-repository decision tracking had operated satisfactorily, I relied on that established harness to guide development.

However, because this was my first GUI app, I sensed it needed another layer of constraints and decision tracking which I dubbed "GUI-invariants" as a specialised form of ADR for GUI decisions such as layout on a more generic level instead of specific decisions.
Getting the agent to consider the invariants proved relatively simple by just adding it to the agent instructions.
Letting it interpret them in a way was harder, because invariants are inherently generic.
So they needed more concrete positive and negative examples and be more prescriptive.

For some reason I had a quirk that I wanted all text to be "properly" type-set such as in TeX.
And why not?
We are custom coding this project and I have tokens to burn, so let's have at it.
The agent was very helpful in this bounded context.
A simple assignment, a starting example algorithm, the specific constraint of the SwiftUI infrastructure, and an oracle to determine whether it did it right.
This allowed me to set a simple goal and I could walk away from it.

These three main projects and several side projects, all shared one important property: they were greenfield projects where I individually was the sole owner and decision maker.

In urban planning terms, greenfield land is effortless to traverse and develop, precisely because it is virgin ground: untouched, open, and unencumbered by history.
Brownfield sites, by contrast, are contaminated swaths of land burdened by buried foundations, undocumented pipelines, and wrapped in a veneer of toxic legacy waste.
Left unattended, software tends to follow an identical trajectory.
When projects expand without discipline, the virgin meadow rapidly deteriorates into a muddy, hazardous swamp that engineers dread to touch.

Yet that decay is avoidable. 
Long-lived systems like the Linux kernel survive decades of churn not through bureaucratic upfront planning, but through uncompromising maintainer gatekeeping: strict subsystem boundaries, relentless review against established patterns, and clear commit rationale.[^software-evolution]
We should learn the lesson and apply it to agentic programming.
Because general purpose models lack this inherent institutional memory and intrinsic motivation, we must codify that gatekeeping directly into the environment: tracking architectural decisions, enforcing invariant suites, and maintaining deterministic mechanical guardrails.

Handled with that discipline, a codebase avoids becoming a toxic brownfield. 
Instead, it becomes a **cultivated glasshouse**: a structured, highly productive environment where automated checks supply continuous nourishment, and recorded context allows development to continue to flourish at speed.

## From the green pasture into the mud

While agents seem to shine on greenfield projects that they manage themselves, which is why vibecoding works so well, they start falling flat when projects become larger or they need to be used in established projects.
This is because the structure agents use to locate where to make changes mirrors the latent patterns ingrained during training.[^code-loc]
What a coincidence, this is also the same for human engineers.

My experience creating the initial draft of the dynamic-fault-injection filter for an open-source "Built on Envoy" project exposed this boundary. 
I had the foundational distribution formula and prior experience running it in a JVM-based sidecar; directing an agent to scaffold the initial Go plugin was straightforward. 
The friction began when integrating it upstream. 
Unlike self-owned projects where you can invent your own shortcuts, I needed to fit cleanly into the project's established discipline: adhering to its battle-tested test patterns, documentation conventions, and repository layouts.

Anyone can staple a quick fix onto a codebase, and just that quickly another wart is added.
The actual challenge is figuring out where the change truly belongs, and doing the painstaking rework so it looks like it was designed that way from the start.
Occasional contributors face this constantly: deciding when to live with an existing pattern, and when to roll up your sleeves and rewrite surrounding code to make it fit properly.

Navigating that friction succeeded primarily because I received generous, precise guidance from the lead maintainer. 
That human feedback supplied the unspoken repository context and architectural intent that neither I nor the model could extract from the code alone. 

I could then feed those maintainer insights directly back into prompts, giving the agent the concrete boundaries it needed to shape the implementation correctly.

However, sitting down quietly with a cup of tea and looking over the code, I realised the code organisation was still a bit off. 
That intuition was only available as latent skill in my head based on code scattered across the repository, and admittedly not applied consistently enough for an untrained eye like mine on the project. 
This is where concrete ADRs and contributing guides earn their keep: they turn what is locked in maintainers' heads into explicit rules agents can actually read.

When tackling bigger projects, and more-insufficient usage limits, I started looking into other models and noticed that the model families tend to have similar behaviour.

To me these are the qualifications of the model families, that I have personal experience with. Even though they are a mathematical probabilistic bag of parameters, they seem to have emergent personalities:

**Hyperactive Junior Engineer Qwen**: The Qwen family (both 3.6 dense, MoE, and 3.8) generates code at blistering speed, rarely pausing to ask questions. It demands highly granular guidance. The 2-bit Saluki quantisation displays this impulsiveness even more acutely.

**Expensive easily-distracted CS Professor Gemini**: Extremely strong at reducing a problem description to its core and finding the academic consensus on the solution fit. It is particularly strong in Go and machine-learning topics, and great at explaining *why* models behave in certain ways. Especially at the architecture level.

However, it is expensive for routine coding, not *that fast or good*, and gets stuck in repetitive read loops, endlessly inspecting files without making changes:
![Gemini stuck in a read-inspection loop](/images/archive/stuck-gemini.png)

Gemini 4.0 Argon is even worse for everyday work: a review of a small piece of text burns through 50% of my weekly Pro allowance. No matter how good it is, it is unusable in this way.

**Dependable, rabbit-holing workhorse GPT**: This is the family that is actually interesting, especially given the pricing strategies. In general, I found GPT to rabbit-hole into a particular solution direction faster and refuse to get out of it, where Gemini could step back. Sol 6.1 (and prior with 5.6 Terra) is the perfect reviewer and Luna (both 5.6 and 6) is the perfect workhorse. Luna is decently smart, but especially because it is virtually free. Both Sol and Luna have the tendency to double down on a chosen path and are very hard to get to think outside the box or take a step back.

Across all these model families, one shared quirk emerged: an almost infuriating default bias toward Python for writing quick scripts. 
Instructing an agent explicitly *not* to use Python invariably yielded TypeScript. 
This is clearly the gravitational pull of public training data, where Python and TypeScript examples dwarf everything else.[^corpus-bias] 
It constantly left me longing for something like `goshell`: a single standalone Go file declaring its dependencies inline, executed directly without scaffolding an entire module.

AI companies love to shout about breakthroughs at the absolute peak of their frontier models. 
Yet what actually changed my day-to-day experience was not a dramatic leap in the raw intelligence of the top tier. 
The real shift was the surge in quality among the *affordable* models: the ones actually in reach for everyday work. 
A model like Sol 6.1 brings reasoning that used to belong exclusively to flagships, but at a price where you can afford to keep it running continuously in tight feedback loops. 
It is not that the absolute Pareto frontier has moved, but that the *affordable* Pareto frontier has moved. 
In that sense, practical agentic development has become far more of an economics game than a science game.[^inference-econ] 
Slashing the price of a monster flagship like Astra or Argon would be an easy way to win the "smartest model" contest, but subsidising those gargantuan inference costs would burn through cash at a staggering rate. 
The provider that wins the practical developer ecosystem won't necessarily be the one with the cleverest paper, but the one that delivers reliable reasoning at a price where hundreds of daily iterations don't break the bank.

Choosing the right model for the task at hand is an important part of being able to work in large code-bases.
However, more is needed.

## The Engineering harness

Looking back at my career on what made successful teams achieve their successes besides their social skills and trust among themselves, was the fact that they spend time on great tooling. Even the best of minds have a limit to how much they can keep in their RAM, how fast they can write new code, and how fast they can diligently review code.

They use tools to prevent discussions about code style by applying linters and auto formatting rules, anywhere from in a check command, a git hook, or in the PR reviews.
This avoids noise.

They have an integral test-suite, that lets them move fast ahead with ideas, while mistakes are still caught.
This integral test suite goes further than mere unit tests and includes tools that have been around in various variations.

It automatically tests for:

* Performance
  * Microbenchmarking
  * Profiling
* Security
  * Static analysis
  * Dynamic pentesting tools
  * Fuzzers
* Quality
  * Linting
  * Coverage and coverage trends
  * Property tests and mutation testing

Notice how far away this is from reams of prose.
Humans with a limited span of attention and limited brain capacity cannot map their every action against a bookwork of architecture and procedural rules, it would be too time consuming.
They use heuristics to drive their "online" work, and "offline" guard rails to prevent their oversights from hitting customer impact.
These same measures apply with agentic coding.

Because agents are LLMs with extra rules, as long as you have compute power, and sufficient supply of well thought out jobs / issues, the agents can produce code faster and without bottlenecks.
This increases the leverage, arm, or multiplier of how much agents benefit from having these guardrails in place.


### Institutional Knowledge 

Guardrails in the form of rigorous automated test suites, as introduced above, carry us only so far.
They cannot remember the *WHY*: why do we have a certain test, why do we implement a feature, why did we choose a particular library.

In smaller, one-person greenfield projects, recording the decisions in simple ADRs was good enough to keep the agent bounded and give me a mechanism for manually pointing out to adhere to the ADRs.
In sprawling brownfields and cultivated glasshouses alike, decisions come from many different directions; the institutional experience of the engineering corps is heterogeneous and in continuous flux.

This is what organisational theorists and software architects also refer to as "Institutional memory". [^inst-mem]

In humans this knowledge is gradually ingrained into the engineer through experience.
If you are lucky they stay on the projects long enough to provide continuity.

Agents have no such context at all.
Actually, engineers are *also* fallible.
They can forget. (Context compaction and fresh sessions)
They can have a lapse of focus or being disturbed. (The "lost-in-the-middle" effect[^lost-middle].)
They can move on to a new project. (Starting brand new chats.)

With agents, that departure happens constantly: every time you open a brand-new chat thread, every unrecorded insight simply evaporates into /dev/null (unless you count the provider retraining on your prompts ;-) ).

Thus we need a system that actually benefits both engineers *and* agents and so I extended the harness in two ways:

First: I introduced **Research Questions**, capturing the *intent* of any change or feature.
Basically it represents the inbox to the project.
Before an implementation starts, stakeholders and engineers discuss goals, trade-offs, and explore alternatives.
This leads to what to build and more importantly what **not** to build.
This is context that very rarely survives into issues, pull-requests, and/or commit messages and is therefore lost.
By recording these deliberations in the research questions, it retains this information while at the same time also allows ADRs to stay small, concrete, and focussed, to stay kind to the limited context windows of Agents.

Secondly: Even with this mechanism in place, I noticed that there was still drift so I added a rule to check nightly or weekly to do a full reconciliation of changes against established ADRs and of code against Research Questions to see whether new data had come in or we actually already made decisions.

This now closes the loop in the harness:

* Agents drive change
* Linters, compilers, and tests protect the quality of the implementation and documentation.
* Reconciliation loops at various levels and scopes (code change, daily, weekly, monthly)

This highlights a fundamental parallel: the core engineering discipline remains identical for human engineers and autonomous agents; only the execution velocity differs.
Improvements introduced for engineers benefit agents even more dramatically.
We must heed what software engineering has demonstrated over the past half-century.

This entire harness stems from a core conviction: **let the neural net focus on high-level reasoning and intent, while letting deterministic CPU tools handle the mechanical verification.** 

Compilers (especially with strong type checkers), linters, and test suites execute in milliseconds with absolute certainty. 
Burning expensive, probabilistic tokens on verification that a local binary handles in microseconds is wasteful in the extreme.

This same conviction _also_ drives my preference for dense languages such as Haskell.
Strict types, purity as a first principle, and a dense syntax help to keep functions small, isolated, and focused.
It allows an engineer to hold a function in *their* working memory, reason about its behaviour, and ensure its correctness.
From the outside it is enough to know that the types match.
Afterwards, type signatures alone dictate how to compose functions, rarely requiring subsequent repairs due to unforeseen defects.

## What is next: leveraging the speedup gains from agentic development and corporate fit

With deterministic guardrails governing both code quality and institutional memory, a curious dynamic emerges.
The bottleneck shifts upwards from code generation to coordinating concurrent, sensible streams of work.

Additionally, with the approach of using ADR and Research Questions as outlined in this article, the institutional memory remains siloed in a single repository, whereas architectural principles, designs, and organisational conventions span the entire organisation and therefore multiple projects.
The current approach still lacks a clean mechanism to unlock that higher-level institutional memory to the agent without flooding its context window with (potentially) irrelevant history.

These two challenges are where an active body of research is currently directed, and what I will focus on next.[^multi-agent-research]


## Conclusion

I like to end with something that has been cycling in my head all along and my AI quoted back to me when reviewing this conclusion:

> software engineering has never been about syntax; it has always been about managing fallible, context-bounded intelligence, whether made of carbon or silicon.
>
> *(Gemini, when repeating back what I told it above.)*

Software *Programming* has traditionally centred on syntax, algorithms, and code generation. 
In that narrow domain, modern models excel: they produce functional routines, wire up libraries, and spit out syntax with astonishing velocity.
On that level, programming as we know it has become very niche.

Conversely, Software *Engineering* has *always* begun where raw programming ends.[^swe-foundation] It encompasses establishing architectural boundaries, aligning technical trade-offs with business priorities, and managing fallible, context-bounded intelligence across time.

Although agents are tireless programmers, they possess zero business context. They lack the knowledge whether a repository is a throwaway prototype that needs to ship by Friday or a core payment engine where a subtle edge case creates legal liability. 
Engineering transcends syntax because it demands commercial **judgement**: knowing when to accept technical debt, how to balance elegance against commercial deadlines, and which problems actually matter to the business. 
Institutional memory (captured in our ADRs and Research Questions) is simply the durable record of those business trade-offs. The agent can generate the syntax, but the engineer must supply the strategic context: the architectural boundaries, the business priorities, and the long-term intent that code alone can never express.

Ultimately, in the context of software engineering, human engineers and autonomous agents are far more alike than we might care to admit.
Both are fallible, context-bounded intelligences prone to lapses in attention, recency bias, and unstated assumptions.
The hard-won disciplines of decades of software engineering (clear boundaries, explicit contracts, automated invariant testing, and durable records of intent) were always meant to safeguard us from our own limitations.
It turns out nicely that these exact same disciplines are also what allow AI to succeed.
Hopefully this marks the time where we are able to bring these practices in daily use.

What we need to figure out next is how to leverage the immense speed gains agentic programming gives us while staying sane as engineers ourselves.

### Postscript
This blog post is itself an example of that dynamic: making something of higher quality in the same amount of time. 
I spent about a day writing this post. 
Using LLMs in a targeted way allowed me to test different phrasing and let the model hunt for inconsistencies in my arguments. 
All text was hand-written (or rather, typed), then reviewed alongside Gemini to refine logical flow, prune meandering thoughts, and source primary references. 
I hope you enjoyed the journey.

[^sycophancy]: M. Sharma et al., ["Towards Understanding Sycophancy in Language Models"](https://doi.org/10.48550/arXiv.2310.13548), arXiv:2310.13548, 2023; E. Perez et al., ["Discovering Language Model Behaviors with Model-Written Evaluations"](https://doi.org/10.48550/arXiv.2212.09251), arXiv:2212.09251, 2022.
[^context-degradation]: C.-P. Hsieh et al., ["RULER: What's the Real Context Size of Your Long-Context Language Models?"](https://doi.org/10.48550/arXiv.2404.06654), arXiv:2404.06654, 2024.
[^static-analysis]: C. Sadowski et al., ["Tricorder: Building a Program Analysis Ecosystem"](https://doi.org/10.1109/ICSE.2015.76), *Proceedings of the 37th International Conference on Software Engineering (ICSE)*, pp. 598–608, 2015.
[^software-evolution]: M.M. Lehman, ["Programs, Life Cycles, and Laws of Software Evolution"](https://doi.org/10.1109/PROC.1980.11805), *Proceedings of the IEEE*, 68(9), pp. 1060–1076, 1980; D.E. Perry and A.L. Wolf, ["Foundations for the Study of Software Architecture"](https://doi.org/10.1145/141874.141884), *ACM SIGSOFT Software Engineering Notes*, 17(4), pp. 40–52, 1992; M.W. Godfrey and Q. Tu, ["Evolution in Open Source Software: A Case Study of Linux"](https://doi.org/10.1109/ICSM.2000.883030), *IEEE International Conference on Software Maintenance (ICSM)*, pp. 131–142, 2000; G. Kroah-Hartman, ["Linux Kernel Development Model"](https://www.kernel.org/doc/html/latest/process/development-process.html), *The Linux Kernel Documentation*, 2018; D. Spinellis and P. Avgeriou, ["Evolution of the Unix System Architecture: An Exploratory Case Study"](https://doi.org/10.1109/TSE.2019.2892149), *IEEE Transactions on Software Engineering*, 47(6), pp. 1134–1163, 2021.
[^code-loc]: R. Hu et al., ["Schrödinger's Code Repository: Evaluating LLMs on Dynamic Codebases"](https://doi.org/10.48550/arXiv.2406.03444), arXiv:2406.03444, 2024; R. Brooks, ["Towards a theory of the cognitive processes in computer programming"](https://doi.org/10.1016/S0020-7373(83)80031-5), *International Journal of Man-Machine Studies*, 18(6), pp. 543–554, 1983.
[^corpus-bias]: D. Kocetkov et al., ["The Stack: 3 TB of Permissively Licensed Source Code"](https://doi.org/10.48550/arXiv.2211.15533), *Transactions on Machine Learning Research*, 2022; F. Cassano et al., ["MultiPL-E: A Scalable and Polyglot Approach to Benchmarking Neural Code Generation"](https://doi.org/10.1109/TSE.2023.3267446), *IEEE Transactions on Software Engineering*, 49(7), pp. 3675–3691, 2023.
[^inference-econ]: M. Zaharia et al., ["The Shift from Models to Compound AI Systems"](https://doi.org/10.1145/3664802), *Communications of the ACM*, 67(6), pp. 24–27, 2024; L. Chen, M. Zaharia, and J. Zou, ["FrugalGPT: How to Use Large Language Models While Reducing Cost and Improving Performance"](https://doi.org/10.48550/arXiv.2305.05176), arXiv:2305.05176, 2023.
[^inst-mem]: E.W. Stein and V. Zwass, ["Actualizing Organizational Memory with Information Systems"](https://doi.org/10.1287/isre.6.2.85), *Information Systems Research*, 6(2), pp. 85–117, 1995; P. Kruchten, H. Obbink, and J. Stafford, ["The Past, Present, and Future for Software Architecture"](https://doi.org/10.1109/MS.2006.59), *IEEE Software*, 23(2), pp. 22–30, 2006.
[^lost-middle]: Nelson F. Liu et al., ["Lost in the Middle: How Language Models Use Long Contexts"](https://doi.org/10.1162/tacl_a_00638), *Transactions of the Association for Computational Linguistics*, 12, pp. 157–173, 2024.
[^multi-agent-research]: T.R. Sumers et al., ["Cognitive Architectures for Language Agents"](https://doi.org/10.48550/arXiv.2309.02427), *Transactions on Machine Learning Research*, 2024; S. Hong et al., ["MetaGPT: Meta Programming for A Multi-Agent Collaborative Framework"](https://doi.org/10.48550/arXiv.2308.00352), *ICLR*, 2024.
[^swe-foundation]: T. Winters, T. Manshreck, and H. Wright, [*Software Engineering at Google: Lessons Learned from Programming Over Time*](https://abseil.io/resources/swe-book), O'Reilly Media, 2020; F.P. Brooks Jr., [*The Mythical Man-Month: Essays on Software Engineering*](https://doi.org/10.1145/800027.808439), Addison-Wesley, 1975.

