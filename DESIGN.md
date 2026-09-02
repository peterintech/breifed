---
name: Briefed
description: A calm, source-first editorial reader for public and personalized RSS news.
colors:
  paper: "oklch(0.982 0.007 86)"
  raised-paper: "oklch(0.994 0.004 86)"
  editorial-ink: "oklch(0.19 0.012 255)"
  quiet-metadata: "oklch(0.53 0.014 255)"
  hairline-rule: "oklch(0.89 0.009 86)"
  forest-primary: "oklch(0.27 0.061 152)"
  forest-hover: "oklch(0.33 0.073 152)"
  forest-selected: "oklch(0.96 0.018 152)"
  danger: "oklch(0.55 0.19 25)"
typography:
  display:
    fontFamily: "Inter, Aptos, Segoe UI, system-ui, sans-serif"
    fontSize: "clamp(2.4rem, 5vw, 4.75rem)"
    fontWeight: 700
    lineHeight: 0.98
    letterSpacing: "-0.055em"
  headline:
    fontFamily: "Inter, Aptos, Segoe UI, system-ui, sans-serif"
    fontSize: "1.25rem"
    fontWeight: 600
    lineHeight: 1.25
    letterSpacing: "-0.025em"
  body:
    fontFamily: "Inter, Aptos, Segoe UI, system-ui, sans-serif"
    fontSize: "1rem"
    fontWeight: 400
    lineHeight: 1.75
  label:
    fontFamily: "Inter, Aptos, Segoe UI, system-ui, sans-serif"
    fontSize: "0.75rem"
    fontWeight: 700
    lineHeight: 1.25
    letterSpacing: "0.09em"
rounded:
  control: "0.7rem"
  image: "0.55rem"
  overlay: "1rem"
  pill: "9999px"
spacing:
  compact: "0.75rem"
  control: "1rem"
  section: "2rem"
  rail: "2.5rem"
components:
  button-primary:
    backgroundColor: "{colors.forest-primary}"
    textColor: "{colors.raised-paper}"
    rounded: "{rounded.pill}"
    padding: "0.75rem 1.5rem"
  button-primary-hover:
    backgroundColor: "{colors.forest-hover}"
    textColor: "{colors.raised-paper}"
  input:
    backgroundColor: "{colors.paper}"
    textColor: "{colors.editorial-ink}"
    rounded: "{rounded.control}"
    padding: "0.75rem 1rem"
  chip-selected:
    backgroundColor: "{colors.forest-selected}"
    textColor: "{colors.forest-primary}"
    rounded: "{rounded.pill}"
    padding: "0.5rem 1rem"
---

# Design System: Briefed

## 1. Overview

**Creative North Star: "The Living News Desk"**

Briefed feels like a composed newspaper desk that updates throughout the day: a strong lead story, structured article rails, visible sources, generous paper-like space, and controls that recede while the reader scans. It is editorial, calm, and credible rather than social or attention-maximizing.

Hierarchy comes from typography, placement, rules, and image scale rather than boxed cards or decoration. Personalization appears only when useful: a focused modal for first-time setup and an overlapping drawer for later changes.

**Key Characteristics:**

- Source-first editorial hierarchy.
- Warm, light reading surfaces.
- One restrained forest-green action color.
- Hairline structure instead of card containers.
- Fast state transitions with reduced-motion support.

## 2. Colors

The palette uses tinted paper neutrals and one deep forest-green voice. OKLCH values are canonical because they are used directly in the Tailwind theme.

### Primary

- **Newsroom Forest** (`oklch(0.27 0.061 152)`): primary actions and decisive selected states.
- **Forest Hover** (`oklch(0.33 0.073 152)`): interactive hover and active navigation.
- **Forest Wash** (`oklch(0.96 0.018 152)`): selected chips and quiet focus support.

### Neutral

- **Paper Warmth** (`oklch(0.982 0.007 86)`): page background.
- **Raised Paper** (`oklch(0.994 0.004 86)`): modal, drawer, and elevated form surfaces.
- **Editorial Ink** (`oklch(0.19 0.012 255)`): headlines and primary text.
- **Quiet Metadata** (`oklch(0.53 0.014 255)`): excerpts, timestamps, and secondary labels.
- **Hairline Rule** (`oklch(0.89 0.009 86)`): rail and article separators.
- **Editorial Danger** (`oklch(0.55 0.19 25)`): validation errors only.

**The One Voice Rule.** Forest green marks action or state and never becomes decoration.

## 3. Typography

**Display Font:** Inter with Aptos, Segoe UI, system-ui, and sans-serif fallbacks  
**Body Font:** The same neutral sans-serif stack

**Character:** Contemporary and highly legible, using weight, scale, and tight headline tracking to create newspaper hierarchy without a decorative display face.

### Hierarchy

- **Display** (700, `clamp(2.4rem, 5vw, 4.75rem)`, `0.98`): the single lead story.
- **Headline** (600, `1.25rem`, `1.25`): timeline stories and major supporting titles.
- **Title** (700, `1.875rem`, tight): modal, drawer, and page titles.
- **Body** (400, `1rem`, `1.75`): excerpts and supporting explanations, usually capped near 54ch.
- **Label** (700, `0.75rem`, `0.09em`, uppercase): topics, section labels, and progress context.

**The Front Page Rule.** Only one story per view receives display scale; every supporting story steps down clearly.

## 4. Elevation

Briefed is flat by default. Hairline rules and tonal surfaces provide structure. The only shadow is the soft ambient overlay shadow (`0 1.5rem 4rem oklch(0.18 0.01 255 / 0.18)`) used by the modal, login panel, and preferences drawer.

**The Temporary Lift Rule.** Resting news content stays flat; only transient or focused surfaces lift above the page.

## 5. Components

### Buttons

- **Shape:** confident pill for primary actions (`9999px`); compact rounded controls use `0.7rem`.
- **Primary:** Newsroom Forest with Raised Paper text and `0.75rem 1.5rem` padding.
- **Hover / Focus:** Forest Hover plus a two-stage forest focus ring; transitions stay under 250ms.
- **Secondary / Ghost:** transparent, text-led, and low-contrast until hover.

### Chips

- **Style:** compact outlined pills or softly rounded selection rows.
- **State:** binary. Selected uses Forest Wash, a Forest border, and text/icon indication so color is not the only cue.

### Cards / Containers

- **Corner Style:** articles are not boxed; images use `0.55rem`, overlays use `1rem`.
- **Background:** continuous Paper Warmth, with Raised Paper reserved for focused surfaces.
- **Shadow Strategy:** none for articles.
- **Border:** one-pixel Hairline Rules divide stories and rails.
- **Internal Padding:** `1rem` for controls and `2rem` for major sections.

### Inputs / Fields

- **Style:** quiet paper surface, one-pixel rule, and `0.7rem` radius.
- **Focus:** Forest border and a visible two-stage focus ring.
- **Error / Disabled:** danger copy plus border treatment; never rely on color alone.

### Navigation

Navigation is horizontal, compact, and text-led. The active topic uses forest text plus a two-pixel underline. Topics scroll horizontally on narrow screens without changing document order.

### Editorial Rails and Overlays

Desktop uses one lead rail, one latest/for-you rail, and one discovery rail. The onboarding modal is centered and full-screen on small devices. The 440px preferences drawer is fixed to the right and overlays—never resizes—the news desk.

## 6. Do's and Don'ts

### Do:

- **Do** lead with real public stories before registration.
- **Do** vary article scale and image treatment to communicate hierarchy.
- **Do** keep source, topic, and publication time visible.
- **Do** use one-pixel rules and the exact paper/ink/forest tokens above.
- **Do** preserve keyboard behavior, visible focus, and reduced-motion support.

### Don't:

- **Don't** add social-media engagement dashboards with reactions, scores, follower counts, or noisy sharing controls.
- **Don't** use boxed card grids that make every story look equally important.
- **Don't** use SaaS landing-page clichés, decorative gradients, glassmorphism, neon accents, or pink presentation backdrops.
- **Don't** force onboarding before the visitor can read the public feed.
- **Don't** add Less / Some / More controls; interest selection is binary.
