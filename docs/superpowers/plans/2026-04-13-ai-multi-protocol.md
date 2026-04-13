# AI Multi-Protocol Interpretation Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add multi-protocol failed-run AI interpretation support while keeping a single active AI configuration.

**Architecture:** Expand the existing `protocol` enum in-place, keep the current settings schema, and route failed-run interpretation requests through provider-specific backend adapters. The frontend remains a single configuration form with richer protocol labels and guidance.

**Tech Stack:** Go HTTP handlers, SQLite/MySQL store layer, Next.js, React, Vitest

---

### Task 1: Backend Protocol Tests

**Files:**
- Modify: `internal/httpapi/ai_test.go`

- [ ] **Step 1: Write failing tests for supported protocol normalization**
- [ ] **Step 2: Run the targeted Go tests and verify they fail for the expected unsupported values**
- [ ] **Step 3: Write failing request/response tests for OpenAI Responses, Anthropic, and Gemini**
- [ ] **Step 4: Run the targeted Go tests and verify request-shape failures**
- [ ] **Step 5: Commit**

### Task 2: Backend Protocol Dispatcher

**Files:**
- Modify: `internal/model/ai.go`
- Modify: `internal/httpapi/ai.go`

- [ ] **Step 1: Add protocol constants and normalization support**
- [ ] **Step 2: Replace the single OpenAI request path with a protocol dispatcher**
- [ ] **Step 3: Implement minimal provider adapters for OpenAI Responses, Anthropic, and Gemini**
- [ ] **Step 4: Run targeted Go tests until green**
- [ ] **Step 5: Commit**

### Task 3: Frontend Settings Tests

**Files:**
- Modify: `web-next/src/components/modules/setting/AI.test.tsx`

- [ ] **Step 1: Write failing tests for expanded protocol values and protocol switching submit behavior**
- [ ] **Step 2: Run the targeted Vitest file and verify it fails**
- [ ] **Step 3: Commit**

### Task 4: Frontend Settings UI

**Files:**
- Modify: `web-next/src/types/index.ts`
- Modify: `web-next/src/components/modules/setting/AI.tsx`
- Modify: `web-next/src/components/modules/setting/index.tsx`

- [ ] **Step 1: Expand the protocol union and default protocol handling**
- [ ] **Step 2: Add protocol option metadata, dynamic labels, and helper text**
- [ ] **Step 3: Run the targeted Vitest file until green**
- [ ] **Step 4: Commit**

### Task 5: Verification

**Files:**
- Modify: `docs/superpowers/specs/2026-04-13-ai-multi-protocol-design.md`
- Modify: `docs/superpowers/plans/2026-04-13-ai-multi-protocol.md`

- [ ] **Step 1: Run focused Go tests for AI interpretation**
- [ ] **Step 2: Run focused frontend tests for AI settings**
- [ ] **Step 3: Review diffs for accidental API or schema drift**
- [ ] **Step 4: Summarize remaining risks**
