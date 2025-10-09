# Session Recovery Guide - Million Dollar Hunter

**Purpose**: This document provides step-by-step instructions for new agent sessions to restore full project context quickly and accurately.

**Target Time**: < 5 minutes for complete context restoration

---

## Quick Start: 3-Step Context Restoration

### Step 1: Read This Document (2 min)
You're already doing it! Continue reading to understand the recovery process.

### Step 2: Read QUICK-REFERENCE.md (2 min)
Get a high-level overview of the project architecture, tech stack, and key concepts.

### Step 3: Review Current Status (1 min)
Check `DEVELOPMENT-STATUS.md` to understand what's been completed and what's in progress.

---

## Full Context Restoration Process

### Phase 1: Essential Context (Priority 1 - Read First)

**Estimated Time**: 5 minutes

Read these documents in order:

1. **QUICK-REFERENCE.md** (2 min)
   - Project overview and architecture summary
   - Technology stack
   - Key file locations
   - Quick navigation guide

2. **DEVELOPMENT-STATUS.md** (2 min)
   - Current development phase
   - Completed work items
   - In-progress work items
   - Recent changes and blockers

3. **AGENT-ASSIGNMENTS.md** (1 min)
   - Your agent role (A, B, C, D, E, or F)
   - Your responsibilities
   - Your dependencies on other agents

**Checkpoint**: After Phase 1, you should know:
- ✅ What the project is building
- ✅ What phase of development we're in
- ✅ What's been completed
- ✅ What you're responsible for

---

### Phase 2: Your Domain Context (Priority 2 - Read Next)

**Estimated Time**: 10-15 minutes

Based on your agent assignment, read your domain-specific documentation:

#### If You Are Agent A (Authentication & Security):
1. `PRD-Million-Dollar-Hunter-Crypto-Dashboard.md` - Goals, Security sections
2. `Technical Development Plan.md` - Section II.E (Security Architecture)
3. `security-hardening.md` - Complete document
4. `database-migration-strategy.md` - For auth-service schema
5. `API-STATUS.md` - Check `/api/v1/auth/*` endpoint status

#### If You Are Agent B (API Gateway & Orchestration):
1. `PRD-Million-Dollar-Hunter-Crypto-Dashboard.md` - Complete document
2. `Technical Development Plan.md` - Section II (Backend Development)
3. `openapi.yaml` - **PRIMARY REFERENCE** for REST API spec
4. `external-api-integrations.md` - Rate limiting algorithms
5. `monitoring-alerting.md` - Observability requirements
6. `API-STATUS.md` - Check all REST endpoint status

#### If You Are Agent C (Portfolio & Aggregation):
1. `PRD-Million-Dollar-Hunter-Crypto-Dashboard.md` - Token Analytics, Wallet Monitoring
2. `Technical Development Plan.md` - Sections II.B, II.D (Microservices, Schemas)
3. `database-migration-strategy.md` - Schema versioning
4. `architecture-decisions.md` - CQRS and Saga patterns
5. `API-STATUS.md` - Check Kafka event schemas and portfolio endpoints

#### If You Are Agent D (Data Ingestion):
1. `PRD-Million-Dollar-Hunter-Crypto-Dashboard.md` - Functional Requirements
2. `Technical Development Plan.md` - Sections II.B, IV.A (ingestion-service, External APIs)
3. `external-api-integrations.md` - **CRITICAL** - rate limits, fallbacks, cost tracking
4. `database-migration-strategy.md` - Write-optimized schemas
5. `API-STATUS.md` - Check external API integration status (Alchemy, Moralis)

#### If You Are Agent E (Market Data):
1. `PRD-Million-Dollar-Hunter-Crypto-Dashboard.md` - Price Tracking sections
2. `Technical Development Plan.md` - Sections II.B, IV.A (market-data-service, CoinGecko)
3. `external-api-integrations.md` - Data freshness, caching strategies
4. `database-migration-strategy.md` - Token prices schema
5. `API-STATUS.md` - Check CoinGecko integration and price endpoints

#### If You Are Agent F (Frontend & Visualization):
1. `PRD-Million-Dollar-Hunter-Crypto-Dashboard.md` - **Complete** (user stories, UX)
2. `Technical Development Plan.md` - Section III (Frontend Development - Complete)
3. `frontend-components.md` - Component library, props, design system
4. `openapi.yaml` - REST API contracts for frontend calls
5. `API-STATUS.md` - Check which backend endpoints are ready

**Checkpoint**: After Phase 2, you should know:
- ✅ Detailed requirements for your domain
- ✅ Technical specifications for your components
- ✅ What APIs or interfaces you depend on
- ✅ Current implementation status of your area

---

### Phase 3: Technical Decisions & History (Priority 3 - Reference as Needed)

**Estimated Time**: 5-10 minutes

Review these documents to understand past decisions and their rationale:

1. **TECHNICAL-DECISIONS.md** (5 min)
   - Browse recent decisions (top of file)
   - Understand why key technologies were chosen
   - See what alternatives were considered

2. **AGENT-HANDOFF.md** (2 min)
   - Handoff protocol when you complete work
   - Communication procedures with other agents

3. **API-STATUS.md** (3 min - if not read in Phase 2)
   - Detailed endpoint implementation status
   - gRPC interface definitions
   - Kafka event schemas

**Checkpoint**: After Phase 3, you should know:
- ✅ Why the architecture is designed this way
- ✅ What technical decisions have been made
- ✅ How to hand off work to other agents

---

### Phase 4: Supporting Documentation (Priority 4 - As Needed)

**Estimated Time**: Variable

Reference these documents when you need specific information:

- **performance-requirements.md** - SLO targets (API latency, frontend metrics)
- **testing-strategy.md** - Test requirements, coverage targets
- **monitoring-alerting.md** - Observability and alerting setup
- **operational-runbook.md** - Health checks, troubleshooting
- **data-privacy-retention.md** - Data handling policies
- **dev-environment-setup.md** - Local development setup
- **requirements-traceability.md** - PRD to implementation mapping

---

## Context Restoration Checklist

Use this checklist to verify you have full context:

### Essential Knowledge
- [ ] I know what Million Dollar Hunter is (crypto portfolio dashboard)
- [ ] I know the current development phase (Phase 0/1/2/3/4)
- [ ] I know which agent I am (A/B/C/D/E/F) or my role
- [ ] I know what's been completed so far
- [ ] I know what I'm supposed to work on

### Domain Knowledge
- [ ] I've read my agent-specific required documentation
- [ ] I know my deliverables and responsibilities
- [ ] I know which other agents I depend on
- [ ] I know which agents depend on me
- [ ] I've checked API-STATUS.md for my relevant interfaces

### Technical Knowledge
- [ ] I know the technology stack for my components
- [ ] I understand the architecture patterns (microservices, CQRS, etc.)
- [ ] I know where to find API specifications (openapi.yaml, proto files)
- [ ] I understand key technical decisions from TECHNICAL-DECISIONS.md

### Operational Knowledge
- [ ] I know how to update documentation after completing work
- [ ] I know the handoff protocol (AGENT-HANDOFF.md)
- [ ] I know how to check for blockers or dependencies
- [ ] I know where to log new technical decisions

**If you checked all boxes**: ✅ You're ready to start working!  
**If you're missing any**: Review the relevant documentation from the phases above.

---

## Common Scenarios

### Scenario 1: "I'm a new agent starting Phase 1 work"

**Context You Need**:
1. Read QUICK-REFERENCE.md (understand project)
2. Read DEVELOPMENT-STATUS.md (confirm Phase 1 status)
3. Read AGENT-ASSIGNMENTS.md (confirm your role)
4. Read your agent-specific docs from Phase 2 above
5. Check TECHNICAL-DECISIONS.md for architecture rationale

**First Actions**:
- Set up your development environment (dev-environment-setup.md)
- Check API-STATUS.md for your endpoints/interfaces
- Review your Phase 1 tasks in DEVELOPMENT-STATUS.md
- Begin implementation

---

### Scenario 2: "I'm continuing work from a previous session"

**Context You Need**:
1. Read DEVELOPMENT-STATUS.md - "Recent Changes" section
2. Check what you marked as "In Progress" last session
3. Review TECHNICAL-DECISIONS.md for any new decisions
4. Check API-STATUS.md for changes to your interfaces

**First Actions**:
- Review your last commit messages
- Check for any PRs or comments from other agents
- Continue implementation where you left off
- Update DEVELOPMENT-STATUS.md when you complete items

---

### Scenario 3: "I need to integrate with another agent's work"

**Context You Need**:
1. Read DEVELOPMENT-STATUS.md to see if dependency is complete
2. Check API-STATUS.md for the interface specification
3. Review AGENT-HANDOFF.md for coordination protocol
4. Check if the other agent left integration notes

**First Actions**:
- Verify the interface is ready (check API-STATUS.md status)
- Review the spec (openapi.yaml for REST, .proto for gRPC)
- If blocked, note in DEVELOPMENT-STATUS.md blockers section
- Test integration with mock/stub if real service not ready

---

### Scenario 4: "I'm debugging or investigating an issue"

**Context You Need**:
1. Read TECHNICAL-DECISIONS.md to understand design rationale
2. Check DEVELOPMENT-STATUS.md for recent changes
3. Review architecture-decisions.md for patterns used
4. Check operational-runbook.md for troubleshooting

**First Actions**:
- Check logs (if services are running)
- Review recent commits
- Check if issue is documented in DEVELOPMENT-STATUS.md
- Review relevant technical decision (TD-XXX) for context

---

### Scenario 5: "I need to make a technical decision"

**Context You Need**:
1. Read TECHNICAL-DECISIONS.md to see similar past decisions
2. Review architecture-decisions.md for architectural constraints
3. Check PRD for product requirements
4. Review Technical Development Plan for technical constraints

**First Actions**:
- Consider alternatives (research, prototype if needed)
- Document in TECHNICAL-DECISIONS.md (use template)
- Update affected documentation (API-STATUS, DEVELOPMENT-STATUS)
- Notify dependent agents via AGENT-HANDOFF protocol

---

## Documentation Update Workflow

After completing work, update these documents in order:

### 1. API-STATUS.md (If you worked on APIs)
- Update endpoint status: In Progress → Completed → Tested
- Update gRPC interface status if you created .proto files
- Update Kafka event schema status if you defined events
- Note any deviations from original spec

### 2. DEVELOPMENT-STATUS.md (Always)
- Move completed items from "In-Progress" to "Completed"
- Remove items from "Pending" if you started them (move to "In-Progress")
- Update "Last Updated" timestamp and your agent designation
- Add entry to "Recent Changes" section with what you completed
- Note any new blockers or dependencies

### 3. TECHNICAL-DECISIONS.md (If you made a decision)
- Add new decision using TD-XXX numbering
- Use the decision template
- Explain rationale and alternatives
- Update "Last Updated" date

### 4. Your Service-Specific Docs (If applicable)
- Update README in your service directory
- Update .proto files with comments
- Update migration scripts documentation
- Update any service-specific configuration docs

### 5. Commit and Hand Off
- Commit all documentation changes
- Follow AGENT-HANDOFF.md protocol
- Tag dependent agents in commit message or PR
- Update this file if you discover better recovery procedures

---

## Troubleshooting Context Issues

### Problem: "I don't understand the project"
**Solution**: Read QUICK-REFERENCE.md first, then PRD sections relevant to your agent

### Problem: "I don't know what to work on"
**Solution**: Read DEVELOPMENT-STATUS.md, find your agent (A-F), check "In-Progress" or "Pending" sections

### Problem: "I don't know the architecture"
**Solution**: Read Technical Development Plan Section I.A (Architecture), then QUICK-REFERENCE.md

### Problem: "I can't find a specification"
**Solution**: Check API-STATUS.md for location of specs (openapi.yaml, .proto files, event schemas)

### Problem: "I don't know why a decision was made"
**Solution**: Read TECHNICAL-DECISIONS.md, search for relevant TD-XXX entry

### Problem: "I don't know if I can start my work"
**Solution**: Check DEVELOPMENT-STATUS.md "Blockers & Dependencies" section for your agent

### Problem: "I completed work but don't know how to hand off"
**Solution**: Read AGENT-HANDOFF.md, follow the protocol, update DEVELOPMENT-STATUS.md

### Problem: "Documentation is out of date"
**Solution**: Update it! All docs are living documents. Update "Last Updated" timestamp.

---

## Tips for Efficient Context Restoration

1. **Bookmark Key Files**: Keep QUICK-REFERENCE.md, DEVELOPMENT-STATUS.md, API-STATUS.md open in tabs

2. **Use Search**: All docs are Markdown. Use Ctrl+F / Cmd+F to find keywords quickly

3. **Read Backwards**: Recent changes are at the top of TECHNICAL-DECISIONS.md and DEVELOPMENT-STATUS.md

4. **Check Timestamps**: Look at "Last Updated" dates to see what's fresh vs stale

5. **Agent-Specific Focus**: Don't try to read everything. Focus on your agent's required docs first.

6. **Verify Completion**: Use the checklist above to confirm you have enough context before starting work

7. **Update as You Go**: Don't wait until the end. Update DEVELOPMENT-STATUS.md when you complete items.

8. **Ask Questions in Commits**: If something is unclear, note it in your commit message for next session

---

## Recovery Time Estimates

| Scenario | Estimated Time | Key Documents |
|----------|---------------|---------------|
| First time joining project | 20-30 min | All Phase 1-3 docs |
| Continuing your own work | 5-10 min | DEVELOPMENT-STATUS, your last commits |
| Starting new phase | 15-20 min | Phase 1-2 docs + phase-specific requirements |
| Integration work | 10-15 min | API-STATUS, AGENT-HANDOFF, relevant specs |
| Debugging | 15-25 min | TECHNICAL-DECISIONS, recent changes, logs |

---

## Document Location Quick Reference

All documentation is in `./docs/` directory:

**Essential for All Agents:**
- `QUICK-REFERENCE.md` - Project overview (READ FIRST)
- `DEVELOPMENT-STATUS.md` - Current progress (READ SECOND)
- `AGENT-ASSIGNMENTS.md` - Agent roles (READ THIRD)
- `API-STATUS.md` - Endpoint implementation status
- `TECHNICAL-DECISIONS.md` - Decision log with rationale
- `AGENT-HANDOFF.md` - Handoff protocol

**Product & Architecture:**
- `PRD-Million-Dollar-Hunter-Crypto-Dashboard.md` - Product requirements
- `Technical Development Plan.md` - Technical architecture
- `million-hunter-development-plan.md` - Multi-agent development plan

**Technical Specifications:**
- `openapi.yaml` - REST API specification
- `architecture-decisions.md` - Architecture patterns
- `database-migration-strategy.md` - Schema versioning

**Agent-Specific:**
- `security-hardening.md` - Agent A
- `external-api-integrations.md` - Agents D, E
- `monitoring-alerting.md` - Agent B
- `frontend-components.md` - Agent F

**Operations:**
- `performance-requirements.md` - SLO targets
- `testing-strategy.md` - Test requirements
- `operational-runbook.md` - Troubleshooting
- `dev-environment-setup.md` - Local setup

---

## Success Criteria

You have successfully restored context when you can answer:

1. **What is this project?** → Crypto portfolio dashboard for single user
2. **What phase are we in?** → Check DEVELOPMENT-STATUS.md
3. **What's my role?** → Check AGENT-ASSIGNMENTS.md for your agent (A-F)
4. **What should I work on?** → Check DEVELOPMENT-STATUS.md for your agent's tasks
5. **Who do I depend on?** → Check AGENT-ASSIGNMENTS.md dependencies
6. **What APIs do I use?** → Check API-STATUS.md for your interfaces
7. **What's been decided?** → Check TECHNICAL-DECISIONS.md for rationale
8. **Where do I start?** → Your agent's Phase 1-4 tasks in DEVELOPMENT-STATUS.md

**If you can answer all 8 questions**: You're ready to code! 🚀

---

## Maintenance

This document should be updated when:
- Recovery procedures are found to be unclear or incomplete
- New documentation is added that should be in the recovery process
- Recovery time estimates are found to be inaccurate
- New scenarios are encountered that should be documented

**Last Updated**: 2025-10-09  
**Maintained By**: All agents (update as needed)

---

*Goal: Any agent should be able to restore full project context in < 5 minutes by following this guide.*
