# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

An **Astro-based website** publishing a French technical monograph titled "Interopérabilité en Écosystème d'Entreprise : Convergence des Architectures d'Intégration". Documents enterprise integration architecture patterns across three domains: Applications, Data, and Events.

**Central Thesis**: Interoperability is a continuum requiring a hybrid strategy (App → Data → Event) — from tight coupling to maximum decoupling, culminating in the "Entreprise Agentique".

## Development Commands

```bash
npm run dev      # Start development server
npm run build    # Build for production (output: ./dist)
npm run preview  # Preview production build locally
```

Deploys automatically to GitHub Pages on push to `main` via `.github/workflows/deploy.yml`.

## Architecture

**Tech Stack**: Astro 5.0 + Tailwind CSS + @tailwindcss/typography

### Content System

- **Source content**: `Chapitres/` — standalone markdown chapter files (03-08 currently)
- **Web content**: `src/content/chapters/` — markdown with frontmatter for Astro's content collection
- **Schema**: `src/content/config.ts` defines required frontmatter (title, chapterNumber, romanNumeral, domain, domainColor, description)

### Component Structure

```
src/
├── layouts/
│   ├── BaseLayout.astro      # HTML shell, meta tags
│   └── ChapterLayout.astro   # Chapter rendering with sidebar nav
├── components/
│   ├── Header.astro          # Site header
│   ├── Footer.astro          # Site footer
│   ├── ChapterNav.astro      # Sidebar chapter navigation
│   ├── ChapterCard.astro     # Chapter preview card
│   └── DomainBadge.astro     # Domain color indicator
└── pages/
    └── chapitres/[...slug].astro  # Dynamic chapter routes
```

### Domain Color System

Three integration domains have assigned colors used throughout the UI:
- `verb` (Applications) — accent-verb
- `noun` (Données) — accent-noun
- `signal` (Événements) — accent-signal
- `neutral` — default for intro/conclusion chapters

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
