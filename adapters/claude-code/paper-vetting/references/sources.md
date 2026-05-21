# Sources & Tool Ladder

Each lens has a cheap → better → best path. Use the cheapest tool that produces enough signal; upgrade only when ambiguity remains.

## Lens A — Team

| Step               | Cheap                     | Better                                   | Best                                           |
| ------------------ | ------------------------- | ---------------------------------------- | ---------------------------------------------- |
| Position / tenure  | Lab page via WebSearch    | Institution faculty directory            | DBLP + faculty page cross-check                |
| Publication record | Google Scholar profile    | OpenAlex Authors API (`/authors/<id>`)   | OpenAlex + DBLP intersect                      |
| Career stage / m-q | First-pub year on Scholar | OpenAlex `works_count` + `summary_stats` | Compute m-quotient = h / years-since-first-pub |
| Field-fit          | Read recent paper titles  | Connected Papers neighborhood            | ResearchRabbit author network                  |

### Notes

- **Google Scholar** is fast but uncurated; junk venues count. Good for senior researchers with maintained profiles, useless for industry researchers.
- **OpenAlex API** (`https://api.openalex.org`) is free, 100k credits/day with a free key. Programmatic, well-documented. Good default when you need >5 lookups.
- **DBLP** (`https://dblp.org`) is the cleanest CS publication list — no spam venues. Use for verification when Google Scholar looks gameable.
- **Connected Papers** (`connectedpapers.com`) — visual neighborhood. Useful for "is this person actually in this subfield, or are they tourists?"

## Lens B — Citation context

| Step                      | Cheap                                | Better                                            | Best                                                  |
| ------------------------- | ------------------------------------ | ------------------------------------------------- | ----------------------------------------------------- |
| Citation count            | Google Scholar paper page            | Semantic Scholar API                              | OpenAlex `cited_by_count` + `counts_by_year`          |
| **Citation context**      | Skim 5 random citing papers manually | **Semantic Scholar Highly-Influential Citations** | **scite.ai Smart Citations** (supporting/contrasting) |
| Field neighborhood        | Manual related-work skim             | Connected Papers                                  | ResearchRabbit + Litmaps                              |
| Post-publication concerns | None                                 | Search PubPeer for title/DOI                      | PubPeer + Retraction Watch + arxiv withdraw flag      |
| Replication record        | Manual                               | Replication Database (psychology only)            | Field-specific replication trackers + scite           |

### Notes

- **scite.ai Smart Citations** is the single highest-value Lens B tool. It classifies each citation as `supporting` / `contrasting` / `mentioning`. Free tier is limited; paid is worth it for serious vetting. Even reading the _examples_ on the public page tells you a lot.
- **Semantic Scholar Highly-Influential Citations** (free, in API and UI) flags citations where the citing paper actually built on the cited work. Strips out ceremonial citations.
- **PubPeer** (`pubpeer.com`) — search by DOI or title. If a paper has a thread, read it.
- **Retraction Watch** (`retractionwatch.com`) — searchable database. Quick check is cheap insurance.
- **Connected Papers** generates a graph of citation neighbors; an isolated paper with high cite count is suspicious (cited everywhere, built on by no one).

## Lens C — Claim-level evidence

| Step                             | Cheap                             | Better                                         | Best                                                         |
| -------------------------------- | --------------------------------- | ---------------------------------------------- | ------------------------------------------------------------ |
| Reproducibility checklist        | Search paper PDF for "checklist"  | Read the NeurIPS / ICML / JMLR self-disclosure | Read it + cross-check claims against the methodology section |
| Code release                     | Search GitHub for paper title     | Papers With Code paper page                    | + check last commit, stars, forks, open issues               |
| Benchmark / leaderboard standing | Read paper's claimed comparison   | Papers With Code leaderboard                   | + check whether benchmark itself is contaminated             |
| ML/Agent contamination check     | Eyeball training data + benchmark | See `references/ml-agent-pitfalls.md`          | Use Dynamic-eval surveys (arXiv 2507.21504, 2503.16416)      |
| Statistical rigor (non-ML)       | Read methods section              | Run Statcheck or p-curve                       | + check preregistration on OSF                               |
| Self-citation rate               | Read references section           | Compute % from same author group               | Flag if ≥30% of citations are team's own prior work          |

### Notes

- The **NeurIPS Paper Checklist** is embedded in NeurIPS, ICML, ICLR papers since 2021. It's literally inside the PDF. **Read it.** Authors are rewarded for honesty about limitations — answering "no" isn't disqualifying, but missing entries are.
- **Papers With Code** (`paperswithcode.com`) — for any ML/LLM paper, this is the fastest way to see whether code was released and whether the paper still holds up against newer work on the same benchmark.
- **OSF preregistration** (`osf.io`) — for non-ML empirical work, look for a preregistration link. Its absence on a high-claim paper is a flag.

## Cross-cutting

| Need                        | Tool                                                                               |
| --------------------------- | ---------------------------------------------------------------------------------- |
| Real-time peer reaction     | X / Twitter search by paper title or arxiv id                                      |
| Aggregated public attention | Altmetric                                                                          |
| Practitioner reception      | Hacker News + r/MachineLearning top critique                                       |
| Conference review text      | OpenReview (when applicable)                                                       |
| Author-disambiguation       | OpenAlex Author IDs                                                                |
| Rapid literature triage     | Elicit / Consensus / Connected Papers (use for _batching_, not single-paper depth) |

## API quick-refs

```
arXiv API:
  http://export.arxiv.org/api/query?id_list=<id>

OpenAlex (work by DOI):
  https://api.openalex.org/works/doi:<doi>

OpenAlex (author by ID):
  https://api.openalex.org/authors/<openalex-id>

Semantic Scholar:
  https://api.semanticscholar.org/graph/v1/paper/<paperId>?fields=citations,influentialCitationCount,...

scite.ai (web only; no public free API as of 2026):
  https://scite.ai/reports/<doi>
```

## Anti-sources

- ResearchGate scores — gameable, no methodology transparency.
- Total publication count alone — productivity ≠ impact.
- Press releases from author's institution — marketing.
- LLM-generated paper summaries that don't link to the abs page — frequently hallucinated.
- Paper's own "limitations" section read in isolation — useful as Lens C signal but always cross-check against external lenses.
