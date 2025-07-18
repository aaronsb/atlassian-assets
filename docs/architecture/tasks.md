# Tasks - Atlassian Assets CLI Toolkit

## Task 01 – SDK Research & Selection [req-001, req-002, req-003]
- [✔] sub-01-a Research available open-source Atlassian Assets SDK libraries (req-003)
- [✔] sub-01-b Evaluate SDK options for CRUD operations support (req-003)
- [✔] sub-01-c Document findings and recommendations (req-003)
- [✔] sub-01-d Select primary SDK for development (req-003)

## Task 02 – CLI Architecture Design [req-001, req-002]
- [✔] sub-02-a Design CLI command structure with MCP translation in mind (req-002)
- [✔] sub-02-b Define configuration management approach (req-002)
- [✔] sub-02-c Plan authentication integration (req-002)
- [✔] sub-02-d Map CLI commands to future MCP tools (req-001)

## Task 03 – CLI Prototype Implementation [req-002]
- [✔] sub-03-a Create Go project structure (req-002)
- [✔] sub-03-b Implement basic CLI framework (req-002)
- [✔] sub-03-c Integrate go-atlassian SDK (req-002)
- [✔] sub-03-d Implement CRUD operations (req-002)

## Task 04 – CLI Enhancement & Intelligence [req-002]
- [✔] sub-04-a Design centralized contextual hints system (req-002)
- [✔] sub-04-b Implement workflow hints JSON configuration (req-002)
- [✔] sub-04-c Normalize all CLI commands with contextual hints (req-002)
- [✔] sub-04-d Add intelligent next-step suggestions to commands (req-002)
- [✔] sub-04-e Create compositional command workflows (req-002)

## Task 05 – MCP Interface Development [req-001]
- [ ] sub-05-a Design MCP tool definitions based on CLI (req-001)
- [ ] sub-05-b Implement MCP server wrapper (req-001)
- [ ] sub-05-c Test MCP interface with AI agents (req-001)
- [ ] sub-05-d Document MCP integration patterns (req-001)

---

## Progress Notes
- 2025-07-17: Local tracking structure established
- 2025-07-17: Starting with Task 01 - SDK research phase
- 2025-07-17: Research completed - found 3 viable SDK options
- 2025-07-17: **Recommendation: go-atlassian** - only library with explicit Assets API support
- 2025-07-17: Updated requirements to reflect MCP as primary goal with CLI as prototype
- 2025-07-17: Refined task structure to include MCP interface development phase
- 2025-07-17: **Tasks 01-03 COMPLETE** - CLI prototype framework implemented
- 2025-07-17: Created complete CLI structure with CRUD operations, configuration management
- 2025-07-17: Ready for testing and actual API implementation
- 2025-07-17: **Task 04 COMPLETE** - Implemented intelligent CLI with contextual hints system
- 2025-07-17: Created centralized workflow hints JSON configuration (`internal/hints/workflow_hints.json`)
- 2025-07-17: Normalized all CLI commands with contextual next-step suggestions
- 2025-07-17: Built compositional command workflows with intelligent guidance
- 2025-07-17: CLI now provides contextual hints for object creation, attribute marketplace, and workflow automation
- 2025-07-17: **Status**: Ready for MCP interface development (Task 05)