# Benchmarks

A statistical engine like argos translates more literally than an LLM. This is expected but it was worth testing to quantify the gap. Four context-heavy sentences translated from English into languages supported by argos (de, es, fr, it):

1. Polysemy: "The lawyer decided to file the case after finding the right file in the cabinet, while the assistant filed the paperwork in alphabetical order, which seemed like a fair arrangement to everyone involved."
2. Idiom vs literal: "When the singer broke a leg during the performance, the stagehand actually broke a leg backstage, so they called an ambulance and brought in an understudy."
3. Metaphor chain: "Time is money, the manager said, so stop cutting corners and burning the midnight oil if you want to hit our targets before the market cools down."
4. Pronoun resolution: "She told him she would meet him at the bank after work, but he went to the wrong one because he thought she meant the branch near the park, not the one downtown."

| Lang | S1 | S2 | S3 | S4 | Argos avg | Gemma 4 avg |
|------|----|----|----|----|-----------|-------------|
| de   | 69% | 48% | 72% | 65% | ~340ms | ~14.6s |
| es   | 79% | 79% | 68% | 78% | ~240ms | ~13.0s |
| fr   | 85% | 57% | 50% | 75% | ~310ms | ~13.9s |
| it   | 60% | 52% | 50% | 56% | ~200ms | ~17.5s |

**Average similarity: 68.8%**, Gemma 4 diverges from argos's literal translation in roughly one third of the output.

**Test hardware**: Apple M2, 16 GB unified memory. Inference times often scale with hardware so obviously expect faster results on GPUs with more compute and memory bandwidth.

Key differences observed:

| Phrase | Argos (literal) | Gemma 4 (context-aware) |
|--------|----------------|------------------------|
| "file the case" (DE) | den Fall einzureichen | den Fall einzureichen |
| "file" (noun, DE) | Akte | Akte |
| "filed the paperwork" (DE) | reichte ein (submitted) | ordnete (organized) |
| "cutting corners" (ES) | cortar esquinas (cut corners) | tomar atajos (take shortcuts) |
| "burning the midnight oil" (FR) | brûler l'huile de minuit (burn midnight oil) | travailler tard (work late) |
| "branch" (IT) | ramo (tree branch) | filiale (bank branch) |
| "understudy" (FR) | sous-étude (lit. sub-study) | remplaçant (replacement) |

Argos translates at 200–400ms per sentence but misses context: it renders "branch" as *ramo* (tree branch) instead of *filiale* (bank branch), and "burning the midnight oil" as *brûler l'huile de minuit* (literal, nonsensical in French) rather than *travailler tard*. Gemma 4 resolves these at the cost of 10–20× longer processing.
