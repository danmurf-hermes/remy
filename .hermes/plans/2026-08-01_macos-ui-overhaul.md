# Remy macOS-Style UI Overhaul

> **For Hermes:** Use kimi-k2.7-code for frontend/UI work.

**Goal:** Make Remy's Wails GUI look like a modern macOS app — frosted glass, SF-style design language, smooth animations, proper typography.

**Approach:** Update CSS variables and component styles across all Svelte files. No structural changes to the HTML/JS — purely visual polish.

**Files to modify:**
- `frontend/src/App.svelte` — global CSS variables, layout
- `frontend/src/lib/Sidebar.svelte` — macOS toolbar styling
- `frontend/src/lib/Chat.svelte` — message area, input
- `frontend/src/lib/MessageBubble.svelte` — bubble design
- `frontend/src/lib/ConversationList.svelte` — Mail-style sidebar
- `frontend/src/lib/Settings.svelte` — System Settings-style tabs
- `frontend/src/lib/MemoryExplorer.svelte` — search, cards
- `frontend/src/lib/TaskManager.svelte` — card design
- `frontend/src/lib/PersonaStudio.svelte` — split panel
- `frontend/src/lib/ActivityLog.svelte` — timeline
- `frontend/src/lib/Toast.svelte` — positioning, animation
- `frontend/src/lib/FactList.svelte`, `EpisodeTimeline.svelte`, `EntityGraph.svelte`, `ScratchpadViewer.svelte` — sub-components

## Design System

### Colors (light)
- Background: `#ffffff` → `rgba(255,255,255,0.72)` with backdrop blur
- Sidebar: `#f5f5f7` → `rgba(246,246,248,0.85)` with backdrop blur
- Card: `rgba(255,255,255,0.5)` with border
- Accent: `#0071e3` (Apple blue)
- Text: `#1d1d1f` primary, `#6e6e73` secondary, `#86868b` tertiary
- Separator: `rgba(0,0,0,0.1)`

### Colors (dark)
- Background: `rgba(28,28,30,0.85)` with backdrop blur
- Sidebar: `rgba(30,30,32,0.9)` with backdrop blur
- Card: `rgba(44,44,46,0.6)` with border
- Accent: `#0a84ff`
- Text: `#f5f5f7` primary, `#a1a1a6` secondary, `#6e6e73` tertiary
- Separator: `rgba(255,255,255,0.08)`

### Typography
- Font: `-apple-system, BlinkMacSystemFont, 'SF Pro', 'Helvetica Neue', sans-serif`
- Mono: `'SF Mono', 'JetBrains Mono', monospace`
- Headings: system font, semibold
- Body: system font, regular

### Effects
- `backdrop-filter: blur(20px) saturate(180%)` on panels
- `box-shadow` with Apple-style spread
- Smooth transitions (0.2s ease)
- Border radius: 8px standard, 12px cards, 20px pills

## Tasks

### Task 1: Global styles & CSS variables
Update `App.svelte` with new design tokens, backdrop blur, proper shadows.

### Task 2: Sidebar
macOS toolbar look — SF Symbols-style icons (use unicode/emoji as fallback), proper spacing, active state with accent background, frosted glass.

### Task 3: Chat & MessageBubble
Redesign bubbles — user messages as blue pills (iMessage style), agent messages as grey cards with avatar. Proper typography, subtle shadows. Input area with rounded pill design.

### Task 4: ConversationList
Apple Mail sidebar — proper list items with hover states, active state with accent highlight, preview text truncation.

### Task 5: Settings
macOS System Settings layout — sidebar tabs on left, content on right. Proper form controls, toggle switches, slider styling.

### Task 6: MemoryExplorer
Clean search bar, card-based results, proper sub-tab navigation.

### Task 7: TaskManager
Card-based task list with proper spacing, clean form design.

### Task 8: PersonaStudio
Split panel with proper card selection, clean editor.

### Task 9: ActivityLog
Timeline with proper entry cards, filter chips, expandable details.

### Task 10: Toast
Positioned notifications with proper animation and frosted glass.

### Task 11: Sub-components (FactList, EpisodeTimeline, EntityGraph, ScratchpadViewer)
Consistent card styling.

## Verification
- `cd frontend && npm test` — all tests pass
- `cd frontend && npx eslint --ext .js,.svelte src/` — no lint errors
- `cd frontend && npx prettier --check src/` — formatting clean
- Visual: open in Wails dev mode and verify macOS look
