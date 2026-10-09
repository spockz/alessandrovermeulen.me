---
{
  "title": "A fresh leaf",
  "date": "2026-10-08T22:00:00+02:00",
  "draft": false,
  "tags": ["web", "ai", "programming"],
}
---

It has been a very long time since I posted a new article. The reasons for this are quite commonplace. Starting a family and building a career captured both my interest and energy, leaving little of either for sharing.

Obviously, I never stopped reading or learning; however, the vehicle for sharing changed to one with less friction: Twitter. Educating my teammates and colleagues satisfied my need to teach and share for a long time. It felt like this was the place where I could contribute. I found so many great articles and ideas that I simply wanted to amplify them, feeling unable to add to their eloquence. Ultimately, when family and career demands grew further, I stopped sharing completely. (Twitter going down the drain also didn't help.)

Fast-forward a few years, and I found myself confronted with burnout. I was and still am lucky enough to recover at my own pace, take the time to get back on my feet, and reflect on the road that brought me here.

The burnout, while not nice to experience, also has been a great boon. It allowed me to reinvent myself, re-establish my sense of self, and plot a new course. Essentially, it meant going back to the roots of doing what I enjoy: learning, creating, and sharing.

All this time, ideas have been floating around my head continuously, ranging from academic research topics and work tasks to household improvements and side projects.

For a long time, I was quite sceptical about using LLMs for programming. My experience stemmed from early research agents around 2018–2020. I shared a common sense of AI fatigue in the news. Some enthusiastic colleagues tried to convince me how great agentic programming with these "new frontier" models was and never really succeeded since I had such an underwhelming experience with LLMs in the past.

Then, three things converged to change my mind: genuine skin in the game, zero skills to investigate on my own, and free access to a frontier model.

I saw [Gamers Nexus: "216,000,000 Spy TVs | The LG Smart TV Problem"](https://youtu.be/6IFVTcM28KA) on Hacker News about LG TVs allegedly containing spyware. I own an LG TV, an earlier model than the article, and wanted to know whether my device was affected as well. I had no experience with this type of digging into systems so I didn't know where to start. Now I had skin in the game and really wanted to figure this out. Coincidentally, I got a link from someone to an OpenAI page and realised I created an account there once because my password manager suggested a login. Curious, I logged in and saw that I had a month of free usage! So why not use it for the investigation?

Long story short, I started by using ChatGPT to explain to me how I should approach this investigation. The approach sounded reasonable. It quickly escalated to using Codex with many agents. Along the way I made many mistakes blowing through five-hour usage limits, etc. But the agents were finding results. Not the ones from the article, but close. (Maybe more on that later.) This was amazing: here was a tool that could do what I could not, and all I needed to do was nudge it forward, left, or right. I was hooked.

In a month, I went from an AI/LLM-sceptic to a guarded optimist. Doing several projects, I discovered how to control agents, the effects of harnesses, their strengths (and weaknesses...), and the tendencies of the different model families. I went from using ACP in Zed to just using Codex and Antigravity on the CLI, only rarely inspecting the code manually. Steering the code and code quality happens with test-coverage, mutation-coverage, and linting rules.

This is a journey I feel I want to share. For that I needed to redo the site. I didn't want to relearn modern Hugo or figure out all the best themes out there and how to configure them; I just wanted results.

There was another impetus. I have a long-standing pet peeve with how my blog looked before, and how sites in general look. Specifically for longer reads. I _love_ the visual of text typeset in [LaTeX](https://www.tug.org/) (specifically the [Knuth–Plass line-breaking algorithm](https://en.wikipedia.org/wiki/Line_wrap_and_word_wrap#Knuth%E2%80%93Plass_algorithm)), and now [Typst](https://typst.app/), but I cannot stand how most text looks in browsers, especially with `text-align: justify`.

The solution was simple. Clone my old site repository, point Codex at it, explain what I wanted to do, and it got cracking. It asked me a few questions on what I liked, what design I wanted, and came back with a few themes. I added my text layout criteria and we went forward implementing it. (It was also straightforward to let it port the algorithm to Go, as I previously had it port the code to Swift for my [KnowledgeBase Viewer](https://github.com/spockz/knowledgebase-viewer).)

The result is what you are reading now. I look forward to sharing more.
