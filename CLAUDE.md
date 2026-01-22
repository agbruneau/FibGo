# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a **French technical monograph** titled "Interopérabilité en Écosystème d'Entreprise : Convergence des Architectures d'Intégration". It documents enterprise integration architecture patterns across three domains: Applications, Data, and Events.

**Central Thesis**: Interoperability is not binary but a continuum requiring a hybrid strategy (App → Data → Event) — from tight coupling to maximum decoupling, culminating in the "Entreprise Agentique".

## Structure

All content is in `Chapitres/`:
- **01-11**: Numbered chapter files (e.g., `01_Introduction_Problematique.md`)
- **Annexes.md**: Glossary, technology comparisons, maturity checklist, bibliography

## Chapter Flow

```
I. Problem → II. Theory → III-V. Solutions (App→Data→Event)
→ VI-VII. Foundations (Standards + Resilience)
→ VIII. Collaboration → IX. Synthesis → X. Case Study → XI. Vision (Entreprise Agentique)
```

## The Three Integration Domains

| Domain | Metaphor | Focus | Chapter |
|--------|----------|-------|---------|
| Applications | Le Verbe | Orchestration, synchronous interactions | III |
| Données | Le Nom | State consistency, data accessibility | IV |
| Événements | Le Signal | Reactivity, maximum temporal decoupling | V |

## Editorial Guidelines

### Language
- **Quebec French professional voice** — use Quebec terms (infonuagique, courriel)
- Avoid untranslated Anglo-Saxon jargon (except recognized technical terms)
- Expert tone; vulgarize complex concepts

### Format
- Target: ~8,000 words per chapter
- Structure: Introduction (10-15%), Development (75-80%), Conclusion (10%)
- Headings: `##` main sections, `###` subsections, `####` sparingly
- Prefer **fluid prose over bullet lists**
- Each chapter ends with a structured **Résumé**
- Cross-references: "Comme établi au chapitre II..." or "Le patron CDC présenté au chapitre IV..."

### Terminology
- First occurrence of acronyms: full form with acronym in parentheses
- 23 architecture patterns documented across chapters III-V
- Refer to Annexes.md glossary for consistent terminology

### Citations
- Prioritize 2023-2026 sources
- Reference: Confluent, Apache, Google Cloud, Anthropic, Microsoft
- Format: "selon [Organisation, Année]" or Author (Year)
